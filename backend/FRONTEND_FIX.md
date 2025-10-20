# 前端访问问题修复说明

## 问题描述

访问 `http://localhost:8099` 时页面一直显示"正在跳转..."，无法正常使用。

## 问题原因

1. **首页自动跳转**：`frontend/src/app/page.tsx` 设置了自动跳转到 `/login`
2. **AuthContext loading 卡住**：在静态导出模式下，React 的 `useEffect` 执行时机问题导致 `loading` 状态可能一直为 `true`
3. **AuthGuard 阻塞**：当 `loading` 为 `true` 时，`AuthGuard` 显示"正在验证身份..."，阻止了登录表单的显示

## 已修复内容

### 1. 修改首页跳转逻辑

**文件**: `frontend/src/app/page.tsx`

```tsx
// 从 Next.js Router 跳转改为浏览器原生跳转
if (typeof window !== 'undefined') {
  window.location.href = '/login.html';
}
```

### 2. 添加 AuthContext 超时保护

**文件**: `frontend/src/contexts/AuthContext.tsx`

```tsx
// 添加3秒超时保护，确保 loading 状态不会永久卡住
const timeout = setTimeout(() => {
  setLoading(false);
}, 3000);
```

## 访问方式

### 方式 1：直接访问登录页（推荐）

```
http://localhost:8099/login
```

或

```
http://localhost:8099/login.html
```

### 方式 2：等待首页自动跳转

访问 `http://localhost:8099/`，会自动跳转到登录页（可能需要等待几秒）。

### 方式 3：访问 Dashboard（需要先登录）

```
http://localhost:8099/dashboard
```

## 默认登录凭据

根据 `config.toml` 配置：

- **用户名**: `admin`
- **密码**: `admin123`
- **邮箱**: `admin@example.com`

## 测试步骤

1. **启动应用**
   ```bash
   ./build/fail2ban-web
   ```

2. **访问登录页**
   ```
   http://localhost:8099/login
   ```

3. **输入凭据登录**
   - 用户名: `admin`
   - 密码: `admin123`

4. **登录成功后自动跳转到 Dashboard**

## 常见问题

### Q1: 页面还是显示"正在跳转"

**解决方案**：
- 清除浏览器缓存（Ctrl+F5 或 Cmd+Shift+R）
- 直接访问 `/login.html` 而不是 `/login`
- 等待 3-5 秒让 JavaScript 加载完成

### Q2: 登录后显示"正在验证身份"

**原因**：AuthContext 正在调用 `/api/v1/auth/profile` 接口

**解决方案**：
- 等待 3 秒超时保护生效
- 检查浏览器控制台是否有网络错误
- 确认后端服务正常运行

### Q3: API 返回 503 错误

**原因**：Fail2Ban 服务未运行（这是正常的，在开发环境中可以忽略）

**影响**：不影响登录和基本功能，只是某些统计数据无法显示

### Q4: 静态资源加载失败

**检查项**：
1. 确认已运行 `make build-full`
2. 查看 `web/` 目录是否包含 `_next/static/` 文件夹
3. 检查浏览器控制台的 404 错误

## 验证修复

运行以下命令验证：

```bash
# 1. 测试登录页面
curl -I http://localhost:8099/login.html

# 应该返回 HTTP/1.1 200 OK

# 2. 测试 API
curl http://localhost:8099/api/v1/auth/profile

# 未登录应该返回 401 Unauthorized（这是正确的）
```

## 下一步优化建议

1. **移除静态导出模式**：将 `next.config.ts` 中的 `output: 'export'` 移除，使用 Next.js 服务端渲染
2. **优化 AuthContext**：减少不必要的 API 调用
3. **添加错误边界**：捕获 React 渲染错误
4. **改进路由策略**：使用 Next.js 的 middleware 处理认证

## 更新记录

- **2025-10-19**
  - ✅ 修复首页跳转问题
  - ✅ 添加 AuthContext 超时保护
  - ✅ 重新构建前端和后端
  - ✅ 验证登录流程

---

**注意**：如果问题仍然存在，请提供浏览器控制台的错误信息和网络请求详情。
