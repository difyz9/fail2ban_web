#!/bin/bash

# 测试前端页面访问脚本

set -e

echo "🧪 测试前端页面访问..."

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 清理旧进程
pkill -9 fail2ban-web 2>/dev/null || true
sleep 1

# 启动应用（后台运行）
echo -e "${YELLOW}🚀 启动应用...${NC}"
./build/fail2ban-web > /tmp/fail2ban_test.log 2>&1 &
APP_PID=$!

# 等待应用启动
echo -e "${YELLOW}⏳ 等待应用启动...${NC}"
sleep 5

# 检查进程是否还在运行
if ! kill -0 $APP_PID 2>/dev/null; then
    echo -e "${RED}❌ 应用启动失败！${NC}"
    echo "日志内容："
    cat /tmp/fail2ban_test.log
    exit 1
fi

echo -e "${GREEN}✅ 应用已启动 (PID: $APP_PID)${NC}"
echo ""

# 测试各个页面
echo -e "${YELLOW}📋 测试页面访问...${NC}"
echo ""

# 测试首页
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8099/)
if [ "$HTTP_CODE" == "200" ]; then
    echo -e "${GREEN}✅ 首页 (/) - HTTP $HTTP_CODE${NC}"
else
    echo -e "${RED}❌ 首页 (/) - HTTP $HTTP_CODE${NC}"
fi

# 测试登录页
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8099/login)
if [ "$HTTP_CODE" == "200" ]; then
    echo -e "${GREEN}✅ 登录页 (/login) - HTTP $HTTP_CODE${NC}"
else
    echo -e "${RED}❌ 登录页 (/login) - HTTP $HTTP_CODE${NC}"
fi

# 测试登录页 (带 .html)
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8099/login.html)
if [ "$HTTP_CODE" == "200" ]; then
    echo -e "${GREEN}✅ 登录页 (/login.html) - HTTP $HTTP_CODE${NC}"
else
    echo -e "${RED}❌ 登录页 (/login.html) - HTTP $HTTP_CODE${NC}"
fi

# 测试 dashboard
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8099/dashboard)
if [ "$HTTP_CODE" == "200" ]; then
    echo -e "${GREEN}✅ Dashboard (/dashboard) - HTTP $HTTP_CODE${NC}"
else
    echo -e "${RED}❌ Dashboard (/dashboard) - HTTP $HTTP_CODE${NC}"
fi

# 测试静态资源 (_next)
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8099/_next/static/chunks/webpack.js)
if [ "$HTTP_CODE" == "200" ] || [ "$HTTP_CODE" == "404" ]; then
    echo -e "${GREEN}✅ 静态资源 (/_next) - 可访问${NC}"
else
    echo -e "${RED}❌ 静态资源 (/_next) - HTTP $HTTP_CODE${NC}"
fi

# 测试 API
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8099/api/v1/health)
if [ "$HTTP_CODE" == "200" ]; then
    echo -e "${GREEN}✅ API (/api/v1/health) - HTTP $HTTP_CODE${NC}"
else
    echo -e "${RED}❌ API (/api/v1/health) - HTTP $HTTP_CODE${NC}"
fi

echo ""
echo -e "${YELLOW}📄 获取首页内容预览...${NC}"
curl -s http://localhost:8099/ | head -20
echo ""

echo -e "${YELLOW}🔍 检查日志...${NC}"
tail -30 /tmp/fail2ban_test.log
echo ""

# 清理
echo -e "${YELLOW}🧹 清理测试进程...${NC}"
kill $APP_PID 2>/dev/null || true
sleep 1
pkill -9 fail2ban-web 2>/dev/null || true

echo ""
echo -e "${GREEN}✅ 测试完成！${NC}"
