package main

import (
	"embed"

	"fail2ban-web/app"
)

//go:embed web
var staticFiles embed.FS

func main() {
	// 创建并启动 fx 应用
	fxApp := app.NewApp(staticFiles)
	fxApp.Run()
}
