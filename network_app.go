package main

import (
	"context"
	"fmt"
)

type NetworkApp struct {
	ctx     context.Context
	capture *NetworkCapture
	proxy   *GoProxyServer
	webview *WebViewCapture
}

func NewNetworkApp() *NetworkApp {
	capture := NewNetworkCapture()
	return &NetworkApp{
		capture: capture,
		proxy:   NewGoProxyServer(8888, capture),
		webview: NewWebViewCapture(capture),
	}
}

func (na *NetworkApp) startup(ctx context.Context) {
	na.ctx = ctx
	na.capture.ctx = ctx
	na.webview.SetContext(ctx)
}

// 获取注入脚本
func (na *NetworkApp) GetInjectionScript() string {
	return na.capture.GetInjectionScript()
}

// 记录请求
func (na *NetworkApp) RecordRequest(reqJSON string) error {
	return na.capture.RecordRequest(reqJSON)
}

// 记录响应
func (na *NetworkApp) RecordResponse(respJSON string) error {
	return na.capture.RecordResponse(respJSON)
}

// 获取所有请求
func (na *NetworkApp) GetAllRequests() []NetworkRequest {
	return na.capture.GetRequests()
}

// 获取所有响应
func (na *NetworkApp) GetAllResponses() []NetworkResponse {
	return na.capture.GetResponses()
}

// 获取所有条目(合并请求和响应)
func (na *NetworkApp) GetAllEntries() []NetworkEntry {
	requests := na.capture.GetRequests()
	responses := na.capture.GetResponses()

	// 合并请求和响应
	respMap := make(map[string]NetworkResponse)
	for _, resp := range responses {
		respMap[resp.URL] = resp
	}

	entries := make([]NetworkEntry, 0)
	for _, req := range requests {
		resp, hasResp := respMap[req.URL]
		entry := NetworkEntry{
			ID:      req.ID,
			URL:     req.URL,
			Method:  req.Method,
			Type:    req.Type,
			Time:    req.Time,
			Domain:  req.Domain,
			Path:    req.Path,
			Request: req,
		}

		if hasResp {
			entry.Status = resp.Status
			entry.StatusText = resp.StatusText
			entry.Size = resp.Size
			entry.Duration = resp.Duration
			entry.Response = resp
		}

		entries = append(entries, entry)
	}

	return entries
}

// 清空记录
func (na *NetworkApp) ClearCapture() {
	na.capture.Clear()
}

// 启动代理服务器
func (na *NetworkApp) StartProxy() error {
	return na.proxy.Start()
}

// 使用指定端口启动代理服务器
func (na *NetworkApp) StartProxyWithPort(port int) error {
	na.proxy = NewGoProxyServer(port, na.capture)
	return na.proxy.Start()
}

// 停止代理服务器
func (na *NetworkApp) StopProxy() error {
	return na.proxy.Stop()
}

// 获取代理状态
func (na *NetworkApp) IsProxyRunning() bool {
	return na.proxy.IsRunning()
}

// 获取代理URL
func (na *NetworkApp) GetProxyURL() string {
	return na.proxy.GetProxyURL()
}

// 获取代理端口
func (na *NetworkApp) GetProxyPort() int {
	return na.proxy.GetPort()
}

// 获取CA证书路径
func (na *NetworkApp) GetCACertPath() string {
	return na.proxy.GetCACertPath()
}

// 导出HAR格式
func (na *NetworkApp) ExportHAR() (string, error) {
	// 简化版HAR导出
	requests := na.capture.GetRequests()
	responses := na.capture.GetResponses()

	return fmt.Sprintf(`{
  "log": {
    "version": "1.2",
    "creator": {"name": "NetworkCapture", "version": "1.0"},
    "entries": %d,
    "requests": %d,
    "responses": %d
  }
}`, len(requests)+len(responses), len(requests), len(responses)), nil
}

// 在Chrome中打开URL
func (na *NetworkApp) OpenInChrome(url string) error {
	proxyURL := na.proxy.GetProxyURL()
	return OpenInChrome(url, proxyURL)
}

// 在Edge中打开URL
func (na *NetworkApp) OpenInEdge(url string) error {
	proxyURL := na.proxy.GetProxyURL()
	return OpenInEdge(url, proxyURL)
}

// 在Firefox中打开URL
func (na *NetworkApp) OpenInFirefox(url string) error {
	proxyURL := na.proxy.GetProxyURL()
	return OpenInFirefox(url, proxyURL)
}
