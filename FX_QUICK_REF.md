# 🚀 Fx 重构 - 快速参考

## 一键启动

```bash
# 编译
make build

# 运行
./build/fail2ban-web

# 测试
curl http://localhost:8092/api/v1/auth/login \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

## 核心变更

### main.go (200行 → 10行)

```go
//go:embed web
var staticFiles embed.FS

func main() {
    fxApp := app.NewApp(staticFiles)
    fxApp.Run()
}
```

## 模块结构

```
app/
├── app.go       → 组装所有模块
├── config.go    → 配置
├── logger.go    → 日志（zap）
├── database.go  → 数据库（GORM）
├── services.go  → 业务服务
├── handlers.go  → HTTP处理器
├── router.go    → 路由（Gin）
└── server.go    → HTTP服务器
```

## 依赖注入示例

```go
// 自动注入依赖
type ServiceParams struct {
    fx.In
    Config *config.Config
    DB     *gorm.DB
    Logger *zap.Logger
}

func NewServices(params ServiceParams) ServiceResult {
    // fx 自动提供所有依赖
    service := NewService(params.DB)
    return ServiceResult{Service: service}
}
```

## 生命周期钩子

```go
lc.Append(fx.Hook{
    OnStart: func(ctx context.Context) error {
        // 启动时执行
        return resource.Init()
    },
    OnStop: func(ctx context.Context) error {
        // 关闭时执行
        return resource.Close()
    },
})
```

## 常用命令

```bash
# 构建
make build

# 运行
make run

# 清理
make clean

# 测试
make test

# 热重载
make dev

# 查看帮助
make help
```

## 文档

- 完整指南: `FX_REFACTOR_GUIDE.md`
- 重构总结: `FX_REFACTOR_SUMMARY.md`
- 旧代码备份: `main_old.go.bak`

## 优势

✅ main.go 减少 95% 代码  
✅ 自动依赖注入  
✅ 生命周期管理  
✅ 模块化设计  
✅ 易于测试  
✅ 详细日志  

## 技术栈

- **Fx**: v1.20.1 (依赖注入)
- **Zap**: v1.26.0 (日志)
- **Gin**: v1.9.1 (HTTP)
- **GORM**: v1.25.5 (ORM)
- **SQLite**: v1.5.4 (数据库)

## 状态

✅ **生产就绪**  
📅 重构日期: 2024-01-15  
🔖 版本: v0.1.0
