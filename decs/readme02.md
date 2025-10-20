# Fail2Ban Web Panel - 快速部署指南

## 🚀 一键部署 (Ubuntu/Debian)

在您的Ubuntu服务器上执行以下命令：

```bash
# 1. 克隆或上传项目到服务器
cd ~/
git clone <your-repo-url> fail2ban_web
cd fail2ban_web

# 2. 运行自动部署脚本
./deploy.sh
```

## 📋 部署脚本会自动完成：

✅ 检查系统依赖 (Go, Fail2Ban)  
✅ 创建数据目录  
✅ 配置sudo权限  
✅ 编译应用程序  
✅ 生成安全配置  
✅ 创建系统服务  
✅ 启动Web面板  

## 🔐 默认登录信息

部署完成后，脚本会显示：
- **用户名**: admin
- **密码**: (自动生成的安全密码)
- **访问地址**: http://your-server-ip:8099

## 🛠️ 管理命令

```bash
# 查看服务状态
sudo systemctl status fail2ban-web

# 查看实时日志
sudo journalctl -u fail2ban-web -f

# 重启服务
sudo systemctl restart fail2ban-web

# 停止服务
sudo systemctl stop fail2ban-web
```

## 🔧 权限问题解决

如果遇到权限错误：

```bash
# 检查sudo配置
sudo cat /etc/sudoers.d/fail2ban-web

# 测试权限
sudo fail2ban-client status

# 查看详细错误
sudo journalctl -u fail2ban-web -n 50
```

## 🌐 防火墙配置

```bash
# 允许Web面板端口
sudo ufw allow 8099

# 或使用Nginx反向代理 (推荐生产环境)
sudo ufw allow 'Nginx Full'
```

## 📈 功能特性

- ✅ **实时监控**: 自动扫描SSH/Nginx日志
- ✅ **智能分析**: 威胁评分和自动封禁
- ✅ **Web管理**: 直观的管理界面
- ✅ **权限安全**: 支持sudo权限管理
- ✅ **配置模板**: 10个预置安全规则
- ✅ **REST API**: 完整的API接口

## 🔗 相关链接

- 详细部署文档: [DEPLOYMENT.md](./DEPLOYMENT.md)
- 项目源码: [GitHub Repository](your-repo-url)

---

**需要帮助？** 请查看详细的 [DEPLOYMENT.md](./DEPLOYMENT.md) 文档或提交Issue。


要查看fail2ban-client封禁的 IP，可以使用以下命令：
bash
sudo fail2ban-client status <jail名称>
例如，要查看与 SSH 相关的封禁 IP，可使用命令：
bash
sudo fail2ban-client status sshd
执行该命令后，会显示相关jail的状态信息，其中在Actions部分的Banned IP list中会列出当前被封禁的 IP 地址。
