package network

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"netcaptor/internal/ai"
	"netcaptor/internal/browser"
	"netcaptor/internal/download"
	"netcaptor/internal/proxy"
	"netcaptor/internal/types"
	"netcaptor/internal/utils"
	"netcaptor/internal/workflow"
)

type NetworkApp struct {
	ctx              context.Context
	capture          *NetworkCapture
	proxy            *proxy.GoProxyServer
	webview          *WebViewCapture
	wsServer         *WebSocketServer
	workflowExecutor *workflow.WorkflowExecutor
	workflowStorage  *workflow.WorkflowStorage
	webhookServer    *WebhookServer
	proxyConfigMgr   *proxy.ProxyConfigManager
	smartProxyMgr    *proxy.SmartProxyManager
}

func NewApp() *NetworkApp {
	capture := NewNetworkCapture()
	proxyConfigMgr := proxy.NewProxyConfigManager()
	smartProxyMgr := proxy.NewSmartProxyManager(proxyConfigMgr)

	app := &NetworkApp{
		capture:        capture,
		proxy:          proxy.NewGoProxyServer(8888, capture),
		webview:        NewWebViewCapture(capture),
		proxyConfigMgr: proxyConfigMgr,
		smartProxyMgr:  smartProxyMgr,
	}

	// 设置GoProxy的智能代理管理器
	// 	app.proxy.smartProxy = smartProxyMgr
	// 	app.proxy.networkApp = app // 设置NetworkApp引用
	// 设置智能代理传输层
	app.proxy.SetupSmartProxy()

	app.wsServer = NewWebSocketServer(app)
	app.workflowExecutor = workflow.NewWorkflowExecutor(app, app.wsServer, app.proxyConfigMgr, app.smartProxyMgr)
	app.webhookServer = NewWebhookServer()

	// 设置代理服务器的WS端口获取函数
	app.proxy.SetWSPortGetter(func() int {
		if app.wsServer != nil {
			return app.wsServer.GetWSPort()
		}
		return 0
	})

	// 初始化存储
	storage, err := workflow.NewWorkflowStorage()
	if err != nil {
		utils.AppLog.Info(fmt.Sprintf("[App] 初始化存储失败: %v", err))
	} else {
		app.workflowStorage = storage
	}

	return app
}

func (na *NetworkApp) Startup(ctx context.Context) {
	na.ctx = ctx
	na.capture.ctx = ctx
	na.webview.SetContext(ctx)

	// 设置 proxy 的 appHandler
	if na.proxy != nil {
		na.proxy.SetAppHandler(na)
	}

	// 自动启动 WebSocket 服务器
	if na.wsServer != nil {
		if err := na.wsServer.Start(); err != nil {
			utils.AppLog.Error(fmt.Sprintf("[App] WebSocket服务器启动失败: %v", err))
		} else {
			utils.AppLog.Info(fmt.Sprintf("[App] WebSocket服务器已启动，端口: %d", na.wsServer.GetWSPort()))
		}
	}

	// 加载AI模型配置
	if na.workflowStorage != nil && na.workflowExecutor != nil && na.workflowExecutor.GetAIService() != nil {
		if models, err := na.workflowStorage.LoadAIModels(); err == nil && len(models) > 0 {
			na.workflowExecutor.GetAIService().UpdateModels(models)
			utils.AppLog.Info(fmt.Sprintf("[App] 已加载 %d 个AI模型配置", len(models)))
		}
	}

	// 加载历史记录最大数量配置
	cm := utils.GetConfigManager()
	var maxEntries int
	if err := cm.GetConfigJSON("max_history_entries", &maxEntries); err == nil && maxEntries > 0 {
		na.capture.SetMaxEntries(maxEntries)
		utils.AppLog.Info(fmt.Sprintf("[App] 已加载历史记录最大数量: %d", maxEntries))
	}
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
func (na *NetworkApp) GetAllRequests() []types.NetworkRequest {
	return na.capture.GetRequests()
}

// 获取所有响应
func (na *NetworkApp) GetAllResponses() []types.NetworkResponse {
	return na.capture.GetResponses()
}

// 获取所有条目(合并请求和响应)
func (na *NetworkApp) GetAllEntries() []types.NetworkEntry {
	return na.capture.GetEntries()
}

// 清空记录
func (na *NetworkApp) ClearCapture() {
	na.capture.Clear()
}

// 启动代理服务器
func (na *NetworkApp) StartProxy() error {
	// 	utils.AppLog.Info(fmt.Sprintf("[NetworkApp] 启动代理服务器，当前规则数量: %d", len(na.proxy.interceptor.GetRules())))
	return na.proxy.Start()
}

// 使用指定端口启动代理服务器
func (na *NetworkApp) StartProxyWithPort(port int) error {
	// 保存旧的规则
	// 	oldRules := na.proxy.interceptor.GetRules()
	// 	na.proxy = proxy.NewGoProxyServer(port, na.capture)

	// 重新设置NetworkApp引用和智能代理管理器
	// 	na.proxy.smartProxy = na.smartProxyMgr
	// 	na.proxy.networkApp = na
	na.proxy.SetupSmartProxy()

	// 恢复规则
	// 	if len(oldRules) > 0 {
	// 		utils.AppLog.Info(fmt.Sprintf("[NetworkApp] 恢复 %d 条拦截规则", len(oldRules)))
	// 		na.proxy.interceptor.SetRules(oldRules)
	// 	}
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

func (na *NetworkApp) GetCACertInfo() *types.CertInfo {
	// 直接从GoProxy服务器获取证书路径和信息
	certPath := na.proxy.GetCACertPath()
	info := &types.CertInfo{
		Path: certPath,
	}

	// 检查文件是否存在
	if stat, err := os.Stat(certPath); err == nil {
		info.Exists = true
		info.CreatedAt = stat.ModTime().Format("2006-01-02 15:04:05")

		// 尝试读取证书内容获取详细信息
		if certPEM, err := os.ReadFile(certPath); err == nil {
			if certBlock, _ := pem.Decode(certPEM); certBlock != nil {
				if cert, err := x509.ParseCertificate(certBlock.Bytes); err == nil {
					info.NotBefore = cert.NotBefore.Format("2006-01-02 15:04:05")
					info.NotAfter = cert.NotAfter.Format("2006-01-02 15:04:05")
					info.Subject = cert.Subject.String()
					info.Issuer = cert.Issuer.String()
				}
			}
		}
	}

	return info
}

// 导出数据
func (na *NetworkApp) ExportData(entriesJSON string) error {
	return ExportToFile(na.ctx, entriesJSON)
}

// 在Chrome中打开URL
func (na *NetworkApp) OpenInChrome(url string) error {
	proxyURL := na.proxy.GetProxyURL()
	utils.AppLog.Info(fmt.Sprintf("[NetworkApp] 🚀 启动Chrome浏览器\n"))
	utils.AppLog.Info(fmt.Sprintf("[NetworkApp] 📍 目标URL: %s\n", url))
	utils.AppLog.Info(fmt.Sprintf("[NetworkApp] 🔗 中间人代理URL: %s\n", proxyURL))
	utils.AppLog.Info(fmt.Sprintf("[NetworkApp] 📊 代理服务器运行状态: %v\n", na.proxy.IsRunning()))
	return browser.OpenInChrome(url, proxyURL)
}

// 在Edge中打开URL
func (na *NetworkApp) OpenInEdge(url string) error {
	proxyURL := na.proxy.GetProxyURL()
	return browser.OpenInEdge(url, proxyURL)
}

// 在Firefox中打开URL
func (na *NetworkApp) OpenInFirefox(url string) error {
	proxyURL := na.proxy.GetProxyURL()
	return browser.OpenInFirefox(url, proxyURL)
}

// 下载响应内容
func (na *NetworkApp) DownloadResponse(url string, filename string) error {
	return download.DownloadFile(na.ctx, url, filename)
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
	reader := utils.NewJSONLReader(filePath)
	if err := reader.Load(); err != nil {
		return nil, fmt.Errorf("加载文件失败: %w", err)
	}

	return map[string]interface{}{
		"keys":       reader.GetKeys(),
		"totalLines": reader.GetLineCount(),
	}, nil
}

// 获取拦截规则
func (na *NetworkApp) GetInterceptRules() []types.InterceptRule {
	return na.proxy.GetInterceptor().GetRulesAsTypes()
}

// 设置拦截规则
func (na *NetworkApp) SetInterceptRules(rules []types.InterceptRule) error {
	utils.AppLog.Info(fmt.Sprintf("[NetworkApp] 设置拦截规则，数量: %d", len(rules)))
	for i, rule := range rules {
		utils.AppLog.Info(fmt.Sprintf("[NetworkApp] 规则 %d: Name=%s, Pattern=%s, Enabled=%v, ActionType=%s", i, rule.Name, rule.URLPattern, rule.Enabled, rule.ActionType))
	}
	return na.proxy.GetInterceptor().SetRulesFromTypes(rules)
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
	na.wsServer.Broadcast(types.WSMessage{
		Type: msgType,
		Data: data,
	})
}

// 执行工作流
func (na *NetworkApp) ExecuteWorkflow(task workflow.WorkflowTask) error {
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
func (na *NetworkApp) SaveWorkflowTask(task workflow.WorkflowTask) error {
	if na.workflowStorage == nil {
		return fmt.Errorf("存储未初始化")
	}
	return na.workflowStorage.SaveTask(task)
}

// 获取工作流任务
func (na *NetworkApp) GetWorkflowTask(id string) (*workflow.WorkflowTask, error) {
	if na.workflowStorage == nil {
		return nil, fmt.Errorf("存储未初始化")
	}
	return na.workflowStorage.GetTask(id)
}

// 获取所有工作流任务
func (na *NetworkApp) GetAllWorkflowTasks() ([]workflow.WorkflowTask, error) {
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
	// 保存到数据库
	cm := utils.GetConfigManager()
	cm.SetConfigJSON("max_history_entries", max)
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

// // UpdateAIModels 更新AI模型配置
func (app *NetworkApp) UpdateAIModels(models []ai.AIModel) error {
	// 更新AIService
	if app.workflowExecutor != nil && app.workflowExecutor.GetAIService() != nil {
		app.workflowExecutor.GetAIService().UpdateModels(models)
	}
	// 保存到数据库
	if app.workflowStorage != nil {
		return app.workflowStorage.SaveAIModels(models)
	}
	return nil
}

// GetAIModels 获取AI模型配置
func (app *NetworkApp) GetAIModels() []ai.AIModel {
	if app.workflowStorage != nil {
		models, err := app.workflowStorage.LoadAIModels()
		if err == nil {
			return models
		}
		utils.AppLog.Error(fmt.Sprintf("加载AI模型失败: %v", err))
	}
	return []ai.AIModel{}
}

// TestAIModel 测试AI模型连接
func (app *NetworkApp) TestAIModel(model ai.AIModel) error {
	utils.AppLog.Info(fmt.Sprintf("[AI测试] 开始测试模型: %s, 供应商: %s\n", model.Name, model.Provider))

	// 创建临时AIService进行测试
	aiService := ai.NewAIService(app.proxyConfigMgr, app.smartProxyMgr)
	aiService.UpdateModels([]ai.AIModel{model})

	err := aiService.TestModel(model)
	if err != nil {
		utils.AppLog.Info(fmt.Sprintf("[AI测试] 测试失败: %v\n", err))
	} else {
		utils.AppLog.Info(fmt.Sprintf("[AI测试] 测试成功\n"))
	}

	return err
}

// CallAI 调用AI接口
func (app *NetworkApp) CallAI(modelIndex int, prompt string, systemPrompt string) (string, error) {
	if app.workflowExecutor == nil || app.workflowExecutor.GetAIService() == nil {
		return "", fmt.Errorf("AI服务未初始化")
	}

	return app.workflowExecutor.GetAIService().CallAI(modelIndex, prompt, systemPrompt)
}

// ===== 代理配置管理方法 =====

// GetProxyConfig 获取代理配置
func (app *NetworkApp) GetProxyConfig() *proxy.ProxyConfig {
	return app.proxyConfigMgr.GetConfig()
}

// SetProxyConfig 设置代理配置
func (app *NetworkApp) SetProxyConfig(config *proxy.ProxyConfig) error {
	return app.proxyConfigMgr.SetConfig(config)
}

// TestProxyConnection 测试代理连接
func (app *NetworkApp) TestProxyConnection() proxy.ProxyTestResult {
	return app.proxyConfigMgr.TestConnection()
}

// TestProxyConnectionWithURL 使用指定URL测试代理连接
func (app *NetworkApp) TestProxyConnectionWithURL(testURL string) proxy.ProxyTestResult {
	return app.proxyConfigMgr.TestConnectionWithURL(testURL)
}

// CreateHTTPClient 创建支持代理的HTTP客户端
func (app *NetworkApp) CreateHTTPClient(timeoutSeconds int) *http.Client {
	timeout := time.Duration(timeoutSeconds) * time.Second
	return app.proxyConfigMgr.CreateHTTPClient(timeout)
}

// ===== 智能代理管理方法 =====

// GetSmartProxyRules 获取智能代理规则
func (app *NetworkApp) GetSmartProxyRules() []*proxy.RouteRule {
	return app.smartProxyMgr.GetRules()
}

// AddSmartProxyRule 添加智能代理规则
func (app *NetworkApp) AddSmartProxyRule(pattern, routeType string) error {
	return app.smartProxyMgr.AddManualRule(pattern, routeType)
}

// RemoveSmartProxyRule 删除智能代理规则
func (app *NetworkApp) RemoveSmartProxyRule(id string) error {
	return app.smartProxyMgr.RemoveRule(id)
}

// ClearAutoLearnedRules 清空自动学习的规则
func (app *NetworkApp) ClearAutoLearnedRules() error {
	return app.smartProxyMgr.ClearAutoLearnedRules()
}

// TestSmartProxyRouting 测试智能代理路由
func (app *NetworkApp) TestSmartProxyRouting(testURL string) string {
	if testURL == "" {
		testURL = "https://www.google.com"
	}

	// 解析URL获取主机名
	parsedURL, err := url.Parse(testURL)
	if err != nil {
		return fmt.Sprintf("URL解析失败: %v", err)
	}

	host := parsedURL.Host
	if host == "" {
		host = parsedURL.Path // 如果没有协议，可能整个URL在Path中
	}

	// 获取路由决策
	routeType := app.smartProxyMgr.DecideRoute(host)

	result := fmt.Sprintf("测试URL: %s\n主机: %s\n路由决策: %s\n", testURL, host, routeType)

	// 检查代理配置
	if app.proxyConfigMgr.GetConfig().Enabled {
		config := app.proxyConfigMgr.GetConfig()
		result += fmt.Sprintf("代理配置: %s:%d (类型: %s)\n", config.Host, config.Port, config.Type)

		if routeType == "proxy" {
			result += "✅ 将通过代理转发请求\n"
		} else {
			result += "⚠️ 将直连访问（不使用代理）\n"
		}
	} else {
		result += "❌ 代理未启用\n"
	}

	// 显示匹配的规则
	rules := app.smartProxyMgr.GetRules()
	result += "\n当前智能代理规则:\n"
	for _, rule := range rules {
		if rule.Enabled {
			result += fmt.Sprintf("- %s -> %s (%s)\n", rule.Pattern, rule.RouteType, rule.Source)
		}
	}

	return result
}

// GetSmartProxyManager 实现 types.AppHandler 接口
func (na *NetworkApp) GetSmartProxyManager() interface{} {
	return na.smartProxyMgr
}

// GetProxyConfigManager 实现 types.AppHandler 接口
func (na *NetworkApp) GetProxyConfigManager() interface{} {
	return na.proxyConfigMgr
}

// GetContext 实现 types.WorkflowAppHandler 接口
func (na *NetworkApp) GetContext() context.Context {
	return na.ctx
}

// SendWSMessage 实现 types.WorkflowAppHandler 接口
func (na *NetworkApp) SendWSMessage(msg types.WSMessage) {
	if na.wsServer != nil {
		na.wsServer.SendMessage(msg)
	}
}
