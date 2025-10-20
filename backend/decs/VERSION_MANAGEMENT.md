# 版本管理说明

## 📋 概述

项目使用语义化版本号规范（Semantic Versioning），格式为：`vMAJOR.MINOR.PATCH`

- **MAJOR（主版本号）**: 不兼容的 API 修改
- **MINOR（次版本号）**: 向下兼容的功能性新增
- **PATCH（补丁版本号）**: 向下兼容的问题修正

---

## 🚀 快速使用

### 1. 最常用：递增补丁版本（推荐）

```bash
# 自动从 v0.0.8 递增到 v0.0.9
make tag-push
```

或使用快速发布（不询问提交信息）：

```bash
make release
```

### 2. 递增次版本号

```bash
# 自动从 v0.0.8 递增到 v0.1.0
make tag-minor
```

### 3. 递增主版本号

```bash
# 自动从 v0.0.8 递增到 v1.0.0
make tag-major
```

---

## 📝 命令详解

### 版本管理命令

| 命令 | 功能 | 示例 |
|------|------|------|
| `make tag-push` | 递增补丁版本号并推送（会询问提交信息） | v0.0.8 → v0.0.9 |
| `make tag-minor` | 递增次版本号并推送（会询问提交信息） | v0.0.8 → v0.1.0 |
| `make tag-major` | 递增主版本号并推送（会询问提交信息） | v0.0.8 → v1.0.0 |
| `make release` | 快速发布（自动递增补丁版本，无提示） | v0.0.8 → v0.0.9 |
| `make get-version` | 查看当前版本号 | 显示：v0.0.8 |
| `make list-tags` | 查看所有版本标签 | 显示所有 tag |

### 标签管理命令

| 命令 | 功能 |
|------|------|
| `make delete-tag` | 删除本地标签（会提示输入标签名） |
| `make delete-tag-remote` | 删除远程标签（会提示输入标签名） |

---

## 💡 使用场景

### 场景 1: 修复 Bug
```bash
# 修复了一个登录问题
make tag-push
# 输入提交信息: Fix login authentication bug
# 结果: v0.0.8 → v0.0.9
```

### 场景 2: 添加新功能
```bash
# 添加了新的 API 接口
make tag-minor
# 输入提交信息: Add new jail management API
# 结果: v0.0.8 → v0.1.0
```

### 场景 3: 重大更新
```bash
# 重构了整个前端架构
make tag-major
# 输入提交信息: Rebuild frontend with Next.js 15
# 结果: v0.0.8 → v1.0.0
```

### 场景 4: 快速发布
```bash
# 紧急修复，不想输入提交信息
make release
# 自动创建: v0.0.9，提交信息: Release v0.0.9
```

---

## 🔍 命令执行流程

以 `make tag-push` 为例：

```
1. 获取当前最新标签（v0.0.8）
   ↓
2. 解析版本号（0.0.8）
   ↓
3. 递增补丁版本号（0.0.9）
   ↓
4. 询问用户输入提交信息
   ↓
5. 创建带注释的标签（git tag -a v0.0.9 -m "..."）
   ↓
6. 推送标签到远程仓库（git push origin v0.0.9）
   ↓
7. 显示成功信息 ✅
```

---

## 📦 实际示例

### 示例 1: 使用 `make tag-push`

```bash
$ make tag-push
获取当前最新版本...
当前版本: v0.0.8
新版本: v0.0.9
请输入提交信息 (默认: Release v0.0.9): Fix web template API response parsing
创建标签: v0.0.9
推送标签到远程仓库...
Enumerating objects: 1, done.
Counting objects: 100% (1/1), done.
Writing objects: 100% (1/1), 186 bytes | 186.00 KiB/s, done.
Total 1 (delta 0), reused 0 (delta 0), pack-reused 0
To github.com:difyz9/fail2ban_web.git
 * [new tag]         v0.0.9 -> v0.0.9
✅ 版本 v0.0.9 已成功推送!
```

### 示例 2: 使用 `make release`

```bash
$ make release
🚀 快速发布新版本...
当前版本: v0.0.9
新版本: v0.0.10
To github.com:difyz9/fail2ban_web.git
 * [new tag]         v0.0.10 -> v0.0.10
✅ 版本 v0.0.10 已成功推送!
```

### 示例 3: 查看版本历史

```bash
$ make list-tags
所有版本标签：
v0.0.10
v0.0.9
v0.0.8
v0.0.7
v0.0.6
v0.0.5
v0.0.4
v0.0.3
v0.0.2
v0.0.1
```

---

## 🛠️ 高级操作

### 删除错误的标签

```bash
# 删除本地标签
make delete-tag
# 输入: v0.0.9

# 删除远程标签
make delete-tag-remote
# 输入: v0.0.9
```

### 手动创建标签（不推荐）

```bash
git tag -a v0.0.9 -m "Release v0.0.9"
git push origin v0.0.9
```

---

## ⚠️ 注意事项

1. **推送前确认代码已提交**
   ```bash
   git status  # 确保没有未提交的更改
   ```

2. **标签创建后无法修改**
   - 如果创建错了，需要先删除再重新创建

3. **版本号不可回退**
   - 一旦推送到远程，建议不要删除
   - 如有问题，继续创建新版本

4. **提交信息建议**
   - 简洁明了
   - 使用英文（可选）
   - 描述主要变更内容

---

## 🔄 完整发布流程

### 推荐工作流

```bash
# 1. 开发并测试功能
git add .
git commit -m "Add new feature"

# 2. 推送代码到主分支
git push origin main

# 3. 创建并推送版本标签
make tag-push  # 或 make release

# 4. 验证标签已推送
make list-tags

# 5. 在 GitHub 上创建 Release（可选）
# 访问: https://github.com/difyz9/fail2ban_web/releases
```

---

## 📊 版本号建议

### 当前阶段（v0.x.x）

- **v0.0.x**: Bug 修复、小改进
- **v0.x.0**: 新功能、重要更新
- **v1.0.0**: 第一个稳定版本

### 稳定后（v1.x.x+）

- **vX.0.0**: 重大架构变更、破坏性更新
- **v1.X.0**: 新功能、兼容性更新
- **v1.0.X**: Bug 修复、补丁

---

## 🎯 最佳实践

1. **小步快跑**: 频繁发布小版本，而不是累积大版本
2. **语义化版本**: 严格遵循版本号含义
3. **变更日志**: 在 Release 中详细说明变更内容
4. **测试先行**: 打标签前确保测试通过
5. **文档同步**: 重要版本需更新 README 和文档

---

## 🆘 常见问题

### Q: 如何查看当前版本？
```bash
make get-version
# 或
git describe --tags
```

### Q: 创建标签时忘记输入提交信息怎么办？
A: 默认会使用 "Release vX.X.X" 作为提交信息

### Q: 可以跳过某个版本号吗？
A: 可以，但不推荐。保持版本连续性便于追踪

### Q: 如何回退到某个版本？
```bash
git checkout v0.0.8  # 切换到指定版本
```

### Q: 标签推送失败怎么办？
A: 检查网络连接和 Git 权限，确认标签在本地和远程都不存在

---

## 📚 相关资源

- [语义化版本 2.0.0](https://semver.org/lang/zh-CN/)
- [Git 标签文档](https://git-scm.com/book/zh/v2/Git-基础-打标签)
- [GitHub Releases](https://docs.github.com/en/repositories/releasing-projects-on-github)

---

**当前版本**: v0.0.8  
**最后更新**: 2024-01-15
