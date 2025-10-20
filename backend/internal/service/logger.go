package service

import (
	"github.com/sirupsen/logrus"
)

// NewLogrusLogger 创建一个 logrus logger 实例
func NewLogrusLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	return logger
}
