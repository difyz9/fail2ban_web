package app

import (
	"context"
	"fail2ban-web/config"
	"fail2ban-web/internal/service"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ServiceParams 服务层依赖参数
type ServiceParams struct {
	fx.In
	Config *config.Config
	DB     *gorm.DB
	Logger *zap.Logger
}

// ServiceResult 服务层输出
type ServiceResult struct {
	fx.Out
	Fail2banService              *service.Fail2BanService
	JailService                  *service.JailService
	SSHService                   *service.SSHService
	NginxService                 *service.NginxService
	DefaultSSHService            *service.DefaultSSHService
	DefaultNginxService          *service.DefaultNginxService
	DefaultNginxAdvancedService  *service.DefaultNginxAdvancedService
	IntelligentService           *service.IntelligentScanService
	DefaultJailService           *service.DefaultJailService
}

// NewServices 创建所有服务
func NewServices(lc fx.Lifecycle, params ServiceParams) ServiceResult {
	// 将 zap.Logger 转换为 logrus 兼容的 logger
	// 注意：这里暂时使用 zap 的 SugaredLogger 来模拟 logrus
	// 更好的做法是重构 service 层使用 zap.Logger
	
	// 初始化服务
	jailService := service.NewJailService(params.DB)
	sshService := service.NewSSHService(params.Config, params.DB)
	nginxService := service.NewNginxService(params.Config, params.DB)
	defaultSSHService := service.NewDefaultSSHService(jailService)
	defaultNginxService := service.NewDefaultNginxServiceWithJail(jailService)
	defaultNginxAdvancedService := service.NewDefaultNginxAdvancedService(jailService)
	defaultJailService := service.NewDefaultJailService(jailService)
	
	// Fail2BanService 需要 logrus.Logger，这里需要适配
	// 临时创建一个 logrus logger
	logrusLogger := service.NewLogrusLogger()
	fail2banService := service.NewFail2BanService(logrusLogger)
	
	// 初始化智能扫描服务
	intelligentService := service.NewIntelligentScanService(
		params.Config,
		params.DB,
		sshService,
		nginxService,
		jailService,
		fail2banService,
	)
	
	// 添加生命周期钩子
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.Logger.Info("Starting intelligent scan service...")
			intelligentService.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info("Stopping intelligent scan service...")
			intelligentService.Stop()
			return nil
		},
	})

	return ServiceResult{
		Fail2banService:             fail2banService,
		JailService:                 jailService,
		SSHService:                  sshService,
		NginxService:                nginxService,
		DefaultSSHService:           defaultSSHService,
		DefaultNginxService:         defaultNginxService,
		DefaultNginxAdvancedService: defaultNginxAdvancedService,
		IntelligentService:          intelligentService,
		DefaultJailService:          defaultJailService,
	}
}

// ServiceModule 服务模块
var ServiceModule = fx.Module("services",
	fx.Provide(NewServices),
)
