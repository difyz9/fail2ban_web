#!/bin/bash
# fail2ban配置错误诊断和修复脚本

echo "=== Fail2Ban 配置错误诊断 ==="

# 1. 检查配置文件语法
echo -e "\n--- 检查配置语法 ---"
sudo fail2ban-client -t 2>&1 | head -20

# 2. 查看详细错误日志
echo -e "\n--- 详细错误日志 ---"
sudo journalctl -u fail2ban.service --no-pager -n 20

# 3. 检查配置文件存在性
echo -e "\n--- 检查配置文件 ---"
echo "jail.conf: $(ls -la /etc/fail2ban/jail.conf 2>/dev/null || echo '不存在')"
echo "jail.local: $(ls -la /etc/fail2ban/jail.local 2>/dev/null || echo '不存在')"

# 4. 备份当前配置
echo -e "\n--- 备份当前配置 ---"
if [ -f /etc/fail2ban/jail.local ]; then
    sudo cp /etc/fail2ban/jail.local /etc/fail2ban/jail.local.backup.$(date +%Y%m%d_%H%M%S)
    echo "已备份jail.local"
fi

# 5. 创建最小可用配置
echo -e "\n--- 创建最小可用配置 ---"
sudo tee /etc/fail2ban/jail.local > /dev/null << 'EOF'
# Fail2Ban 最小可用配置

[DEFAULT]
# 基础设置
bantime = 3600
findtime = 600
maxretry = 3

# 忽略本地IP
ignoreip = 127.0.0.1/8 ::1 192.168.0.0/16 10.0.0.0/8 172.16.0.0/12

# 基础动作
banaction = iptables-multiport

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
EOF

echo "已创建最小可用配置"

# 6. 测试新配置
echo -e "\n--- 测试新配置 ---"
sudo fail2ban-client -t

if [ $? -eq 0 ]; then
    echo "✓ 配置语法正确"
    
    # 7. 重启服务
    echo -e "\n--- 重启服务 ---"
    sudo systemctl restart fail2ban
    sleep 2
    
    # 8. 检查服务状态
    echo -e "\n--- 检查服务状态 ---"
    if sudo systemctl is-active --quiet fail2ban; then
        echo "✓ fail2ban服务运行正常"
        echo -e "\n--- 当前jail状态 ---"
        sudo fail2ban-client status
    else
        echo "✗ fail2ban服务仍然失败"
        echo "查看日志: sudo journalctl -u fail2ban.service -f"
    fi
else
    echo "✗ 配置语法错误，请检查配置文件"
fi

echo -e "\n=== 诊断完成 ==="