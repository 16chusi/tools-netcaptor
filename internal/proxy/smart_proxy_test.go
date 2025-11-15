package proxy

import (
	"path/filepath"
	"testing"
)

func TestSmartProxyManager(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "smart_proxy_rules.json")

	// 创建代理配置管理器
	proxyConfigMgr := &ProxyConfigManager{
		config: &ProxyConfig{
			Enabled: true,
			Type:    "http",
			Host:    "proxy.example.com",
			Port:    8080,
		},
	}

	// 创建智能代理管理器
	smartProxy := &SmartProxyManager{
		configPath:     configPath,
		rules:          make([]*RouteRule, 0),
		proxyConfigMgr: proxyConfigMgr,
	}

	// 添加默认规则
	smartProxy.addDefaultRules()

	// 测试本地地址直连
	route := smartProxy.DecideRoute("localhost")
	if route != "direct" {
		t.Errorf("localhost 应该直连，实际: %s", route)
	}

	route = smartProxy.DecideRoute("127.0.0.1")
	if route != "direct" {
		t.Errorf("127.0.0.1 应该直连，实际: %s", route)
	}

	// 测试未知域名（默认直连）
	route = smartProxy.DecideRoute("unknown.com")
	if route != "direct" {
		t.Errorf("未知域名应该直连，实际: %s", route)
	}

	// 测试添加手动规则
	err := smartProxy.AddManualRule("google.com", "proxy")
	if err != nil {
		t.Fatalf("添加手动规则失败: %v", err)
	}

	route = smartProxy.DecideRoute("google.com")
	if route != "proxy" {
		t.Errorf("google.com 应该走代理，实际: %s", route)
	}

	// 测试通配符规则
	err = smartProxy.AddManualRule("*.openai.com", "proxy")
	if err != nil {
		t.Fatalf("添加通配符规则失败: %v", err)
	}

	route = smartProxy.DecideRoute("api.openai.com")
	if route != "proxy" {
		t.Errorf("api.openai.com 应该走代理，实际: %s", route)
	}

	route = smartProxy.DecideRoute("openai.com")
	if route != "proxy" {
		t.Errorf("openai.com 应该走代理，实际: %s", route)
	}

	// 测试自动学习
	smartProxy.LearnFromFailure("github.com", "connection timeout")
	route = smartProxy.DecideRoute("github.com")
	if route != "proxy" {
		t.Errorf("github.com 学习后应该走代理，实际: %s", route)
	}

	// 测试规则优先级（手动 > 自动）
	smartProxy.LearnFromFailure("google.com", "connection failed") // 尝试添加自动规则
	route = smartProxy.DecideRoute("google.com")
	if route != "proxy" {
		t.Errorf("google.com 手动规则应该优先，实际: %s", route)
	}

	// 验证规则数量
	rules := smartProxy.GetRules()
	expectedCount := 3 + 2 + 1 // 默认规则 + 手动规则 + 自动学习规则
	if len(rules) != expectedCount {
		t.Errorf("期望 %d 条规则，实际 %d 条", expectedCount, len(rules))
	}
}

func TestPatternMatching(t *testing.T) {
	smartProxy := &SmartProxyManager{}

	testCases := []struct {
		pattern  string
		host     string
		expected bool
	}{
		{"google.com", "google.com", true},
		{"google.com", "www.google.com", false},
		{"*.google.com", "www.google.com", true},
		{"*.google.com", "api.google.com", true},
		{"*.google.com", "google.com", true},
		{"*.openai.com", "api.openai.com", true},
		{"*.openai.com", "chat.openai.com", true},
		{"*.openai.com", "openai.com", true},
		{"*.openai.com", "example.com", false},
		{"localhost", "localhost", true},
		{"localhost", "127.0.0.1", false},
	}

	for _, tc := range testCases {
		result := smartProxy.matchPattern(tc.pattern, tc.host)
		if result != tc.expected {
			t.Errorf("matchPattern('%s', '%s') = %v，期望 %v", tc.pattern, tc.host, result, tc.expected)
		}
	}
}
