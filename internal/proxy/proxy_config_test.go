package proxy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProxyConfigManager(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "proxy_config.json")

	// 创建代理配置管理器
	manager := &ProxyConfigManager{
		configPath: configPath,
		config: &ProxyConfig{
			Enabled: false,
			Type:    "http",
			Host:    "",
			Port:    8080,
			Auth: ProxyAuth{
				Enabled:  false,
				Username: "",
				Password: "",
			},
			Bypass: "localhost\n127.0.0.1\n*.local",
		},
	}

	// 测试默认配置
	if manager.config.Enabled {
		t.Error("默认配置应该禁用代理")
	}

	if manager.config.Type != "http" {
		t.Error("默认代理类型应该是http")
	}

	// 测试保存配置
	manager.config.Enabled = true
	manager.config.Host = "proxy.example.com"
	manager.config.Port = 3128
	manager.config.Auth.Enabled = true
	manager.config.Auth.Username = "testuser"
	manager.config.Auth.Password = "testpass"

	err := manager.SaveConfig()
	if err != nil {
		t.Fatalf("保存配置失败: %v", err)
	}

	// 验证配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("配置文件未创建")
	}

	// 测试加载配置
	newManager := &ProxyConfigManager{
		configPath: configPath,
		config: &ProxyConfig{
			Enabled: false,
			Type:    "http",
			Host:    "",
			Port:    8080,
			Auth: ProxyAuth{
				Enabled:  false,
				Username: "",
				Password: "",
			},
			Bypass: "localhost\n127.0.0.1\n*.local",
		},
	}

	err = newManager.LoadConfig()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证加载的配置
	if !newManager.config.Enabled {
		t.Error("加载的配置应该启用代理")
	}

	if newManager.config.Host != "proxy.example.com" {
		t.Errorf("期望主机名 'proxy.example.com'，实际 '%s'", newManager.config.Host)
	}

	if newManager.config.Port != 3128 {
		t.Errorf("期望端口 3128，实际 %d", newManager.config.Port)
	}

	if !newManager.config.Auth.Enabled {
		t.Error("加载的配置应该启用认证")
	}

	if newManager.config.Auth.Username != "testuser" {
		t.Errorf("期望用户名 'testuser'，实际 '%s'", newManager.config.Auth.Username)
	}
}

func TestShouldBypass(t *testing.T) {
	manager := &ProxyConfigManager{
		config: &ProxyConfig{
			Enabled: true,
			Bypass:  "localhost\n127.0.0.1\n*.local\n192.168.*",
		},
	}

	testCases := []struct {
		host     string
		expected bool
	}{
		{"localhost", true},
		{"127.0.0.1", true},
		{"test.local", true},
		{"192.168.1.1", true},
		{"192.168.100.50", true},
		{"google.com", false},
		{"example.com", false},
		{"10.0.0.1", false},
	}

	for _, tc := range testCases {
		result := manager.ShouldBypass(tc.host)
		if result != tc.expected {
			t.Errorf("ShouldBypass('%s') = %v，期望 %v", tc.host, result, tc.expected)
		}
	}
}

func TestGetProxyURL(t *testing.T) {
	// 测试无认证的代理URL
	manager := &ProxyConfigManager{
		config: &ProxyConfig{
			Enabled: true,
			Type:    "http",
			Host:    "proxy.example.com",
			Port:    8080,
			Auth: ProxyAuth{
				Enabled: false,
			},
		},
	}

	expected := "http://proxy.example.com:8080"
	result := manager.GetProxyURL()
	if result != expected {
		t.Errorf("GetProxyURL() = '%s'，期望 '%s'", result, expected)
	}

	// 测试有认证的代理URL
	manager.config.Auth.Enabled = true
	manager.config.Auth.Username = "user"
	manager.config.Auth.Password = "pass"

	expected = "http://user:pass@proxy.example.com:8080"
	result = manager.GetProxyURL()
	if result != expected {
		t.Errorf("GetProxyURL() = '%s'，期望 '%s'", result, expected)
	}

	// 测试禁用代理时的URL
	manager.config.Enabled = false
	result = manager.GetProxyURL()
	if result != "" {
		t.Errorf("禁用代理时GetProxyURL()应该返回空字符串，实际 '%s'", result)
	}
}

func TestTestConnectionWithURL(t *testing.T) {
	manager := &ProxyConfigManager{
		config: &ProxyConfig{
			Enabled: false,
		},
	}

	// 测试代理未启用的情况
	result := manager.TestConnectionWithURL("https://www.google.com")
	if result.Success {
		t.Error("代理未启用时应该返回失败")
	}
	if result.Message != "代理未启用或配置不完整" {
		t.Errorf("期望错误消息 '代理未启用或配置不完整'，实际 '%s'", result.Message)
	}
}
