#!/bin/bash
# fail2ban状态分析脚本

echo "=== Fail2Ban 状态分析 ==="

# 1. 检查当前sshd jail配置
echo -e "\n--- SSHD Jail 配置 ---"
echo "封禁时间: $(sudo fail2ban-client get sshd bantime)秒"
echo "查找时间: $(sudo fail2ban-client get sshd findtime)秒" 
echo "最大重试: $(sudo fail2ban-client get sshd maxretry)次"

# 2. 检查当前状态
echo -e "\n--- 当前状态 ---"
sudo fail2ban-client status sshd

# 3. 查看最近的fail2ban日志
echo -e "\n--- 最近的封禁记录 ---"
sudo tail -20 /var/log/fail2ban.log | grep -E "(Ban|Unban)" | tail -10

# 4. 分析SSH攻击日志
echo -e "\n--- SSH攻击分析 ---"
echo "最近1小时的SSH失败登录:"
sudo grep "$(date -d '1 hour ago' +'%b %d %H')" /var/log/auth.log 2>/dev/null | grep -i "failed\|invalid" | wc -l || echo "无法访问auth.log"

# 5. 推荐配置
echo -e "\n--- 建议优化配置 ---"
cat << 'EOF'
当前bantime=600秒(10分钟)太短，建议:

1. 增加封禁时间:
   bantime = 3600    # 1小时
   或
   bantime = 86400   # 24小时

2. 使用递增封禁时间:
   bantime.increment = true
   bantime.factor = 2
   bantime.maxtime = 86400

3. 调整检测参数:
   findtime = 600    # 10分钟内
   maxretry = 3      # 失败3次就封禁

修改配置文件: /etc/fail2ban/jail.local
EOF

# 6. 生成优化配置示例
echo -e "\n--- 优化配置示例 ---"
cat << 'EOF'
创建 /etc/fail2ban/jail.local:

[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3
bantime.increment = true
bantime.factor = 2
bantime.maxtime = 86400

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
banaction = iptables-multiport
EOF

echo -e "\n=== 分析完成 ==="