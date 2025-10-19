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
	// 检查是否是 Next.js 构建产物结构（web/ 目录下直接有 HTML 文件）
	if _, err := fs.Stat(staticFiles, "web/index.html"); err == nil {
		// Next.js 静态导出模式
		log.Println("Detected Next.js build output, serving as static files")
		
		// 提供整个 web 目录作为静态文件服务
		webFS, err := fs.Sub(staticFiles, "web")
		if err != nil {
			log.Fatal("Failed to load web files:", err)
		}
		
		// 使用 NoRoute 处理所有静态文件请求（包括 _next, HTML, 图片等）
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			
			log.Printf("[DEBUG] NoRoute handling: %s", path)
			
			// 默认路径指向 index.html
			if path == "/" {
				path = "/index.html"
			} else if len(path) > 1 && path[len(path)-1] == '/' {
				// 移除尾部斜杠，尝试访问同名 HTML 文件
				path = path[:len(path)-1] + ".html"
			} else {
				// 尝试多种文件路径
				// 1. 原始路径
				if _, err := fs.Stat(webFS, path[1:]); err != nil {
					log.Printf("[DEBUG] File not found (original): %s, error: %v", path[1:], err)
					// 2. 添加 .html 后缀
					if _, err := fs.Stat(webFS, path[1:]+".html"); err == nil {
						path = path + ".html"
						log.Printf("[DEBUG] Found with .html suffix: %s", path)
					} else {
						// 3. 尝试目录下的 index.html
						if _, err := fs.Stat(webFS, path[1:]+"/index.html"); err == nil {
							path = path + "/index.html"
							log.Printf("[DEBUG] Found index.html in directory: %s", path)
						}
					}
				} else {
					log.Printf("[DEBUG] File found (original): %s", path[1:])
				}
			}
			
			log.Printf("[DEBUG] Attempting to open: %s", path[1:])
			
			// 读取文件
			file, err := webFS.Open(path[1:])
			if err != nil {
				log.Printf("[DEBUG] Failed to open file: %s, error: %v", path[1:], err)
				// 如果找不到文件，返回 404.html 或默认 404
				file404, err := webFS.Open("404.html")
				if err == nil {
					defer file404.Close()
					content, _ := io.ReadAll(file404)
					c.Data(http.StatusNotFound, "text/html; charset=utf-8", content)
					return
				}
				c.String(http.StatusNotFound, "404 page not found")
				return
			}
			defer file.Close()
			
			// 读取文件内容
			content, err := io.ReadAll(file)
			if err != nil {
				c.String(http.StatusInternalServerError, "Error reading file")
				return
			}
			
			// 根据文件扩展名设置 Content-Type
			contentType := "application/octet-stream"
			switch {
			case len(path) > 5 && path[len(path)-5:] == ".html":
				contentType = "text/html; charset=utf-8"
			case len(path) > 4 && path[len(path)-4:] == ".css":
				contentType = "text/css; charset=utf-8"
			case len(path) > 3 && (path[len(path)-3:] == ".js" || path[len(path)-3:] == ".mjs"):
				contentType = "application/javascript; charset=utf-8"
			case len(path) > 4 && path[len(path)-4:] == ".svg":
				contentType = "image/svg+xml"
			case len(path) > 4 && path[len(path)-4:] == ".png":
				contentType = "image/png"
			case len(path) > 4 && path[len(path)-4:] == ".jpg" || len(path) > 5 && path[len(path)-5:] == ".jpeg":
				contentType = "image/jpeg"
			case len(path) > 4 && path[len(path)-4:] == ".ico":
				contentType = "image/x-icon"
			case len(path) > 5 && path[len(path)-5:] == ".json":
				contentType = "application/json; charset=utf-8"
			case len(path) > 4 && path[len(path)-4:] == ".txt":
				contentType = "text/plain; charset=utf-8"
			}
			
			c.Data(http.StatusOK, contentType, content)
		})
	} else {
		// 传统的 templates + static 模式
		log.Println("Using traditional templates + static files mode")
		
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
