package app

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ServerParams 服务器依赖参数
type ServerParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Logger    *zap.Logger
	Router    *gin.Engine
}

// RegisterServer 注册并启动 HTTP 服务器
func RegisterServer(params ServerParams) {
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.Logger.Info("Starting HTTP server on :8092...")
			params.Logger.Info("Access the management panel at http://localhost:8092")
			
			go func() {
				if err := params.Router.Run(":8092"); err != nil {
					params.Logger.Fatal("Failed to start HTTP server", zap.Error(err))
				}
			}()
			
			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info("HTTP server stopped")
			return nil
		},
	})
}
