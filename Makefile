# Makefile for Fail2Ban Web Panel

# 变量定义
APP_NAME = fail2ban-web
BUILD_DIR = build
MAIN_FILE = main.go
BINARY_NAME = $(APP_NAME)

# Go 相关变量
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod

# Docker 相关变量
DOCKER_IMAGE = $(APP_NAME):latest
DOCKER_CONTAINER = $(APP_NAME)-container

# 默认目标
.PHONY: all
all: clean build-frontend build

# 构建前端资源
.PHONY: build-frontend
build-frontend:
	@echo "构建前端资源..."
	@chmod +x build_frontend.sh
	@./build_frontend.sh

# 构建应用程序
.PHONY: build
build:
	@echo "构建应用程序..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 完整构建（前端+后端）
.PHONY: build-full
build-full: build-frontend build
	@echo "✅ 完整构建完成！"

# 运行应用程序
.PHONY: run
run:
	@echo "运行应用程序..."
	$(GOCMD) run $(MAIN_FILE)

# 清理构建文件
.PHONY: clean
clean:
	@echo "清理构建文件..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# 运行测试
.PHONY: test
test:
	@echo "运行测试..."
	$(GOTEST) -v ./...

# 下载依赖
.PHONY: deps
deps:
	@echo "下载依赖..."
	$(GOMOD) download
	$(GOMOD) tidy

# 格式化代码
.PHONY: fmt
fmt:
	@echo "格式化代码..."
	$(GOCMD) fmt ./...

# 代码检查
.PHONY: vet
vet:
	@echo "代码检查..."
	$(GOCMD) vet ./...

# 构建 Linux 版本
.PHONY: build-linux
build-linux:
	@echo "构建 Linux 版本..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_FILE)

# 构建 Windows 版本
.PHONY: build-windows
build-windows:
	@echo "构建 Windows 版本..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows.exe $(MAIN_FILE)

# 构建 macOS 版本
.PHONY: build-darwin
build-darwin:
	@echo "构建 macOS 版本..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin $(MAIN_FILE)

# 构建所有平台版本
.PHONY: build-all
build-all: build-frontend build-linux build-windows build-darwin
	@echo "✅ 所有平台构建完成！"

# Docker 构建
.PHONY: docker-build
docker-build:
	@echo "构建 Docker 镜像..."
	docker build -t $(DOCKER_IMAGE) .

# Docker 运行
.PHONY: docker-run
docker-run:
	@echo "运行 Docker 容器..."
	docker run -d --name $(DOCKER_CONTAINER) -p 8099:8099 $(DOCKER_IMAGE)

# Docker 停止
.PHONY: docker-stop
docker-stop:
	@echo "停止 Docker 容器..."
	docker stop $(DOCKER_CONTAINER) || true
	docker rm $(DOCKER_CONTAINER) || true

# Docker Compose 启动
.PHONY: compose-up
compose-up:
	@echo "启动 Docker Compose..."
	docker-compose up -d

# Docker Compose 停止
.PHONY: compose-down
compose-down:
	@echo "停止 Docker Compose..."
	docker-compose down

# 安装依赖工具
.PHONY: install-tools
install-tools:
	@echo "安装开发工具..."
	$(GOGET) -u github.com/cosmtrek/air@latest
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 热重载开发
.PHONY: dev
dev:
	@echo "启动热重载开发服务器..."
	air

g_push:
	@echo "推送到远程仓库..."
	git add .
	git commit -m "Auto commit"
	git push origin main

# 获取当前版本号
.PHONY: get-version
get-version:
	@git tag --sort=-v:refname | head -1 || echo "v0.0.0"

# 自动版本递增并推送 tag
.PHONY: tpush
tpush:
	@echo "获取当前最新版本..."
	@CURRENT_TAG=$$(git tag --sort=-v:refname | head -1 || echo "v0.0.0"); \
	echo "当前版本: $$CURRENT_TAG"; \
	VERSION=$${CURRENT_TAG#v}; \
	IFS='.' read -r MAJOR MINOR PATCH <<< "$$VERSION"; \
	NEW_PATCH=$$((PATCH + 1)); \
	NEW_TAG="v$$MAJOR.$$MINOR.$$NEW_PATCH"; \
	echo "新版本: $$NEW_TAG"; \
	read -p "请输入提交信息 (默认: Release $$NEW_TAG): " COMMIT_MSG; \
	COMMIT_MSG=$${COMMIT_MSG:-"Release $$NEW_TAG"}; \
	echo "创建标签: $$NEW_TAG"; \
	git tag -a $$NEW_TAG -m "$$COMMIT_MSG"; \
	echo "推送标签到远程仓库..."; \
	git push origin $$NEW_TAG; \
	echo "✅ 版本 $$NEW_TAG 已成功推送!"

# 自动版本递增（主版本号）
.PHONY: tag-major
tag-major:
	@echo "获取当前最新版本..."
	@CURRENT_TAG=$$(git tag --sort=-v:refname | head -1 || echo "v0.0.0"); \
	echo "当前版本: $$CURRENT_TAG"; \
	VERSION=$${CURRENT_TAG#v}; \
	IFS='.' read -r MAJOR MINOR PATCH <<< "$$VERSION"; \
	NEW_MAJOR=$$((MAJOR + 1)); \
	NEW_TAG="v$$NEW_MAJOR.0.0"; \
	echo "新版本: $$NEW_TAG"; \
	read -p "请输入提交信息 (默认: Release $$NEW_TAG): " COMMIT_MSG; \
	COMMIT_MSG=$${COMMIT_MSG:-"Release $$NEW_TAG"}; \
	echo "创建标签: $$NEW_TAG"; \
	git tag -a $$NEW_TAG -m "$$COMMIT_MSG"; \
	echo "推送标签到远程仓库..."; \
	git push origin $$NEW_TAG; \
	echo "✅ 版本 $$NEW_TAG 已成功推送!"

# 自动版本递增（次版本号）
.PHONY: tag-minor
tag-minor:
	@echo "获取当前最新版本..."
	@CURRENT_TAG=$$(git tag --sort=-v:refname | head -1 || echo "v0.0.0"); \
	echo "当前版本: $$CURRENT_TAG"; \
	VERSION=$${CURRENT_TAG#v}; \
	IFS='.' read -r MAJOR MINOR PATCH <<< "$$VERSION"; \
	NEW_MINOR=$$((MINOR + 1)); \
	NEW_TAG="v$$MAJOR.$$NEW_MINOR.0"; \
	echo "新版本: $$NEW_TAG"; \
	read -p "请输入提交信息 (默认: Release $$NEW_TAG): " COMMIT_MSG; \
	COMMIT_MSG=$${COMMIT_MSG:-"Release $$NEW_TAG"}; \
	echo "创建标签: $$NEW_TAG"; \
	git tag -a $$NEW_TAG -m "$$COMMIT_MSG"; \
	echo "推送标签到远程仓库..."; \
	git push origin $$NEW_TAG; \
	echo "✅ 版本 $$NEW_TAG 已成功推送!"

# 快速发布（自动递增补丁版本，不询问提交信息）
.PHONY: release
release:
	@echo "🚀 快速发布新版本..."
	@CURRENT_TAG=$$(git tag --sort=-v:refname | head -1 || echo "v0.0.0"); \
	echo "当前版本: $$CURRENT_TAG"; \
	VERSION=$${CURRENT_TAG#v}; \
	IFS='.' read -r MAJOR MINOR PATCH <<< "$$VERSION"; \
	NEW_PATCH=$$((PATCH + 1)); \
	NEW_TAG="v$$MAJOR.$$MINOR.$$NEW_PATCH"; \
	echo "新版本: $$NEW_TAG"; \
	git tag -a $$NEW_TAG -m "Release $$NEW_TAG"; \
	git push origin $$NEW_TAG; \
	echo "✅ 版本 $$NEW_TAG 已成功推送!"

# 查看所有标签
.PHONY: list-tags
list-tags:
	@echo "所有版本标签："
	@git tag --sort=-v:refname

# 删除本地标签
.PHONY: delete-tag
delete-tag:
	@read -p "请输入要删除的标签名称: " TAG_NAME; \
	git tag -d $$TAG_NAME; \
	echo "本地标签 $$TAG_NAME 已删除"

# 删除远程标签
.PHONY: delete-tag-remote
delete-tag-remote:
	@read -p "请输入要删除的远程标签名称: " TAG_NAME; \
	git push origin --delete $$TAG_NAME; \
	echo "远程标签 $$TAG_NAME 已删除"

# 生产构建
.PHONY: build-prod
build-prod: build-frontend
	@echo "生产环境构建..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 $(GOBUILD) -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ 生产环境构建完成！"

# 帮助信息
.PHONY: help
help:
	@echo "可用的命令："
	@echo ""
	@echo "📦 构建相关:"
	@echo "  build-frontend - 构建前端资源到 web 目录"
	@echo "  build         - 构建应用程序（仅后端）"
	@echo "  build-full    - 完整构建（前端 + 后端）"
	@echo "  build-linux   - 构建 Linux 版本"
	@echo "  build-windows - 构建 Windows 版本"
	@echo "  build-darwin  - 构建 macOS 版本"
	@echo "  build-all     - 构建所有平台版本（包含前端）"
	@echo "  build-prod    - 生产环境构建（包含前端）"
	@echo ""
	@echo "🚀 运行相关:"
	@echo "  run           - 运行应用程序"
	@echo "  dev           - 热重载开发"
	@echo ""
	@echo "🧹 清理和测试:"
	@echo "  clean         - 清理构建文件"
	@echo "  test          - 运行测试"
	@echo "  fmt           - 格式化代码"
	@echo "  vet           - 代码检查"
	@echo ""
	@echo "📚 依赖管理:"
	@echo "  deps          - 下载依赖"
	@echo "  install-tools - 安装开发工具"
	@echo ""
	@echo "🐳 Docker 相关:"
	@echo "  docker-build  - 构建 Docker 镜像"
	@echo "  docker-run    - 运行 Docker 容器"
	@echo "  docker-stop   - 停止 Docker 容器"
	@echo "  compose-up    - 启动 Docker Compose"
	@echo "  compose-down  - 停止 Docker Compose"
	@echo ""
	@echo "🏷️  版本标签管理:"
	@echo "  get-version       - 获取当前版本号"
	@echo "  tpush          - 自动递增补丁版本号并推送 (v0.0.X)"
	@echo "  tag-minor         - 自动递增次版本号并推送 (v0.X.0)"
	@echo "  tag-major         - 自动递增主版本号并推送 (vX.0.0)"
	@echo "  release           - 快速发布（自动递增补丁版本）"
	@echo "  list-tags         - 查看所有版本标签"
	@echo "  delete-tag        - 删除本地标签"
	@echo "  delete-tag-remote - 删除远程标签"
	@echo ""
	@echo "📝 Git 相关:"
	@echo "  g_push        - 快速提交并推送"
	@echo ""
	@echo "❓ 其他:"
	@echo "  help          - 显示帮助信息"