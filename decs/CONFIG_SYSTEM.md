# 配置系统重构说明

## 概述

项目配置系统已完全参考 [geekai](https://github.com/yangjian102621/geekai/blob/main/api/core/config.go) 的实现方式进行重构，使用 TOML 格式的配置文件。

## 配置系统架构

### 1. 配置文件结构

**config.toml** - TOML 格式的配置文件

```toml
Listen = ":8092"
StaticDir = "./web/static"
StaticURL = "http://localhost:8092/static"
DBPath = "fail2ban_web.db"

[Admin]
  Username = "admin"
  Password = "admin123"
  Email = "admin@example.com"
```

### 2. 核心类型定义 (core/types.go)

```go
// AppConfig 应用配置
type AppConfig struct {
	Path      string `toml:"-"` // 配置文件路径，不保存到 TOML
	Listen    string // 监听地址，如 ":8092" 或 "0.0.0.0:8092"
	StaticDir string // 静态资源目录
	StaticURL string // 静态资源 URL
	DBPath    string // 数据库文件路径
	Admin     AdminConfig
}

// AdminConfig 管理员配置
type AdminConfig struct {
	Username string
	Password string
	Email    string
}
```

### 3. 配置加载逻辑 (core/config.go)

参考 geekai 的三个核心函数：

#### NewDefaultConfig()
创建默认配置，当配置文件不存在时使用：

```go
func NewDefaultConfig() *AppConfig {
	return &AppConfig{
		Listen:    ":8092",
		StaticDir: "./web/static",
		StaticURL: "http://localhost:8092/static",
		DBPath:    "fail2ban_web.db",
		Admin: AdminConfig{
			Username: "admin",
			Password: "admin123",
			Email:    "admin@example.com",
		},
	}
}
```

#### LoadConfig(configFile string)
加载配置文件，如果文件不存在则创建默认配置：

```go
func LoadConfig(configFile string) (*AppConfig, error) {
	var config *AppConfig
	_, err := os.Stat(configFile)
	if err != nil {
		// 配置文件不存在，创建默认配置
		log.Println("Creating new config file:", configFile)
		config = NewDefaultConfig()
		config.Path = configFile
		
		// 保存默认配置到文件
		err := SaveConfig(config)
		if err != nil {
			return nil, err
		}
		
		return config, nil
	}
	
	// 读取配置文件
	_, err = toml.DecodeFile(configFile, &config)
	if err != nil {
		return nil, err
	}
	
	config.Path = configFile
	return config, nil
}
```

#### SaveConfig(config *AppConfig)
保存配置到文件：

```go
func SaveConfig(config *AppConfig) error {
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	if err := encoder.Encode(config); err != nil {
		return err
	}
	
	return os.WriteFile(config.Path, buf.Bytes(), 0644)
}
```

### 4. main.go 中的配置加载

```go
func main() {
	// 从环境变量读取配置文件路径
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config.toml"
	}
	log.Println("Loading config file:", configFile)
	
	app := fx.New(
		// 提供核心配置
		fx.Provide(func() (*core.AppConfig, error) {
			return core.LoadConfig(configFile)
		}),
		// ...
	)
}
```

## 使用方式

### 1. 首次运行

首次运行时，如果 `config.toml` 不存在，程序会自动创建默认配置：

```bash
$ ./fail2ban_web
2025/10/19 17:31:17 Loading config file: config.toml
2025/10/19 17:31:17 Creating new config file: config.toml
```

### 2. 修改配置

直接编辑 `config.toml` 文件：

```toml
Listen = ":8080"  # 修改监听端口
StaticDir = "./static"
StaticURL = "http://localhost:8080/static"
DBPath = "my_database.db"

[Admin]
  Username = "myadmin"       # 修改管理员用户名
  Password = "mypassword"    # 修改管理员密码
  Email = "admin@mydomain.com"
```

### 3. 使用自定义配置文件

通过环境变量指定配置文件路径：

```bash
# 使用自定义配置文件
export CONFIG_FILE="my_config.toml"
./fail2ban_web

# 或者一行命令
CONFIG_FILE="production.toml" ./fail2ban_web
```

### 4. 在 Handler 中使用配置

所有 Handler 都通过 `h.App.Config` 访问配置：

```go
type AuthHandler struct {
	BaseHandler  // 包含 App *core.AppServer
}

func (h *AuthHandler) Login(c *gin.Context) {
	// 访问管理员配置
	username := h.App.Config.Admin.Username
	password := h.App.Config.Admin.Password
	email := h.App.Config.Admin.Email
	
	// 访问服务器配置
	listen := h.App.Config.Listen
	dbPath := h.App.Config.DBPath
}
```

## 与 geekai 的对比

### 相同点

1. **TOML 格式** - 使用 TOML 格式的配置文件
2. **三函数模式** - NewDefaultConfig(), LoadConfig(), SaveConfig()
3. **自动创建** - 配置文件不存在时自动创建默认配置
4. **环境变量支持** - 通过 `CONFIG_FILE` 环境变量指定配置文件
5. **Path 字段** - 配置结构中包含 Path 字段（`toml:"-"` 标记不序列化）

### 不同点

1. **配置项简化** - 我们只包含必需的配置项（Listen, StaticDir, StaticURL, DBPath, Admin）
2. **数据库类型** - geekai 使用 MySQL，我们使用 SQLite
3. **无 Redis** - 我们暂不需要 Redis 配置
4. **无支付系统** - 我们不需要支付相关配置

## 配置项说明

### 基础配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| Listen | string | ":8092" | HTTP 服务器监听地址 |
| StaticDir | string | "./web/static" | 静态资源目录路径 |
| StaticURL | string | "http://localhost:8092/static" | 静态资源访问 URL |
| DBPath | string | "fail2ban_web.db" | SQLite 数据库文件路径 |

### 管理员配置 [Admin]

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| Username | string | "admin" | 管理员用户名 |
| Password | string | "admin123" | 管理员密码 |
| Email | string | "admin@example.com" | 管理员邮箱 |

## 安全建议

1. **修改默认密码**
   ```toml
   [Admin]
     Password = "your-strong-password-here"
   ```

2. **配置文件权限**
   ```bash
   chmod 600 config.toml  # 只有所有者可读写
   ```

3. **不要提交配置文件**
   确保 `.gitignore` 包含：
   ```
   config.toml
   *.db
   ```

4. **使用环境变量**
   对于敏感信息，可以考虑使用环境变量替代配置文件

## 依赖包

需要安装 TOML 解析库：

```bash
go get github.com/BurntSushi/toml
```

在 `go.mod` 中：

```go
require (
	github.com/BurntSushi/toml v1.5.0
	// ...
)
```

## 示例配置

### 开发环境配置 (config.toml)

```toml
Listen = ":8092"
StaticDir = "./web/static"
StaticURL = "http://localhost:8092/static"
DBPath = "fail2ban_web.db"

[Admin]
  Username = "admin"
  Password = "admin123"
  Email = "admin@localhost"
```

### 生产环境配置 (production.toml)

```toml
Listen = "0.0.0.0:80"
StaticDir = "/var/www/fail2ban/static"
StaticURL = "https://fail2ban.example.com/static"
DBPath = "/var/lib/fail2ban/fail2ban_web.db"

[Admin]
  Username = "administrator"
  Password = "very-strong-password-here"
  Email = "admin@example.com"
```

## 配置更新流程

1. **停止服务**
   ```bash
   pkill fail2ban_web
   ```

2. **修改配置文件**
   ```bash
   vi config.toml
   ```

3. **重启服务**
   ```bash
   ./fail2ban_web
   ```

## 配置热重载

目前配置不支持热重载，修改配置后需要重启应用。未来可以考虑添加配置热重载功能。

## 故障排除

### 配置文件损坏

如果配置文件格式错误：

```bash
# 删除损坏的配置文件
rm config.toml

# 重新运行程序，它会创建新的默认配置
./fail2ban_web
```

### 权限问题

```bash
# 确保配置文件可读
chmod 644 config.toml

# 确保程序有写入权限（用于创建配置文件）
chmod +w .
```

### 无法加载配置

检查日志输出：
```
2025/10/19 17:31:17 Loading config file: config.toml
2025/10/19 17:31:17 Creating new config file: config.toml
```

## 参考资料

- [geekai 配置系统](https://github.com/yangjian102621/geekai/blob/main/api/core/config.go)
- [TOML 规范](https://toml.io/)
- [github.com/BurntSushi/toml](https://github.com/BurntSushi/toml)

---

**最后更新：** 2025-10-19  
**配置版本：** v1.0  
**参考项目：** geekai
