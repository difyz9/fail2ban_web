# 🚀 快速版本发布指南

## 一键发布新版本

### 最常用命令

```bash
# 快速发布（自动递增补丁版本 v0.0.8 → v0.0.9）
make release

# 或带提交信息的发布
make tag-push
```

### 其他版本类型

```bash
# 次版本号递增（v0.0.8 → v0.1.0）
make tag-minor

# 主版本号递增（v0.0.8 → v1.0.0）
make tag-major
```

### 查看版本信息

```bash
# 当前版本
make get-version

# 所有版本
make list-tags
```

## 📖 详细说明

查看 [VERSION_MANAGEMENT.md](./VERSION_MANAGEMENT.md) 了解：
- ✅ 完整的版本管理流程
- ✅ 语义化版本号规范
- ✅ 常见问题解答
- ✅ 最佳实践建议

## 🎯 推荐工作流

```bash
# 1. 提交代码
git add .
git commit -m "Your changes"
git push origin main

# 2. 发布新版本
make release

# 3. 完成！✨
```

---

**当前版本**: v0.0.8 | [查看所有版本](https://github.com/difyz9/fail2ban_web/tags)
