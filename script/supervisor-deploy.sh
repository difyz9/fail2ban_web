#!/bin/bash

# Supervisor 部署脚本
set -e

# 获取项目名称
if [ -f "go.mod" ]; then
    PROJECT_NAME=$(grep '^module ' go.mod | awk '{print $2}' | sed 's/.*\///')
else
    PROJECT_NAME="golang-gin-demo"
fi

echo "开始部署 $PROJECT_NAME with Supervisor..."

# 检查是否以 root 权限运行
if [ "$EUID" -ne 0 ]; then
    echo "请使用 sudo 运行此脚本"
    exit 1
fi

# 安装 Supervisor（如果未安装）
if ! command -v supervisorctl &> /dev/null; then
    echo "安装 Supervisor..."
    apt-get update
    apt-get install -y supervisor
    systemctl enable supervisor
    systemctl start supervisor
fi

# 停止现有服务
echo "停止现有服务..."
supervisorctl stop $PROJECT_NAME || true

# 读取 .env 文件并生成环境变量字符串
echo "读取环境变量配置..."
ENV_VARS="GIN_MODE=release"

if [ -f ".env" ]; then
    echo "发现 .env 文件，读取环境变量..."
    while IFS='=' read -r key value || [[ -n "$key" ]]; do
        # 跳过注释行和空行
        if [[ ! "$key" =~ ^#.*$ ]] && [[ -n "$key" ]] && [[ -n "$value" ]]; then
            # 移除可能的空格和引号
            key=$(echo "$key" | xargs)
            value=$(echo "$value" | xargs | sed 's/^"//;s/"$//')
            if [[ -n "$key" && -n "$value" ]]; then
                ENV_VARS="$ENV_VARS,$key=$value"
                echo "添加环境变量: $key"
            fi
        fi
    done < .env
else
    echo "警告: .env 文件不存在，仅使用默认环境变量"
fi

echo "环境变量配置: $ENV_VARS"

# 生成 Supervisor 配置文件
echo "生成 Supervisor 配置文件..."
cat >/tmp/$PROJECT_NAME.conf <<EOL
[program:$PROJECT_NAME]
directory = /app/$PROJECT_NAME
command = /app/$PROJECT_NAME/$PROJECT_NAME
autostart = true ; 在 supervisord 启动的时候也自动启动
startsecs = 5 ; 启动 5 秒后没有异常退出，就当作已经正常启动了
autorestart = true ; 程序异常退出后自动重启
startretries = 3 ; 启动失败自动重试次数，默认是 3
user = root ; 用哪个用户启动
redirect_stderr = true ; 把 stderr 重定向到 stdout，默认 false
stdout_logfile_maxbytes = 20MB ; stdout 日志文件大小，默认 50MB
stdout_logfile_backups = 20 ; stdout 日志文件备份数
stdout_logfile = /var/log/$PROJECT_NAME.log ; 日志文件
environment=$ENV_VARS ; 设置环境变量
EOL

# 移动配置文件到正确位置
mv /tmp/$PROJECT_NAME.conf /etc/supervisor/conf.d/$PROJECT_NAME.conf

# 更新 Supervisor 配置
echo "更新 Supervisor 配置..."
supervisorctl reread
supervisorctl update

# 启动服务
echo "启动服务..."
supervisorctl start $PROJECT_NAME

# 等待服务启动
echo "等待服务启动..."
sleep 10

# 查看服务状态
echo "查看服务状态..."
supervisorctl status $PROJECT_NAME

# 测试服务是否正常
echo "测试服务健康状态..."
for i in {1..30}; do
    if curl -f http://localhost:8080/health &>/dev/null; then
        echo "服务启动成功！"
        echo "访问地址: http://$(hostname -I | awk '{print $1}'):8080"
        echo "健康检查: http://$(hostname -I | awk '{print $1}'):8080/health"
        exit 0
    fi
    echo "等待服务启动... ($i/30)"
    sleep 2
done

echo "服务启动失败！"
echo "查看日志："
tail -50 /var/log/$PROJECT_NAME.log
echo "查看 Supervisor 状态："
supervisorctl status $PROJECT_NAME
exit 1
