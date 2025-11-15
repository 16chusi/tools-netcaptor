package network

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"netcaptor/internal/utils"
	"os"
	"path/filepath"
	"sync"
)

// WebhookServer Webhook服务器
type WebhookServer struct {
	port     int
	running  bool
	mu       sync.Mutex
	listener net.Listener
}

// WebhookRequest Webhook请求参数
type WebhookRequest struct {
	Action string `json:"action"`
	File   string `json:"file"`
	Data   string `json:"data"`
	Type   string `json:"type"`
}

// NewWebhookServer 创建Webhook服务器
func NewWebhookServer() *WebhookServer {
	return &WebhookServer{}
}

// Start 启动服务器
func (ws *WebhookServer) Start() error {
	ws.mu.Lock()
	if ws.running {
		ws.mu.Unlock()
		return fmt.Errorf("Webhook服务器已在运行")
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		ws.mu.Unlock()
		return fmt.Errorf("启动Webhook服务器失败: %w", err)
	}

	ws.listener = listener
	ws.port = listener.Addr().(*net.TCPAddr).Port
	ws.running = true
	ws.mu.Unlock()

	utils.AppLog.Info(fmt.Sprintf("[Webhook] 服务器启动在 http://127.0.0.1:%d", ws.port))

	mux := http.NewServeMux()
	mux.HandleFunc("/", ws.handleTestPage)
	mux.HandleFunc("/webhook", ws.handleWebhook)
	mux.HandleFunc("/api/test", ws.handleAPITest)
	mux.HandleFunc("/api/data", ws.handleAPIData)

	go func() {
		if err := http.Serve(listener, mux); err != nil {
			utils.AppLog.Info(fmt.Sprintf("[Webhook] 服务器错误: %v", err))
		}
	}()

	return nil
}

// Stop 停止服务器
func (ws *WebhookServer) Stop() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if !ws.running {
		return nil
	}

	if ws.listener != nil {
		ws.listener.Close()
	}

	ws.running = false
	utils.AppLog.Info(fmt.Sprintf("[Webhook] 服务器已停止"))
	return nil
}

// IsRunning 是否运行中
func (ws *WebhookServer) IsRunning() bool {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.running
}

// GetPort 获取端口
func (ws *WebhookServer) GetPort() int {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.port
}

// handleWebhook 处理Webhook请求
func (ws *WebhookServer) handleWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == "GET" {
		query := r.URL.Query()
		req := WebhookRequest{
			Action: query.Get("action"),
			File:   query.Get("file"),
			Data:   query.Get("data"),
			Type:   query.Get("type"),
		}

		if req.Data == "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "ok",
				"message": "Webhook服务运行中",
				"port":    ws.port,
			})
			return
		}

		if req.Action == "" {
			req.Action = "print"
		}
		if req.Type == "" {
			req.Type = "txt"
		}

		var result string
		var processErr error

		switch req.Action {
		case "save":
			result, processErr = ws.handleSave(req)
		case "print":
			result, processErr = ws.handlePrint(req)
		default:
			http.Error(w, "未知的action: "+req.Action, http.StatusBadRequest)
			return
		}

		if processErr != nil {
			http.Error(w, processErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": result,
		})
		return
	}

	if r.Method != "POST" {
		http.Error(w, "只支持GET和POST请求", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "读取请求失败", http.StatusBadRequest)
		return
	}

	var req WebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "解析JSON失败", http.StatusBadRequest)
		return
	}

	if req.Action == "" {
		req.Action = "print"
	}
	if req.Type == "" {
		req.Type = "txt"
	}

	var result string
	var processErr error

	switch req.Action {
	case "save":
		result, processErr = ws.handleSave(req)
	case "print":
		result, processErr = ws.handlePrint(req)
	default:
		http.Error(w, "未知的action: "+req.Action, http.StatusBadRequest)
		return
	}

	if processErr != nil {
		http.Error(w, processErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": result,
	})
}

// handleSave 保存到文件
func (ws *WebhookServer) handleSave(req WebhookRequest) (string, error) {
	if req.File == "" {
		return "", fmt.Errorf("缺少file参数")
	}

	data, err := ws.decodeData(req.Data, req.Type)
	if err != nil {
		return "", fmt.Errorf("解码数据失败: %w", err)
	}

	filename := req.File
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(".", filename)
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}
	if _, err := f.WriteString("\n"); err != nil {
		return "", fmt.Errorf("写入换行符失败: %w", err)
	}

	utils.AppLog.Info(fmt.Sprintf("[Webhook] 追加文件: %s (%d bytes)", filename, len(data)))
	return fmt.Sprintf("文件已追加: %s", filename), nil
}

// handlePrint 打印到控制台
func (ws *WebhookServer) handlePrint(req WebhookRequest) (string, error) {
	data, err := ws.decodeData(req.Data, req.Type)
	if err != nil {
		return "", fmt.Errorf("解码数据失败: %w", err)
	}

	utils.AppLog.Info(fmt.Sprintf("[Webhook] 打印数据 (%d bytes):\n%s", len(data), string(data)))
	return "数据已打印到控制台", nil
}

// decodeData 解码数据
func (ws *WebhookServer) decodeData(data, dataType string) ([]byte, error) {
	switch dataType {
	case "base64":
		return base64.StdEncoding.DecodeString(data)
	case "hex":
		return hex.DecodeString(data)
	case "txt", "json", "":
		return []byte(data), nil
	default:
		return nil, fmt.Errorf("不支持的数据类型: %s", dataType)
	}
}

// handleTestPage 测试页面
func (ws *WebhookServer) handleTestPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>测试页面</title>
    <style>
        body { font-family: Arial; padding: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; }
        button { padding: 12px 24px; margin: 10px; border: none; border-radius: 5px; cursor: pointer; font-size: 14px; }
        .btn-primary { background: #007bff; color: white; }
        #result { margin-top: 20px; padding: 15px; background: #f8f9fa; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🧪 网络请求测试页面</h1>
        <p>这个页面运行在 http://localhost:%d</p>
        <button class="btn-primary" onclick="testFetch()">测试 Fetch</button>
        <button class="btn-primary" onclick="testXHR()">测试 XHR</button>
        <button class="btn-primary" onclick="testWebhook()">测试 Webhook</button>
        <div id="result"></div>
    </div>
    <script>
        function log(msg) {
            document.getElementById('result').innerHTML += '<div>' + msg + '</div>';
        }
        
        async function testFetch() {
            log('发送 Fetch 请求...');
            try {
                const resp = await fetch('/api/test');
                const data = await resp.json();
                log('✅ Fetch 成功: ' + JSON.stringify(data));
            } catch (e) {
                log('❌ Fetch 失败: ' + e.message);
            }
        }
        
        function testXHR() {
            log('发送 XHR 请求...');
            const xhr = new XMLHttpRequest();
            xhr.open('GET', '/api/data');
            xhr.onload = function() {
                log('✅ XHR 成功: ' + xhr.responseText);
            };
            xhr.onerror = function() {
                log('❌ XHR 失败');
            };
            xhr.send();
        }
        
        async function testWebhook() {
            log('发送 Webhook 请求...');
            try {
                const resp = await fetch('/webhook', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        data: 'Webhook 测试消息: ' + new Date().toISOString()
                    })
                });
                const data = await resp.json();
                log('✅ Webhook 成功: ' + data.message);
            } catch (e) {
                log('❌ Webhook 失败: ' + e.message);
            }
        }
        
        log('页面加载完成,可以开始测试');
    </script>
</body>
</html>
`, ws.port)
}

// handleAPITest API测试接口
func (ws *WebhookServer) handleAPITest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, `{"message":"Hello from API","timestamp":%d}`, r.Context().Value("time"))
}

// handleAPIData API数据接口
func (ws *WebhookServer) handleAPIData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, `{"data":"Test data","method":"%s"}`, r.Method)
}
