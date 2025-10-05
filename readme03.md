# Fail2Ban 学习笔记

## 1. Fail2Ban 基础概念

### 什么是 Fail2Ban？
Fail2Ban 是一个入侵防御软件框架，可以保护计算机服务器免受暴力破解攻击。它通过分析日志文件（如 `/var/log/auth.log`、`/var/log/nginx/access.log` 等）来检测可疑活动，并在发现攻击模式时通过更新防火墙规则临时或永久封禁攻击者 IP。

### 核心组件
- **fail2ban-server**: 后台守护进程，负责监控日志和执行封禁操作
- **fail2ban-client**: 命令行工具，用于与服务器交互
- **filter**: 过滤器，定义如何解析日志并识别攻击模式
- **action**: 动作，定义检测到攻击时的响应措施
- **jail**: 监狱，组合过滤器和动作来保护特定服务

## 2. 配置详解

### 配置文件结构
- `/etc/fail2ban/fail2ban.conf` - 全局配置
- `/etc/fail2ban/jail.conf` - 默认jail配置
- `/etc/fail2ban/jail.d/*.conf` - 自定义jail配置（优先级更高）
- `/etc/fail2ban/filter.d/*.conf` - 过滤器定义
- `/etc/fail2ban/action.d/*.conf` - 动作定义

### 关键配置参数
```ini
[DEFAULT]
# 封禁时间（秒）
bantime = 3600

# 查找时间窗口（秒）
findtime = 600

# 最大失败尝试次数
maxretry = 3

# 忽略的IP地址（不会被封禁）
ignoreip = 127.0.0.1/8 ::1

# 日志路径
logtarget = /var/log/fail2ban.log

# 日志级别
loglevel = INFO
```

### jail配置示例
```ini
[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 5
bantime = 3600
```

### 优化建议（基于实践）
- **递增封禁时间**: 重复攻击者应获得更长的封禁时间
  ```ini
  [DEFAULT]
  bantime.increment = true
  bantime.factor = 1
  bantime.formula = ban.Time * (1<<(ban.Count if ban.Count<20 else 20)) * banFactor
  bantime.multipliers = 1 2 4 8 16 32 64 128 256 512 1024
  ```
- **永久封禁**: 对严重攻击者设置极长封禁时间
  ```ini
  [recidive]
  enabled = true
  filter = recidive
  logpath = /var/log/fail2ban.log
  bantime = 604800  # 1周
  findtime = 86400  # 1天
  maxretry = 5
  ```

## 3. 常用命令

### 服务管理
```bash
# 启动/停止/重启服务
sudo systemctl start fail2ban
sudo systemctl stop fail2ban
sudo systemctl restart fail2ban

# 检查状态
sudo systemctl status fail2ban

# 重新加载配置
sudo fail2ban-client reload
```

### 监控命令
```bash
# 查看所有jail状态
sudo fail2ban-client status

# 查看特定jail状态
sudo fail2ban-client status sshd

# 查看封禁IP列表
sudo fail2ban-client status sshd | grep "Banned IP list"

# 查看fail2ban日志
sudo tail -f /var/log/fail2ban.log
```

### IP管理命令
```bash
# 手动封禁IP
sudo fail2ban-client set sshd banip 192.168.1.100

# 手动解封IP
sudo fail2ban-client set sshd unbanip 192.168.1.100

# 测试过滤器
sudo fail2ban-regex /var/log/auth.log /etc/fail2ban/filter.d/sshd.conf
```

## 4. 自定义过滤器和动作

### 创建自定义过滤器
```bash
# 创建PHP攻击过滤器
sudo nano /etc/fail2ban/filter.d/nginx-php-attack.conf
```

```ini
[Definition]
failregex = ^<HOST> .*(\.php|\.asp|\.aspx|\.jsp)
ignoreregex =
```

### 创建自定义动作
```bash
# 创建通知动作
sudo nano /etc/fail2ban/action.d/email-notification.conf
```

```ini
[Definition]
actionban = echo "IP: <ip> banned at `date`" | mail -s "Fail2Ban notification" admin@example.com
actionunban = echo "IP: <ip> unbanned at `date`" | mail -s "Fail2Ban notification" admin@example.com
```

## 5. 进阶功能与实践经验

### 智能分析与封禁
- **威胁评分系统**: 根据攻击类型、频率和严重性评分
- **多因素分析**: 结合多个日志源进行综合分析
- **递增惩罚**: 根据攻击者历史记录调整封禁时间

### 白名单管理
- 总是为内部网络添加白名单: `ignoreip = 127.0.0.1/8 10.0.0.0/8`
- 为合法扫描工具添加白名单: `ignoreip += 8.8.8.8`

### 日志分析技巧
- 结合 `grep`, `awk` 和 `sed` 分析日志模式
- 使用 `fail2ban-regex` 测试过滤器有效性
- 定期审查日志找出新的攻击模式

### 常见攻击模式
1. **SSH暴力破解**: 大量失败登录尝试
2. **Web攻击**: PHP、管理页面扫描
3. **WordPress攻击**: xmlrpc.php、wp-login.php尝试
4. **扫描器识别**: 短时间内访问多个404页面

## 6. 整合 Fail2Ban Web 管理面板

### 关键功能
1. **实时监控**: 查看当前封禁IP和状态
2. **IP地理分析**: 提供攻击来源地理信息
3. **智能防护**: 基于复杂规则自动封禁
4. **白名单保护**: 防止误封重要IP
5. **日志可视化**: 图表展示攻击趋势

### 配置建议
- **封禁时间**: 设置较长时间（1小时以上）
- **日志路径**: 确保Web应用有权访问
- **权限处理**: 使用sudo配置正确权限

### 安全最佳实践
1. **多层防御**: Fail2Ban只是安全策略的一部分
2. **定期更新**: 保持软件和规则最新
3. **备份配置**: 保留已验证有效的配置
4. **监控系统**: 设置失败告警机制
5. **日志轮转**: 防止日志文件过大

## 7. 故障排除

### 常见问题
1. **服务无法启动**: 检查配置文件语法
   ```bash
   sudo fail2ban-client -t
   ```

2. **过滤器无效**: 测试正则表达式
   ```bash
   sudo fail2ban-regex /var/log/auth.log /etc/fail2ban/filter.d/sshd.conf
   ```

3. **权限问题**: 确保日志文件可读
   ```bash
   sudo chmod 640 /var/log/auth.log
   sudo usermod -a -G adm fail2ban
   ```

4. **封禁不生效**: 检查防火墙规则
   ```bash
   sudo iptables -L -n
   ```

5. **IP解封过快**: 检查bantime设置
   ```bash
   sudo fail2ban-client get sshd bantime
   ```

## 8. 实用脚本

### 分析fail2ban状态
```bash
#!/bin/bash
echo -e "\n--- Fail2Ban状态概览 ---"
echo "总共jail数: $(sudo fail2ban-client status | grep "Number of jail" | cut -d: -f2)"
echo "已启用jail列表:"
sudo fail2ban-client status | grep "Jail list" | cut -d: -f2

for jail in $(sudo fail2ban-client status | grep "Jail list" | cut -d: -f2 | tr ',' ' '); do
    echo -e "\n--- $jail 详情 ---"
    sudo fail2ban-client status $jail
done

echo -e "\n--- 最近封禁记录 ---"
sudo tail -20 /var/log/fail2ban.log | grep "Ban"
```

### IP地理位置分析
```bash
#!/bin/bash
# 提取被封禁的IP
banned_ips=$(sudo fail2ban-client status | grep "Jail list" | cut -d: -f2 | tr ',' ' ' | xargs -I{} sudo fail2ban-client status {} | grep "Banned IP list" | cut -d: -f2)

# 分析每个IP
for ip in $banned_ips; do
    geo=$(curl -s "http://ip-api.com/json/$ip")
    country=$(echo $geo | jq -r ".country")
    city=$(echo $geo | jq -r ".city")
    isp=$(echo $geo | jq -r ".isp")
    echo "$ip: $country, $city, $isp"
    sleep 1
done
```

## 总结

Fail2Ban是一个强大而灵活的安全工具，通过日志分析和自动响应机制，能有效减少服务器面临的暴力破解风险。结合Web管理面板，可以更直观地监控和管理系统安全状态，提高安全防护效率。实践中，合理配置封禁参数、定制专用规则、实施智能分析，才能发挥Fail2Ban的最大价值。

---

**注意**：在README中设置图片宽高，您可以使用HTML语法而不是Markdown语法：

```html
<img src="img/451759627900_.pic.jpg" width="400" height="600" alt="描述文本">
```

或者使用HTML属性和样式：

```html
<img src="img/451759627900_.pic.jpg" style="width: 80%; max-width: 800px;" alt="描述文本">
```

Markdown本身不支持直接设置图片尺寸，但GitHub和大多数Markdown渲染器支持内嵌HTML。