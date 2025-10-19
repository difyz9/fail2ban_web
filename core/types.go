package core

// AppConfig 应用配置
type AppConfig struct {
	Listen    string
	StaticDir string
	StaticURL string
	DBPath    string
}

// SystemConfig 系统配置（可以从数据库加载）
type SystemConfig struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}
