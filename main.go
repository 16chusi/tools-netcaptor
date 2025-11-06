package main

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// 全局日志实例
var AppLog logger.Logger

func main() {
	// 初始化日志系统
	if err := InitLogger(); err != nil {
		fmt.Printf("初始化日志系统失败: %v\n", err)
		os.Exit(1)
	}
	// Create an instance of the app structure
	app := NewApp()
	networkApp := NewNetworkApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "NetCaptor",
		Width:  AppConfig.WindowWidth,
		Height: AppConfig.WindowHeight,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			AppLog.Info("NetCaptor 应用启动")

			app.startup(ctx)
			networkApp.startup(ctx)
		},
		Bind: []interface{}{
			app,
			networkApp,
		},
		Logger: AppLog, // 使用文件日志器
	})

	if err != nil {
		AppLog.Error("应用启动失败: " + err.Error())
	}
}
