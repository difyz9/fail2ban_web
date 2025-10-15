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
all: clean build

# 构建应用程序
.PHONY: build
build:
	@echo "构建应用程序..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

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
build-all: build-linux build-windows build-darwin

# Docker 构建
.PHONY: docker-build
docker-build:
	@echo "构建 Docker 镜像..."
	docker build -t $(DOCKER_IMAGE) .

# Docker 运行
.PHONY: docker-run
docker-run:
	@echo "运行 Docker 容器..."
	docker run -d --name $(DOCKER_CONTAINER) -p 8092:8092 $(DOCKER_IMAGE)

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

# 生产构建
.PHONY: build-prod
build-prod:
	@echo "生产环境构建..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 $(GOBUILD) -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

# 帮助信息
.PHONY: help
help:
	@echo "可用的命令："
	@echo "  build         - 构建应用程序"
	@echo "  run           - 运行应用程序"
	@echo "  clean         - 清理构建文件"
	@echo "  test          - 运行测试"
	@echo "  deps          - 下载依赖"
	@echo "  fmt           - 格式化代码"
	@echo "  vet           - 代码检查"
	@echo "  build-linux   - 构建 Linux 版本"
	@echo "  build-windows - 构建 Windows 版本"
	@echo "  build-darwin  - 构建 macOS 版本"
	@echo "  build-all     - 构建所有平台版本"
	@echo "  docker-build  - 构建 Docker 镜像"
	@echo "  docker-run    - 运行 Docker 容器"
	@echo "  docker-stop   - 停止 Docker 容器"
	@echo "  compose-up    - 启动 Docker Compose"
	@echo "  compose-down  - 停止 Docker Compose"
	@echo "  dev           - 热重载开发"
	@echo "  build-prod    - 生产环境构建"
	@echo "  help          - 显示帮助信息"