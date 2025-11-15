package proxy

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"netcaptor/internal/types"
	"netcaptor/internal/utils"
)

type ProxyServer struct {
	port        int
	listener    net.Listener
	capture     types.CaptureHandler
	certManager *CertManager
	running     bool
	mu          sync.Mutex
}

func NewProxyServer(port int, capture types.CaptureHandler) *ProxyServer {
	certManager, err := NewCertManager()
	if err != nil {
		utils.AppLog.Info(fmt.Sprintf("证书管理器初始化失败: %v\n", err))
	}
	return &ProxyServer{
		port:        port,
		capture:     capture,
		certManager: certManager,
	}
}

func (ps *ProxyServer) Start() error {
	ps.mu.Lock()
	if ps.running {
		ps.mu.Unlock()
		return fmt.Errorf("代理服务器已在运行")
	}
	ps.mu.Unlock()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", ps.port))
	if err != nil {
		return fmt.Errorf("启动代理服务器失败: %v", err)
	}

	ps.mu.Lock()
	ps.listener = listener
	ps.running = true
	ps.mu.Unlock()

	go ps.acceptConnections()
	return nil
}

func (ps *ProxyServer) Stop() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if !ps.running {
		return nil
	}

	ps.running = false
	if ps.listener != nil {
		return ps.listener.Close()
	}
	return nil
}

func (ps *ProxyServer) IsRunning() bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.running
}

func (ps *ProxyServer) GetPort() int {
	return ps.port
}

func (ps *ProxyServer) acceptConnections() {
	for {
		conn, err := ps.listener.Accept()
		if err != nil {
			ps.mu.Lock()
			running := ps.running
			ps.mu.Unlock()
			if !running {
				return
			}
			continue
		}
		go ps.handleConnection(conn)
	}
}

func (ps *ProxyServer) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	reader := bufio.NewReader(clientConn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		return
	}

	// 修正请求URL
	if req.URL.Scheme == "" {
		req.URL.Scheme = "http"
	}
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}

	// 记录请求
	ps.recordRequest(req)

	if req.Method == http.MethodConnect {
		ps.handleHTTPS(clientConn, req)
	} else {
		ps.handleHTTP(clientConn, req)
	}
}

func (ps *ProxyServer) handleHTTP(clientConn net.Conn, req *http.Request) {
	// 转发HTTP请求
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// 读取响应体用于记录
	bodyBytes, _ := io.ReadAll(resp.Body)

	// 记录响应
	ps.recordResponseWithBody(req.URL.String(), resp, bodyBytes)

	// 写回状态行
	clientConn.Write([]byte(fmt.Sprintf("HTTP/%d.%d %d %s\r\n", resp.ProtoMajor, resp.ProtoMinor, resp.StatusCode, resp.Status)))

	// 写回响应头
	for k, v := range resp.Header {
		for _, vv := range v {
			clientConn.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, vv)))
		}
	}

	// 空行分隔头和体
	clientConn.Write([]byte("\r\n"))

	// 写回响应体
	clientConn.Write(bodyBytes)
}

func (ps *ProxyServer) handleHTTPS(clientConn net.Conn, req *http.Request) {
	if ps.certManager == nil {
		// 降级为普通隧道
		ps.handleHTTPSTunnel(clientConn, req)
		return
	}

	// 告诉客户端隧道已建立
	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// 获取主机名
	host, _, _ := net.SplitHostPort(req.Host)
	if host == "" {
		host = req.Host
	}

	// 获取证书
	cert, err := ps.certManager.GetCertForHost(host)
	if err != nil {
		return
	}

	// TLS握手
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
	}
	tlsConn := tls.Server(clientConn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		return
	}
	defer tlsConn.Close()

	// 读取解密后的HTTP请求
	reader := bufio.NewReader(tlsConn)
	httpReq, err := http.ReadRequest(reader)
	if err != nil {
		return
	}

	// 修正URL
	httpReq.URL.Scheme = "https"
	httpReq.URL.Host = req.Host

	// 记录请求
	ps.recordRequest(httpReq)

	// 转发到真实服务器
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, _ := io.ReadAll(resp.Body)

	// 记录响应
	ps.recordResponseWithBody(httpReq.URL.String(), resp, bodyBytes)

	// 写回状态行
	tlsConn.Write([]byte(fmt.Sprintf("HTTP/%d.%d %d %s\r\n", resp.ProtoMajor, resp.ProtoMinor, resp.StatusCode, resp.Status)))

	// 写回响应头
	for k, v := range resp.Header {
		for _, vv := range v {
			tlsConn.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, vv)))
		}
	}

	// 空行分隔头和体
	tlsConn.Write([]byte("\r\n"))

	// 写回响应体
	tlsConn.Write(bodyBytes)
}

func (ps *ProxyServer) handleHTTPSTunnel(clientConn net.Conn, req *http.Request) {
	// 普通HTTPS隧道(不解密)
	targetConn, err := net.DialTimeout("tcp", req.Host, 10*time.Second)
	if err != nil {
		clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer targetConn.Close()

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	done := make(chan struct{}, 2)
	go func() {
		io.Copy(targetConn, clientConn)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(clientConn, targetConn)
		done <- struct{}{}
	}()
	<-done
}

func (ps *ProxyServer) recordRequest(req *http.Request) {
	// 跳过HTTPS CONNECT请求的记录
	if req.Method == http.MethodConnect {
		return
	}

	headers := make(map[string]string)
	for k, v := range req.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	body := ""
	if req.Body != nil {
		bodyBytes := make([]byte, 1000)
		n, _ := req.Body.Read(bodyBytes)
		body = string(bodyBytes[:n])
	}

	_ = types.NetworkRequest{
		ID:      generateID(),
		URL:     req.URL.String(),
		Method:  req.Method,
		Headers: headers,
		Body:    body,
		Type:    "proxy",
		Time:    time.Now().UnixMilli(),
		Domain:  req.Host,
		Path:    req.URL.Path,
	}

	// TODO: Fix this - ps.capture.AddRequest(&reqData)

}

func (ps *ProxyServer) recordResponseWithBody(url string, resp *http.Response, bodyBytes []byte) {
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// 限制记录的body大小
	body := string(bodyBytes)
	if len(body) > 5000 {
		body = body[:5000]
	}

	contentType := resp.Header.Get("Content-Type")

	respData := types.NetworkResponse{
		ID:          generateID(),
		URL:         url,
		Status:      resp.StatusCode,
		StatusText:  resp.Status,
		Headers:     headers,
		Body:        body,
		Size:        len(bodyBytes),
		ContentType: contentType,
	}

	ps.capture.AddResponse(&respData)

}

// 生成自签名证书用于HTTPS代理
func (ps *ProxyServer) generateCert() (*tls.Certificate, error) {
	// 简化版本,实际使用需要完整的证书生成逻辑
	return nil, fmt.Errorf("证书生成未实现")
}

// 获取代理配置信息
func (ps *ProxyServer) GetProxyURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", ps.port)
}

// 获取PAC脚本
func (ps *ProxyServer) GetPACScript() string {
	return fmt.Sprintf(`function FindProxyForURL(url, host) {
    return "PROXY 127.0.0.1:%d";
}`, ps.port)
}

// generateID 生成唯一ID
