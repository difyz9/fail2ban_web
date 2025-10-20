# Fx 架构对比文档

## 概述

本文档对比两种 Fx 依赖注入架构的实现方式：
1. **模块化风格** (crypto-wallet-backend 风格) - 已废弃
2. **单文件风格** (geekai 风格) - **当前使用**

## 两种实现方式对比

### 1. 模块化风格（已废弃）

**文件结构：**
```
app/
├── app.go          # 主模块组装
├── config.go       # 配置模块
├── database.go     # 数据库模块
├── handlers.go     # Handler 模块
├── logger.go       # 日志模块
├── router.go       # 路由模块
├── server.go       # 服务器模块
└── services.go     # 服务模块
main.go             # 简单入口
```

**main.go 代码：**
```go
package main

import (
    "embed"
    "fail2ban-web/app"
    "go.uber.org/fx"
)

//go:embed web
var staticFiles embed.FS

func main() {
    fx.New(
        fx.Supply(staticFiles),
        app.NewApp(),
    ).Run()
}
```

**app/app.go 代码：**
```go
package app

import "go.uber.org/fx"

func NewApp() fx.Option {
    return fx.Options(
        ConfigModule,
        LoggerModule,
        DatabaseModule,
        ServicesModule,
        HandlersModule,
        RouterModule,
        ServerModule,
    )
}
```

**优点：**
- ✅ 代码组织清晰，模块职责分明
- ✅ 每个模块独立，易于单元测试
- ✅ 模块间依赖关系清晰
- ✅ 主入口文件非常简洁

**缺点：**
- ❌ 需要维护额外的 app 目录和多个文件
- ❌ 模块之间的依赖关系不够直观
- ❌ 不符合 geekai 项目的实现风格
- ❌ 学习曲线略高，需要理解 fx.Module 概念

---

### 2. 单文件风格（当前使用）

**文件结构：**
```
main.go            # 所有 Fx 组装逻辑
internal/
├── handler/       # Handler 实现
├── service/       # Service 实现
└── model/         # 数据模型
```

**main.go 代码：**
```go
package main

import (
    "context"
    "fail2ban-web/config"
    "fail2ban-web/internal/handler"
    "fail2ban-web/internal/service"
    "go.uber.org/fx"
)

func main() {
    app := fx.New(
        // 提供配置
        fx.Provide(func() *config.Config {
            return config.LoadConfig()
        }),

        // 提供日志
        fx.Provide(func() *logrus.Logger {
            logger := logrus.New()
            // ...
            return logger
        }),

        // 提供数据库
        fx.Provide(func() (*gorm.DB, error) {
            // ...
        }),

        // 创建服务
        fx.Provide(service.NewFail2BanService),
        fx.Provide(service.NewJailService),
        // ...

        // 创建 Handler
        fx.Provide(handler.NewAuthHandler),
        fx.Provide(handler.NewFail2BanHandler),
        // ...

        // 注册路由
        fx.Invoke(registerRoutes),

        // 生命周期管理
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

func registerRoutes(
    lifecycle fx.Lifecycle,
    cfg *config.Config,
    authHandler *handler.AuthHandler,
    // ...
) {
    // 路由注册逻辑
}
```

**优点：**
- ✅ 所有依赖关系在一个文件中，一目了然
- ✅ 符合 geekai 项目的实现风格
- ✅ 减少文件数量，降低复杂度
- ✅ Fx 的使用更加直观，学习成本低
- ✅ 依赖注入流程清晰可见

**缺点：**
- ❌ main.go 文件较大（约 350 行）
- ❌ 所有配置集中在一个文件，修改时可能不够灵活

---

## 核心差异分析

### 1. 依赖提供方式

**模块化风格：**
```go
var ServicesModule = fx.Module("services",
    fx.Provide(
        service.NewFail2BanService,
        service.NewJailService,
        service.NewSSHService,
        // ...
    ),
)
```

**单文件风格：**
```go
fx.Provide(service.NewFail2BanService),
fx.Provide(service.NewJailService),
fx.Provide(service.NewSSHService),
// ...
```

### 2. 路由注册方式

**模块化风格：**
```go
// 在 app/router.go 中
var RouterModule = fx.Module("router",
    fx.Invoke(func(
        lifecycle fx.Lifecycle,
        cfg *config.Config,
        handlers *Handlers,
    ) {
        r := gin.Default()
        registerRoutes(r, handlers)
        // ...
    }),
)
```

**单文件风格：**
```go
// 在 main.go 中
fx.Invoke(registerRoutes),

// registerRoutes 直接接收所有 handler
func registerRoutes(
    lifecycle fx.Lifecycle,
    cfg *config.Config,
    authHandler *handler.AuthHandler,
    fail2banHandler *handler.Fail2BanHandler,
    // ...
) {
    // 路由注册逻辑
}
```

### 3. 生命周期管理

**模块化风格：**
```go
// 在 app/server.go 中
var ServerModule = fx.Module("server",
    fx.Invoke(func(lifecycle fx.Lifecycle, r *gin.Engine, cfg *config.Config) {
        lifecycle.Append(fx.Hook{
            OnStart: func(ctx context.Context) error {
                go r.Run(":8099")
                return nil
            },
        })
    }),
)
```

**单文件风格：**
```go
// 在 main.go 中
fx.Provide(NewAppLifeCycle),
fx.Invoke(func(lifecycle fx.Lifecycle, lc *AppLifecycle) {
    lifecycle.Append(fx.Hook{
        OnStart: lc.OnStart,
        OnStop:  lc.OnStop,
    })
}),
```

## 迁移指南

### 从模块化风格迁移到单文件风格

1. **备份现有代码：**
   ```bash
   mv main.go main_modules_style.go.bak
   ```

2. **创建新的 main.go：**
   - 将 `app/config.go` 中的配置提供逻辑移到 main.go
   - 将 `app/logger.go` 中的日志提供逻辑移到 main.go
   - 将 `app/database.go` 中的数据库提供逻辑移到 main.go
   - 将 `app/services.go` 中的服务提供改为独立的 fx.Provide
   - 将 `app/handlers.go` 中的 Handler 提供改为独立的 fx.Provide
   - 将 `app/router.go` 中的路由注册改为 fx.Invoke
   - 将 `app/server.go` 中的服务器启动改为生命周期钩子

3. **删除 app 目录：**
   ```bash
   rm -rf app/
   ```

4. **测试运行：**
   ```bash
   go run main.go
   ```

## 最佳实践建议

### 选择单文件风格的场景：
- ✅ 中小型项目（< 20 个服务和 Handler）
- ✅ 团队成员对 Fx 不太熟悉
- ✅ 需要与 geekai 等项目保持一致的代码风格
- ✅ 希望快速理解整个应用的依赖关系

### 选择模块化风格的场景：
- ✅ 大型项目（> 50 个服务和 Handler）
- ✅ 团队成员熟悉 Fx 的高级特性
- ✅ 需要高度模块化和可测试性
- ✅ 不同团队负责不同模块

## 性能对比

两种实现方式在**运行时性能上没有区别**，因为它们最终生成的依赖图完全相同。差异仅在于：
- **开发体验**：单文件风格更直观，模块化风格更灵活
- **代码组织**：单文件风格集中，模块化风格分散
- **学习曲线**：单文件风格更低，模块化风格需要理解 fx.Module

## 参考项目

### geekai 项目
- **仓库：** https://github.com/yangjian102621/geekai
- **特点：** 使用单文件风格，所有 fx.Provide 和 fx.Invoke 都在 main.go 中
- **适用场景：** AI 聊天应用，中等复杂度

### crypto-wallet-backend 项目
- **特点：** 使用模块化风格，将不同职责拆分到独立的模块文件
- **适用场景：** 加密货币钱包后端，高复杂度

## 总结

本项目**最终采用单文件风格（geekai 风格）**，原因如下：

1. **项目规模适中**：约 15 个服务和 7 个 Handler，单文件完全可控
2. **代码一致性**：与参考项目 geekai 保持一致
3. **维护简单**：所有依赖关系在一个文件中，修改方便
4. **学习成本低**：团队成员更容易理解和上手

如果项目未来规模扩大（超过 50 个组件），可以考虑迁移回模块化风格。

---

**最后更新：** 2025-10-19
**当前版本：** geekai 风格（单文件）
**备份文件：** 
- `main_old.go.bak` - 原始版本（无 Fx）
- `main_modules_style.go.bak` - 模块化风格版本
