## 部署流程

### 1. 配置 GitHub Secrets

在 GitHub 仓库的 Settings -> Environments -> dev 环境中配置：

```
SERVER_HOST=你的服务器IP
SERVER_PORT=22
SERVER_USER=你的用户名
SERVER_PASSWORD=你的密码
```

### 2. 推送代码触发部署

```bash
git add .
git commit -m "deploy with supervisor"
git push origin main
```

### 3. 自动部署过程

GitHub Actions 会自动：
1. 运行测试
2. 编译 Go 二进制文件
3. 打包部署文件
4. 传输到服务器
5. 使用 Supervisor 部署和启动服务
