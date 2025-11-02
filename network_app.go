package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type NetworkApp struct {
	ctx              context.Context
	capture          *NetworkCapture
	proxy            *GoProxyServer
	webview          *WebViewCapture
	wsServer         *WebSocketServer
	workflowExecutor *WorkflowExecutor
	workflowStorage  *WorkflowStorage
	webhookServer    *WebhookServer
}

func NewNetworkApp() *NetworkApp {
	capture := NewNetworkCapture()
	app := &NetworkApp{
		capture: capture,
		proxy:   NewGoProxyServer(8888, capture),
		webview: NewWebViewCapture(capture),
	}
	app.wsServer = NewWebSocketServer(app)
	app.workflowExecutor = NewWorkflowExecutor(app)
	app.webhookServer = NewWebhookServer()

	// 初始化存储
	storage, err := NewWorkflowStorage()
	if err != nil {
		log.Printf("[App] 初始化存储失败: %v", err)
	} else {
		app.workflowStorage = storage
	}

	return app
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
	return na.capture.GetEntries()
}

// 清空记录
func (na *NetworkApp) ClearCapture() {
	na.capture.Clear()
}

// 启动代理服务器
func (na *NetworkApp) StartProxy() error {
	log.Printf("[NetworkApp] 启动代理服务器，当前规则数量: %d", len(na.proxy.interceptor.GetRules()))
	return na.proxy.Start()
}

// 使用指定端口启动代理服务器
func (na *NetworkApp) StartProxyWithPort(port int) error {
	// 保存旧的规则
	oldRules := na.proxy.interceptor.GetRules()
	na.proxy = NewGoProxyServer(port, na.capture)
	// 恢复规则
	if len(oldRules) > 0 {
		log.Printf("[NetworkApp] 恢复 %d 条拦截规则", len(oldRules))
		na.proxy.interceptor.SetRules(oldRules)
	}
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

// 导出数据
func (na *NetworkApp) ExportData(entriesJSON string) error {
	return ExportToFile(na.ctx, entriesJSON)
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

// 下载响应内容
func (na *NetworkApp) DownloadResponse(url string, filename string) error {
	return DownloadFile(na.ctx, url, filename)
}

// 选择下载目录
func (na *NetworkApp) SelectDownloadDirectory() (string, error) {
	path, err := runtime.OpenDirectoryDialog(na.ctx, runtime.OpenDialogOptions{
		Title: "选择下载目录",
	})
	return path, err
}

// 选择JSONL文件
func (na *NetworkApp) SelectJSONLFile() (string, error) {
	path, err := runtime.OpenFileDialog(na.ctx, runtime.OpenDialogOptions{
		Title: "选择JSONL文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSONL Files (*.jsonl)", Pattern: "*.jsonl"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	return path, err
}

// 选择保存文件路径
func (na *NetworkApp) SelectSaveFilePath(defaultFilename string) (string, error) {
	path, err := runtime.SaveFileDialog(na.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
		Title:           "选择保存文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSONL Files (*.jsonl)", Pattern: "*.jsonl"},
			{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	return path, err
}

// 加载JSONL文件
func (na *NetworkApp) LoadJSONLFile(filePath string) (map[string]interface{}, error) {
	reader := NewJSONLReader(filePath)
	if err := reader.Load(); err != nil {
		return nil, fmt.Errorf("加载文件失败: %w", err)
	}

	return map[string]interface{}{
		"keys":       reader.GetKeys(),
		"totalLines": reader.GetLineCount(),
	}, nil
}

// 获取拦截规则
func (na *NetworkApp) GetInterceptRules() []InterceptRule {
	return na.proxy.interceptor.GetRules()
}

// 设置拦截规则
func (na *NetworkApp) SetInterceptRules(rules []InterceptRule) error {
	log.Printf("[NetworkApp] 设置拦截规则，数量: %d", len(rules))
	for i, rule := range rules {
		log.Printf("[NetworkApp] 规则 %d: Name=%s, Pattern=%s, Enabled=%v, ActionType=%s", i, rule.Name, rule.URLPattern, rule.Enabled, rule.ActionType)
	}
	return na.proxy.interceptor.SetRules(rules)
}

// 显示信息对话框
func (na *NetworkApp) ShowInfoDialog(title, message string) {
	runtime.MessageDialog(na.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   title,
		Message: message,
	})
}

// 显示错误对话框
func (na *NetworkApp) ShowErrorDialog(title, message string) {
	runtime.MessageDialog(na.ctx, runtime.MessageDialogOptions{
		Type:    runtime.ErrorDialog,
		Title:   title,
		Message: message,
	})
}

// 显示确认对话框
func (na *NetworkApp) ShowQuestionDialog(title, message string) (string, error) {
	return runtime.MessageDialog(na.ctx, runtime.MessageDialogOptions{
		Type:    runtime.QuestionDialog,
		Title:   title,
		Message: message,
	})
}

// 启动 WebSocket 服务器
func (na *NetworkApp) StartWebSocketServer() error {
	return na.wsServer.Start()
}

// 停止 WebSocket 服务器
func (na *NetworkApp) StopWebSocketServer() error {
	return na.wsServer.Stop()
}

// 获取 WebSocket 端口
func (na *NetworkApp) GetWebSocketPort() int {
	return na.wsServer.GetWSPort()
}

// 获取 WebSocket 运行状态
func (na *NetworkApp) IsWebSocketRunning() bool {
	return na.wsServer.IsRunning()
}

// 发送消息到浏览器插件
func (na *NetworkApp) SendToExtension(msgType string, data map[string]interface{}) {
	na.wsServer.Broadcast(WSMessage{
		Type: msgType,
		Data: data,
	})
}

// 执行工作流
func (na *NetworkApp) ExecuteWorkflow(task WorkflowTask) error {
	go func() {
		if err := na.workflowExecutor.Execute(task); err != nil {
			if na.ctx != nil {
				runtime.EventsEmit(na.ctx, "workflow_error", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}()
	return nil
}

// 停止工作流
func (na *NetworkApp) StopWorkflow() {
	na.workflowExecutor.Stop()
}

// 获取工作流运行状态
func (na *NetworkApp) IsWorkflowRunning() bool {
	return na.workflowExecutor.IsRunning()
}

// 保存工作流任务
func (na *NetworkApp) SaveWorkflowTask(task WorkflowTask) error {
	if na.workflowStorage == nil {
		return fmt.Errorf("存储未初始化")
	}
	return na.workflowStorage.SaveTask(task)
}

// 获取工作流任务
func (na *NetworkApp) GetWorkflowTask(id string) (*WorkflowTask, error) {
	if na.workflowStorage == nil {
		return nil, fmt.Errorf("存储未初始化")
	}
	return na.workflowStorage.GetTask(id)
}

// 获取所有工作流任务
func (na *NetworkApp) GetAllWorkflowTasks() ([]WorkflowTask, error) {
	if na.workflowStorage == nil {
		return nil, fmt.Errorf("存储未初始化")
	}
	return na.workflowStorage.GetAllTasks()
}

// 删除工作流任务
func (na *NetworkApp) DeleteWorkflowTask(id string) error {
	if na.workflowStorage == nil {
		return fmt.Errorf("存储未初始化")
	}
	return na.workflowStorage.DeleteTask(id)
}

// 启动 Webhook 服务器
func (na *NetworkApp) StartWebhookServer() error {
	return na.webhookServer.Start()
}

// 停止 Webhook 服务器
func (na *NetworkApp) StopWebhookServer() error {
	return na.webhookServer.Stop()
}

// 获取 Webhook 运行状态
func (na *NetworkApp) IsWebhookRunning() bool {
	return na.webhookServer.IsRunning()
}

// 获取 Webhook 端口
func (na *NetworkApp) GetWebhookPort() int {
	return na.webhookServer.GetPort()
}

// 设置历史记录最大数量
func (na *NetworkApp) SetMaxHistoryEntries(max int) {
	na.capture.SetMaxEntries(max)
}

// 获取历史记录最大数量
func (na *NetworkApp) GetMaxHistoryEntries() int {
	return na.capture.GetMaxEntries()
}

// 导出到文件
func ExportToFile(ctx context.Context, data string) error {
	// 弹出保存对话框
	savePath, err := runtime.SaveFileDialog(ctx, runtime.SaveDialogOptions{
		DefaultFilename: "network-capture.json",
		Title:           "导出网络数据",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})

	if err != nil {
		return err
	}

	if savePath == "" {
		return nil // 用户取消
	}

	// 写入文件
	return os.WriteFile(savePath, []byte(data), 0644)
}
