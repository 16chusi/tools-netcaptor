package types

import "time"

// types.NetworkRequest 网络请求结构
type NetworkRequest struct {
	ID              string            `json:"id"`
	URL             string            `json:"url"`
	Method          string            `json:"method"`
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
	RequestHeaders  map[string]string `json:"requestHeaders"`
	RequestBody     string            `json:"requestBody"`
	ResponseHeaders map[string]string `json:"responseHeaders"`
	ResponseBody    string            `json:"responseBody"`
	StatusCode      int               `json:"statusCode"`
	Timestamp       time.Time         `json:"timestamp"`
	Duration        int64             `json:"duration"`
	ContentType     string            `json:"contentType"`
	Size            int64             `json:"size"`
	Type            string            `json:"type"`
	Time            int64             `json:"time"`
	Domain          string            `json:"domain"`
	Path            string            `json:"path"`
}

// NetworkResponse 网络响应结构
type NetworkResponse struct {
	ID          string            `json:"id"`
	URL         string            `json:"url"`
	Status      int               `json:"status"`
	StatusText  string            `json:"statusText"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body,omitempty"`
	Size        int               `json:"size"`
	Duration    int64             `json:"duration"`
	ContentType string            `json:"contentType"`
}

// NetworkEntry 网络条目结构
type NetworkEntry struct {
	ID         string          `json:"id"`
	URL        string          `json:"url"`
	Method     string          `json:"method"`
	Status     int             `json:"status"`
	StatusText string          `json:"statusText"`
	Type       string          `json:"type"`
	Size       int             `json:"size"`
	Time       int64           `json:"time"`
	Duration   int64           `json:"duration"`
	Domain     string          `json:"domain"`
	Path       string          `json:"path"`
	Request    NetworkRequest  `json:"request"`
	Response   NetworkResponse `json:"response"`
}

// types.WSMessage WebSocket 消息结构
type WSMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
}

// types.CertInfo 证书信息
type CertInfo struct {
	Path      string `json:"path"`
	Exists    bool   `json:"exists"`
	CreatedAt string `json:"createdAt"`
	ExpiresAt string `json:"expiresAt"`
	NotBefore string `json:"notBefore"`
	NotAfter  string `json:"notAfter"`
	Issuer    string `json:"issuer"`
	Subject   string `json:"subject"`
}

// types.InterceptRule 拦截规则
type InterceptRule struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	URLPattern      string `json:"urlPattern"`
	Enabled         bool   `json:"enabled"`
	ActionType      string `json:"actionType"`
	StatusCode      int    `json:"statusCode,omitempty"`
	ContentType     string `json:"contentType,omitempty"`
	Body            string `json:"body,omitempty"`
	Headers         string `json:"headers,omitempty"`
	FindText        string `json:"findText,omitempty"`
	ReplaceText     string `json:"replaceText,omitempty"`
	UseRegex        bool   `json:"useRegex,omitempty"`
	ReplaceAll      bool   `json:"replaceAll,omitempty"`
	ResponseContent string `json:"responseContent,omitempty"`
	RedirectURL     string `json:"redirectUrl,omitempty"`
	WebhookURL      string `json:"webhookUrl,omitempty"`
	WebhookEnabled  bool   `json:"webhookEnabled,omitempty"`
	SaveToFile      bool   `json:"saveToFile,omitempty"`
	SaveFilePath    string `json:"saveFilePath,omitempty"`
	SaveFormat      string `json:"saveFormat,omitempty"`
}
