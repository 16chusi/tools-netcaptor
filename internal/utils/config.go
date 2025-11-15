package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

type Config struct {
	LogLevel string
	LogMode  string // file, console, both
	LogFile  string

	// 窗口配置
	WindowWidth  int
	WindowHeight int
}

var AppConfig *Config

// LoadConfig 加载配置文件，如果不存在则创建默认配置
func LoadConfig() error {
	configPath := "config.ini"

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 创建默认配置文件
		if err := createDefaultConfig(configPath); err != nil {
			return fmt.Errorf("创建默认配置文件失败: %v", err)
		}
	}

	// 加载配置文件
	cfg, err := ini.Load(configPath)
	if err != nil {
		return fmt.Errorf("加载配置文件失败: %v", err)
	}

	// 解析配置
	AppConfig = &Config{
		LogLevel: cfg.Section("log").Key("level").MustString("INFO"),
		LogMode:  cfg.Section("log").Key("mode").MustString("file"),
		LogFile:  cfg.Section("log").Key("file").MustString("logs/netcaptor.log"),

		WindowWidth:  cfg.Section("window").Key("width").MustInt(1400),
		WindowHeight: cfg.Section("window").Key("height").MustInt(900),
	}

	// 环境变量优先级更高
	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		AppConfig.LogLevel = strings.ToUpper(envLevel)
	}

	return nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(configPath string) error {
	defaultConfig := `# NetCaptor 配置文件

[log]
# 日志级别: DEBUG, INFO, WARNING, ERROR
level = INFO

# 日志输出方式: file(文件), console(控制台), both(两者)
mode = file

# 日志文件路径
file = logs/netcaptor.log

[window]
# 窗口默认宽度
width = 1400

# 窗口默认高度
height = 900

[proxy]
# 代理服务器端口
port = 8080
`

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}
