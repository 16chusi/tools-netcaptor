package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"netcaptor/internal/utils"
)

// ProxyTestResult 代理测试结果
type ProxyTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ProxyAuth 代理认证配置
type ProxyAuth struct {
	Enabled  bool   `json:"enabled"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Enabled bool      `json:"enabled"`
	Type    string    `json:"type"` // http, https, socks5
	Host    string    `json:"host"`
	Port    int       `json:"port"`
	Auth    ProxyAuth `json:"auth"`
	Bypass  string    `json:"bypass"` // 绕过代理的地址列表，换行分隔
}

// ProxyConfigManager 代理配置管理器
type ProxyConfigManager struct {
	configPath string
	config     *ProxyConfig
}

// NewProxyConfigManager 创建代理配置管理器
func NewProxyConfigManager() *ProxyConfigManager {
	manager := &ProxyConfigManager{
		configPath: "", // 不再使用文件路径
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

	// 从SQLite加载配置
	manager.LoadConfig()

	return manager
}

// LoadConfig 加载配置
func (pcm *ProxyConfigManager) LoadConfig() error {
	cm := utils.GetConfigManager()
	err := cm.GetConfigJSON("proxy_config", pcm.config)
	if err != nil {
		return nil // 配置不存在，使用默认配置
	}
	return nil
}

// SaveConfig 保存配置
func (pcm *ProxyConfigManager) SaveConfig() error {
	cm := utils.GetConfigManager()
	return cm.SetConfigJSON("proxy_config", pcm.config)
}

// GetConfig 获取配置
func (pcm *ProxyConfigManager) GetConfig() *ProxyConfig {
	return pcm.config
}

// SetConfig 设置配置
func (pcm *ProxyConfigManager) SetConfig(config *ProxyConfig) error {
	pcm.config = config
	return pcm.SaveConfig()
}

// TestConnection 测试代理连接
func (pcm *ProxyConfigManager) TestConnection() ProxyTestResult {
	return pcm.TestConnectionWithURL("http://httpbin.org/ip")
}

// TestConnectionWithURL 使用指定URL测试代理连接
func (pcm *ProxyConfigManager) TestConnectionWithURL(testURL string) ProxyTestResult {
	if !pcm.config.Enabled || pcm.config.Host == "" || pcm.config.Port == 0 {
		return ProxyTestResult{Success: false, Message: "代理未启用或配置不完整"}
	}

	// 构建代理URL
	var proxyURL string
	if pcm.config.Auth.Enabled && pcm.config.Auth.Username != "" {
		proxyURL = fmt.Sprintf("%s://%s:%s@%s:%d",
			pcm.config.Type,
			url.QueryEscape(pcm.config.Auth.Username),
			url.QueryEscape(pcm.config.Auth.Password),
			pcm.config.Host,
			pcm.config.Port)
	} else {
		proxyURL = fmt.Sprintf("%s://%s:%d",
			pcm.config.Type,
			pcm.config.Host,
			pcm.config.Port)
	}

	// 解析代理URL
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return ProxyTestResult{Success: false, Message: fmt.Sprintf("代理URL格式错误: %v", err)}
	}

	// 创建HTTP客户端
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxy),
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).DialContext,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	// 测试连接
	resp, err := client.Get(testURL)
	if err != nil {
		return ProxyTestResult{Success: false, Message: fmt.Sprintf("连接失败: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return ProxyTestResult{Success: true, Message: fmt.Sprintf("代理连接成功，访问 %s 返回 HTTP %d", testURL, resp.StatusCode)}
	}

	return ProxyTestResult{Success: false, Message: fmt.Sprintf("访问 %s 失败: HTTP %d", testURL, resp.StatusCode)}
}

// ShouldBypass 检查地址是否应该绕过代理
func (pcm *ProxyConfigManager) ShouldBypass(host string) bool {
	if !pcm.config.Enabled || pcm.config.Bypass == "" {
		return false
	}

	bypassList := strings.Split(pcm.config.Bypass, "\n")
	for _, pattern := range bypassList {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		// 简单的通配符匹配
		if strings.Contains(pattern, "*") {
			// 将通配符转换为正则表达式风格的匹配
			pattern = strings.ReplaceAll(pattern, "*", "")
			if strings.Contains(host, pattern) {
				return true
			}
		} else if host == pattern {
			return true
		}
	}

	return false
}

// GetProxyURL 获取代理URL（用于HTTP客户端）
func (pcm *ProxyConfigManager) GetProxyURL() string {
	if !pcm.config.Enabled {
		return ""
	}

	if pcm.config.Auth.Enabled && pcm.config.Auth.Username != "" {
		return fmt.Sprintf("%s://%s:%s@%s:%d",
			pcm.config.Type,
			url.QueryEscape(pcm.config.Auth.Username),
			url.QueryEscape(pcm.config.Auth.Password),
			pcm.config.Host,
			pcm.config.Port)
	}

	return fmt.Sprintf("%s://%s:%d",
		pcm.config.Type,
		pcm.config.Host,
		pcm.config.Port)
}

// CreateHTTPClient 创建支持代理的HTTP客户端
func (pcm *ProxyConfigManager) CreateHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
		DialContext: (&net.Dialer{
			Timeout: timeout,
		}).DialContext,
	}

	// 如果启用了代理，设置代理
	if pcm.config.Enabled && pcm.config.Host != "" && pcm.config.Port > 0 {
		proxyURL := pcm.GetProxyURL()
		if proxyURL != "" {
			if proxy, err := url.Parse(proxyURL); err == nil {
				transport.Proxy = http.ProxyURL(proxy)
			}
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// CreateSmartHTTPClient 创建智能路由的HTTP客户端
func (pcm *ProxyConfigManager) CreateSmartHTTPClient(targetURL string, smartProxy *SmartProxyManager, timeout time.Duration) *http.Client {
	// 提取主机名
	host := extractHostFromURL(targetURL)

	// 决定路由方式
	routeType := smartProxy.DecideRoute(host)

	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
		DialContext: (&net.Dialer{
			Timeout: timeout,
		}).DialContext,
	}

	// 根据路由决策设置代理
	if routeType == "proxy" && pcm.config.Enabled && pcm.config.Host != "" && pcm.config.Port > 0 {
		proxyURL := pcm.GetProxyURL()
		if proxyURL != "" {
			if proxy, err := url.Parse(proxyURL); err == nil {
				transport.Proxy = http.ProxyURL(proxy)
			}
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// extractHostFromURL 从URL中提取主机名
func extractHostFromURL(targetURL string) string {
	if u, err := url.Parse(targetURL); err == nil {
		return u.Host
	}
	return targetURL
}
