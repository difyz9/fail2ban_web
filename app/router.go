package app

import (
	"embed"
	"fail2ban-web/internal/handler"
	"fail2ban-web/internal/middleware"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// RouterParams 路由依赖参数
type RouterParams struct {
	fx.In
	Logger               *zap.Logger
	AuthHandler          *handler.AuthHandler
	Fail2banHandler      *handler.Fail2BanHandler
	JailHandler          *handler.JailHandler
	DefaultConfigHandler *handler.DefaultConfigHandler
	SSHHandler           *handler.SSHHandler
	NginxHandler         *handler.NginxHandler
	IntelligentHandler   *handler.IntelligentHandler
	StaticFiles          embed.FS `name:"staticFiles"`
}

// NewRouter 创建 Gin 路由器
func NewRouter(params RouterParams) *gin.Engine {
	// 设置为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 创建 Gin 路由器
	r := gin.Default()

	// 添加中间件
	r.Use(middleware.CORSMiddleware())

	// 设置静态文件
	setupStaticFiles(r, params.StaticFiles)

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
		auth.POST("/login", params.AuthHandler.Login)
		auth.POST("/refresh", params.AuthHandler.RefreshToken)
		auth.GET("/profile", params.AuthHandler.GetProfile)
	}

	// 需要认证的API路由
	authenticated := api.Group("")
	// authenticated.Use(authMiddleware.JWTAuth()) // 暂时禁用认证
	{
		// 健康检查
		authenticated.GET("/health", params.Fail2banHandler.HealthCheck)

		// 统计信息
		authenticated.GET("/stats", params.Fail2banHandler.GetStats)

		// 系统信息
		authenticated.GET("/system-info", params.Fail2banHandler.GetSystemInfo)

		// 版本信息
		authenticated.GET("/version", params.Fail2banHandler.GetVersion)

		// 被禁IP管理
		authenticated.GET("/banned-ips", params.Fail2banHandler.GetBannedIPs)
		authenticated.POST("/unban", params.Fail2banHandler.UnbanIP)
		authenticated.POST("/ban", params.Fail2banHandler.BanIP)

		// 日志查看
		authenticated.GET("/logs", params.Fail2banHandler.GetLogs)

		// Jail 配置管理
		jails := authenticated.Group("/jails")
		{
			jails.GET("", params.JailHandler.GetJails)
			jails.GET("/:name", params.JailHandler.GetJail)
			jails.GET("/:name/status", params.Fail2banHandler.GetJailStatus)
			jails.POST("", params.JailHandler.CreateJail)
			jails.PUT("/:name", params.JailHandler.UpdateJail)
			jails.DELETE("/:name", params.JailHandler.DeleteJail)
			jails.POST("/:name/toggle", params.JailHandler.ToggleJail)
		}

		// 默认配置管理
		defaults := authenticated.Group("/defaults")
		{
			defaults.GET("/info", params.DefaultConfigHandler.GetDefaultConfigInfo)
			defaults.POST("/nginx/install", params.DefaultConfigHandler.InstallNginxDefaults)
			defaults.GET("/nginx/filters", params.DefaultConfigHandler.GetNginxFilterTemplates)
			defaults.GET("/nginx/jail-config", params.DefaultConfigHandler.GetNginxJailConfig)
			defaults.GET("/nginx/export", params.DefaultConfigHandler.ExportNginxConfig)
		}

		// SSH监控管理
		ssh := authenticated.Group("/ssh")
		{
			ssh.GET("/stats", params.SSHHandler.GetSSHStats)
			ssh.GET("/logs", params.SSHHandler.GetSSHLogs)
			ssh.GET("/status", params.SSHHandler.GetSSHJailStatus)
			ssh.POST("/ban", params.SSHHandler.BanSSHIP)
			ssh.POST("/unban", params.SSHHandler.UnbanSSHIP)
			ssh.GET("/defaults", params.SSHHandler.GetSSHDefaults)
			ssh.POST("/defaults/install", params.SSHHandler.InstallSSHDefaults)
		}

		// Nginx监控管理
		nginx := authenticated.Group("/nginx")
		{
			nginx.GET("/stats", params.NginxHandler.GetNginxStats)
			nginx.GET("/logs", params.NginxHandler.GetNginxLogs)
			nginx.GET("/status", params.NginxHandler.GetNginxJailStatus)
			nginx.POST("/ban", params.NginxHandler.BanNginxIP)
			nginx.POST("/unban", params.NginxHandler.UnbanNginxIP)
			nginx.GET("/defaults", params.NginxHandler.GetNginxDefaults)
			nginx.GET("/defaults/advanced", params.NginxHandler.GetNginxAdvancedDefaults)
			nginx.POST("/defaults/install", params.NginxHandler.InstallNginxDefaults)
			nginx.POST("/defaults/advanced/install", params.NginxHandler.InstallNginxAdvancedDefaults)
		}

		// 智能分析管理
		intelligent := authenticated.Group("/intelligent")
		{
			intelligent.GET("/threats", params.IntelligentHandler.GetCurrentThreats)
			intelligent.GET("/scan-result", params.IntelligentHandler.GetScanResult)
			intelligent.GET("/stats", params.IntelligentHandler.GetThreatStats)
			intelligent.POST("/ban", params.IntelligentHandler.ManualBanIP)
			intelligent.POST("/analyze-log", params.IntelligentHandler.AnalyzeLogFile)
			intelligent.POST("/analyze-access-log", params.IntelligentHandler.AnalyzeAccessLog)
		}
	}

	params.Logger.Info("Router configured successfully")
	return r
}

// setupStaticFiles 设置静态文件处理
func setupStaticFiles(r *gin.Engine, staticFiles embed.FS) {
	// 设置静态文件
	staticFS, _ := fs.Sub(staticFiles, "web/static")
	r.StaticFS("/static", http.FS(staticFS))

	// 加载 HTML 模板
	templ := template.Must(template.New("").ParseFS(staticFiles, "web/templates/*.html"))
	r.SetHTMLTemplate(templ)
}

// RouterModule 路由模块
var RouterModule = fx.Module("router",
	fx.Provide(NewRouter),
)
