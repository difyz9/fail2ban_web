# 使用官方 Go 镜像作为构建环境
FROM golang:1.21-alpine AS builder

# 安装必要的包
RUN apk add --no-cache git gcc musl-dev

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# 使用 Alpine Linux 作为运行环境
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates fail2ban iptables rsyslog

# 创建非root用户
RUN addgroup -g 1000 appgroup && \
    adduser -D -s /bin/sh -u 1000 -G appgroup appuser

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 创建必要的目录
RUN mkdir -p /var/log/fail2ban /var/lib/fail2ban /etc/fail2ban

# 设置权限
RUN chown -R appuser:appgroup /root/

# 暴露端口
EXPOSE 8080

# 添加健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# 切换到非root用户
USER appuser

# 运行应用程序
CMD ["./main"]