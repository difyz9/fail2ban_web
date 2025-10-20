# fail2ban_web

一款用于可视化与管理 Fail2Ban 日志与封禁记录的轻量级 Web 管理工具（Go + Gin + GORM + SQLite）。

项目基于 geekai 的工程结构改造，使用 uber-go/fx 实现依赖注入，配置采用 TOML（参照 geekai 的 core/config.go 实现）。

**✨ 特色**：前端使用 Next.js 15，构建后通过 Go embed 嵌入到单一二进制文件，实现真正的一键部署！

## 📚 文档导航

- [前端构建详细指南](FRONTEND_BUILD.md) - 完整的前端构建和 embed 使用说明
- [配置系统说明](CONFIG_SYSTEM.md) - TOML 配置文件详解
- [Embed 完成总结](EMBED_SUMMARY.md) - 前端嵌入功能实现总结

---

## 关键特性

- 基于 Go 的单二进制部署（Gin + fx）
- 使用 GORM + SQLite 存储数据（轻量化，无外部 DB 依赖）
- TOML 配置文件（`config.toml`）自动生成与加载
- 简单的管理员认证（配置文件中定义 Admin 用户）
- Web 前端（静态资源位于 `web/static`，模板在 `web/templates`）
- 可通过 Docker 部署（仓库包含 `Dockerfile` 与 `docker-compose.yml`）

---

## 目录结构（摘录）

```
fail2ban_web/
├── frontend/              # Next.js 前端源码
│   ├── src/              # React 组件和页面
│   ├── public/           # 静态资源
│   └── package.json      # 前端依赖
├── web/                  # 前端构建产物（自动生成，通过 embed 嵌入）
├── cmd/                  # 工具与辅助命令
├── core/                 # 应用核心（AppServer、config、types 等）
│   ├── app_server.go     # 应用服务器
│   ├── config.go         # 配置加载（TOML）
│   └── types.go          # 核心类型定义
├── internal/
│   ├── handler/          # HTTP 处理器
│   ├── service/          # 业务服务层
│   ├── middleware/       # 中间件
│   └── model/            # 数据模型
├── build_frontend.sh     # 前端构建脚本
├── Dockerfile            # Docker 镜像构建
├── docker-compose.yml    # Docker Compose 配置
├── Makefile             # 构建命令
└── main.go              # 应用入口（fx 依赖注入）
```

---

## 快速开始（本地）

### 前置项

- Go 1.24+
- Node.js 18+ 和 npm（用于构建前端）
- (可选) Docker

### 克隆仓库

```bash
git clone https://github.com/difyz9/fail2ban_web
cd fail2ban_web
```

### 构建

项目采用前后端分离架构，前端使用 Next.js，后端使用 Go。

#### 完整构建（推荐）

一键构建前端和后端：

```bash
make build-full
```

这个命令会：
1. 构建 Next.js 前端项目（`frontend/` 目录）
2. 将构建产物复制到 `web/` 目录
3. 通过 Go embed 将前端资源嵌入到二进制文件
4. 生成可执行文件 `build/fail2ban-web`

#### 仅构建前端

```bash
make build-frontend
```

#### 仅构建后端

```bash
make build
```

#### 生产环境构建

```bash
make build-prod
```

生产构建会自动构建前端并优化二进制文件大小（使用 `-ldflags "-s -w"`）。

### 运行

首次运行会自动创建 `config.toml`：

```bash
./build/fail2ban-web
# 或者如果是直接用 go build 构建的
./fail2ban_web

# 输出示例：
# 2025/10/19 17:31:17 Loading config file: config.toml
# 2025/10/19 17:31:17 Creating new config file: config.toml
```

程序默认监听 `:8099`，访问 `http://localhost:8099` 即可使用 Web 界面。

### 使用自定义配置文件

```bash
CONFIG_FILE="production.toml" ./fail2ban_web
```

---

## 配置（`config.toml`）

程序使用 TOML 文件作为配置，默认配置示例：

```toml
Listen = ":8099"
StaticDir = "./web/static"
StaticURL = "http://localhost:8099/static"
DBPath = "fail2ban_web.db"

[Admin]
  Username = "admin"
  Password = "admin123"
  Email = "admin@example.com"
```

敏感信息建议不要写入仓库，编辑后请确保权限设置合理：

```bash
chmod 600 config.toml
```

---

## Docker 部署

仓库包含 `Dockerfile` 与 `docker-compose.yml`，可以通过 Docker 快速部署：

```bash
# 使用本地构建镜像
docker build -t fail2ban_web:latest .

# 或者使用 docker-compose
docker-compose up -d --build
```

注意：容器化部署时请将 `config.toml`、数据库文件与静态目录挂载到宿主机以保持持久化。

---

## 开发指南

### 项目架构

- **前端**：Next.js 15 + React + TypeScript，位于 `frontend/` 目录
- **后端**：Go + Gin + fx（依赖注入）+ GORM + SQLite
- **配置**：TOML 格式配置文件（参照 geekai 实现）
- **部署**：单二进制文件，前端资源通过 Go embed 嵌入

### 开发流程

1. **前端开发**

```bash
cd frontend
npm install
npm run dev  # 开发服务器运行在 http://localhost:3000
```

2. **后端开发**

```bash
# 使用热重载（需要先安装 air）
make dev

# 或者直接运行
go run main.go
```

3. **完整构建测试**

```bash
make build-full
./build/fail2ban-web
```

### 代码组织

- 代码使用 fx 进行依赖注入，应用入口在 `main.go`
- 新增配置项请修改 `core/types.go` 的 `AppConfig` 并在 `core/config.go` 中处理默认值与保存
- 处理器位于 `internal/handler`，服务逻辑在 `internal/service`
- 前端页面在 `frontend/src/app`，组件在 `frontend/src/components`

### 前端构建说明

- 前端构建产物会自动复制到 `web/` 目录
- `web/` 目录下的所有文件通过 `//go:embed web` 嵌入到 Go 二进制文件
- 生产部署时只需要单个二进制文件，无需额外的静态资源目录

---

## 常见操作

### 构建命令

```bash
# 查看所有可用命令
make help

# 构建前端
make build-frontend

# 完整构建（前端 + 后端）
make build-full

# 生产环境构建
make build-prod

# 构建所有平台版本
make build-all
```

### 开发调试

```bash
# 前端开发
cd frontend && npm run dev

# 后端热重载
make dev

# 运行测试
make test

# 代码格式化
make fmt
```

### 查看日志

```bash
# 直接运行时 stdout
./build/fail2ban-web

# 后台运行并查看日志（macOS / Linux）
nohup ./build/fail2ban-web > fail2ban_web.log 2>&1 &
tail -f fail2ban_web.log
```

### 重置配置

```bash
# 删除配置并重置默认
rm config.toml
./build/fail2ban-web
```

---

## 故障排查

- 如果程序无法启动，先检查 `config.toml` 是否存在且 TOML 格式正确。
- 检查数据库文件权限，确保进程有读写权限。
- 若出现依赖问题，运行 `go mod tidy` 以补全模块。

---

## 贡献与许可

欢迎提交 Issue 与 Pull Request。请在 PR 中描述变更及测试步骤。

---
