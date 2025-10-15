package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"fail2ban-web/config"
	"fail2ban-web/internal/handler"
	"fail2ban-web/internal/middleware"
	"fail2ban-web/internal/model"
	"fail2ban-web/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//go:embed web
var staticFiles embed.FS

func main() {
	// 初始化配置
	cfg := config.LoadConfig()

	// 初始化日志
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// 初始化数据库
	db, err := initDB()
	if err != nil {
		log.Fatal("初始化数据库失败:", err)
	}

	// 初始化服务
	fail2banService := service.NewFail2BanService(logger)
	jailService := service.NewJailService(db)
	sshService := service.NewSSHService(cfg, db)
	nginxService := service.NewNginxService(cfg, db)
	defaultSSHService := service.NewDefaultSSHService(jailService)
	defaultNginxService := service.NewDefaultNginxServiceWithJail(jailService)
	defaultNginxAdvancedService := service.NewDefaultNginxAdvancedService(jailService)
	
	// 初始化智能扫描服务
	intelligentService := service.NewIntelligentScanService(cfg, db, sshService, nginxService, jailService, fail2banService)
	
	// 启动智能扫描服务
	intelligentService.Start()
	defer intelligentService.Stop()

	// 初始化中间件
	authMiddleware := middleware.NewJWTMiddleware(cfg.JWT.Secret)

	// 初始化handlers
	authHandler := handler.NewAuthHandler(cfg)
	fail2banHandler := handler.NewFail2BanHandler(fail2banService)
	jailHandler := handler.NewJailHandler(jailService)
	defaultConfigHandler := handler.NewDefaultConfigHandler()
	sshHandler := handler.NewSSHHandler(sshService, defaultSSHService)
	nginxHandler := handler.NewNginxHandler(nginxService, defaultNginxService, defaultNginxAdvancedService)
	intelligentHandler := handler.NewIntelligentHandler(intelligentService)

	// 设置为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 创建 Gin 路由器
	r := gin.Default()

	// 添加中间件
	r.Use(middleware.CORSMiddleware())

	// 设置静态文件
	setupStaticFiles(r)

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
		auth.POST("/refresh", authMiddleware.JWTAuth(), authHandler.RefreshToken)
		auth.GET("/profile", authMiddleware.JWTAuth(), authHandler.GetProfile)
	}

	// 需要认证的API路由
	authenticated := api.Group("")
	authenticated.Use(authMiddleware.JWTAuth())
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

	// 启动服务器
	log.Println("服务器启动在端口 :8092")
	log.Println("访问 http://localhost:8092 打开管理面板")
	if err := r.Run(":8092"); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}

// setupStaticFiles 设置静态文件处理
func setupStaticFiles(r *gin.Engine) {
	// 设置静态文件
	staticFS, _ := fs.Sub(staticFiles, "web/static")
	r.StaticFS("/static", http.FS(staticFS))

	// 加载 HTML 模板
	templ := template.Must(template.New("").ParseFS(staticFiles, "web/templates/*.html"))
	r.SetHTMLTemplate(templ)
}

// initDB 初始化数据库
func initDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("fail2ban_web.db"), &gorm.Config{})
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

	// 创建默认Nginx配置（如果不存在）
	defaultJailService := service.NewDefaultJailService(service.NewJailService(db))
	if err := createDefaultNginxConfig(defaultJailService); err != nil {
		log.Printf("创建默认Nginx配置失败: %v", err)
	}

	return db, nil
}

// createDefaultNginxConfig 创建默认Nginx配置
func createDefaultNginxConfig(defaultJailService *service.DefaultJailService) error {
	// 检查是否已有nginx相关配置
	// 这里简单检查，实际可以更精确
	log.Println("检查默认Nginx配置...")
	
	// 可以在这里添加检查逻辑，如果需要的话
	// 现在暂时跳过自动创建，让用户手动安装
	log.Println("默认Nginx配置准备就绪，用户可通过Web界面安装")
	
	return nil
}