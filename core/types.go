package core

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

// SystemConfig 系统配置（可以从数据库加载）
type SystemConfig struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}
