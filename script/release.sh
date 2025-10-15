#!/bin/bash

# 快速发布脚本
# 用法: ./scripts/release.sh <version> [--skip-tests]
# 示例: ./scripts/release.sh 1.0.0
# 示例: ./scripts/release.sh 1.0.0 --skip-tests

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认选项
SKIP_TESTS=false

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        --help|-h)
            echo "用法: $0 <version> [选项]"
            echo "选项:"
            echo "  --skip-tests    跳过测试步骤"
            echo "  --help, -h      显示此帮助信息"
            echo ""
            echo "示例:"
            echo "  $0 1.0.0"
            echo "  $0 1.0.0 --skip-tests"
            exit 0
            ;;
        -*)
            echo "未知选项: $1"
            echo "使用 --help 查看帮助"
            exit 1
            ;;
        *)
            if [ -z "$VERSION" ]; then
                VERSION=$1
            else
                echo "错误: 提供了多个版本号"
                exit 1
            fi
            shift
            ;;
    esac
done

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查参数
if [ -z "$VERSION" ]; then
    print_error "请提供版本号"
    echo "用法: $0 <version> [选项]"
    echo "示例: $0 1.0.0"
    echo "示例: $0 v1.0.0"
    echo "示例: $0 1.0.0 --skip-tests"
    echo "使用 --help 查看详细帮助"
    exit 1
fi

# 处理版本号格式，支持带v或不带v的输入
if [[ $VERSION == v* ]]; then
    # 如果已经带v前缀，直接使用
    TAG_NAME="$VERSION"
    VERSION_NUMBER="${VERSION#v}"  # 去掉v前缀用于验证
else
    # 如果不带v前缀，自动添加
    VERSION_NUMBER="$VERSION"
    TAG_NAME="v${VERSION}"
fi

# 验证版本号格式 (支持三位或四位版本号)
if ! [[ $VERSION_NUMBER =~ ^[0-9]+\.[0-9]+\.[0-9]+(\.[0-9]+)?(-[a-zA-Z0-9.-]+)?$ ]]; then
    print_error "版本号格式不正确，请使用版本格式 (例如: 1.0.0, 0.0.1.5, v1.0.0, 1.0.0-alpha)"
    exit 1
fi

print_info "开始发布版本: ${TAG_NAME}"



# 检查工作区是否干净
if ! git diff-index --quiet HEAD --; then
    print_warning "工作区有未提交的更改"
    echo "未提交的文件:"
    git status --porcelain
    echo
    read -p "是否继续发布? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "发布已取消"
        exit 0
    fi
fi

# 检查标签是否已存在
if git rev-parse "$TAG_NAME" >/dev/null 2>&1; then
    print_error "标签 $TAG_NAME 已存在"
    echo "现有标签:"
    git tag --list "v*" | tail -10
    exit 1
fi

# 获取当前分支
CURRENT_BRANCH=$(git symbolic-ref --short HEAD)
print_info "当前分支: $CURRENT_BRANCH"

# 确保在main分支（可选）
if [ "$CURRENT_BRANCH" != "main" ] && [ "$CURRENT_BRANCH" != "master" ]; then
    print_warning "当前不在主分支 (main/master)"
    read -p "是否继续在 $CURRENT_BRANCH 分支发布? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "发布已取消"
        exit 0
    fi
fi



# 创建标签
print_info "创建标签 $TAG_NAME..."
git tag -a "$TAG_NAME" -m "Release version $VERSION_NUMBER" || {
    print_error "创建标签失败"
    exit 1
}

print_success "标签 $TAG_NAME 创建成功"

# 推送标签
print_info "推送标签到远程仓库..."
git push origin "$TAG_NAME" || {
    print_error "推送标签失败"
    print_info "删除本地标签..."
    git tag -d "$TAG_NAME"
    exit 1
}

print_success "标签推送成功"
