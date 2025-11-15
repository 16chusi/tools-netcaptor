package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"netcaptor/internal/types"
	"netcaptor/internal/utils"
)

// interceptRule 内部拦截规则（包含编译后的正则）
type interceptRule struct {
	types.InterceptRule
	compiledURLRegex  *regexp.Regexp
	compiledFindRegex *regexp.Regexp
}

type Interceptor struct {
	rules []interceptRule
	mu    sync.RWMutex
}

func NewInterceptor() *Interceptor {
	return &Interceptor{
		rules: []interceptRule{},
	}
}

func (i *Interceptor) SetRules(rules []interceptRule) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	for idx := range rules {
		if rules[idx].ActionType == "findReplace" && rules[idx].UseRegex && rules[idx].FindText != "" {
			compiled, err := regexp.Compile(rules[idx].FindText)
			if err != nil {
				return err
			}
			rules[idx].compiledFindRegex = compiled
		}
	}

	i.rules = rules
	return nil
}

// SetRulesFromTypes 从types.InterceptRule设置规则
func (i *Interceptor) SetRulesFromTypes(rules []types.InterceptRule) error {
	internalRules := make([]interceptRule, len(rules))
	for idx, rule := range rules {
		internalRules[idx] = interceptRule{InterceptRule: rule}
		if rule.ActionType == "findReplace" && rule.UseRegex && rule.FindText != "" {
			compiled, err := regexp.Compile(rule.FindText)
			if err != nil {
				return err
			}
			internalRules[idx].compiledFindRegex = compiled
		}
	}
	return i.SetRules(internalRules)
}

func (i *Interceptor) GetRules() []interceptRule {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.rules
}

// GetRulesAsTypes 获取types.InterceptRule格式的规则
func (i *Interceptor) GetRulesAsTypes() []types.InterceptRule {
	i.mu.RLock()
	defer i.mu.RUnlock()

	result := make([]types.InterceptRule, len(i.rules))
	for idx, rule := range i.rules {
		result[idx] = rule.InterceptRule
	}
	return result
}

func (i *Interceptor) matchURL(url string, rule interceptRule) bool {
	if !rule.Enabled {
		return false
	}
	return matchWildcard(url, rule.URLPattern)
}

func matchWildcard(text, pattern string) bool {
	if !strings.Contains(pattern, "*") {
		return text == pattern
	}

	// 转义特殊字符，但保留 *
	parts := strings.Split(pattern, "*")
	for i := range parts {
		parts[i] = regexp.QuoteMeta(parts[i])
	}
	regexPattern := "^" + strings.Join(parts, ".*") + "$"

	matched, _ := regexp.MatchString(regexPattern, text)
	return matched
}

// InterceptRequest 拦截并可能转发请求
func (i *Interceptor) InterceptRequest(req *http.Request) (*http.Response, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	url := req.URL.String()
	if url == "" {
		url = "http://" + req.Host + req.URL.Path
		if req.URL.RawQuery != "" {
			url += "?" + req.URL.RawQuery
		}
	}

	for _, rule := range i.rules {
		if !i.matchURL(url, rule) {
			continue
		}

		// 检查是否是请求转发类型
		if rule.ActionType == "forwardRequest" && rule.ForwardRequest != nil {
			utils.AppLog.Info(fmt.Sprintf("[Interceptor] ✓ 匹配到转发规则: %s, 转发到: %s", rule.Name, rule.ForwardRequest.TargetURL))
			return forwardRequest(req, rule.ForwardRequest)
		}
	}

	return nil, nil
}

func (i *Interceptor) InterceptResponse(resp *http.Response, req *http.Request) error {
	i.mu.RLock()
	defer i.mu.RUnlock()

	url := req.URL.String()
	if url == "" {
		url = "http://" + req.Host + req.URL.Path
		if req.URL.RawQuery != "" {
			url += "?" + req.URL.RawQuery
		}
	}

	utils.AppLog.Debug(fmt.Sprintf("[Interceptor] 检查URL: %s, 规则数量: %d", url, len(i.rules)))

	for _, rule := range i.rules {
		utils.AppLog.Debug(fmt.Sprintf("[Interceptor] 测试规则: %s, Pattern: %s, Enabled: %v", rule.Name, rule.URLPattern, rule.Enabled))
		if !i.matchURL(url, rule) {
			utils.AppLog.Debug(fmt.Sprintf("[Interceptor] URL不匹配: %s vs %s", url, rule.URLPattern))
			continue
		}

		utils.AppLog.Info(fmt.Sprintf("[Interceptor] ✓ 匹配到规则: %s, ActionType: %s", rule.Name, rule.ActionType))

		// Read original body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body.Close()

		bodyStr := string(bodyBytes)
		modified := false

		// Apply transformation based on action type
		switch rule.ActionType {
		case "redirect":
			if rule.RedirectURL != "" {
				resp.StatusCode = 302
				resp.Status = "302 Found"
				resp.Header.Set("Location", rule.RedirectURL)
				resp.Body = io.NopCloser(bytes.NewReader([]byte{}))
				resp.ContentLength = 0
				modified = true
			}

		case "responseReplace":
			if rule.ResponseContent != "" {
				bodyStr = rule.ResponseContent
				modified = true
			}

		case "findReplace":
			if rule.FindText != "" {
				if rule.UseRegex && rule.compiledFindRegex != nil {
					if rule.ReplaceAll {
						bodyStr = rule.compiledFindRegex.ReplaceAllString(bodyStr, rule.ReplaceText)
					} else {
						bodyStr = rule.compiledFindRegex.ReplaceAllStringFunc(bodyStr, func(s string) string {
							return rule.ReplaceText
						})
					}
				} else {
					if rule.ReplaceAll {
						bodyStr = strings.ReplaceAll(bodyStr, rule.FindText, rule.ReplaceText)
					} else {
						bodyStr = strings.Replace(bodyStr, rule.FindText, rule.ReplaceText, 1)
					}
				}
				modified = true
			}

		case "remoteHTTP":
			if rule.RemoteHTTP != nil && rule.RemoteHTTP.URL != "" {
				result, err := callRemoteHTTP(rule.RemoteHTTP, url, bodyStr)
				if err != nil {
					utils.AppLog.Info(fmt.Sprintf("[Interceptor] RemoteHTTP调用失败: %v", err))
				} else if rule.RemoteHTTP.UseResponse {
					bodyStr = result
					modified = true
					utils.AppLog.Info(fmt.Sprintf("[Interceptor] ✓ 使用RemoteHTTP响应替换 (长度: %d)", len(result)))
				}
			}
		}

		// Update response body
		if modified && rule.ActionType != "redirect" {
			newBody := []byte(bodyStr)
			resp.Body = io.NopCloser(bytes.NewReader(newBody))
			resp.ContentLength = int64(len(newBody))
			resp.Header.Del("Content-Encoding")
		} else if !modified {
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// Enhanced Webhook (sync or async)
		if rule.WebhookEnabled && rule.WebhookURL != "" {
			if rule.WebhookSync {
				// 同步调用，可以用响应替换
				result, err := sendWebhookSync(rule.WebhookURL, url, bodyStr)
				if err != nil {
					utils.AppLog.Info(fmt.Sprintf("[Interceptor] Webhook同步调用失败: %v", err))
				} else if result != "" {
					bodyStr = result
					newBody := []byte(bodyStr)
					resp.Body = io.NopCloser(bytes.NewReader(newBody))
					resp.ContentLength = int64(len(newBody))
					resp.Header.Del("Content-Encoding")
					utils.AppLog.Info(fmt.Sprintf("[Interceptor] ✓ 使用Webhook响应替换 (长度: %d)", len(result)))
				}
			} else {
				// 异步通知
				go sendWebhook(rule.WebhookURL, url, bodyStr)
			}
		}

		// Save to file (async)
		if (rule.ActionType == "saveToFile" || rule.SaveToFile) && rule.SaveFilePath != "" {
			utils.AppLog.Info(fmt.Sprintf("[Interceptor] 触发保存到文件: %s", rule.SaveFilePath))
			go saveResponseToFile(rule.SaveFilePath, bodyStr, rule.SaveFormat)
		}

		break // Only apply first matching rule
	}

	return nil
}

func sendWebhook(webhookURL, requestURL, body string) {
	payload := map[string]string{
		"url":  requestURL,
		"body": body,
	}
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		utils.AppLog.Info(fmt.Sprintf("[Interceptor] Webhook发送失败: %v", err))
		return
	}
	defer resp.Body.Close()
	utils.AppLog.Info(fmt.Sprintf("[Interceptor] ✓ Webhook已发送到: %s", webhookURL))
}

func sendWebhookSync(webhookURL, requestURL, body string) (string, error) {
	payload := map[string]string{
		"url":  requestURL,
		"body": body,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("webhook返回状态码: %d", resp.StatusCode)
	}

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	utils.AppLog.Info(fmt.Sprintf("[Interceptor] ✓ Webhook同步调用成功，响应长度: %d", len(result)))
	return string(result), nil
}

func forwardRequest(originalReq *http.Request, config *types.ForwardConfig) (*http.Response, error) {
	// 解析目标URL
	targetURL, err := url.Parse(config.TargetURL)
	if err != nil {
		return nil, fmt.Errorf("无效的目标URL: %w", err)
	}

	// 构建新的URL
	newURL := *originalReq.URL
	newURL.Scheme = targetURL.Scheme
	newURL.Host = targetURL.Host

	if config.KeepPath {
		// 保持原路径
		newURL.Path = originalReq.URL.Path
	} else {
		// 使用目标URL的路径
		newURL.Path = targetURL.Path
	}

	// 读取原始请求体
	var bodyBytes []byte
	if originalReq.Body != nil {
		bodyBytes, _ = io.ReadAll(originalReq.Body)
		originalReq.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// 创建新请求
	newReq, err := http.NewRequest(originalReq.Method, newURL.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	// 复制原始请求头
	for key, values := range originalReq.Header {
		for _, value := range values {
			newReq.Header.Add(key, value)
		}
	}

	// 替换Host头
	if config.ReplaceHost {
		newReq.Host = targetURL.Host
		newReq.Header.Set("Host", targetURL.Host)
	}

	// 应用自定义请求头
	if len(config.ReplaceHeaders) > 0 {
		for key, value := range config.ReplaceHeaders {
			newReq.Header.Set(key, value)
		}
	}

	// 设置超时
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30000 // 默认30秒
	}
	client := &http.Client{Timeout: time.Duration(timeout) * time.Millisecond}

	// 发送请求
	utils.AppLog.Info(fmt.Sprintf("[Interceptor] 🔄 转发请求: %s → %s", originalReq.URL.String(), newURL.String()))
	resp, err := client.Do(newReq)
	if err != nil {
		return nil, fmt.Errorf("转发请求失败: %w", err)
	}

	utils.AppLog.Info(fmt.Sprintf("[Interceptor] ✓ 转发成功，状态码: %d", resp.StatusCode))
	return resp, nil
}

func callRemoteHTTP(config *types.RemoteHTTPConfig, requestURL, originalBody string) (string, error) {
	// 设置默认值
	method := config.Method
	if method == "" {
		method = "POST"
	}
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5000
	}

	// 准备请求体
	var reqBody io.Reader
	if config.SendOriginal {
		reqBody = bytes.NewBufferString(originalBody)
	} else if config.BodyTemplate != "" {
		// 简单的模板替换
		body := strings.ReplaceAll(config.BodyTemplate, "{{url}}", requestURL)
		body = strings.ReplaceAll(body, "{{body}}", originalBody)
		reqBody = bytes.NewBufferString(body)
	} else {
		// 默认发送JSON格式
		payload := map[string]string{
			"url":  requestURL,
			"body": originalBody,
		}
		jsonData, _ := json.Marshal(payload)
		reqBody = bytes.NewBuffer(jsonData)
	}

	// 创建请求
	client := &http.Client{Timeout: time.Duration(timeout) * time.Millisecond}
	req, err := http.NewRequest(method, config.URL, reqBody)
	if err != nil {
		return "", err
	}

	// 设置请求头
	if len(config.Headers) > 0 {
		for k, v := range config.Headers {
			req.Header.Set(k, v)
		}
	} else {
		req.Header.Set("Content-Type", "application/json")
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("远程HTTP返回状态码: %d", resp.StatusCode)
	}

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	utils.AppLog.Info(fmt.Sprintf("[Interceptor] ✓ RemoteHTTP调用成功: %s, 响应长度: %d", config.URL, len(result)))
	return string(result), nil
}

func saveResponseToFile(filePath, body, format string) {
	utils.AppLog.Info(fmt.Sprintf("[Interceptor] saveResponseToFile 被调用: filePath=%s, bodyLen=%d, format=%s", filePath, len(body), format))

	if format == "" {
		format = "jsonl"
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err == nil {
		utils.AppLog.Info(fmt.Sprintf("[Interceptor] 文件已存在，当前大小: %d 字节", fileInfo.Size()))
	} else {
		utils.AppLog.Info(fmt.Sprintf("[Interceptor] 文件不存在，将创建新文件"))
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		utils.AppLog.Info(fmt.Sprintf("[Interceptor] 保存文件失败: %v", err))
		return
	}
	defer file.Close()

	var content string
	if format == "jsonl" {
		content = body + "\n"
	} else {
		content = body + "\n"
	}

	n, err := file.WriteString(content)
	if err != nil {
		utils.AppLog.Info(fmt.Sprintf("[Interceptor] 写入文件失败: %v", err))
	} else {
		utils.AppLog.Info(fmt.Sprintf("[Interceptor] ✓ 响应已追加到: %s (写入: %d 字节)", filePath, n))
		// 再次检查文件大小
		if fileInfo2, err := os.Stat(filePath); err == nil {
			utils.AppLog.Info(fmt.Sprintf("[Interceptor] 写入后文件大小: %d 字节", fileInfo2.Size()))
		}
	}
}
