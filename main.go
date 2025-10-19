package main

import (
	"context"
	"embed"
	"fail2ban-web/config"
	"fail2ban-web/core"
	"fail2ban-web/internal/handler"
	"fail2ban-web/internal/middleware"
	"fail2ban-web/internal/model"
	"fail2ban-web/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed web
var staticFiles embed.FS

// AppLifecycle 应用程序生命周期
type AppLifecycle struct {
}

// OnStart 应用程序启动时执行
func (l *AppLifecycle) OnStart(context.Context) error {
	log.Println("Application started successfully")
	return nil
}

// OnStop 应用程序停止时执行
func (l *AppLifecycle) OnStop(context.Context) error {
	log.Println("Application stopping...")
	return nil
}

func NewAppLifeCycle() *AppLifecycle {
	return &AppLifecycle{}
}

func main() {
	app := fx.New(
		// 提供配置
		fx.Provide(func() *core.AppConfig {
			return core.LoadConfig()
		}),
		
		// 提供旧的 config.Config 以兼容现有服务
		fx.Provide(func() *config.Config {
			return config.LoadConfig()
		}),

		// 提供日志
		fx.Provide(func() *logrus.Logger {
			logger := logrus.New()
			logger.SetLevel(logrus.InfoLevel)
			logger.SetFormatter(&logrus.TextFormatter{
				FullTimestamp: true,
			})
			return logger
		}),

		// 创建应用服务器
		fx.Provide(core.NewServer),

		// 初始化服务器
		fx.Invoke(func(s *core.AppServer) {
			s.Init(staticFiles)
		}),

		// 提供数据库连接
		fx.Provide(func(config *core.AppConfig) (*gorm.DB, error) {
			gormLogger := logger.Default.LogMode(logger.Info)
			db, err := gorm.Open(sqlite.Open(config.DBPath), &gorm.Config{
				Logger: gormLogger,
			})
			if err != nil {
				return nil, err
			}

			// 自动迁移
			err = db.AutoMigrate(
				&model.BannedIP{},
				&model.Fail2banJail{},
			)
			if err != nil {
				return nil, err
			}

			log.Println("Database initialized successfully")
			return db, nil
		}),

		// 创建服务
		fx.Provide(func(logger *logrus.Logger) *service.Fail2BanService {
			return service.NewFail2BanService(logger)
		}),
		fx.Provide(service.NewJailService),
		fx.Provide(service.NewSSHService),
		fx.Provide(service.NewNginxService),
		fx.Provide(func(jailService *service.JailService) *service.DefaultSSHService {
			return service.NewDefaultSSHService(jailService)
		}),
		fx.Provide(func(jailService *service.JailService) *service.DefaultNginxService {
			return service.NewDefaultNginxServiceWithJail(jailService)
		}),
		fx.Provide(service.NewDefaultNginxAdvancedService),
		fx.Provide(service.NewDefaultJailService),
		fx.Provide(service.NewIntelligentScanService),

		// 启动智能扫描服务
		fx.Invoke(func(lifecycle fx.Lifecycle, intelligentService *service.IntelligentScanService) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					log.Println("Starting intelligent scan service...")
					intelligentService.Start()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					log.Println("Stopping intelligent scan service...")
					intelligentService.Stop()
					return nil
				},
			})
		}),

		// 创建 Handler
		fx.Provide(handler.NewAuthHandler),
		fx.Provide(handler.NewFail2BanHandler),
		fx.Provide(handler.NewJailHandler),
		fx.Provide(handler.NewDefaultConfigHandler),
		fx.Provide(handler.NewSSHHandler),
		fx.Provide(handler.NewNginxHandler),
		fx.Provide(handler.NewIntelligentHandler),

		// 注册路由
		fx.Invoke(registerRoutes),

		// 启动服务器
		fx.Invoke(func(s *core.AppServer, db *gorm.DB) {
			go func() {
				err := s.Run(db)
				if err != nil {
					log.Fatal(err)
				}
			}()
		}),

		// 生命周期
		fx.Provide(NewAppLifeCycle),
		fx.Invoke(func(lifecycle fx.Lifecycle, lc *AppLifecycle) {
			lifecycle.Append(fx.Hook{
				OnStart: lc.OnStart,
				OnStop:  lc.OnStop,
			})
		}),
	)

	// 启动应用程序
	go func() {
		if err := app.Start(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	// 监听退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 关闭应用程序
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Stop(ctx); err != nil {
		log.Fatal(err)
	}
}

// registerRoutes 注册所有路由
func registerRoutes(
	s *core.AppServer,
	authHandler *handler.AuthHandler,
	fail2banHandler *handler.Fail2BanHandler,
	jailHandler *handler.JailHandler,
	defaultConfigHandler *handler.DefaultConfigHandler,
	sshHandler *handler.SSHHandler,
	nginxHandler *handler.NginxHandler,
	intelligentHandler *handler.IntelligentHandler,
) {
	r := s.Engine

	// 添加中间件
	r.Use(middleware.CORSMiddleware())

	// 公共路由
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Fail2Ban 管理面板",
		})
	})

	// 登录页面
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "登录 - Fail2Ban 管理面板",
		})
	})

	// API 路由组
	api := r.Group("/api/v1")

	// 认证相关路由（不需要认证）
	auth := api.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.GET("/profile", authHandler.GetProfile)
	}

	// 需要认证的API路由
	authenticated := api.Group("")
	// authenticated.Use(authMiddleware.JWTAuth()) // 暂时禁用认证
	{
		// 健康检查
		authenticated.GET("/health", fail2banHandler.HealthCheck)

		// 统计信息
		authenticated.GET("/stats", fail2banHandler.GetStats)

		// 系统信息
		authenticated.GET("/system-info", fail2banHandler.GetSystemInfo)

		// 版本信息
		authenticated.GET("/version", fail2banHandler.GetVersion)

		// 被禁IP管理
		authenticated.GET("/banned-ips", fail2banHandler.GetBannedIPs)
		authenticated.POST("/unban", fail2banHandler.UnbanIP)
		authenticated.POST("/ban", fail2banHandler.BanIP)

		// 日志查看
		authenticated.GET("/logs", fail2banHandler.GetLogs)

		// Jail 配置管理
		jails := authenticated.Group("/jails")
		{
			jails.GET("", jailHandler.GetJails)
			jails.GET("/:name", jailHandler.GetJail)
			jails.GET("/:name/status", fail2banHandler.GetJailStatus)
			jails.POST("", jailHandler.CreateJail)
			jails.PUT("/:name", jailHandler.UpdateJail)
			jails.DELETE("/:name", jailHandler.DeleteJail)
			jails.POST("/:name/toggle", jailHandler.ToggleJail)
		}

		// 默认配置管理
		defaults := authenticated.Group("/defaults")
		{
			defaults.GET("/info", defaultConfigHandler.GetDefaultConfigInfo)
			defaults.POST("/nginx/install", defaultConfigHandler.InstallNginxDefaults)
			defaults.GET("/nginx/filters", defaultConfigHandler.GetNginxFilterTemplates)
			defaults.GET("/nginx/jail-config", defaultConfigHandler.GetNginxJailConfig)
			defaults.GET("/nginx/export", defaultConfigHandler.ExportNginxConfig)
		}

		// SSH监控管理
		ssh := authenticated.Group("/ssh")
		{
			ssh.GET("/stats", sshHandler.GetSSHStats)
			ssh.GET("/logs", sshHandler.GetSSHLogs)
			ssh.GET("/status", sshHandler.GetSSHJailStatus)
			ssh.POST("/ban", sshHandler.BanSSHIP)
			ssh.POST("/unban", sshHandler.UnbanSSHIP)
			ssh.GET("/defaults", sshHandler.GetSSHDefaults)
			ssh.POST("/defaults/install", sshHandler.InstallSSHDefaults)
		}

		// Nginx监控管理
		nginx := authenticated.Group("/nginx")
		{
			nginx.GET("/stats", nginxHandler.GetNginxStats)
			nginx.GET("/logs", nginxHandler.GetNginxLogs)
			nginx.GET("/status", nginxHandler.GetNginxJailStatus)
			nginx.POST("/ban", nginxHandler.BanNginxIP)
			nginx.POST("/unban", nginxHandler.UnbanNginxIP)
			nginx.GET("/defaults", nginxHandler.GetNginxDefaults)
			nginx.GET("/defaults/advanced", nginxHandler.GetNginxAdvancedDefaults)
			nginx.POST("/defaults/install", nginxHandler.InstallNginxDefaults)
			nginx.POST("/defaults/advanced/install", nginxHandler.InstallNginxAdvancedDefaults)
		}

		// 智能分析管理
		intelligent := authenticated.Group("/intelligent")
		{
			intelligent.GET("/threats", intelligentHandler.GetCurrentThreats)
			intelligent.GET("/scan-result", intelligentHandler.GetScanResult)
			intelligent.GET("/stats", intelligentHandler.GetThreatStats)
			intelligent.POST("/ban", intelligentHandler.ManualBanIP)
			intelligent.POST("/analyze-log", intelligentHandler.AnalyzeLogFile)
			intelligent.POST("/analyze-access-log", intelligentHandler.AnalyzeAccessLog)
		}
	}
}
