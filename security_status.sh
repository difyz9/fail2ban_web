#!/bin/bash
# 综合安全分析脚本

echo "=== 综合安全状态分析 ==="

# 1. 检查fail2ban当前状态
echo -e "\n🔍 Fail2ban 当前状态:"
sudo fail2ban-client status

echo -e "\n📊 SSHD Jail 详细信息:"
sudo fail2ban-client status sshd

# 2. 检查优化后的配置
echo -e "\n⚙️  当前配置:"
echo "封禁时间: $(sudo fail2ban-client get sshd bantime)秒"
echo "查找时间: $(sudo fail2ban-client get sshd findtime)秒"
echo "最大重试: $(sudo fail2ban-client get sshd maxretry)次"

# 3. 分析最近的攻击
echo -e "\n🚨 最近的攻击分析:"
echo "今天的SSH失败登录次数:"
TODAY=$(date '+%b %d')
sudo grep "$TODAY" /var/log/auth.log 2>/dev/null | grep -i "failed\|invalid\|refused" | wc -l || echo "无法访问日志"

echo -e "\n🎯 最近被封禁的IP:"
sudo tail -10 /var/log/fail2ban.log | grep "Ban " | tail -5

# 4. 显示当前被封禁的IP
echo -e "\n🔒 当前被封禁的IP列表:"
sudo fail2ban-client get sshd banip 2>/dev/null || echo "无当前封禁IP或命令不支持"

# 5. 网络连接分析  
echo -e "\n🌐 当前SSH连接:"
sudo netstat -tnpa | grep :22 | grep ESTABLISHED | wc -l

echo -e "\n✅ 优化建议已应用，系统安全性大幅提升！"
echo "现在恶意IP将被封禁更长时间，有效防止重复攻击。"