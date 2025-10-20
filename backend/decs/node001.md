# Fail2Ban 学习笔记

## 目录
1. [概述](#概述)
2. [核心概念](#核心概念)
3. [安装与配置](#安装与配置)
4. [配置文件详解](#配置文件详解)
5. [过滤器配置](#过滤器配置)
6. [动作配置](#动作配置)
7. [实战案例](#实战案例)
8. [监控与调试](#监控与调试)
9. [最佳实践](#最佳实践)

## 概述

### 什么是 Fail2Ban
Fail2Ban 是一个入侵防御软件框架，通过监控系统日志文件（如 `/var/log/auth.log`，`/var/log/apache/access.log` 等）来检测恶意行为，如暴力破解、密码猜测等，并自动更新防火墙规则来阻止恶意IP地址。

### 工作原理
1. **监控日志**：定期扫描指定的日志文件
2. **模式匹配**：使用正则表达式匹配攻击模式
3. **触发禁令**：当失败次数超过阈值时触发禁令
4. **执行动作**：更新防火墙规则阻止IP
5. **解除禁令**：经过设定的时间后自动解除禁令

## 核心概念

### 主要组件
- **Jail**：定义监控服务、过滤规则、动作和参数的配置单元
- **Filter**：包含用于识别攻击的正则表达式模式
- **Action**：定义检测到攻击时执行的操作
- **Ban Time**：IP被封锁的持续时间
- **Max Retry**：触发禁令前的最大失败尝试次数
- **Find Time**：统计失败尝试的时间窗口

## 安装与配置

### 安装方法

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install fail2ban
```

**CentOS/RHEL:**
```bash
sudo yum install epel-release
sudo yum install fail2ban
# 或使用 dnf
sudo dnf install fail2ban
```

**从源码安装:**
```bash
git clone https://github.com/fail2ban/fail2ban.git
cd fail2ban
sudo python setup.py install
```

### 服务管理
```bash
# 启动服务
sudo systemctl start fail2ban

# 停止服务
sudo systemctl stop fail2ban

# 重启服务
sudo systemctl restart fail2ban

# 查看状态
sudo systemctl status fail2ban

# 开机自启
sudo systemctl enable fail2ban
```

## 配置文件详解

### 配置文件结构
```
/etc/fail2ban/
├── fail2ban.conf          # 主配置文件
├── jail.conf              # 默认监狱配置
├── jail.d/                # 监狱配置目录
├── filter.d/              # 过滤器目录
├── action.d/              # 动作目录
└── action.d/              # 其他动作文件
```

### 主配置文件 (fail2ban.conf)
```ini
[Definition]
# 日志级别
loglevel = INFO

# 日志文件位置
logtarget = /var/log/fail2ban.log

# socket 文件位置
socket = /var/run/fail2ban/fail2ban.sock

# pid 文件位置
pidfile = /var/run/fail2ban/fail2ban.pid
```

### 监狱配置 (jail.local)
**不要直接修改 jail.conf**，创建 `jail.local` 进行自定义配置：

```ini
[DEFAULT]
# 通用设置
bantime = 3600
findtime = 600
maxretry = 5
backend = auto

# 邮件通知
destemail = admin@yourdomain.com
sender = fail2ban@yourdomain.com
mta = sendmail

# 动作设置
action = %(action_)s

[sshd]
enabled = true
port = ssh
logpath = /var/log/auth.log
maxretry = 3
bantime = 86400

[sshd-ddos]
enabled = true
port = ssh
logpath = /var/log/auth.log
maxretry = 2
findtime = 60
bantime = 3600

[apache-auth]
enabled = true
port = http,https
logpath = /var/log/apache2/*error.log
maxretry = 3
bantime = 600

[nginx-http-auth]
enabled = true
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 3
bantime = 600
```

## 过滤器配置

### 自定义过滤器示例
创建 `/etc/fail2ban/filter.d/sshd-custom.conf`：

```ini
[Definition]
# 失败登录尝试
failregex = ^%(__prefix_line)s(?:error: PAM: )?Authentication failure for .* from <HOST>( via \S+)?\s*$
            ^%(__prefix_line)s(?:error: PAM: )?User not known to the underlying authentication module for .* from <HOST>\s*$
            ^%(__prefix_line)sFailed password for (?:invalid user |)(?P<user>\S+) from <HOST> port \d+ ssh2\s*$
            ^%(__prefix_line)sReceived disconnect from <HOST> port \d+:11: (?:Bye|Authentication failed)\s*$
            ^%(__prefix_line)sDisconnected from authenticating user (?P<user>\S+) <HOST> port \d+ \[preauth\]\s*$

# 忽略规则
ignoreregex =

[Init]
# 日志行最大字符数
maxlines = 1
```

### 测试过滤器
```bash
# 测试过滤器规则
fail2ban-regex /var/log/auth.log /etc/fail2ban/filter.d/sshd.conf

# 测试自定义日志
fail2ban-regex /path/to/logfile /etc/fail2ban/filter.d/your-filter.conf
```

## 动作配置

### 默认动作
- `iptables-multiport`：使用 iptables 封锁多个端口
- `iptables-allports`：使用 iptables 封锁所有端口
- `shorewall`：使用 Shorewall 防火墙
- `tcpwrapper`：使用 /etc/hosts.deny

### 自定义动作示例
创建 `/etc/fail2ban/action.d/iptables-custom.conf`：

```ini
[Definition]
# 动作前缀
actionstart = <iptables> -N f2b-<name>
              <iptables> -A f2b-<name> -j <returntype>
              <iptables> -I <chain> -p <protocol> -j f2b-<name>

# 封锁IP
actionban = <iptables> -I f2b-<name> 1 -s <ip> -j <blocktype>

# 解除封锁
actionunban = <iptables> -D f2b-<name> -s <ip> -j <blocktype>

# 动作停止
actionstop = <iptables> -D <chain> -p <protocol> -j f2b-<name>
             <iptables> -X f2b-<name>

[Init]
# 初始化
name = default
iptables = iptables
chain = INPUT
protocol = tcp
blocktype = REJECT --reject-with icmp-port-unreachable
returntype = RETURN
```

## 实战案例

### 案例1：保护 SSH 服务
```ini
[sshd]
enabled = true
port = ssh
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
findtime = 600
ignoreip = 127.0.0.1/8 192.168.1.0/24
```

### 案例2：保护 WordPress
创建过滤器 `/etc/fail2ban/filter.d/wordpress.conf`：
```ini
[Definition]
failregex = ^<HOST> -.*POST.*wp-login.php.* 200$
            ^<HOST> -.*POST.*xmlrpc.php.* 200$
ignoreregex =
```

监狱配置：
```ini
[wordpress]
enabled = true
port = http,https
logpath = /var/log/apache2/access.log
maxretry = 3
findtime = 300
bantime = 3600
```

### 案例3：防止暴力破解 FTP
```ini
[vsftpd]
enabled = true
port = ftp,ftp-data,ftps,ftps-data
logpath = /var/log/vsftpd.log
maxretry = 3
bantime = 1800
findtime = 300
```

### 案例4：保护 MySQL 数据库
```ini
[mysqld-auth]
enabled = true
port = 3306
logpath = /var/log/mysql/error.log
maxretry = 3
bantime = 3600
findtime = 600
```

### 案例5：自定义端口扫描检测
创建过滤器 `/etc/fail2ban/filter.d/portscan.conf`：
```ini
[Definition]
failregex = ^%(__prefix_line)sUDP scan from <HOST>
            ^%(__prefix_line)sTCP scan from <HOST>
ignoreregex =
```

监狱配置：
```ini
[portscan]
enabled = true
port = all
logpath = /var/log/syslog
maxretry = 2
findtime = 30
bantime = 604800
action = iptables-allports[name=portscan, protocol=all]
```

## 监控与调试

### 常用命令
```bash
# 查看 Fail2Ban 状态
sudo fail2ban-client status

# 查看特定监狱状态
sudo fail2ban-client status sshd

# 手动解封 IP
sudo fail2ban-client set sshd unbanip 192.168.1.100

# 手动封禁 IP
sudo fail2ban-client set sshd banip 192.168.1.100

# 重新加载配置
sudo fail2ban-client reload

# 查看详细日志
sudo tail -f /var/log/fail2ban.log
```

### 日志分析
```bash
# 查看被封锁的 IP
sudo fail2ban-client status sshd

# 实时监控 Fail2Ban 日志
sudo tail -f /var/log/fail2ban.log | grep -i 'ban\|unban'

# 检查 iptables 规则
sudo iptables -L -n
sudo iptables -L f2b-sshd -n
```

### 调试技巧
```bash
# 增加日志详细程度
# 在 fail2ban.conf 中设置：
loglevel = DEBUG

# 测试正则表达式
fail2ban-regex /var/log/auth.log /etc/fail2ban/filter.d/sshd.conf

# 检查配置文件语法
fail2ban-client -t

# 监控特定监狱
watch 'fail2ban-client status sshd'
```

## 最佳实践

### 安全建议
1. **白名单重要IP**：将可信IP添加到 `ignoreip`
2. **合理设置封锁时间**：根据服务重要性调整 `bantime`
3. **监控 Fail2Ban 状态**：定期检查被封IP和系统状态
4. **备份配置**：定期备份自定义配置
5. **测试规则**：部署前充分测试过滤规则

### 性能优化
```ini
[DEFAULT]
# 使用更高效的后端
backend = auto

# 调整查找时间间隔
findtime = 600

# 合理设置最大重试次数
maxretry = 3

# 使用数据库后端（可选）
dbfile = /var/lib/fail2ban/fail2ban.sqlite3
dbpurgeage = 1d
```

### 高级配置示例
```ini
[DEFAULT]
# 全局设置
bantime = 3600
findtime = 600
maxretry = 3
backend = systemd
ignoreip = 127.0.0.1/8 10.0.0.0/8 172.16.0.0/12 192.168.0.0/16
destemail = admin@yourdomain.com
sender = fail2ban@yourserver.com
action = %(action_mwl)s

# 使用更严格的 SSH 配置
[sshd]
enabled = true
port = ssh
logpath = %(sshd_log)s
maxretry = 2
bantime = 86400
findtime = 300

# 保护 Web 应用
[nginx-botsearch]
enabled = true
port = http,https
logpath = /var/log/nginx/access.log
maxretry = 10
findtime = 600
bantime = 3600
```

### 故障排除
1. **服务无法启动**：检查配置文件语法
2. **IP未被封锁**：验证过滤器和日志路径
3. **误封问题**：调整过滤规则和重试次数
4. **性能问题**：优化查找间隔和日志文件大小

---

**注意**：在生产环境中部署前，请务必在测试环境中验证所有配置，避免因配置错误导致服务中断。