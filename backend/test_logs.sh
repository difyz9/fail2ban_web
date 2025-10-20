#!/bin/bash

# 测试脚本：验证Fail2Ban Web Panel的日志分析功能

echo "=== Fail2Ban Web Panel 日志分析测试 ==="
echo

# 检查Nginx日志
echo "1. 检查Nginx日志文件..."
NGINX_LOGS=(
    "/var/log/nginx/access.log"
    "/var/log/nginx/default.access.log"
    "/usr/local/nginx/logs/access.log"
)

for log_file in "${NGINX_LOGS[@]}"; do
    if [[ -f "$log_file" ]]; then
        echo "✓ 找到Nginx日志: $log_file"
        echo "  文件大小: $(ls -lh "$log_file" | awk '{print $5}')"
        echo "  最后修改: $(ls -l "$log_file" | awk '{print $6, $7, $8}')"
        echo "  最新5行:"
        tail -n 5 "$log_file" | sed 's/^/    /'
        echo
        break
    else
        echo "✗ 未找到: $log_file"
    fi
done

# 检查SSH日志
echo "2. 检查SSH日志文件..."
SSH_LOGS=(
    "/var/log/auth.log"
    "/var/log/secure"
    "/var/log/messages"
)

for log_file in "${SSH_LOGS[@]}"; do
    if [[ -f "$log_file" ]]; then
        echo "✓ 找到SSH日志: $log_file"
        echo "  文件大小: $(ls -lh "$log_file" | awk '{print $5}')"
        echo "  最后修改: $(ls -l "$log_file" | awk '{print $6, $7, $8}')"
        echo "  SSH相关的最新5行:"
        grep -i ssh "$log_file" | tail -n 5 | sed 's/^/    /'
        echo
        break
    else
        echo "✗ 未找到: $log_file"
    fi
done

# 检查权限
echo "3. 检查文件权限..."
for log_file in "${NGINX_LOGS[@]}" "${SSH_LOGS[@]}"; do
    if [[ -f "$log_file" ]]; then
        permissions=$(ls -l "$log_file" | awk '{print $1, $3, $4}')
        echo "  $log_file: $permissions"
        
        if [[ -r "$log_file" ]]; then
            echo "    ✓ 可读"
        else
            echo "    ✗ 不可读 - 需要sudo权限"
        fi
    fi
done

echo

# 生成测试日志条目
echo "4. 生成测试Nginx日志条目..."
cat << 'EOF' > /tmp/test_nginx.log
192.168.1.100 - - [03/Oct/2025:22:45:01 +0000] "GET / HTTP/1.1" 200 615 "-" "Mozilla/5.0"
10.0.0.1 - - [03/Oct/2025:22:45:02 +0000] "POST /login HTTP/1.1" 401 82 "-" "curl/7.68.0"
192.168.1.200 - - [03/Oct/2025:22:45:03 +0000] "GET /admin HTTP/1.1" 404 162 "-" "sqlmap/1.0"
172.16.0.1 - - [03/Oct/2025:22:45:04 +0000] "GET /?id=1' OR '1'='1 HTTP/1.1" 400 400 "-" "BadBot/1.0"
203.0.113.1 - - [03/Oct/2025:22:45:05 +0000] "GET /wp-admin HTTP/1.1" 404 162 "-" "scanner"
EOF

echo "✓ 测试日志已生成: /tmp/test_nginx.log"
echo "内容预览:"
cat /tmp/test_nginx.log | sed 's/^/  /'
echo

# 生成测试SSH日志条目
echo "5. 生成测试SSH日志条目..."
cat << 'EOF' > /tmp/test_ssh.log
Oct  3 22:45:01 ubuntu sshd[1234]: Failed password for admin from 192.168.1.100 port 22 ssh2
Oct  3 22:45:02 ubuntu sshd[1235]: Failed password for root from 10.0.0.1 port 22 ssh2
Oct  3 22:45:03 ubuntu sshd[1236]: Failed password for admin from 192.168.1.100 port 22 ssh2
Oct  3 22:45:04 ubuntu sshd[1237]: Failed password for test from 172.16.0.1 port 22 ssh2
Oct  3 22:45:05 ubuntu sshd[1238]: Accepted password for ubuntu from 192.168.1.50 port 22 ssh2
EOF

echo "✓ 测试日志已生成: /tmp/test_ssh.log"
echo "内容预览:"
cat /tmp/test_ssh.log | sed 's/^/  /'
echo

# 建议的环境变量配置
echo "6. 建议的环境变量配置..."
echo "如果日志路径不同，请在.env文件中设置:"
echo
echo "# Nginx日志路径"
for log_file in "${NGINX_LOGS[@]}"; do
    if [[ -f "$log_file" ]]; then
        echo "NGINX_ACCESS_LOG=$log_file"
        break
    fi
done

echo
echo "# SSH日志路径"
for log_file in "${SSH_LOGS[@]}"; do
    if [[ -f "$log_file" ]]; then
        echo "SSH_LOG_PATH=$log_file"
        break
    fi
done

echo
echo "# 如果需要sudo权限访问日志"
echo "FAIL2BAN_FORCE_SUDO=true"
echo

echo "=== 测试完成 ==="
echo
echo "提示:"
echo "1. 如果日志文件权限不足，请运行应用时使用sudo或配置适当的权限"
echo "2. 确保日志文件路径在配置中正确设置"
echo "3. 查看应用日志以确认日志解析是否正常工作"
echo "4. 运行: sudo journalctl -u fail2ban-web -f"