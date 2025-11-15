package proxy

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	"netcaptor/internal/types"
	"netcaptor/internal/utils"
)

type GoProxyServer struct {
	port        int
	proxy       *goproxy.ProxyHttpServer
	capture     types.CaptureHandler
	interceptor *Interceptor
	running     bool
	requestMap  map[string]*types.NetworkRequest
	smartProxy  *SmartProxyManager
	appHandler  types.AppHandler // 使用接口
	getWSPort   func() int       // 获取WebSocket端口的函数
}

func NewGoProxyServer(port int, capture types.CaptureHandler) *GoProxyServer {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false

	gps := &GoProxyServer{
		port:        port,
		proxy:       proxy,
		capture:     capture,
		interceptor: NewInterceptor(),
		requestMap:  make(map[string]*types.NetworkRequest),
	}

	// 设置自定义CA证书
	certManager, err := NewCertManager()
	if err == nil {
		// 使用我们的CA证书替代默认证书
		goproxy.GoproxyCa = *certManager.GetCACert()
	}

	// 启用HTTPS MITM
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	// 拦截请求
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		host := req.Host
		if host == "" && req.URL != nil {
			host = req.URL.Host
		}

		utils.AppLog.Debug(fmt.Sprintf("[GoProxy Request] 收到请求: %s %s (Host: %s)\n", req.Method, req.URL.String(), host))

		// 调试智能代理状态
		utils.AppLog.Info(fmt.Sprintf("[GoProxy Request] 🔍 本地智能代理管理器状态: %v\n", gps.smartProxy != nil))
		utils.AppLog.Info(fmt.Sprintf("[GoProxy Request] 🔍 NetworkApp状态: %v\n", gps.appHandler != nil))
		if gps.appHandler != nil {
			utils.AppLog.Info(fmt.Sprintf("[GoProxy Request] 🔍 NetworkApp.smartProxyMgr状态: %v\n", gps.appHandler.GetSmartProxyManager() != nil))
			utils.AppLog.Info(fmt.Sprintf("[GoProxy Request] 🔍 NetworkApp.proxyConfigMgr状态: %v\n", gps.appHandler.GetProxyConfigManager() != nil))
		}

		// 临时创建智能代理管理器进行测试
		if gps.appHandler != nil {
			if proxyMgr, ok := gps.appHandler.GetProxyConfigManager().(*ProxyConfigManager); ok {
				// 创建临时的智能代理管理器
				tempSmartProxy := NewSmartProxyManager(proxyMgr)

				routeType := tempSmartProxy.DecideRoute(host)
				utils.AppLog.Info(fmt.Sprintf("[GoProxy Request] 🧪 临时智能路由决策: %s -> %s\n", host, routeType))

				// 打印所有规则
				rules := tempSmartProxy.GetRules()
				utils.AppLog.Info(fmt.Sprintf("[GoProxy Request] 📋 当前规则数量: %d\n", len(rules)))
				for i, rule := range rules {
					utils.AppLog.Debug(fmt.Sprintf("[GoProxy Request] 规则%d: %s -> %s (%s, enabled: %v)\n",
						i, rule.Pattern, rule.RouteType, rule.Source, rule.Enabled))
				}

				// 检查代理配置
				config := proxyMgr.GetConfig()
				utils.AppLog.Debug(fmt.Sprintf("[GoProxy Request] 代理配置: enabled=%v, host=%s, port=%d\n",
					config.Enabled, config.Host, config.Port))

				if routeType == "proxy" && config.Enabled {
					proxyURL := fmt.Sprintf("http://%s:%d", config.Host, config.Port)
					utils.AppLog.Info(fmt.Sprintf("[GoProxy Request] ✅ 需要使用上游代理: %s\n", proxyURL))

					// 动态设置全局传输层的代理
					parsedURL, err := url.Parse(proxyURL)
					if err == nil {
						gps.proxy.Tr.Proxy = http.ProxyURL(parsedURL)
						utils.AppLog.Info(fmt.Sprintf("[GoProxy Request] 🔄 全局代理设置成功\n"))
					} else {
						utils.AppLog.Info(fmt.Sprintf("[GoProxy Request] ❌ 代理URL解析失败: %v\n", err))
					}
				} else {
					utils.AppLog.Info(fmt.Sprintf("[GoProxy Request] ℹ️ 直连访问: %s (routeType=%s, proxyEnabled=%v)\n", host, routeType, config.Enabled))
					// 设置为直连
					gps.proxy.Tr.Proxy = nil
				}
			}
		}

		gps.recordRequest(req)
		return req, nil
	})

	// 拦截CONNECT请求（HTTPS）
	proxy.OnRequest().HandleConnect(goproxy.FuncHttpsHandler(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		utils.AppLog.Info(fmt.Sprintf("[GoProxy CONNECT] 🔒 HTTPS连接请求: %s\n", host))

		// 检查是否是Google相关域名
		if strings.Contains(host, "google") {
			utils.AppLog.Info(fmt.Sprintf("[GoProxy CONNECT] 🎯 检测到Google域名: %s\n", host))
		}

		return goproxy.OkConnect, host
	}))

	// 拦截响应
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp != nil {
			// Apply intercept rules
			gps.interceptor.InterceptResponse(resp, ctx.Req)
			gps.recordResponse(ctx.Req, resp)
		}
		return resp
	})

	return gps
}

// SetupSmartProxy 设置智能代理（必须在设置smartProxy字段后调用）
func (gps *GoProxyServer) SetupSmartProxy() {
	utils.AppLog.Info(fmt.Sprintf("[GoProxy] 🔧 开始设置智能代理传输层\n"))
	utils.AppLog.Info(fmt.Sprintf("[GoProxy] 🔍 smartProxy指针: %p\n", gps.smartProxy))

	// 设置自定义传输层来支持智能代理
	gps.proxy.Tr = &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			host := req.Host
			if host == "" && req.URL != nil {
				host = req.URL.Host
			}

			utils.AppLog.Info(fmt.Sprintf("[GoProxy Transport] 🚀 传输层被调用: %s (Host: %s)\n", req.URL.String(), host))

			if gps.smartProxy != nil {
				routeType := gps.smartProxy.DecideRoute(host)
				utils.AppLog.Info(fmt.Sprintf("[GoProxy Transport] 智能路由决策: %s -> %s\n", host, routeType))

				if routeType == "proxy" && gps.smartProxy.proxyConfigMgr != nil {
					config := gps.smartProxy.proxyConfigMgr.GetConfig()
					if config.Enabled {
						proxyURL := fmt.Sprintf("http://%s:%d", config.Host, config.Port)
						utils.AppLog.Info(fmt.Sprintf("[GoProxy Transport] ✅ 使用代理转发: %s\n", proxyURL))

						parsedURL, err := url.Parse(proxyURL)
						if err != nil {
							utils.AppLog.Info(fmt.Sprintf("[GoProxy Transport] ❌ 代理URL解析失败: %v\n", err))
							return nil, err
						}
						return parsedURL, nil
					} else {
						utils.AppLog.Info(fmt.Sprintf("[GoProxy Transport] ⚠️ 代理未启用，使用直连\n"))
					}
				} else {
					utils.AppLog.Info(fmt.Sprintf("[GoProxy Transport] ℹ️ 直连访问: %s\n", host))
				}
			} else {
				utils.AppLog.Info(fmt.Sprintf("[GoProxy Transport] ⚠️ 智能代理管理器未初始化\n"))
			}

			utils.AppLog.Info(fmt.Sprintf("[GoProxy Transport] 🔄 使用直连\n"))
			return nil, nil // 直连
		},
	}

	utils.AppLog.Info(fmt.Sprintf("[GoProxy] ✅ 智能代理传输层设置完成\n"))
}

func (gps *GoProxyServer) Start() error {
	gps.running = true
	utils.AppLog.Info(fmt.Sprintf("[GoProxy] 🚀 代理服务器启动中，端口: %d\n", gps.port))

	// 创建包装handler，处理API请求和代理请求
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 处理API请求
		if r.URL.Path == "/api/ws-port" {
			gps.handleAPIWSPort(w, r)
			return
		}
		if r.URL.Path == "/api/status" {
			gps.handleAPIStatus(w, r)
			return
		}

		// 其他请求交给代理处理
		gps.proxy.ServeHTTP(w, r)
	})

	go func() {
		addr := fmt.Sprintf(":%d", gps.port)
		utils.AppLog.Info(fmt.Sprintf("[GoProxy] 📡 代理服务器监听地址: %s\n", addr))
		err := http.ListenAndServe(addr, handler)
		if err != nil {
			utils.AppLog.Info(fmt.Sprintf("[GoProxy] ❌ 代理服务器启动失败: %v\n", err))
		}
	}()
	utils.AppLog.Info(fmt.Sprintf("[GoProxy] ✅ 代理服务器已启动: http://127.0.0.1:%d\n", gps.port))
	return nil
}

// handleAPIWSPort 处理获取WS端口的API请求
func (gps *GoProxyServer) handleAPIWSPort(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	wsPort := 0
	if gps.getWSPort != nil {
		wsPort = gps.getWSPort()
	}

	fmt.Fprintf(w, `{"wsPort":%d}`, wsPort)
}

// handleAPIStatus 处理获取状态的API请求
func (gps *GoProxyServer) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	wsPort := 0
	if gps.getWSPort != nil {
		wsPort = gps.getWSPort()
	}

	fmt.Fprintf(w, `{"proxy":true,"ws":%v,"wsPort":%d}`, wsPort > 0, wsPort)
}

// SetWSPortGetter 设置获取WebSocket端口的函数
func (gps *GoProxyServer) SetWSPortGetter(getter func() int) {
	gps.getWSPort = getter
}

func (gps *GoProxyServer) Stop() error {
	gps.running = false
	return nil
}

func (gps *GoProxyServer) IsRunning() bool {
	return gps.running
}

func (gps *GoProxyServer) GetPort() int {
	return gps.port
}

func (gps *GoProxyServer) GetProxyURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", gps.port)
}

func (gps *GoProxyServer) recordRequest(req *http.Request) {
	headers := make(map[string]string)
	for k, v := range req.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	body := ""
	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		body = string(bodyBytes)
		if len(body) > 1000 {
			body = body[:1000]
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	url := req.URL.String()
	if req.URL.Scheme == "" {
		url = "http://" + req.Host + req.URL.Path
		if req.URL.RawQuery != "" {
			url += "?" + req.URL.RawQuery
		}
	}

	reqID := generateID()
	reqData := types.NetworkRequest{
		ID:      reqID,
		URL:     url,
		Method:  req.Method,
		Headers: headers,
		Body:    body,
		Type:    "proxy",
		Time:    time.Now().UnixMilli(),
		Domain:  req.Host,
		Path:    req.URL.Path,
	}

	// 保存到 requestMap 以便响应时查找
	gps.requestMap[url] = &reqData

	gps.capture.AddRequest(&reqData)
}

func (gps *GoProxyServer) recordResponse(req *http.Request, resp *http.Response) {
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)
	fullBody := body
	if len(body) > 5000 {
		body = body[:5000]
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	url := req.URL.String()
	if req.URL.Scheme == "" {
		url = "http://" + req.Host + req.URL.Path
		if req.URL.RawQuery != "" {
			url += "?" + req.URL.RawQuery
		}
	}

	contentType := resp.Header.Get("Content-Type")

	// 从 requestMap 查找请求信息
	var reqData types.NetworkRequest
	var duration int64 = 0
	if reqPtr, ok := gps.requestMap[url]; ok {
		reqData = *reqPtr
		duration = time.Now().UnixMilli() - reqData.Time
		delete(gps.requestMap, url)
	} else {
		// 如果 requestMap 中没有，创建一个默认请求
		reqData = types.NetworkRequest{
			ID:     generateID(),
			URL:    url,
			Method: req.Method,
			Type:   "proxy",
			Time:   time.Now().UnixMilli(),
			Domain: req.Host,
			Path:   req.URL.Path,
		}
	}

	respData := types.NetworkResponse{
		ID:          generateID(),
		URL:         url,
		Status:      resp.StatusCode,
		StatusText:  resp.Status,
		Headers:     headers,
		Body:        fullBody,
		Size:        len(bodyBytes),
		Duration:    duration,
		ContentType: contentType,
	}

	gps.capture.AddResponse(&respData)

	// 总是创建 entry
	if true {
		entry := types.NetworkEntry{
			ID:         reqData.ID,
			URL:        url,
			Method:     reqData.Method,
			Status:     resp.StatusCode,
			StatusText: resp.Status,
			Type:       reqData.Type,
			Size:       len(bodyBytes),
			Time:       reqData.Time,
			Duration:   duration,
			Domain:     reqData.Domain,
			Path:       reqData.Path,
			Request:    reqData,
			Response:   respData,
		}
		gps.capture.AddEntry(&entry)
	}

}

func (gps *GoProxyServer) GetCACertPath() string {
	// 导出goproxy的CA证书
	homeDir, _ := os.UserHomeDir()
	certDir := filepath.Join(homeDir, ".netcaptor", "certs")
	os.MkdirAll(certDir, 0755)

	certPath := filepath.Join(certDir, "netcaptor-ca.crt")

	// 保存证书
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: goproxy.GoproxyCa.Certificate[0],
	})

	os.WriteFile(certPath, certPEM, 0644)
	return certPath
}

// SetAppHandler 设置应用处理器
func (gps *GoProxyServer) SetAppHandler(handler types.AppHandler) {
	gps.appHandler = handler
}

// GetInterceptor 获取拦截器
func (gps *GoProxyServer) GetInterceptor() *Interceptor {
	return gps.interceptor
}

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
