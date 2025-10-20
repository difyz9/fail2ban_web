# Web 模板更新说明

## 更新日期
2024年

## 更新目的
将 web 目录下的传统模板文件（login.html 和 app.js）与后端的统一 API 响应格式同步。

## 后端统一响应格式
```json
{
  "success": true,
  "data": { ... },
  "error": "",
  "message": "操作成功"
}
```

## 主要更新内容

### 1. 登录页面 (web/templates/login.html)

#### 更新前
```javascript
const data = await response.json();
if (response.ok) {
    localStorage.setItem('token', data.token);
    localStorage.setItem('username', data.username);
}
```

#### 更新后
```javascript
const result = await response.json();
if (response.ok && result.success) {
    const { token, user, expires_at } = result.data;
    localStorage.setItem('token', token);
    localStorage.setItem('username', user.username);
    localStorage.setItem('user_role', user.role);
    localStorage.setItem('token_expires_at', expires_at);
}
```

**改进点：**
- ✅ 检查 `result.success` 状态
- ✅ 从 `result.data` 中解构数据
- ✅ 存储更多用户信息（role, expires_at）
- ✅ 使用 `result.error` 或 `result.message` 显示错误信息

---

### 2. 应用主脚本 (web/static/js/app.js)

#### 新增统一响应处理方法

```javascript
// 处理统一响应格式
async handleApiResponse(response) {
    if (!response || !response.ok) {
        throw new Error('API请求失败');
    }
    
    const result = await response.json();
    
    // 统一响应格式: { success: true, data: {...}, error: "", message: "" }
    if (result.success) {
        return result.data;
    } else {
        throw new Error(result.error || result.message || 'API请求失败');
    }
}
```

**作用：**
- 统一处理所有 API 响应
- 自动提取 `data` 字段
- 统一错误处理逻辑

---

#### 更新的方法列表

| 序号 | 方法名 | 功能 | 改进点 |
|------|--------|------|--------|
| 1 | `loadStats()` | 加载统计数据 | 使用 `handleApiResponse` 提取数据 |
| 2 | `loadRecentBans()` | 加载最近被禁IP | 使用 `handleApiResponse` 提取数据 |
| 3 | `loadSystemInfo()` | 加载系统信息 | 使用 `handleApiResponse` 提取数据 |
| 4 | `refreshBannedIPs()` | 刷新被禁IP列表 | 使用 `handleApiResponse` 提取数据 |
| 5 | `unbanIP()` | 解禁IP地址 | 使用 `handleApiResponse` 处理结果 |
| 6 | `refreshRules()` | 刷新规则列表 | 使用 `handleApiResponse` 提取数据 |
| 7 | `installNginxDefaults()` | 安装Nginx默认配置 | 使用 `handleApiResponse` 处理结果 |
| 8 | `showDefaultConfigInfo()` | 显示默认配置信息 | 使用 `handleApiResponse` 提取数据 |
| 9 | `exportNginxConfig()` | 导出Nginx配置 | 使用 `handleApiResponse` 提取数据 |
| 10 | `toggleJail()` | 切换jail状态 | 使用 `handleApiResponse` 处理结果 |

---

## 更新模式

### 更新前的典型代码
```javascript
const response = await this.authenticatedFetch(`${this.baseURL}/endpoint`);
if (!response || !response.ok) {
    throw new Error('请求失败');
}
const data = await response.json();
// 直接使用 data
```

### 更新后的典型代码
```javascript
const response = await this.authenticatedFetch(`${this.baseURL}/endpoint`);
const data = await this.handleApiResponse(response);
// 直接使用 data（已从 result.data 中提取）
```

---

## 兼容性说明

### Token 存储键名
- **Next.js 前端**: 使用 `'auth_token'`
- **传统 Web 模板**: 使用 `'token'`

两者互不干扰，可以共存。

### API 端点
所有端点保持不变：
- `/api/v1/auth/login` - 登录
- `/api/v1/stats` - 统计数据
- `/api/v1/banned-ips` - 被禁IP列表
- `/api/v1/jails` - 规则列表
- 等等...

---

## 测试建议

### 1. 登录流程测试
```bash
# 测试登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your_password"}'
```

预期响应：
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "username": "admin",
      "role": "admin"
    },
    "expires_at": "2024-01-01T12:00:00Z"
  },
  "message": "登录成功"
}
```

### 2. 仪表板数据测试
```bash
# 测试统计数据（需要 token）
curl http://localhost:8080/api/v1/stats \
  -H "Authorization: Bearer YOUR_TOKEN"
```

预期响应：
```json
{
  "success": true,
  "data": {
    "bannedCount": 42,
    "todayBlocks": 15,
    "activeRules": 10,
    "systemStatus": "正常"
  },
  "message": ""
}
```

---

## 错误处理

### 统一错误格式
```json
{
  "success": false,
  "data": null,
  "error": "详细错误信息",
  "message": "用户友好的错误提示"
}
```

### 前端处理逻辑
```javascript
try {
    const data = await this.handleApiResponse(response);
    // 处理成功数据
} catch (error) {
    // error.message 包含 result.error 或 result.message
    this.showError(`操作失败: ${error.message}`);
}
```

---

## 向后兼容性

✅ **完全兼容**
- 所有旧的 API 端点保持不变
- 只是响应格式增加了包装层
- 前端通过 `handleApiResponse` 自动解包

---

## 维护建议

### 1. 新增 API 调用
当添加新的 API 调用时，请遵循以下模式：

```javascript
async newApiMethod() {
    try {
        const response = await this.authenticatedFetch(`${this.baseURL}/new-endpoint`);
        const data = await this.handleApiResponse(response);
        // 使用 data
    } catch (error) {
        console.error('操作失败:', error);
        this.showError(`操作失败: ${error.message}`);
    }
}
```

### 2. 错误日志记录
所有方法都保留了 `console.error` 调用以便调试：
```javascript
console.error('具体操作失败:', error);
```

### 3. 用户友好提示
使用 `this.showError()` 和 `this.showSuccess()` 显示操作结果

---

## 总结

本次更新确保了 web 目录下的传统模板与后端统一响应格式完全同步，提升了：

1. **代码一致性** - 所有 API 调用使用统一的响应处理
2. **错误处理** - 更清晰的错误信息传递
3. **可维护性** - 集中式响应处理逻辑
4. **用户体验** - 更完整的用户信息存储和错误提示

所有更改都是向后兼容的，不会影响现有功能。
