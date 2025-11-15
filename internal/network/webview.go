package network

import (
	"context"
)

type WebViewCapture struct {
	ctx        context.Context
	capture    *NetworkCapture
	currentURL string
}

func NewWebViewCapture(capture *NetworkCapture) *WebViewCapture {
	return &WebViewCapture{
		capture: capture,
	}
}

func (wc *WebViewCapture) SetContext(ctx context.Context) {
	wc.ctx = ctx
}

// 导航到指定URL
func (wc *WebViewCapture) NavigateToURL(url string) error {
	wc.currentURL = url
	// Wails 会在前端处理导航
	return nil
}

// 获取当前URL
func (wc *WebViewCapture) GetCurrentURL() string {
	return wc.currentURL
}

// 执行JavaScript注入
func (wc *WebViewCapture) InjectScript() string {
	return wc.capture.GetInjectionScript()
}
