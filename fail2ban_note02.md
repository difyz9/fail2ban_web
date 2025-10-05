# Fail2ban 学习笔记

## 目录
- [1. 简介](#1-简介)
- [2. 工作原理](#2-工作原理)
- [3. 安装配置](#3-安装配置)
- [4. 核心概念](#4-核心概念)
- [5. 配置详解](#5-配置详解)
- [6. 常用命令](#6-常用命令)
- [7. 实战案例](#7-实战案例)
- [8. 最佳实践](#8-最佳实践)

---

## 1. 简介

### 什么是 Fail2ban
Fail2ban 是一个入侵防御软件框架，用于保护计算机服务器免受暴力攻击。它通过监控服务日志文件，检测恶意行为模式，并采取相应的防御措施（通常是更新防火墙规则来封禁攻击者的 IP 地址）。

### 主要特性
- **日志监控**: 实时监控各种服务的日志文件
- **模式匹配**: 使用正则表达式识别失败的登录尝试
- **自动封禁**: 自动封禁触发规则的 IP 地址
- **多服务支持**: 支持 SSH、Apache、Nginx、Postfix 等多种服务
- **灵活配置**: 高度可配置的规则和动作
- **自动解封**: 可设置封禁时间，到期自动解封

### 应用场景
- SSH 暴力破解防护
- Web 服务器攻击防护
- 邮件服务器防护
- FTP 服务防护
- 防止 DDoS 攻击

---

## 2. 工作原理

### 基本流程
```
日志文件 → 监控进程 → 模式匹配 → 触发条件 → 执行动作 → 更新防火墙
```

### 详细步骤
1. **监控日志**: Fail2ban 持续监控指定的日志文件
2. **匹配规则**: 使用正则表达式匹配失败登录等恶意行为
3. **计数统计**: 统计指定时间窗口内的失败次数
4. **触发阈值**: 达到设定的失败次数阈值
5. **执行动作**: 调用 action（通常是 iptables 规则）
6. **封禁 IP**: 将恶意 IP 添加到防火墙黑名单
7. **定时解封**: 封禁时间到期后自动解封

### 架构组件
- **Client**: 命令行工具，用于与 fail2ban 服务交互
- **Server**: 主守护进程，负责监控和管理
- **Filter**: 过滤器，定义日志匹配规则
- **Action**: 动作，定义封禁/解封操作
- **Jail**: 监狱，组合 filter 和 action 的配置单元

---

## 3. 安装配置

### 在 Debian/Ubuntu 上安装
```bash
# 更新包列表
sudo apt update

# 安装 fail2ban
sudo apt install fail2ban

# 启动服务
sudo systemctl start fail2ban

# 设置开机自启
sudo systemctl enable fail2ban

# 检查状态
sudo systemctl status fail2ban
```

### 在 CentOS/RHEL 上安装
```bash
# 安装 EPEL 仓库
sudo yum install epel-release

# 安装 fail2ban
sudo yum install fail2ban fail2ban-systemd

# 启动服务
sudo systemctl start fail2ban

# 设置开机自启
sudo systemctl enable fail2ban

# 检查状态
sudo systemctl status fail2ban
```

### 基本配置文件结构
```
/etc/fail2ban/
├── fail2ban.conf          # 主配置文件（不建议直接修改）
├── fail2ban.local         # 本地主配置（覆盖 .conf）
├── jail.conf              # jail 配置文件（不建议直接修改）
├── jail.local             # 本地 jail 配置（推荐使用）
├── jail.d/                # jail 配置目录
├── filter.d/              # 过滤器配置目录
├── action.d/              # 动作配置目录
└── paths-*.conf           # 路径配置文件
```

### 初始配置
```bash
# 复制默认配置文件
sudo cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local

# 编辑本地配置
sudo nano /etc/fail2ban/jail.local
```

---

## 4. 核心概念

### Filter (过滤器)
过滤器定义了如何从日志文件中识别恶意行为。

**示例**: `/etc/fail2ban/filter.d/sshd.conf`
```ini
[Definition]
# 定义失败模式
failregex = ^%(__prefix_line)s(?:error: PAM: )?[aA]uthentication (?:failure|error) for .* from <HOST>( via \S+)?\s*$
            ^%(__prefix_line)sFailed (?:password|publickey) for .* from <HOST>(?: port \d*)?(?: ssh\d*)?$
            ^%(__prefix_line)sROOT LOGIN REFUSED.* FROM <HOST>\s*$

# 定义忽略模式（可选）
ignoreregex =
```

### Action (动作)
动作定义了当检测到恶意行为时应该执行什么操作。

**示例**: `/etc/fail2ban/action.d/iptables-multiport.conf`
```ini
[Definition]
# 封禁动作
actionban = <iptables> -I f2b-<name> 1 -s <ip> -j <blocktype>

# 解封动作
actionunban = <iptables> -D f2b-<name> -s <ip> -j <blocktype>

# 启动时动作
actionstart = <iptables> -N f2b-<name>
              <iptables> -A f2b-<name> -j <returntype>
              <iptables> -I <chain> -p <protocol> -m multiport --dports <port> -j f2b-<name>

# 停止时动作
actionstop = <iptables> -D <chain> -p <protocol> -m multiport --dports <port> -j f2b-<name>
             <iptables> -F f2b-<name>
             <iptables> -X f2b-<name>
```

### Jail (监狱)
Jail 是 fail2ban 的核心配置单元，组合了 filter 和 action。

**基本参数**:
- `enabled`: 是否启用该 jail
- `port`: 保护的端口
- `filter`: 使用的过滤器名称
- `logpath`: 监控的日志文件路径
- `maxretry`: 最大重试次数
- `findtime`: 查找时间窗口（秒）
- `bantime`: 封禁时间（秒）
- `action`: 执行的动作

---

## 5. 配置详解

### 全局配置 (jail.local)

```ini
[DEFAULT]
# 忽略的 IP 地址（白名单）
ignoreip = 127.0.0.1/8 ::1 192.168.1.0/24

# 封禁时间（秒）-1 表示永久封禁
bantime = 3600

# 查找时间窗口（秒）
findtime = 600

# 最大重试次数
maxretry = 5

# 后端（监控方式）
backend = auto

# 邮件相关配置
destemail = admin@example.com
sender = fail2ban@example.com
mta = sendmail

# 默认 action
action = %(action_)s
# action_: 仅封禁
# action_mw: 封禁并发送邮件
# action_mwl: 封禁并发送包含日志的邮件
```

### SSH 保护配置

```ini
[sshd]
enabled = true
port = ssh,22
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
findtime = 600
bantime = 3600
```

### Apache/Nginx 保护配置

```ini
[apache-auth]
enabled = true
port = http,https
filter = apache-auth
logpath = /var/log/apache*/*error.log
maxretry = 5
bantime = 3600

[nginx-http-auth]
enabled = true
filter = nginx-http-auth
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 5
bantime = 3600

[nginx-noscript]
enabled = true
port = http,https
filter = nginx-noscript
logpath = /var/log/nginx/access.log
maxretry = 6
bantime = 3600

[nginx-badbots]
enabled = true
port = http,https
filter = nginx-badbots
logpath = /var/log/nginx/access.log
maxretry = 2
bantime = 86400
```

### 自定义 Filter 示例

创建文件 `/etc/fail2ban/filter.d/custom-app.conf`:
```ini
[Definition]
# 匹配登录失败
failregex = ^.*Login failed for user .* from <HOST>.*$
            ^.*Invalid authentication attempt from <HOST>.*$

# 忽略特定模式
ignoreregex = ^.*successful login.*$

# 日期格式
datepattern = ^%%Y-%%m-%%d %%H:%%M:%%S
```

### 自定义 Action 示例

创建文件 `/etc/fail2ban/action.d/custom-notify.conf`:
```ini
[Definition]
# 封禁时发送通知
actionban = curl -X POST https://api.example.com/alert \
            -d "ip=<ip>&action=ban&jail=<name>"

# 解封时发送通知
actionunban = curl -X POST https://api.example.com/alert \
              -d "ip=<ip>&action=unban&jail=<name>"
```

---

## 6. 常用命令

### 服务管理
```bash
# 启动服务
sudo systemctl start fail2ban

# 停止服务
sudo systemctl stop fail2ban

# 重启服务
sudo systemctl restart fail2ban

# 重新加载配置
sudo systemctl reload fail2ban

# 查看状态
sudo systemctl status fail2ban

# 查看日志
sudo tail -f /var/log/fail2ban.log
```

### fail2ban-client 命令

#### 查看状态
```bash
# 查看所有 jail 状态
sudo fail2ban-client status

# 查看特定 jail 状态
sudo fail2ban-client status sshd

# 查看被封禁的 IP
sudo fail2ban-client get sshd banip

# 查看当前被封禁的 IP 列表
sudo fail2ban-client status sshd | grep "Banned IP"
```

#### 手动封禁/解封
```bash
# 手动封禁 IP
sudo fail2ban-client set sshd banip 192.168.1.100

# 手动解封 IP
sudo fail2ban-client set sshd unbanip 192.168.1.100

# 解封所有 IP
sudo fail2ban-client unban --all
```

#### Jail 管理
```bash
# 启动特定 jail
sudo fail2ban-client start sshd

# 停止特定 jail
sudo fail2ban-client stop sshd

# 重新加载 jail
sudo fail2ban-client reload sshd

# 查看 jail 配置
sudo fail2ban-client get sshd logpath
sudo fail2ban-client get sshd maxretry
sudo fail2ban-client get sshd bantime
```

### 测试配置
```bash
# 测试配置文件语法
sudo fail2ban-client -t

# 测试正则表达式
sudo fail2ban-regex /var/log/auth.log /etc/fail2ban/filter.d/sshd.conf

# 详细测试输出
sudo fail2ban-regex /var/log/auth.log /etc/fail2ban/filter.d/sshd.conf --print-all-matched
```

### iptables 命令（查看封禁规则）
```bash
# 查看所有 fail2ban 规则
sudo iptables -L -n

# 查看特定链
sudo iptables -L f2b-sshd -n -v

# 查看被封禁的 IP 数量
sudo iptables -L f2b-sshd -n | grep -c DROP
```

---

## 7. 实战案例

### 案例 1: 加强 SSH 保护

**需求**: 防止 SSH 暴力破解，3 次失败尝试即封禁 1 小时

**配置**: `/etc/fail2ban/jail.local`
```ini
[sshd]
enabled = true
port = 22
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
findtime = 600
bantime = 3600
action = iptables[name=SSH, port=ssh, protocol=tcp]
         sendmail-whois[name=SSH, dest=admin@example.com]
```

**验证**:
```bash
# 查看状态
sudo fail2ban-client status sshd

# 测试（从另一台机器）
ssh wronguser@your-server  # 故意输入错误密码 3 次

# 查看封禁列表
sudo fail2ban-client status sshd
```

### 案例 2: WordPress 登录保护

**创建过滤器**: `/etc/fail2ban/filter.d/wordpress-auth.conf`
```ini
[Definition]
failregex = ^<HOST> .* "POST /wp-login.php
            ^<HOST> .* "POST /xmlrpc.php

ignoreregex =
```

**配置 Jail**: `/etc/fail2ban/jail.local`
```ini
[wordpress-auth]
enabled = true
filter = wordpress-auth
logpath = /var/log/nginx/access.log
port = http,https
maxretry = 3
findtime = 600
bantime = 7200
```

**重启服务**:
```bash
sudo systemctl restart fail2ban
sudo fail2ban-client status wordpress-auth
```

### 案例 3: 防止端口扫描

**创建过滤器**: `/etc/fail2ban/filter.d/port-scan.conf`
```ini
[Definition]
failregex = ^.*Denied .* from <HOST>.*$
ignoreregex =
```

**配置 Jail**: `/etc/fail2ban/jail.local`
```ini
[port-scan]
enabled = true
filter = port-scan
logpath = /var/log/syslog
maxretry = 5
findtime = 300
bantime = 86400
```

### 案例 4: 配置邮件通知

**全局配置**: `/etc/fail2ban/jail.local`
```ini
[DEFAULT]
# 邮件设置
destemail = admin@example.com
sender = fail2ban@example.com
sendername = Fail2Ban

# 使用带邮件的 action
action = %(action_mwl)s
```

**配置发送邮件**:
```bash
# 安装邮件工具
sudo apt install mailutils

# 测试邮件
echo "Test" | mail -s "Test Subject" admin@example.com
```

### 案例 5: 白名单配置

**添加信任的 IP**: `/etc/fail2ban/jail.local`
```ini
[DEFAULT]
# 忽略特定 IP 和网段
ignoreip = 127.0.0.1/8 
           ::1 
           192.168.1.0/24 
           10.0.0.0/8
           203.0.113.50

# 自己的办公室 IP
ignoreip = %(ignoreip)s 203.0.113.0/24
```

---

## 8. 最佳实践

### 安全配置建议

1. **合理设置封禁时间**
   ```ini
   # 短期封禁（适用于一般情况）
   bantime = 3600  # 1小时
   
   # 长期封禁（适用于严重攻击）
   bantime = 86400  # 24小时
   
   # 永久封禁（谨慎使用）
   bantime = -1
   ```

2. **调整重试次数**
   ```ini
   # SSH 服务（更严格）
   maxretry = 3
   
   # Web 服务（适当宽松）
   maxretry = 5
   ```

3. **使用递增封禁时间**
   ```ini
   [recidive]
   enabled = true
   filter = recidive
   logpath = /var/log/fail2ban.log
   action = iptables-allports[name=recidive]
   bantime = 604800  # 7天
   findtime = 86400  # 24小时
   maxretry = 5
   ```

### 监控和维护

1. **定期检查日志**
   ```bash
   # 查看 fail2ban 日志
   sudo tail -f /var/log/fail2ban.log
   
   # 查看今天的封禁记录
   sudo grep "Ban" /var/log/fail2ban.log | grep "$(date '+%Y-%m-%d')"
   ```

2. **监控封禁统计**
   ```bash
   # 创建监控脚本 /usr/local/bin/fail2ban-stats.sh
   #!/bin/bash
   echo "=== Fail2ban 统计 ==="
   echo "当前活动的 Jails:"
   sudo fail2ban-client status
   echo ""
   echo "各 Jail 详细状态:"
   for jail in $(sudo fail2ban-client status | grep "Jail list" | sed -E 's/^[^:]+:[ \t]+//' | sed 's/,//g')
   do
       echo "--- $jail ---"
       sudo fail2ban-client status $jail
   done
   ```

3. **设置定期清理**
   ```bash
   # 添加到 crontab
   # 每周日凌晨 2 点清理所有封禁
   0 2 * * 0 /usr/bin/fail2ban-client unban --all
   ```

### 性能优化

1. **使用 systemd backend**
   ```ini
   [DEFAULT]
   backend = systemd
   ```

2. **优化日志扫描**
   ```ini
   [sshd]
   # 使用特定日志文件而非通配符
   logpath = /var/log/auth.log
   
   # 限制日志扫描范围
   maxlines = 10
   ```

3. **使用 dbpurgeage 清理旧数据**
   ```ini
   [DEFAULT]
   dbpurgeage = 86400  # 24小时后清理数据库
   ```

### 故障排查

1. **配置测试不通过**
   ```bash
   # 检查配置语法
   sudo fail2ban-client -t
   
   # 查看详细错误
   sudo fail2ban-client -vvv start
   ```

2. **Filter 不工作**
   ```bash
   # 测试正则表达式
   sudo fail2ban-regex /var/log/auth.log /etc/fail2ban/filter.d/sshd.conf
   
   # 查看匹配到的内容
   sudo fail2ban-regex /var/log/auth.log /etc/fail2ban/filter.d/sshd.conf --print-all-matched
   ```

3. **IP 未被封禁**
   ```bash
   # 检查 jail 是否运行
   sudo fail2ban-client status sshd
   
   # 检查日志路径是否正确
   sudo fail2ban-client get sshd logpath
   
   # 手动测试封禁
   sudo fail2ban-client set sshd banip 1.2.3.4
   sudo iptables -L f2b-sshd -n
   ```

4. **服务启动失败**
   ```bash
   # 查看系统日志
   sudo journalctl -u fail2ban -n 50
   
   # 查看 fail2ban 日志
   sudo tail -100 /var/log/fail2ban.log
   
   # 检查权限
   ls -la /var/run/fail2ban/
   ls -la /var/log/fail2ban.log
   ```

### 高级配置技巧

1. **配置递增封禁**
   ```ini
   # 对重复违规者加重处罚
   [recidive]
   enabled = true
   filter = recidive
   logpath = /var/log/fail2ban.log
   action = iptables-allports[name=recidive, protocol=all]
   bantime = 1w
   findtime = 1d
   maxretry = 3
   ```

2. **配置多个动作**
   ```ini
   [sshd]
   enabled = true
   filter = sshd
   action = iptables[name=SSH, port=ssh, protocol=tcp]
            sendmail-whois[name=SSH, dest=admin@example.com]
            cloudflare[cfuser="user@example.com", cftoken="XXX"]
   ```

3. **使用自定义脚本**
   ```ini
   # 在 /etc/fail2ban/action.d/custom-script.conf
   [Definition]
   actionban = /usr/local/bin/ban-notify.sh <ip> <name>
   actionunban = /usr/local/bin/unban-notify.sh <ip> <name>
   ```

### 与其他工具集成

1. **与 CloudFlare 集成**
   - 安装 cloudflare action
   - 配置 API token
   - 在 jail 中使用 cloudflare action

2. **与 Slack 集成**
   ```bash
   # 创建通知脚本
   #!/bin/bash
   SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
   MESSAGE="Fail2ban: IP $1 banned in jail $2"
   curl -X POST -H 'Content-type: application/json' \
        --data "{\"text\":\"$MESSAGE\"}" $SLACK_WEBHOOK
   ```

3. **与监控系统集成**
   - 使用 Prometheus exporter
   - 配置 Grafana 仪表板
   - 设置告警规则

---

## 总结

Fail2ban 是一个强大而灵活的入侵防御工具，通过合理配置可以有效保护服务器免受各种暴力攻击。

**关键要点**:
- ✅ 始终使用 `.local` 文件进行配置
- ✅ 定期检查日志和封禁状态
- ✅ 合理设置白名单避免误封
- ✅ 使用递增封禁对付惯犯
- ✅ 配置邮件或其他通知方式
- ✅ 定期测试和更新规则
- ✅ 备份配置文件

**注意事项**:
- ⚠️ 不要封禁自己的 IP
- ⚠️ 注意日志文件路径的正确性
- ⚠️ 永久封禁要谨慎使用
- ⚠️ 测试环境先验证再上生产
- ⚠️ 定期清理封禁列表避免过多规则影响性能

**相关资源**:
- 官方网站: https://www.fail2ban.org
- GitHub: https://github.com/fail2ban/fail2ban
- 文档: https://fail2ban.readthedocs.io
- Wiki: https://www.fail2ban.org/wiki/

---

**更新日期**: 2025-10-04
**版本**: 1.0