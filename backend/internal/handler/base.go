package handler

import (
	"fail2ban-web/core"

	"gorm.io/gorm"
)

// BaseHandler 基础 Handler，所有 Handler 都应该包含这个
type BaseHandler struct {
	App *core.AppServer
	DB  *gorm.DB
}
