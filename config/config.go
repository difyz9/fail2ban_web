package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Fail2Ban Fail2BanConfig
	Admin    AdminConfig
}

type ServerConfig struct {
	Port string
	Host string
	Mode string
}

type DatabaseConfig struct {
	Path string
}

type JWTConfig struct {
	Secret     string
	ExpireTime int // 小时
}

type Fail2BanConfig struct {
	LogPath        string
	ConfigPath     string
	SocketPath     string
	NginxAccessLog string
	NginxErrorLog  string
	SSHLogPath     string
	ForceSudo      bool   // 强制使用sudo
	SudoUser       string // sudo用户
	DevMode        bool   // 开发模式
}

type AdminConfig struct {
	Username string
	Password string
	Email    string
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8092"),
			Host: getEnv("HOST", "0.0.0.0"),
			Mode: getEnv("GIN_MODE", "release"),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "./fail2ban_web.db"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
			ExpireTime: getEnvAsInt("JWT_EXPIRE_TIME", 24),
		},
		Fail2Ban: Fail2BanConfig{
			LogPath:        getEnv("FAIL2BAN_LOG_PATH", "/var/log/fail2ban.log"),
			ConfigPath:     getEnv("FAIL2BAN_CONFIG_PATH", "/etc/fail2ban"),
			SocketPath:     getEnv("FAIL2BAN_SOCKET_PATH", "/var/run/fail2ban/fail2ban.sock"),
			NginxAccessLog: getEnv("NGINX_ACCESS_LOG", "/var/log/nginx/access.log"),
			NginxErrorLog:  getEnv("NGINX_ERROR_LOG", "/var/log/nginx/error.log"),
			SSHLogPath:     getEnv("SSH_LOG_PATH", "/var/log/auth.log"),
			ForceSudo:      getEnvAsBool("FAIL2BAN_FORCE_SUDO", false),
			SudoUser:       getEnv("SUDO_USER", ""),
			DevMode:        getEnvAsBool("DEV_MODE", false),
		},
		Admin: AdminConfig{
			Username: getEnv("ADMIN_USERNAME", "admin"),
			Password: getEnv("ADMIN_PASSWORD", "admin123"),
			Email:    getEnv("ADMIN_EMAIL", "admin@fail2ban.local"),
		},
	}
}

// getEnv 获取环境变量，如果不存在则使用默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量作为整数，如果不存在或转换失败则使用默认值
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量作为布尔值，如果不存在或转换失败则使用默认值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}