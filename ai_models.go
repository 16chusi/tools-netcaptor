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
	models []AIModel
}

// NewAIService 创建AI服务
func NewAIService() *AIService {
	return &AIService{
		models: []AIModel{},
	}
}

// CallAI 调用AI接口
func (s *AIService) CallAI(modelIndex int, prompt string, systemPrompt string) (string, error) {
	return s.CallAIWithRetry(modelIndex, prompt, systemPrompt, 3, 2)
}

// CallAIWithRetry 调用AI接口（带重试）
func (s *AIService) CallAIWithRetry(modelIndex int, prompt string, systemPrompt string, retryCount int, retryDelay int) (string, error) {
	if modelIndex >= len(s.models) {
		return "", fmt.Errorf("模型索引超出范围")
	}

	model := s.models[modelIndex]
	
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

	var lastErr error
	
	// 重试逻辑
	for attempt := 0; attempt <= retryCount; attempt++ {
		if attempt > 0 {
			fmt.Printf("[AI服务] 第 %d 次重试，等待 %d 秒...\n", attempt, retryDelay)
			time.Sleep(time.Duration(retryDelay) * time.Second)
		}
		
		// 根据供应商调用不同的API
		var result string
		var err error
		
		switch model.Provider {
		case "openai":
			result, err = s.callOpenAI(model, request)
		case "anthropic":
			result, err = s.callAnthropic(model, request)
		case "azure":
			result, err = s.callAzureOpenAI(model, request)
		default:
			result, err = s.callCustomAPI(model, request)
		}
		
		if err == nil {
			if attempt > 0 {
				fmt.Printf("[AI服务] 重试成功，第 %d 次尝试\n", attempt+1)
			}
			return result, nil
		}
		
		lastErr = err
		fmt.Printf("[AI服务] 第 %d 次尝试失败: %v\n", attempt+1, err)
		
		// 检查是否是可重试的错误
		if !isRetryableError(err) {
			fmt.Printf("[AI服务] 不可重试的错误，停止重试\n")
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
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "network") {
		return true
	}
	
	// API限制相关错误
	if strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "too many requests") ||
		strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "504") {
		return true
	}
	
	// 服务器临时错误
	if strings.Contains(errStr, "internal server error") ||
		strings.Contains(errStr, "service unavailable") {
		return true
	}
	
	return false
}

// callOpenAI 调用OpenAI API
func (s *AIService) callOpenAI(model AIModel, request AIRequest) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"
	if model.BaseURL != "" {
		url = model.BaseURL + "/chat/completions"
	}

	return s.makeHTTPRequest(url, model.APIKey, request, "Bearer")
}

// callAnthropic 调用Anthropic API
func (s *AIService) callAnthropic(model AIModel, request AIRequest) (string, error) {
	url := "https://api.anthropic.com/v1/messages"
	if model.BaseURL != "" {
		url = model.BaseURL + "/messages"
	}

	// Anthropic使用不同的请求格式
	anthropicRequest := map[string]interface{}{
		"model":      request.Model,
		"max_tokens": request.MaxTokens,
		"messages":   request.Messages,
	}

	jsonData, err := json.Marshal(anthropicRequest)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", model.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API调用失败: %s", string(body))
	}

	var response AIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("未收到有效响应")
}

// callAzureOpenAI 调用Azure OpenAI API
func (s *AIService) callAzureOpenAI(model AIModel, request AIRequest) (string, error) {
	if model.BaseURL == "" {
		return "", fmt.Errorf("Azure OpenAI需要配置BaseURL")
	}

	url := model.BaseURL + "/openai/deployments/" + request.Model + "/chat/completions?api-version=2023-12-01-preview"
	return s.makeHTTPRequest(url, model.APIKey, request, "api-key")
}

// callCustomAPI 调用自定义API
func (s *AIService) callCustomAPI(model AIModel, request AIRequest) (string, error) {
	if model.BaseURL == "" {
		return "", fmt.Errorf("自定义API需要配置BaseURL")
	}

	url := model.BaseURL + "/chat/completions"
	return s.makeHTTPRequest(url, model.APIKey, request, "Bearer")
}

// makeHTTPRequest 发送HTTP请求
func (s *AIService) makeHTTPRequest(url, apiKey string, request AIRequest, authType string) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

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

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API调用失败: %s", string(body))
	}

	var response AIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("未收到有效响应")
}

// UpdateModels 更新模型配置
func (s *AIService) UpdateModels(models []AIModel) {
	s.models = models
}

// TestModel 测试模型连接
func (s *AIService) TestModel(model AIModel) error {
	fmt.Printf("[AI服务] 测试模型连接: %s (%s)\n", model.Name, model.Provider)
	
	testRequest := AIRequest{
		Model: model.Name,
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: 10,
	}

	var err error
	switch model.Provider {
	case "openai":
		fmt.Printf("[AI服务] 调用OpenAI API\n")
		_, err = s.callOpenAI(model, testRequest)
	case "anthropic":
		fmt.Printf("[AI服务] 调用Anthropic API\n")
		_, err = s.callAnthropic(model, testRequest)
	case "azure":
		fmt.Printf("[AI服务] 调用Azure OpenAI API\n")
		_, err = s.callAzureOpenAI(model, testRequest)
	default:
		fmt.Printf("[AI服务] 调用自定义API\n")
		_, err = s.callCustomAPI(model, testRequest)
	}
	
	if err != nil {
		fmt.Printf("[AI服务] API调用失败: %v\n", err)
	} else {
		fmt.Printf("[AI服务] API调用成功\n")
	}
	
	return err
}
