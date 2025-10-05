# Nginx 日志解析和恶意请求拦截配置说明

## 概述
本配置包含了10个预定义的 Fail2Ban jail 规则，专门用于保护 Nginx 服务器免受各种网络攻击。

## 包含的规则

### 1. nginx-http-auth
- **用途**: HTTP基本认证失败保护
- **监控**: /var/log/nginx/error.log
- **触发**: 5次失败尝试/10分钟
- **封禁**: 1小时

### 2. nginx-limit-req
- **用途**: 限速模块触发保护
- **监控**: /var/log/nginx/error.log
- **触发**: 10次触发/1分钟
- **封禁**: 10分钟

### 3. nginx-botsearch
- **用途**: 恶意爬虫和搜索机器人保护
- **监控**: /var/log/nginx/access.log
- **触发**: 3次恶意请求/10分钟
- **封禁**: 2小时
- **检测**: PHP文件访问、路径遍历、SQL注入尝试

### 4. nginx-noscript
- **用途**: 脚本文件访问保护
- **监控**: /var/log/nginx/access.log
- **触发**: 6次脚本访问尝试/10分钟
- **封禁**: 1小时

### 5. nginx-bad-request
- **用途**: 恶意HTTP请求保护
- **监控**: /var/log/nginx/access.log
- **触发**: 3次错误请求/5分钟
- **封禁**: 30分钟
- **检测**: 400、444、499错误码

### 6. nginx-noproxy
- **用途**: 代理滥用保护
- **监控**: /var/log/nginx/access.log
- **触发**: 2次代理尝试/2分钟
- **封禁**: 1小时

### 7. nginx-req-limit
- **用途**: 请求频率限制
- **监控**: /var/log/nginx/error.log
- **触发**: 5次限制触发/1分钟
- **封禁**: 5分钟

### 8. nginx-cc-ddos
- **用途**: CC/DDoS攻击保护
- **监控**: /var/log/nginx/access.log
- **触发**: 20次请求/1分钟
- **封禁**: 4小时
- **注意**: 此规则较为严格，建议测试后使用

### 9. nginx-404 (默认禁用)
- **用途**: 404错误频率保护
- **监控**: /var/log/nginx/access.log
- **触发**: 10次404错误/10分钟
- **封禁**: 10分钟
- **注意**: 默认禁用，避免误封正常用户

### 10. nginx-forbidden
- **用途**: 403禁止访问保护
- **监控**: /var/log/nginx/access.log
- **触发**: 5次403错误/5分钟
- **封禁**: 30分钟

## 安装步骤

1. **通过Web界面安装**:
   - 访问管理面板的"规则管理"页面
   - 点击"安装Nginx默认配置"按钮
   - 系统会自动创建所有规则配置

2. **手动安装**:
   ```bash
   # 1. 下载配置文件
   curl -o nginx-jails.conf http://your-panel-url/api/v1/defaults/nginx/jail-config
   
   # 2. 复制到fail2ban配置目录
   sudo cp nginx-jails.conf /etc/fail2ban/jail.d/
   
   # 3. 重启fail2ban服务
   sudo systemctl restart fail2ban
   ```

## 自定义配置

### 修改日志路径
如果你的Nginx日志文件位置不同，需要修改相应的 `logpath` 参数：

```ini
# 示例：自定义日志路径
[nginx-http-auth]
logpath = /var/log/nginx/your-custom-error.log

[nginx-botsearch]
logpath = /var/log/nginx/your-custom-access.log
```

### 调整封禁时间和重试次数
根据你的需求调整参数：

```ini
# 更严格的设置
maxretry = 3      # 降低重试次数
findtime = 300    # 缩短检测时间窗口
bantime = 7200    # 延长封禁时间

# 更宽松的设置
maxretry = 10     # 增加重试次数
findtime = 1200   # 延长检测时间窗口
bantime = 600     # 缩短封禁时间
```

## 白名单设置

为了避免误封信任的IP地址，建议设置白名单：

```ini
# 在 /etc/fail2ban/jail.local 中添加
[DEFAULT]
ignoreip = 127.0.0.1/8 192.168.1.0/24 your.trusted.ip.address
```

## 监控和维护

### 检查状态
```bash
# 查看所有jail状态
sudo fail2ban-client status

# 查看特定jail状态
sudo fail2ban-client status nginx-http-auth

# 查看被封IP列表
sudo fail2ban-client get nginx-http-auth banip
```

### 手动解封
```bash
# 解封特定IP
sudo fail2ban-client set nginx-http-auth unbanip 192.168.1.100
```

### 日志监控
```bash
# 监控fail2ban日志
sudo tail -f /var/log/fail2ban.log

# 监控nginx错误日志
sudo tail -f /var/log/nginx/error.log
```

## 注意事项

1. **测试环境**: 建议先在测试环境中验证配置
2. **监控误封**: 部署后密切监控是否有误封正常用户
3. **性能影响**: 大量规则可能对性能有轻微影响
4. **日志轮转**: 确保nginx日志正确轮转，避免日志文件过大
5. **备份配置**: 修改前备份原有配置文件

## 故障排除

### 常见问题

1. **规则不生效**:
   - 检查日志文件路径是否正确
   - 确认nginx日志格式匹配过滤器规则
   - 查看fail2ban日志中的错误信息

2. **误封问题**:
   - 添加可信IP到白名单
   - 调整规则参数（增加重试次数、延长检测时间）
   - 禁用过于严格的规则

3. **性能问题**:
   - 优化日志文件大小
   - 调整检测频率
   - 考虑硬件升级

## 高级配置

### 自定义动作
可以自定义封禁动作，例如发送通知邮件：

```ini
[nginx-http-auth]
action = iptables-multiport[name=nginx-http-auth, port="http,https", protocol=tcp]
         sendmail[name=nginx-http-auth, dest=admin@example.com]
```

### 地理位置封禁
结合GeoIP数据库实现基于地理位置的封禁：

```ini
[nginx-geoblock]
enabled = true
filter = nginx-geoblock
action = iptables-multiport[name=nginx-geoblock, port="http,https", protocol=tcp]
```

通过这些配置，你的Nginx服务器将获得全面的安全保护，有效抵御各种网络攻击。