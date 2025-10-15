# Fail2Ban Web Panel 部署指南

## 服务器环境配置

### 1. 系统要求
- Ubuntu 18.04+ 或其他支持systemd的Linux发行版
- Go 1.19+
- Fail2Ban已安装并运行
- Nginx已安装(可选，用于反向代理)

### 2. 权限配置

#### 方式一：使用sudo (推荐)
创建sudoers配置文件，允许应用用户执行fail2ban-client命令：

```bash
# 创建sudoers文件
sudo visudo -f /etc/sudoers.d/fail2ban-web

# 添加以下内容 (将ubuntu替换为实际用户名)
ubuntu ALL=(ALL) NOPASSWD: /usr/bin/fail2ban-client
```

#### 方式二：将用户添加到fail2ban组
```bash
# 创建fail2ban组
sudo groupadd fail2ban

# 将用户添加到组
sudo usermod -a -G fail2ban ubuntu

# 修改socket权限
sudo chgrp fail2ban /var/run/fail2ban/fail2ban.sock
sudo chmod g+rw /var/run/fail2ban/fail2ban.sock
```

### 3. 环境变量配置

创建环境配置文件 `.env`:

```bash
# 服务器配置
PORT=8092
HOST=0.0.0.0
GIN_MODE=release

# 数据库配置
DB_PATH=/var/lib/fail2ban-web/fail2ban_web.db

# JWT配置
JWT_SECRET=your-very-secure-jwt-secret-key-here
JWT_EXPIRE_TIME=24

# Fail2Ban配置
FAIL2BAN_LOG_PATH=/var/log/fail2ban.log
FAIL2BAN_CONFIG_PATH=/etc/fail2ban
FAIL2BAN_SOCKET_PATH=/var/run/fail2ban/fail2ban.sock
NGINX_ACCESS_LOG=/var/log/nginx/access.log
NGINX_ERROR_LOG=/var/log/nginx/error.log
SSH_LOG_PATH=/var/log/auth.log

# 强制使用sudo (如果权限配置有问题)
FAIL2BAN_FORCE_SUDO=true

# 管理员账户
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your-secure-password
ADMIN_EMAIL=admin@yourserver.com
```

### 4. 系统服务配置

创建systemd服务文件 `/etc/systemd/system/fail2ban-web.service`:

```ini
[Unit]
Description=Fail2Ban Web Management Panel
After=network.target fail2ban.service
Requires=fail2ban.service

[Service]
Type=simple
User=ubuntu
Group=ubuntu
WorkingDirectory=/home/ubuntu/app/fail2ban_web
ExecStart=/home/ubuntu/app/fail2ban_web/fail2ban-web
Restart=always
RestartSec=5
Environment=GIN_MODE=release
EnvironmentFile=-/home/ubuntu/app/fail2ban_web/.env

# 安全配置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/fail2ban-web

[Install]
WantedBy=multi-user.target
```

### 5. 部署步骤

```bash
# 1. 创建应用目录
sudo mkdir -p /var/lib/fail2ban-web
sudo chown ubuntu:ubuntu /var/lib/fail2ban-web

# 2. 编译应用
cd /home/ubuntu/app/fail2ban_web
go build -o fail2ban-web main.go

# 3. 设置权限
chmod +x fail2ban-web

# 4. 配置sudoers (选择方式一)
sudo visudo -f /etc/sudoers.d/fail2ban-web

# 5. 启动服务
sudo systemctl daemon-reload
sudo systemctl enable fail2ban-web
sudo systemctl start fail2ban-web

# 6. 检查状态
sudo systemctl status fail2ban-web
```

### 6. Nginx反向代理配置 (可选)

创建Nginx配置 `/etc/nginx/sites-available/fail2ban-web`:

```nginx
server {
    listen 80;
    server_name your-server-domain.com;

    location / {
        proxy_pass http://127.0.0.1:8092;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

启用配置:
```bash
sudo ln -s /etc/nginx/sites-available/fail2ban-web /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 7. 防火墙配置

```bash
# 允许应用端口
sudo ufw allow 8092

# 如果使用Nginx反向代理
sudo ufw allow 'Nginx Full'
```

### 8. 日志和监控

```bash
# 查看应用日志
sudo journalctl -u fail2ban-web -f

# 查看fail2ban状态
sudo fail2ban-client status

# 测试权限
sudo fail2ban-client ping
```

## 故障排除

### 权限错误
如果出现权限错误：
1. 检查sudoers配置
2. 确认用户在正确的组中
3. 检查socket文件权限
4. 设置环境变量 `FAIL2BAN_FORCE_SUDO=true`

### 服务无法启动
1. 检查fail2ban服务是否运行
2. 检查应用可执行文件权限
3. 检查数据库目录权限
4. 查看systemd日志

### 网络访问问题
1. 检查防火墙设置
2. 确认端口未被占用
3. 检查Nginx配置(如果使用)

## 安全建议

1. 修改默认管理员密码
2. 使用强JWT密钥
3. 配置HTTPS (推荐)
4. 定期更新系统和应用
5. 监控应用日志
6. 限制网络访问(如使用VPN)