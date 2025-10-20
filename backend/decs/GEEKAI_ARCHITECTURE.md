# Fail2Ban Web 项目架构重构说明

## 项目概述

本项目参考 [geekai](https://github.com/yangjian102621/geekai) 项目的架构风格，使用 `uber-go/fx` 依赖注入框架进行了全面重构。

## 核心架构

### 1. 核心包 (core/)

参考 geekai 的 `api/core` 结构，创建了独立的核心包：

```
core/
├── types.go        # 核心类型定义（AppConfig, SystemConfig）
├── app_server.go   # 应用服务器（AppServer）
└── config.go       # 配置加载
```

#### AppServer 结构

```go
type AppServer struct {
    Debug     bool
    Config    *AppConfig
    Engine    *gin.Engine
    Logger    *logrus.Logger
    SysConfig *SystemConfig
}
```

**主要功能：**
- `NewServer()` - 创建应用服务器
- `Init()` - 初始化中间件和静态文件
- `Run()` - 启动HTTP服务器
- 内置中间件：CORS、错误处理、静态资源

### 2. Handler 结构

所有 Handler 都继承 `BaseHandler`，与 geekai 保持一致：

```go
// internal/handler/base.go
type BaseHandler struct {
    App *core.AppServer
    DB  *gorm.DB
}

// 具体的 Handler 实现
type AuthHandler struct {
    BaseHandler
}

func NewAuthHandler(app *core.AppServer, db *gorm.DB) *AuthHandler {
    return &AuthHandler{
        BaseHandler: BaseHandler{
            App: app,
            DB:  db,
        },
    }
}
```

### 3. main.go 结构

参考 geekai 的单文件风格，所有 fx 组装逻辑都在 main.go 中：

```go
func main() {
    app := fx.New(
        // 1. 提供配置
        fx.Provide(func() *core.AppConfig {
            return core.LoadConfig()
        }),
        
        // 2. 提供日志
        fx.Provide(func() *logrus.Logger {
            // ...
        }),
        
        // 3. 创建应用服务器
        fx.Provide(core.NewServer),
        
        // 4. 初始化服务器
        fx.Invoke(func(s *core.AppServer) {
            s.Init(staticFiles)
        }),
        
        // 5. 提供数据库
        fx.Provide(func(config *core.AppConfig) (*gorm.DB, error) {
            // ...
        }),
        
        // 6. 创建服务
        fx.Provide(service.NewFail2BanService),
        fx.Provide(service.NewJailService),
        // ...
        
        // 7. 创建 Handler
        fx.Provide(handler.NewAuthHandler),
        fx.Provide(handler.NewFail2BanHandler),
        // ...
        
        // 8. 注册路由
        fx.Invoke(registerRoutes),
        
        // 9. 启动服务器
        fx.Invoke(func(s *core.AppServer, db *gorm.DB) {
            go func() {
                err := s.Run(db)
                if err != nil {
                    log.Fatal(err)
                }
            }()
        }),
        
        // 10. 生命周期管理
        fx.Provide(NewAppLifeCycle),
        fx.Invoke(func(lifecycle fx.Lifecycle, lc *AppLifecycle) {
            lifecycle.Append(fx.Hook{
                OnStart: lc.OnStart,
                OnStop:  lc.OnStop,
            })
        }),
    )
    
    // 应用启动和信号处理
    // ...
}
```

### 4. 路由注册

使用独立的 `registerRoutes` 函数，接收所有 Handler：

```go
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
    
    // 公共路由
    r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", gin.H{
            "title": "Fail2Ban 管理面板",
        })
    })
    
    // API 路由组
    api := r.Group("/api/v1")
    
    // 认证路由
    auth := api.Group("/auth")
    {
        auth.POST("/login", authHandler.Login)
        auth.POST("/refresh", authHandler.RefreshToken)
        auth.GET("/profile", authHandler.GetProfile)
    }
    
    // 其他路由...
}
```

## 关键特性

### 1. 依赖注入

使用 `fx.Provide` 提供依赖，`fx.Invoke` 消费依赖：

```go
// 提供服务
fx.Provide(service.NewFail2BanService),

// 使用服务
fx.Invoke(func(s *service.Fail2BanService) {
    // 使用服务
}),
```

### 2. 生命周期管理

通过 `fx.Lifecycle` 管理应用生命周期：

```go
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
})
```

### 3. 配置管理

使用环境变量或默认值进行配置：

```go
func LoadConfig() *AppConfig {
    listen := os.Getenv("LISTEN_ADDR")
    if listen == "" {
        listen = ":8099"
    }
    
    dbPath := os.Getenv("DB_PATH")
    if dbPath == "" {
        dbPath = "fail2ban_web.db"
    }
    
    return &AppConfig{
        Listen:    listen,
        DBPath:    dbPath,
        StaticDir: "./web/static",
        StaticURL: "http://localhost:8099/static",
    }
}
```

### 4. 统一响应格式

在 `core.AppServer` 中提供统一的响应方法：

```go
type ApiResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, ApiResponse{
        Code:    0,
        Message: "success",
        Data:    data,
    })
}

func Error(c *gin.Context, message string) {
    c.JSON(http.StatusOK, ApiResponse{
        Code:    -1,
        Message: message,
        Data:    nil,
    })
}
```

## 项目目录结构

```
fail2ban_web/
├── main.go                 # 主入口，所有 fx 组装逻辑
├── core/                   # 核心包（参考 geekai api/core）
│   ├── app_server.go       # 应用服务器
│   ├── config.go           # 配置加载
│   └── types.go            # 核心类型
├── config/                 # 配置包（旧的，逐步迁移）
│   └── config.go
├── internal/
│   ├── handler/            # HTTP 处理器
│   │   ├── base.go         # 基础 Handler
│   │   ├── auth.go         # 认证 Handler
│   │   ├── fail2ban.go     # Fail2Ban Handler
│   │   └── ...
│   ├── service/            # 业务逻辑服务
│   │   ├── fail2ban.go
│   │   ├── jail.go
│   │   └── ...
│   ├── model/              # 数据模型
│   └── middleware/         # 中间件
├── web/                    # 前端资源
│   ├── static/
│   └── templates/
├── go.mod
└── go.sum
```

## 与 geekai 的对比

### 相同点

1. **核心包结构** - 使用独立的 `core` 包封装核心功能
2. **AppServer 设计** - 统一的应用服务器结构
3. **Handler 模式** - BaseHandler + 具体 Handler
4. **fx 使用方式** - 单文件 main.go，所有依赖组装在一起
5. **路由注册** - 使用 fx.Invoke 注册路由
6. **生命周期管理** - 使用 fx.Lifecycle 管理启动/停止

### 不同点

1. **数据库** - geekai 使用 MySQL，我们使用 SQLite
2. **配置系统** - geekai 使用 TOML，我们使用环境变量
3. **业务逻辑** - Fail2Ban 管理 vs AI 聊天应用
4. **中间件** - 我们简化了认证中间件

## 运行要求

### 环境变量

```bash
# 服务器监听地址（可选，默认:8099）
export LISTEN_ADDR=":8099"

# 数据库路径（可选，默认 fail2ban_web.db）
export DB_PATH="fail2ban_web.db"

# 管理员凭据（可选，默认 admin/admin）
export ADMIN_USERNAME="admin"
export ADMIN_PASSWORD="admin"
export ADMIN_EMAIL="admin@example.com"

# 静态资源配置（可选）
export STATIC_DIR="./web/static"
export STATIC_URL="http://localhost:8099/static"
```

### 依赖

- Go 1.24.1+
- Fail2Ban 服务已安装
- sudo 权限（用于执行 fail2ban-client 命令）

## 编译和运行

```bash
# 编译
go build -o fail2ban_web

# 运行
./fail2ban_web

# 或直接运行
go run main.go
```

## 版本历史

### 当前版本 (geekai 风格)
- ✅ 使用 `core` 包封装核心功能
- ✅ 参考 geekai 的 AppServer 设计
- ✅ 单文件 main.go 风格
- ✅ BaseHandler 模式
- ✅ 统一的路由注册
- ✅ 生命周期管理

### 之前版本
- `main_fx_inline.go.bak` - Fx 单文件风格（无 core 包）
- `main_modules_style.go.bak` - 模块化风格（app/ 目录）
- `main_old.go.bak` - 原始版本（无 Fx）

## 最佳实践

1. **Handler 设计** - 所有 Handler 继承 BaseHandler
2. **服务注入** - 通过构造函数注入依赖
3. **错误处理** - 使用统一的响应格式
4. **生命周期** - 合理使用 fx.Lifecycle 管理资源
5. **配置管理** - 优先使用环境变量

## 参考资源

- [geekai 项目](https://github.com/yangjian102621/geekai)
- [uber-go/fx 文档](https://uber-go.github.io/fx/)
- [Gin Web Framework](https://gin-gonic.com/)

---

**最后更新：** 2025-10-19  
**当前版本：** geekai 风格（带 core 包）  
**参考项目：** https://github.com/yangjian102621/geekai/tree/main/api
