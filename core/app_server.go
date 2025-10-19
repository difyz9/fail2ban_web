package core

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AppServer 应用服务器
type AppServer struct {
	Debug     bool
	Config    *AppConfig
	Engine    *gin.Engine
	Logger    *logrus.Logger
	SysConfig *SystemConfig // 系统配置缓存
}

// NewServer 创建应用服务器
func NewServer(config *AppConfig, logger *logrus.Logger) *AppServer {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	
	return &AppServer{
		Debug:  false,
		Config: config,
		Engine: gin.Default(),
		Logger: logger,
	}
}

// Init 初始化服务器中间件和静态文件
func (s *AppServer) Init(staticFiles embed.FS) {
	// 添加中间件
	s.Engine.Use(corsMiddleware())
	s.Engine.Use(errorHandler(s.Logger))
	
	// 设置静态文件
	setupStaticFiles(s.Engine, staticFiles)
}

// Run 运行服务器
func (s *AppServer) Run(db *gorm.DB) error {
	// 这里可以从数据库加载系统配置
	s.SysConfig = &SystemConfig{
		Title:       "Fail2Ban 管理面板",
		Description: "Fail2Ban Web管理界面",
		Version:     "1.0.0",
	}
	
	s.Logger.Infof("Starting HTTP server on %s", s.Config.Listen)
	s.Logger.Infof("Access the management panel at http://localhost%s", s.Config.Listen)
	
	return s.Engine.Run(s.Config.Listen)
}

// corsMiddleware 跨域中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin")
			c.Header("Access-Control-Max-Age", "172800")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		
		if method == http.MethodOptions {
			c.JSON(http.StatusOK, "ok")
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// errorHandler 全局异常处理中间件
func errorHandler(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("Handler Panic: %v", r)
				debug.PrintStack()
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Internal server error",
					"data":    nil,
				})
				c.Abort()
			}
		}()
		
		c.Next()
	}
}

// setupStaticFiles 设置静态文件和模板
func setupStaticFiles(r *gin.Engine, staticFiles embed.FS) {
	// 设置静态文件
	staticFS, err := fs.Sub(staticFiles, "web/static")
	if err != nil {
		log.Fatal("Failed to load static files:", err)
	}
	r.StaticFS("/static", http.FS(staticFS))
	
	// 加载 HTML 模板
	templ := template.Must(template.New("").ParseFS(staticFiles, "web/templates/*.html"))
	r.SetHTMLTemplate(templ)
}

// ApiResponse 统一API响应结构
type ApiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, ApiResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, message string) {
	c.JSON(http.StatusOK, ApiResponse{
		Code:    -1,
		Message: message,
		Data:    nil,
	})
}

// ErrorWithCode 带状态码的错误响应
func ErrorWithCode(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, ApiResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}
