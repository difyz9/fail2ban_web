#!/bin/bash

# 前端资源构建脚本
# 用于打包前端代码到 web 目录，然后通过 embed 嵌入到 Go 二进制文件

set -e

echo "🚀 开始前端资源构建..."

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT=$(pwd)
WEB_DIR="${PROJECT_ROOT}/web"
FRONTEND_SRC="${PROJECT_ROOT}/frontend"
STATIC_DIR="${WEB_DIR}/static"
TEMPLATES_DIR="${WEB_DIR}/templates"

echo -e "${BLUE}📂 项目根目录: ${PROJECT_ROOT}${NC}"
echo -e "${BLUE}📂 Web 目录: ${WEB_DIR}${NC}"

# 检查 web 目录是否存在
if [ ! -d "$WEB_DIR" ]; then
    echo -e "${RED}❌ web 目录不存在，正在创建...${NC}"
    mkdir -p "$WEB_DIR"
fi

# 如果存在独立的 frontend 源码目录，则构建
if [ -d "$FRONTEND_SRC" ]; then
    echo -e "${YELLOW}📦 检测到 frontend 源码目录，开始构建...${NC}"
    
    cd "$FRONTEND_SRC"
    
    # 检查是否有 package.json (Node.js 项目)
    if [ -f "package.json" ]; then
        echo -e "${BLUE}🔍 检测到 Node.js 项目${NC}"
        
        # 检查 node_modules 是否存在
        if [ ! -d "node_modules" ]; then
            echo -e "${YELLOW}📥 安装依赖...${NC}"
            npm install
        fi
        
        # 运行构建命令
        echo -e "${YELLOW}🔨 执行 npm run build...${NC}"
        npm run build
        
        # 清空并重新创建 web 目录
        echo -e "${BLUE}🧹 清理 web 目录...${NC}"
        rm -rf "$WEB_DIR"/*
        
        # 复制构建产物到 web 目录
        if [ -d "out" ]; then
            echo -e "${GREEN}✅ 复制构建产物 (out) 到 web 目录...${NC}"
            cp -r out/* "$WEB_DIR/"
            
            # 如果存在 public/index.html，用它覆盖 Next.js 生成的 index.html
            if [ -f "public/index.html" ]; then
                echo -e "${GREEN}✅ 使用自定义 index.html 覆盖默认文件...${NC}"
                cp public/index.html "$WEB_DIR/index.html"
            fi
        elif [ -d "dist" ]; then
            echo -e "${GREEN}✅ 复制构建产物 (dist) 到 web 目录...${NC}"
            cp -r dist/* "$WEB_DIR/"
        elif [ -d "build" ]; then
            echo -e "${GREEN}✅ 复制构建产物 (build) 到 web 目录...${NC}"
            cp -r build/* "$WEB_DIR/"
        else
            echo -e "${RED}❌ 未找到构建产物目录 (out, dist 或 build)${NC}"
            exit 1
        fi
    else
        echo -e "${YELLOW}⚠️  frontend 目录下没有 package.json，跳过构建${NC}"
    fi
    
    cd "$PROJECT_ROOT"
else
    echo -e "${YELLOW}ℹ️  未检测到 frontend 源码目录，使用现有 web 资源${NC}"
fi

# 统计文件信息
echo ""
echo -e "${GREEN}✅ 前端资源构建完成！${NC}"
echo ""
echo -e "${BLUE}📊 资源统计:${NC}"
echo -e "  总文件数: $(find $WEB_DIR -type f 2>/dev/null | wc -l | xargs)"
echo -e "  HTML 文件: $(find $WEB_DIR -name "*.html" 2>/dev/null | wc -l | xargs)"
echo -e "  CSS 文件:  $(find $WEB_DIR -name "*.css" 2>/dev/null | wc -l | xargs)"
echo -e "  JS 文件:   $(find $WEB_DIR -name "*.js" 2>/dev/null | wc -l | xargs)"
echo ""
echo -e "${GREEN}🎉 所有资源已准备就绪，可以通过 embed 打包到二进制文件！${NC}"
echo -e "${BLUE}💡 下一步: 运行 'make build' 或 'make build-prod' 构建完整应用${NC}"
