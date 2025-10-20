#!/bin/bash

# 快速验证 Embed 功能是否正常工作

echo "🔍 快速验证 Embed 功能..."
echo ""

# 1. 检查 web 目录
if [ ! -d "web" ]; then
    echo "❌ web 目录不存在，请先运行：make build-frontend"
    exit 1
fi

echo "✅ web 目录存在"
HTML_COUNT=$(find web -name "*.html" | wc -l | xargs)
echo "   - HTML 文件: $HTML_COUNT 个"

# 2. 检查二进制文件
if [ ! -f "build/fail2ban-web" ]; then
    echo "❌ 二进制文件不存在，请先运行：make build"
    exit 1
fi

echo "✅ 二进制文件存在"
FILE_SIZE=$(ls -lh build/fail2ban-web | awk '{print $5}')
echo "   - 文件大小: $FILE_SIZE"

# 3. 检查是否包含 embed 标记
if grep -q "//go:embed web" main.go; then
    echo "✅ main.go 包含 embed 配置"
else
    echo "❌ main.go 缺少 embed 配置"
    exit 1
fi

# 4. 检查 core/app_server.go
if grep -q "Detected Next.js build output" core/app_server.go; then
    echo "✅ app_server.go 包含 Next.js 检测逻辑"
else
    echo "⚠️  app_server.go 可能缺少 Next.js 支持"
fi

echo ""
echo "🎉 所有检查通过！"
echo ""
echo "📝 下一步："
echo "   1. 启动应用: ./build/fail2ban-web"
echo "   2. 访问前端: http://localhost:8099"
echo "   3. 运行测试: ./test_frontend.sh"
