# 登录流程测试文档

## 后端返回数据结构

### 1. 登录接口 POST /api/v1/auth/login

**请求：**
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@fail2ban.local",
      "role": "admin",
      "is_active": true,
      "created_at": "0001-01-01T00:00:00Z",
      "updated_at": "0001-01-01T00:00:00Z"
    },
    "expires_at": 1760573452
  },
  "message": "Login successful"
}
```

### 2. 获取用户信息 GET /api/v1/auth/profile

**响应：**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@fail2ban.local",
    "role": "admin",
    "is_active": true,
    "created_at": "0001-01-01T00:00:00Z",
    "updated_at": "0001-01-01T00:00:00Z"
  },
  "message": "Profile retrieved successfully"
}
```

### 3. 刷新Token POST /api/v1/auth/refresh

**响应：**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1760573452
  },
  "message": "Token refreshed successfully"
}
```

## 前端数据处理流程

### 1. apiClient 自动解包

`apiClient.post<LoginResponse>()` 会自动：
1. 检查 `response.data.success` 是否为 `true`
2. 如果成功，返回 `response.data.data`（即 LoginResponse 对象）
3. 如果失败，抛出错误

### 2. authService.login()

```typescript
const response = await apiClient.post<LoginResponse>('/auth/login', credentials);
// response 直接就是 LoginResponse 类型：
// {
//   token: string,
//   user: User,
//   expires_at: number
// }

this.saveAuthData({
  token: response.token,
  user: response.user,
});
```

### 3. Cookie 存储

- `auth_token`: 存储 JWT token
- `user_info`: 存储用户信息的 JSON 字符串

### 4. AuthContext 初始化

- 从 Cookie 读取 `auth_token`
- 如果存在，调用 `/auth/profile` 获取最新用户信息
- 设置到 React Context 中

## 测试步骤

1. **启动后端**：
   ```bash
   cd /Users/apple/opt/difyz10/1014/fail2ban_web
   go run main.go
   ```

2. **启动前端**：
   ```bash
   cd /Users/apple/opt/difyz10/1014/fail2ban_web/frontend
   npm run dev
   ```

3. **测试登录**：
   - 访问 http://localhost:3000/login
   - 输入 admin/admin123
   - 点击登录
   - 应该自动跳转到 /dashboard

4. **验证 Cookie**：
   - 打开浏览器开发者工具
   - 检查 Application > Cookies
   - 应该看到 `auth_token` 和 `user_info`

5. **验证自动刷新**：
   - 等待 15 分钟或修改代码缩短刷新间隔
   - 检查 Network 标签，应该看到 /auth/refresh 请求

## 已修复的问题

✅ AuthContext 中的 Cookie key 从 `'token'` 改为 `'auth_token'`
✅ RefreshToken 响应类型包含 `expires_at` 字段
✅ apiClient 正确处理统一的 ApiResponse 格式
✅ authService 正确解析 LoginResponse 数据
