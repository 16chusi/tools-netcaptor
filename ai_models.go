package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AIModel AI模型配置
type AIModel struct {
	Provider string `json:"provider"`
	Name     string `json:"name"`
	APIKey   string `json:"apiKey"`
	BaseURL  string `json:"baseUrl"`
}

// AIRequest AI请求结构
type AIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIResponse AI响应结构
type AIResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice 选择结构
type Choice struct {
	Message Message `json:"message"`
}

// AIService AI服务
type AIService struct {
	models         []AIModel
	proxyConfigMgr *ProxyConfigManager
	smartProxyMgr  *SmartProxyManager
}

// NewAIService 创建AI服务
func NewAIService(proxyConfigMgr *ProxyConfigManager, smartProxyMgr *SmartProxyManager) *AIService {
	// 添加默认模型配置，避免模型索引超出范围
	defaultModels := []AIModel{
		{
			Provider: "openai",
			Name:     "gpt-3.5-turbo",
			APIKey:   "",
			BaseURL:  "https://api.openai.com/v1",
		},
	}

	AppLog.Info(fmt.Sprintf("[AI服务] 初始化AI服务，默认模型数量: %d", len(defaultModels)))

	return &AIService{
		models:         defaultModels,
		proxyConfigMgr: proxyConfigMgr,
		smartProxyMgr:  smartProxyMgr,
	}
}

// CallAI 调用AI接口
func (s *AIService) CallAI(modelIndex int, prompt string, systemPrompt string) (string, error) {
	return s.CallAIWithRetry(modelIndex, prompt, systemPrompt, 3, 2)
}

// CallAIWithRetry 调用AI接口（带重试）
func (s *AIService) CallAIWithRetry(modelIndex int, prompt string, systemPrompt string, retryCount int, retryDelay int) (string, error) {
	return s.CallAIWithCustomSettings(modelIndex, prompt, systemPrompt, retryCount, retryDelay, nil)
}

// AICustomSettings AI自定义设置
type AICustomSettings struct {
	ThinkingMode string  `json:"thinking_mode,omitempty"`
	TopP         float64 `json:"top_p,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
	MaxTokens    int     `json:"max_tokens,omitempty"`
}

// CallAIWithImages 调用AI接口处理图片内容
func (s *AIService) CallAIWithImages(modelIndex int, prompt string, imageUrls []string, customSettings *AICustomSettings) (string, error) {
	LogDebug(fmt.Sprintf("[AI服务] 图片处理请求 - 模型索引: %d, 可用模型数量: %d", modelIndex, len(s.models)))

	if len(s.models) == 0 {
		LogError("[AI服务] ❌ 没有配置任何AI模型")
		return "", fmt.Errorf("没有配置AI模型，请先在设置中配置AI模型")
	}

	LogDebug(fmt.Sprintf("[AI服务] 可用模型列表:"))
	for i, model := range s.models {
		LogDebug(fmt.Sprintf("[AI服务] 模型 %d: %s (%s)", i, model.Name, model.Provider))
	}

	if modelIndex >= len(s.models) {
		LogError(fmt.Sprintf("[AI服务] ❌ 模型索引 %d 超出范围，可用模型: %d 个", modelIndex, len(s.models)))
		return "", fmt.Errorf("模型索引超出范围，请选择有效的模型 (0-%d)", len(s.models)-1)
	}

	model := s.models[modelIndex]

	// 构建图片消息格式
	messages := []map[string]interface{}{
		{
			"role": "user",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": prompt,
				},
			},
		},
	}

	// 添加图片内容
	userContent := messages[0]["content"].([]map[string]interface{})
	for _, imageUrl := range imageUrls {
		imageContent := map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]interface{}{
				"url": imageUrl,
			},
		}
		userContent = append(userContent, imageContent)
	}
	messages[0]["content"] = userContent

	// 构建请求体
	requestBody := map[string]interface{}{
		"model":    model.Name,
		"messages": messages,
	}

	// 应用自定义设置
	if customSettings != nil {
		if customSettings.Temperature > 0 {
			requestBody["temperature"] = customSettings.Temperature
		}
		if customSettings.TopP > 0 {
			requestBody["top_p"] = customSettings.TopP
		}
		if customSettings.MaxTokens > 0 {
			requestBody["max_tokens"] = customSettings.MaxTokens
		}
	}

	// 发送请求（这里需要根据实际的API实现）
	// 暂时返回模拟结果
	return "图片分析结果", nil
}

// CallAIWithCustomSettings 调用AI接口（带重试和自定义设置）
func (s *AIService) CallAIWithCustomSettings(modelIndex int, prompt string, systemPrompt string, retryCount int, retryDelay int, customSettings *AICustomSettings) (string, error) {
	AppLog.Info(fmt.Sprintf("[AI服务] 文本处理请求 - 模型索引: %d, 可用模型数量: %d", modelIndex, len(s.models)))

	if len(s.models) == 0 {
		AppLog.Info(fmt.Sprintf("[AI服务] ❌ 没有配置任何AI模型"))
		return "", fmt.Errorf("没有配置AI模型，请先在设置中配置AI模型")
	}

	if modelIndex >= len(s.models) {
		AppLog.Info(fmt.Sprintf("[AI服务] ❌ 模型索引 %d 超出范围，可用模型: %d 个", modelIndex, len(s.models)))
		return "", fmt.Errorf("模型索引超出范围，请选择有效的模型 (0-%d)", len(s.models)-1)
	}

	model := s.models[modelIndex]
	AppLog.Info(fmt.Sprintf("[AI服务] 使用模型: %s (%s)", model.Name, model.Provider))

	// 构建请求
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}

	request := AIRequest{
		Model:       model.Name,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	// 应用自定义设置
	if customSettings != nil {
		if customSettings.Temperature > 0 {
			request.Temperature = customSettings.Temperature
		}
		if customSettings.MaxTokens > 0 {
			request.MaxTokens = customSettings.MaxTokens
		}
	}

	var lastErr error

	// 重试逻辑 - 使用指数退避策略
	for attempt := 0; attempt <= retryCount; attempt++ {
		if attempt > 0 {
			// 指数退避：第1次重试等待retryDelay秒，第2次等待retryDelay*2秒，以此类推
			waitTime := time.Duration(retryDelay*(1<<(attempt-1))) * time.Second
			if waitTime > 60*time.Second {
				waitTime = 60 * time.Second // 最大等待60秒
			}
			AppLog.Info(fmt.Sprintf("[AI服务] 第 %d 次重试，等待 %v...", attempt, waitTime))
			time.Sleep(waitTime)
		}

		AppLog.Info(fmt.Sprintf("[AI服务] 开始第 %d 次尝试调用AI", attempt+1))

		// 统一使用OpenAI兼容接口
		result, err := s.callOpenAICompatible(model, request)

		if err == nil {
			if attempt > 0 {
				AppLog.Info(fmt.Sprintf("[AI服务] ✅ 重试成功，第 %d 次尝试", attempt+1))
			} else {
				AppLog.Info(fmt.Sprintf("[AI服务] ✅ 首次调用成功"))
			}
			return result, nil
		}

		lastErr = err
		AppLog.Info(fmt.Sprintf("[AI服务] ❌ 第 %d 次尝试失败: %v", attempt+1, err))

		// 检查是否是可重试的错误
		if !isRetryableError(err) {
			AppLog.Info(fmt.Sprintf("[AI服务] 不可重试的错误，停止重试"))
			break
		}
	}

	return "", fmt.Errorf("AI调用失败，已重试 %d 次: %v", retryCount, lastErr)
}

// isRetryableError 判断是否为可重试的错误
func isRetryableError(err error) bool {
	errStr := strings.ToLower(err.Error())

	// 网络相关错误
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "dial") ||
		strings.Contains(errStr, "reset") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "temporary failure") {
		return true
	}

	// API限制相关错误
	if strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "too many requests") ||
		strings.Contains(errStr, "quota exceeded") ||
		strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "504") ||
		strings.Contains(errStr, "gateway timeout") {
		return true
	}

	// 服务器临时错误
	if strings.Contains(errStr, "internal server error") ||
		strings.Contains(errStr, "service unavailable") ||
		strings.Contains(errStr, "bad gateway") ||
		strings.Contains(errStr, "server error") {
		return true
	}

	return false
}

// callOpenAICompatible 调用OpenAI兼容接口
func (s *AIService) callOpenAICompatible(model AIModel, request AIRequest) (string, error) {
	// 构建API URL
	baseURL := model.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	// 移除末尾的斜杠
	baseURL = strings.TrimSuffix(baseURL, "/")
	url := baseURL + "/chat/completions"

	return s.makeHTTPRequest(url, model.APIKey, request, "Bearer")
}

// makeHTTPRequest 发送HTTP请求
func (s *AIService) makeHTTPRequest(url, apiKey string, request AIRequest, authType string) (string, error) {
	// 验证和清理请求参数
	cleanedRequest := s.cleanAIRequest(request, url)

	jsonData, err := json.Marshal(cleanedRequest)
	if err != nil {
		return "", err
	}

	AppLog.Info(fmt.Sprintf("[AI服务] 请求参数: %s", string(jsonData)))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if authType == "Bearer" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	} else {
		req.Header.Set("api-key", apiKey)
	}

	// 根据不同的API端点使用智能路由和适当的超时时间
	apiURL := "https://api.openai.com"
	timeout := 60 * time.Second // 默认60秒超时

	// 针对不同API设置不同的超时时间
	if strings.Contains(url, "azure") {
		apiURL = url
		timeout = 90 * time.Second // Azure API可能需要更长时间
	} else if strings.Contains(url, "bigmodel.cn") {
		apiURL = "https://open.bigmodel.cn"
		timeout = 120 * time.Second // 智谱AI需要更长超时时间
	} else if strings.Contains(url, "anthropic.com") {
		apiURL = "https://api.anthropic.com"
		timeout = 90 * time.Second
	}

	AppLog.Info(fmt.Sprintf("[AI服务] 发送请求到: %s (超时: %v)", url, timeout))

	client := s.proxyConfigMgr.CreateSmartHTTPClient(apiURL, s.smartProxyMgr, timeout)
	resp, err := client.Do(req)
	if err != nil {
		AppLog.Info(fmt.Sprintf("[AI服务] 请求失败: %v", err))
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	AppLog.Info(fmt.Sprintf("[AI服务] 收到响应，状态码: %d", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		AppLog.Info(fmt.Sprintf("[AI服务] API调用失败，响应: %s", string(body)))
		return "", fmt.Errorf("API调用失败 (状态码: %d): %s", resp.StatusCode, string(body))
	}

	var response AIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		AppLog.Info(fmt.Sprintf("[AI服务] 响应解析失败: %v, 原始响应: %s", err, string(body)))
		return "", fmt.Errorf("响应解析失败: %v", err)
	}

	if len(response.Choices) > 0 {
		AppLog.Info(fmt.Sprintf("[AI服务] 成功获取AI响应"))
		return response.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("未收到有效响应")
}

// cleanAIRequest 清理和验证AI请求参数
func (s *AIService) cleanAIRequest(request AIRequest, url string) AIRequest {
	cleaned := request

	// 通用参数验证
	if cleaned.Temperature <= 0 || cleaned.Temperature > 2 {
		cleaned.Temperature = 0.7
	}

	if cleaned.MaxTokens <= 0 || cleaned.MaxTokens > 8192 {
		cleaned.MaxTokens = 2000
	}

	// 确保消息不为空
	if len(cleaned.Messages) == 0 {
		cleaned.Messages = []Message{
			{Role: "user", Content: "Hello"},
		}
	}

	// 检查消息内容长度
	for i, msg := range cleaned.Messages {
		if len(msg.Content) > 20000 {
			cleaned.Messages[i].Content = msg.Content[:20000] + "..."
		}
	}

	return cleaned
}

// UpdateModels 更新模型配置
func (s *AIService) UpdateModels(models []AIModel) {
	LogInfo(fmt.Sprintf("[AI服务] 更新模型配置，新模型数量: %d", len(models)))
	s.models = models

	// 如果没有配置模型，保留默认模型
	if len(s.models) == 0 {
		LogWarning("[AI服务] ⚠️ 没有提供模型配置，保持默认模型")
	} else {
		for i, model := range s.models {
			LogDebug(fmt.Sprintf("[AI服务] 模型 %d: %s (%s)", i, model.Name, model.Provider))
		}
	}
}

// GetModels 获取当前模型配置
func (s *AIService) GetModels() []AIModel {
	return s.models
}

// TestModel 测试模型连接
func (s *AIService) TestModel(model AIModel) error {
	AppLog.Info(fmt.Sprintf("[AI服务] 测试模型连接: %s", model.Name))

	testRequest := AIRequest{
		Model: model.Name,
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: 10,
	}

	_, err := s.callOpenAICompatible(model, testRequest)

	if err != nil {
		AppLog.Info(fmt.Sprintf("[AI服务] API调用失败: %v", err))
	} else {
		AppLog.Info(fmt.Sprintf("[AI服务] API调用成功"))
	}

	return err
}
