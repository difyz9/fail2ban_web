package app

import (
	"fail2ban-web/config"

	"go.uber.org/fx"
)

// NewConfig 创建配置
func NewConfig() *config.Config {
	return config.LoadConfig()
}

// ConfigModule 配置模块
var ConfigModule = fx.Module("config",
	fx.Provide(NewConfig),
)
