# fail2ban_web

一款用于可视化与管理 Fail2Ban 日志与封禁记录的轻量级 Web 管理工具（Go + Gin + GORM + SQLite）。

项目基于 geekai 的工程结构改造，使用 uber-go/fx 实现依赖注入，配置采用 TOML（参照 geekai 的 core/config.go 实现）。

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

- `cmd/` - 工具与辅助命令
- `core/` - 应用核心（AppServer、config、types 等）
- `internal/handler/` - HTTP 处理器（登录、默认、nginx、ssh 等）
- `internal/service/` - 业务服务层（jail、nginx、ssh、whitelist、user）
- `web/` - 前端静态资源与模板
- `Dockerfile`, `docker-compose.yml` - 容器化部署配置

---

## 快速开始（本地）

### 前置项

- Go 1.24+
- (可选) Docker

### 克隆仓库

```bash
git clone <repo-url> fail2ban_web
cd fail2ban_web
```

### 构建

```bash
go build -o fail2ban_web
```

### 运行

首次运行会自动创建 `config.toml`：

```bash
./fail2ban_web
# 输出示例：
# 2025/10/19 17:31:17 Loading config file: config.toml
# 2025/10/19 17:31:17 Creating new config file: config.toml
```

程序默认监听 `:8092`，静态资源 URL 为 `http://localhost:8092/static`。

### 使用自定义配置文件

```bash
CONFIG_FILE="production.toml" ./fail2ban_web
```

---

## 配置（`config.toml`）

程序使用 TOML 文件作为配置，默认配置示例：

```toml
Listen = ":8092"
StaticDir = "./web/static"
StaticURL = "http://localhost:8092/static"
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

- 代码使用 fx 进行依赖注入，应用入口在 `main.go`。
- 新增配置项请修改 `core/types.go` 的 `AppConfig` 并在 `core/config.go` 中处理默认值与保存。
- 处理器位于 `internal/handler`，服务逻辑在 `internal/service`。
- 前端模板位于 `web/templates`，静态资源在 `web/static`。

---

## 常见操作

- 生成可执行文件：

```bash
go build -o fail2ban_web
```

- 查看日志：

```bash
# 直接运行时 stdout
./fail2ban_web

# 后台运行并查看日志（macOS / Linux）
nohup ./fail2ban_web > fail2ban_web.log 2>&1 &
tail -f fail2ban_web.log
```

- 删除配置并重置默认：

```bash
rm config.toml
./fail2ban_web
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
