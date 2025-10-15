package app

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger 创建日志记录器
func NewLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	
	return logger, nil
}

// LoggerModule 日志模块
var LoggerModule = fx.Module("logger",
	fx.Provide(NewLogger),
)
