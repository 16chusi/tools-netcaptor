package types

import "context"

// CaptureHandler 网络捕获处理接口
type CaptureHandler interface {
	AddRequest(req *NetworkRequest)
	GetInjectionScript() string
	GetRequests() []NetworkRequest
	AddResponse(resp *NetworkResponse)
	AddEntry(entry *NetworkEntry)
	GetEntries() []NetworkEntry
}

// AppHandler 应用处理接口
type AppHandler interface {
	GetSmartProxyManager() interface{}
	GetProxyConfigManager() interface{}
}

// WorkflowAppHandler workflow 需要的应用接口
type WorkflowAppHandler interface {
	GetContext() context.Context
	SendWSMessage(msg WSMessage)
}

// WebSocketHandler WebSocket 处理接口
type WebSocketHandler interface {
	SendMessage(msg WSMessage)
	BroadcastMessage(msg WSMessage)
	IsRunning() bool
	HasClients() bool
}
