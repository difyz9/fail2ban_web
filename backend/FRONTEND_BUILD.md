# 前端构建与 Embed 使用指南

## 概述

本项目采用前后端分离架构，前端使用 Next.js 15，后端使用 Go。前端构建产物通过 Go 的 `embed` 特性嵌入到二进制文件中，实现单文件部署。

## 构建流程

### 1. 前端构建

前端项目位于 `frontend/` 目录，使用 Next.js 静态导出模式。

#### 配置文件 (`frontend/next.config.ts`)

```typescript
export default {
  eslint: {
    ignoreDuringBuilds: true,  // 生产构建时忽略 ESLint 错误
  },
  typescript: {
    ignoreBuildErrors: true,    // 生产构建时忽略 TypeScript 错误
  },
  output: 'export',              // 导出静态 HTML
  distDir: 'out',                // 输出目录为 out
};
```

#### 构建脚本 (`build_frontend.sh`)

自动化前端构建流程：
1. 检测 `frontend/` 目录
2. 运行 `npm run build`
3. 清空 `web/` 目录
4. 复制 `frontend/out/` 到 `web/`

使用方法：
```bash
./build_frontend.sh
```

### 2. 后端 Embed

#### main.go 中的 Embed 配置

```go
//go:embed web
var staticFiles embed.FS
```

这会将整个 `web/` 目录嵌入到编译后的二进制文件中。

#### 静态文件服务 (`core/app_server.go`)

系统自动检测并选择合适的服务模式：

**Next.js 模式** （检测到 `web/index.html`）：
- 使用 `NoRoute` 处理所有请求
- 支持路径重写（如 `/login` → `/login.html`）
- 支持嵌套目录（如 `/dashboard` → `/dashboard.html`）
- 自动返回 `404.html` 处理未找到的页面

**传统模式** （有 `web/templates/` 和 `web/static/`）：
- 使用 Gin 的模板系统
- 静态文件通过 `/static` 路径访问

### 3. 完整构建

#### 使用 Makefile

```bash
# 只构建前端
make build-frontend

# 只构建后端
make build

# 完整构建（前端 + 后端）
make build-full

# 生产环境构建（优化体积）
make build-prod
```

#### Makefile 配置

```makefile
# 完整构建
.PHONY: build-full
build-full: build-frontend build
	@echo "✅ 完整构建完成！"

# 生产构建
.PHONY: build-prod
build-prod: build-frontend
	@echo "生产环境构建..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 $(GOBUILD) -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ 生产环境构建完成！"
```

## 目录结构

```
fail2ban_web/
├── frontend/              # Next.js 前端源码
│   ├── src/
│   ├── public/
│   ├── next.config.ts    # Next.js 配置
│   └── package.json
│
├── web/                  # 前端构建产物（自动生成，通过 embed 嵌入）
│   ├── index.html
│   ├── login.html
│   ├── dashboard/
│   │   ├── index.html
│   │   ├── jails.html
│   │   └── ...
│   ├── _next/
│   │   └── static/      # Next.js 静态资源
│   └── ...
│
├── build_frontend.sh     # 前端构建脚本
├── Makefile             # 构建命令
└── main.go              # 应用入口（embed 配置）
```

## 路由处理

### 前端路由映射

Next.js 导出的文件结构会自动映射到 URL 路径：

| URL 路径 | 文件路径 | 说明 |
|---------|----------|------|
| `/` | `web/index.html` | 首页 |
| `/login` | `web/login.html` | 登录页 |
| `/dashboard` | `web/dashboard.html` | Dashboard 主页 |
| `/dashboard/jails` | `web/dashboard/jails.html` | Jails 管理 |
| `/dashboard/bans` | `web/dashboard/bans.html` | 封禁管理 |

### API 路由

所有 API 请求使用 `/api/v1/` 前缀：

```go
api := r.Group("/api/v1")
api.GET("/health", ...)
api.POST("/auth/login", ...)
```

## 开发流程

### 前端开发

```bash
cd frontend
npm install
npm run dev  # 开发服务器：http://localhost:3000
```

前端开发时可以独立运行，通过配置 `apiClient.ts` 的 API 地址连接到后端。

### 后端开发

```bash
# 方式1：热重载（需要安装 air）
make dev

# 方式2：直接运行
go run main.go
```

后端运行在 `:8099`（可在 `config.toml` 中配置）。

### 联调测试

1. 构建前端：
```bash
make build-frontend
```

2. 运行后端：
```bash
go run main.go
```

3. 访问：`http://localhost:8099`

## 部署

### 本地部署

```bash
# 完整构建
make build-full

# 运行
./build/fail2ban-web
```

### Docker 部署

```bash
# 构建镜像
docker build -t fail2ban-web:latest .

# 运行容器
docker run -d -p 8099:8099 fail2ban-web:latest
```

### 生产部署

```bash
# 生产构建（优化体积）
make build-prod

# 生成的二进制文件位于
ls -lh build/fail2ban-web

# 部署到服务器
scp build/fail2ban-web user@server:/path/to/app/
```

## 文件大小

| 构建类型 | 大小 | 说明 |
|---------|------|------|
| 普通构建 | ~23MB | 包含调试信息 |
| 生产构建 | ~18MB | 使用 `-ldflags "-s -w"` 优化 |
| 压缩后 | ~6MB | 使用 `upx` 压缩 |

## 常见问题

### 1. 前端页面无法访问

**问题**：访问 `/dashboard` 返回 404 或 500

**解决**：
- 确认已运行 `make build-frontend`
- 检查 `web/` 目录下是否有对应的 HTML 文件
- 查看应用启动日志，确认是否检测到 Next.js 构建产物：
  ```
  2025/10/19 20:23:38 Detected Next.js build output, serving as static files
  ```

### 2. 静态资源 404

**问题**：页面加载但样式和脚本无法加载

**解决**：
- 确认 `web/_next/` 目录存在
- 检查浏览器控制台的资源路径
- 验证 embed 是否正确包含了所有文件：
  ```bash
  ls -R web/
  ```

### 3. 构建后体积过大

**解决方案**：
```bash
# 使用生产构建
make build-prod

# 进一步压缩（需要安装 upx）
upx --best --lzma build/fail2ban-web
```

### 4. 开发时修改前端不生效

**原因**：embed 在编译时嵌入，需要重新构建

**解决**：
```bash
# 方式1：完整重新构建
make build-full

# 方式2：只重新构建前端，然后重新编译 Go
make build-frontend
go build -o build/fail2ban-web main.go
```

## 性能优化

### 1. 前端优化

- 使用 Next.js 的静态导出模式
- 启用代码分割
- 压缩图片和资源

### 2. 后端优化

- 使用 `embed` 减少文件 I/O
- 启用 Gzip 压缩中间件
- 合理设置缓存头

### 3. 部署优化

- 使用 CDN 分发静态资源（可选）
- 启用 HTTP/2
- 配置反向代理（Nginx）

## 测试

### 自动化测试

```bash
# 运行测试脚本
./test_frontend.sh
```

测试脚本会自动：
1. 启动应用
2. 测试各个页面的访问
3. 检查 API 响应
4. 清理测试进程

### 手动测试

```bash
# 启动应用
./build/fail2ban-web

# 在另一个终端测试
curl http://localhost:8099/
curl http://localhost:8099/login
curl http://localhost:8099/dashboard
curl http://localhost:8099/api/v1/health
```

## 相关文件

- `build_frontend.sh` - 前端构建脚本
- `test_frontend.sh` - 前端测试脚本
- `Makefile` - 构建命令
- `main.go` - embed 配置
- `core/app_server.go` - 静态文件服务
- `frontend/next.config.ts` - Next.js 配置
- `.gitignore` - 忽略 `web/` 构建产物

## 最后更新

- 日期：2025-10-19
- 版本：v1.0
- 作者：difyz9

---

**注意**：`web/` 目录是自动生成的，不应提交到 Git。已在 `.gitignore` 中配置忽略。
