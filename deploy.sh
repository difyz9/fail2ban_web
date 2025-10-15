#!/bin/bash

# Fail2Ban Web Panel 自动部署脚本
# 适用于Ubuntu服务器

set -e  # 遇到错误时退出

echo "=== Fail2Ban Web Panel 部署脚本 ==="
echo

# 检查是否为root用户
if [[ $EUID -eq 0 ]]; then
   echo "请不要以root用户运行此脚本"
   echo "使用普通用户运行: ./deploy.sh"
   exit 1
fi

# 获取当前用户
CURRENT_USER=$(whoami)
APP_DIR=$(pwd)
DATA_DIR="/var/lib/fail2ban-web"

echo "当前用户: $CURRENT_USER"
echo "应用目录: $APP_DIR"
echo

# 检查必要的依赖
echo "检查系统依赖..."

# 检查Go
if ! command -v go &> /dev/null; then
    echo "错误: 未找到Go编译器"
    echo "请先安装Go: https://golang.org/doc/install"
    exit 1
fi

# 检查fail2ban
if ! command -v fail2ban-client &> /dev/null; then
    echo "错误: 未找到fail2ban"
    echo "请先安装fail2ban: sudo apt install fail2ban"
    exit 1
fi

# 检查fail2ban服务状态
if ! sudo systemctl is-active --quiet fail2ban; then
    echo "警告: fail2ban服务未运行"
    echo "正在启动fail2ban..."
    sudo systemctl start fail2ban
fi

echo "✓ 系统依赖检查完成"
echo

# 创建数据目录
echo "创建数据目录..."
sudo mkdir -p "$DATA_DIR"
sudo chown "$CURRENT_USER:$CURRENT_USER" "$DATA_DIR"
echo "✓ 数据目录创建完成: $DATA_DIR"
echo

# 配置sudo权限
echo "配置fail2ban sudo权限..."
SUDOERS_FILE="/etc/sudoers.d/fail2ban-web"

if [[ ! -f "$SUDOERS_FILE" ]]; then
    echo "$CURRENT_USER ALL=(ALL) NOPASSWD: /usr/bin/fail2ban-client" | sudo tee "$SUDOERS_FILE" > /dev/null
    sudo chmod 440 "$SUDOERS_FILE"
    echo "✓ Sudo权限配置完成"
else
    echo "✓ Sudo权限已存在"
fi

# 测试sudo权限
echo "测试fail2ban权限..."
if sudo fail2ban-client ping &> /dev/null; then
    echo "✓ Fail2ban权限测试成功"
else
    echo "警告: Fail2ban权限测试失败，应用将使用sudo模式"
fi
echo

# 编译应用
echo "编译应用..."
if go build -o fail2ban-web main.go; then
    echo "✓ 应用编译成功"
    chmod +x fail2ban-web
else
    echo "错误: 应用编译失败"
    exit 1
fi
echo

# 创建环境配置文件
ENV_FILE="$APP_DIR/.env"
if [[ ! -f "$ENV_FILE" ]]; then
    echo "创建环境配置文件..."
    cat > "$ENV_FILE" << EOF
# 服务器配置
PORT=8092
HOST=0.0.0.0
GIN_MODE=release

# 数据库配置
DB_PATH=$DATA_DIR/fail2ban_web.db

# JWT配置
JWT_SECRET=fail2ban-web-$(openssl rand -hex 16)
JWT_EXPIRE_TIME=24

# Fail2Ban配置
FAIL2BAN_LOG_PATH=/var/log/fail2ban.log
FAIL2BAN_CONFIG_PATH=/etc/fail2ban
FAIL2BAN_SOCKET_PATH=/var/run/fail2ban/fail2ban.sock
NGINX_ACCESS_LOG=/var/log/nginx/access.log
NGINX_ERROR_LOG=/var/log/nginx/error.log
SSH_LOG_PATH=/var/log/auth.log
FAIL2BAN_FORCE_SUDO=true

# 管理员账户
ADMIN_USERNAME=admin
ADMIN_PASSWORD=$(openssl rand -base64 12)
ADMIN_EMAIL=admin@localhost
EOF
    echo "✓ 环境配置文件创建完成: $ENV_FILE"
    echo
    echo "=== 重要信息 ==="
    echo "管理员用户名: admin"
    echo "管理员密码: $(grep ADMIN_PASSWORD $ENV_FILE | cut -d'=' -f2)"
    echo "请记录此密码！"
    echo "==============="
    echo
else
    echo "✓ 环境配置文件已存在"
fi

# 创建systemd服务
echo "创建systemd服务..."
SERVICE_FILE="/etc/systemd/system/fail2ban-web.service"

sudo tee "$SERVICE_FILE" > /dev/null << EOF
[Unit]
Description=Fail2Ban Web Management Panel
After=network.target fail2ban.service
Requires=fail2ban.service

[Service]
Type=simple
User=$CURRENT_USER
Group=$CURRENT_USER
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/fail2ban-web
Restart=always
RestartSec=5
Environment=GIN_MODE=release
EnvironmentFile=-$APP_DIR/.env

# 安全配置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$DATA_DIR

[Install]
WantedBy=multi-user.target
EOF

echo "✓ Systemd服务文件创建完成"
echo

# 启动服务
echo "启动服务..."
sudo systemctl daemon-reload
sudo systemctl enable fail2ban-web

if sudo systemctl start fail2ban-web; then
    echo "✓ 服务启动成功"
else
    echo "错误: 服务启动失败"
    echo "查看日志: sudo journalctl -u fail2ban-web -f"
    exit 1
fi

# 检查服务状态
sleep 2
if sudo systemctl is-active --quiet fail2ban-web; then
    echo "✓ 服务运行正常"
else
    echo "警告: 服务可能未正常运行"
    echo "检查状态: sudo systemctl status fail2ban-web"
fi

echo
echo "=== 部署完成 ==="
echo "服务地址: http://$(hostname -I | awk '{print $1}'):8092"
echo "管理面板: http://$(hostname -I | awk '{print $1}'):8092/login"
echo
echo "管理命令:"
echo "  查看状态: sudo systemctl status fail2ban-web"
echo "  查看日志: sudo journalctl -u fail2ban-web -f"
echo "  重启服务: sudo systemctl restart fail2ban-web"
echo "  停止服务: sudo systemctl stop fail2ban-web"
echo
echo "如需配置防火墙:"
echo "  sudo ufw allow 8092"
echo
echo "如需修改配置，请编辑: $ENV_FILE"
echo "修改后重启服务: sudo systemctl restart fail2ban-web"
echo