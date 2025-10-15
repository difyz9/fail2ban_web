package app

import (
	"embed"

	"go.uber.org/fx"
)

// NewStaticFiles 提供静态文件
func NewStaticFiles(files embed.FS) embed.FS {
	return files
}

// NewApp 创建 fx 应用
func NewApp(staticFiles embed.FS) *fx.App {
	return fx.New(
		// 提供静态文件
		fx.Provide(
			fx.Annotate(
				func() embed.FS { return staticFiles },
				fx.ResultTags(`name:"staticFiles"`),
			),
		),

		// 核心模块
		ConfigModule,
		LoggerModule,
		DatabaseModule,

		// 业务模块
		ServiceModule,
		HandlerModule,
		RouterModule,

		// 启动 HTTP 服务器
		fx.Invoke(RegisterServer),
	)
}
