package core

import (
	"bytes"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

// NewDefaultConfig 创建默认配置
func NewDefaultConfig() *AppConfig {
	return &AppConfig{
		Listen:    ":8099",
		StaticDir: "./web/static",
		StaticURL: "http://localhost:8099/static",
		DBPath:    "fail2ban_web.db",
		Admin: AdminConfig{
			Username: "admin",
			Password: "admin123",
			Email:    "admin@example.com",
		},
	}
}

// LoadConfig 加载配置文件
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

// SaveConfig 保存配置到文件
func SaveConfig(config *AppConfig) error {
	buf := new(bytes.Buffer)
	encoder := toml.NewEncoder(buf)
	if err := encoder.Encode(config); err != nil {
		return err
	}
	
	return os.WriteFile(config.Path, buf.Bytes(), 0644)
}
