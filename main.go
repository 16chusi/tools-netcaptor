package main

import (
	"context"
	"embed"
	"fmt"
	"os"

	"netcaptor/internal/network"
	"netcaptor/internal/utils"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 初始化日志系统
	if err := utils.InitLogger(); err != nil {
		fmt.Printf("初始化日志系统失败: %v\n", err)
		os.Exit(1)
	}
	// Create an instance of the app structure
	app := NewApp()
	networkApp := network.NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "NetCaptor",
		Width:  utils.AppConfig.WindowWidth,
		Height: utils.AppConfig.WindowHeight,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			utils.AppLog.Info("NetCaptor 应用启动")

			app.startup(ctx)
			networkApp.Startup(ctx)
		},
		Bind: []interface{}{
			app,
			networkApp,
		},
		Logger: utils.AppLog, // 使用文件日志器
	})

	if err != nil {
		utils.AppLog.Error("应用启动失败: " + err.Error())
	}
}
