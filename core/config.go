package core

import (
	"log"
	"os"
)

// LoadConfig 加载配置
func LoadConfig() *AppConfig {
	listen := os.Getenv("LISTEN_ADDR")
	if listen == "" {
		listen = ":8092"
	}
	
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "fail2ban_web.db"
	}
	
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./web/static"
	}
	
	staticURL := os.Getenv("STATIC_URL")
	if staticURL == "" {
		staticURL = "http://localhost:8092/static"
	}
	
	config := &AppConfig{
		Listen:    listen,
		DBPath:    dbPath,
		StaticDir: staticDir,
		StaticURL: staticURL,
	}
	
	log.Printf("Loaded configuration: %+v", config)
	return config
}
