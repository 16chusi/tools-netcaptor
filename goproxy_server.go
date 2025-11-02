package main

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/elazarl/goproxy"
)

type GoProxyServer struct {
	port        int
	proxy       *goproxy.ProxyHttpServer
	capture     *NetworkCapture
	interceptor *Interceptor
	running     bool
	requestMap  map[string]*NetworkRequest
}

func NewGoProxyServer(port int, capture *NetworkCapture) *GoProxyServer {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false

	gps := &GoProxyServer{
		port:        port,
		proxy:       proxy,
		capture:     capture,
		interceptor: NewInterceptor(),
		requestMap:  make(map[string]*NetworkRequest),
	}

	// 启用HTTPS MITM
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

	// 拦截请求
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		gps.recordRequest(req)
		return req, nil
	})

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

func (gps *GoProxyServer) Start() error {
	gps.running = true
	go func() {
		addr := fmt.Sprintf(":%d", gps.port)
		http.ListenAndServe(addr, gps.proxy)
	}()
	return nil
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
	reqData := NetworkRequest{
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

	gps.capture.mu.Lock()
	gps.capture.requests = append(gps.capture.requests, reqData)
	gps.capture.mu.Unlock()
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
	var reqData NetworkRequest
	var duration int64 = 0
	if reqPtr, ok := gps.requestMap[url]; ok {
		reqData = *reqPtr
		duration = time.Now().UnixMilli() - reqData.Time
		delete(gps.requestMap, url)
	} else {
		// 如果 requestMap 中没有，创建一个默认请求
		reqData = NetworkRequest{
			ID:     generateID(),
			URL:    url,
			Method: req.Method,
			Type:   "proxy",
			Time:   time.Now().UnixMilli(),
			Domain: req.Host,
			Path:   req.URL.Path,
		}
	}

	respData := NetworkResponse{
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

	gps.capture.mu.Lock()
	gps.capture.responses = append(gps.capture.responses, respData)

	// 总是创建 entry
	if true {
		entry := NetworkEntry{
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
		gps.capture.entries = append(gps.capture.entries, entry)
	}

	// 限制历史记录数量，保留最近30条
	if len(gps.capture.entries) > gps.capture.maxEntries {
		gps.capture.entries = gps.capture.entries[len(gps.capture.entries)-gps.capture.maxEntries:]
	}
	if len(gps.capture.requests) > gps.capture.maxEntries {
		gps.capture.requests = gps.capture.requests[len(gps.capture.requests)-gps.capture.maxEntries:]
	}
	if len(gps.capture.responses) > gps.capture.maxEntries {
		gps.capture.responses = gps.capture.responses[len(gps.capture.responses)-gps.capture.maxEntries:]
	}

	gps.capture.mu.Unlock()
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
