package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/logger"
)

var ConsoleLogger logger.Logger

// InitLogger 初始化日志系统
func InitLogger() error {
	// 加载配置
	if err := LoadConfig(); err != nil {
		return err
	}

	// 创建日志目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		return err
	}

	// 根据配置创建日志器
	switch strings.ToLower(AppConfig.LogMode) {
	case "console":
		AppLog = logger.NewDefaultLogger()
		ConsoleLogger = AppLog
	case "file":
		AppLog = logger.NewFileLogger(AppConfig.LogFile)
	case "both":
		AppLog = logger.NewFileLogger(AppConfig.LogFile)
		ConsoleLogger = logger.NewDefaultLogger()
	default:
		AppLog = logger.NewFileLogger(AppConfig.LogFile)
	}

	LogInfo(fmt.Sprintf("NetCaptor 启动，日志级别: %s, 输出方式: %s", AppConfig.LogLevel, AppConfig.LogMode))
	return nil
}

// LogDebug 只在 DEBUG 级别时输出日志
func LogDebug(message string) {
	if shouldLog("DEBUG") {
		output("[DEBUG] " + message)
	}
}

// LogInfo 在 INFO 及以上级别时输出日志
func LogInfo(message string) {
	if shouldLog("INFO") {
		output(message)
	}
}

// LogWarning 在 WARNING 及以上级别时输出日志
func LogWarning(message string) {
	if shouldLog("WARNING") {
		AppLog.Warning(message)
		if ConsoleLogger != nil {
			ConsoleLogger.Warning(message)
		}
	}
}

// LogError 在 ERROR 及以上级别时输出日志
func LogError(message string) {
	if shouldLog("ERROR") {
		AppLog.Error(message)
		if ConsoleLogger != nil {
			ConsoleLogger.Error(message)
		}
	}
}

// output 根据配置输出日志
func output(message string) {
	AppLog.Info(message)
	if ConsoleLogger != nil {
		ConsoleLogger.Info(message)
	}
}

// shouldLog 检查是否应该输出指定级别的日志
func shouldLog(level string) bool {
	currentLevel := strings.ToUpper(AppConfig.LogLevel)
	targetLevel := strings.ToUpper(level)

	// 定义日志级别优先级
	levels := map[string]int{
		"DEBUG":   0,
		"INFO":    1,
		"WARNING": 2,
		"ERROR":   3,
	}

	currentPriority, exists1 := levels[currentLevel]
	targetPriority, exists2 := levels[targetLevel]

	if !exists1 || !exists2 {
		return true // 如果级别不存在，默认输出
	}

	return currentPriority <= targetPriority
}
