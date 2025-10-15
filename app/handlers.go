package app

import (
	"fail2ban-web/config"
	"fail2ban-web/internal/handler"
	"fail2ban-web/internal/service"

	"go.uber.org/fx"
)

// HandlerParams Handler 依赖参数
type HandlerParams struct {
	fx.In
	Config                       *config.Config
	Fail2banService              *service.Fail2BanService
	JailService                  *service.JailService
	SSHService                   *service.SSHService
	NginxService                 *service.NginxService
	DefaultSSHService            *service.DefaultSSHService
	DefaultNginxService          *service.DefaultNginxService
	DefaultNginxAdvancedService  *service.DefaultNginxAdvancedService
	IntelligentService           *service.IntelligentScanService
}

// HandlerResult Handler 输出
type HandlerResult struct {
	fx.Out
	AuthHandler          *handler.AuthHandler
	Fail2banHandler      *handler.Fail2BanHandler
	JailHandler          *handler.JailHandler
	DefaultConfigHandler *handler.DefaultConfigHandler
	SSHHandler           *handler.SSHHandler
	NginxHandler         *handler.NginxHandler
	IntelligentHandler   *handler.IntelligentHandler
}

// NewHandlers 创建所有 handlers
func NewHandlers(params HandlerParams) HandlerResult {
	return HandlerResult{
		AuthHandler:          handler.NewAuthHandler(params.Config),
		Fail2banHandler:      handler.NewFail2BanHandler(params.Fail2banService),
		JailHandler:          handler.NewJailHandler(params.JailService),
		DefaultConfigHandler: handler.NewDefaultConfigHandler(),
		SSHHandler:           handler.NewSSHHandler(params.SSHService, params.DefaultSSHService),
		NginxHandler:         handler.NewNginxHandler(params.NginxService, params.DefaultNginxService, params.DefaultNginxAdvancedService),
		IntelligentHandler:   handler.NewIntelligentHandler(params.IntelligentService),
	}
}

// HandlerModule Handler 模块
var HandlerModule = fx.Module("handlers",
	fx.Provide(NewHandlers),
)
