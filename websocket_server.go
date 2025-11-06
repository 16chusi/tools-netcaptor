package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type WebSocketServer struct {
	app      *NetworkApp
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	mu       sync.RWMutex
	wsPort   int
	server   *http.Server
	running  bool
}

type WSMessage struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data,omitempty"`
}

func NewWebSocketServer(app *NetworkApp) *WebSocketServer {
	return &WebSocketServer{
		app: app,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

func (ws *WebSocketServer) Start() error {
	if ws.running {
		return nil
	}

	ws.wsPort = ws.getRandomPort()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ws.HandleWebSocket)

	ws.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", ws.wsPort),
		Handler: ws.corsMiddleware(mux),
	}

	ws.running = true

	go func() {
		log.Printf("[WebSocket] 服务器启动在端口 %d", ws.wsPort)
		if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[WebSocket] 服务器错误: %v", err)
			ws.running = false
		}
	}()

	return nil
}

func (ws *WebSocketServer) Stop() error {
	if !ws.running || ws.server == nil {
		return nil
	}

	ws.running = false
	err := ws.server.Close()
	ws.server = nil
	log.Printf("[WebSocket] 服务器已停止")
	return err
}

func (ws *WebSocketServer) getRandomPort() int {
	rand.Seed(time.Now().UnixNano())
	return 10000 + rand.Intn(10000)
}

func (ws *WebSocketServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (ws *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] 升级失败: %v", err)
		return
	}

	ws.mu.Lock()
	ws.clients[conn] = true
	ws.mu.Unlock()

	log.Printf("[WebSocket] 客户端已连接,当前连接数: %d", len(ws.clients))

	defer func() {
		ws.mu.Lock()
		delete(ws.clients, conn)
		ws.mu.Unlock()
		conn.Close()
		log.Printf("[WebSocket] 客户端已断开,当前连接数: %d", len(ws.clients))
	}()

	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] 读取错误: %v", err)
			}
			break
		}

		ws.handleMessage(conn, msg)
	}
}

func (ws *WebSocketServer) handleMessage(conn *websocket.Conn, msg WSMessage) {
	log.Printf("[WebSocket] 收到消息: %s", msg.Type)

	switch msg.Type {
	case "connection":
		ws.sendToClient(conn, WSMessage{
			Type: "connection_ack",
			Data: map[string]interface{}{"status": "connected"},
		})

	case "element_clicked":
		if ws.app.ctx != nil {
			runtime.EventsEmit(ws.app.ctx, "element_clicked", msg.Data)
		}

	case "page_loaded":
		if ws.app.ctx != nil {
			runtime.EventsEmit(ws.app.ctx, "page_loaded", msg.Data)
		}

	case "action_result":
		log.Printf("[WebSocket服务器] 收到action_result消息: %+v", msg.Data)
		// 转发给工作流执行器
		if ws.app.workflowExecutor != nil {
			log.Printf("[WebSocket服务器] 转发消息给工作流执行器")
			ws.app.workflowExecutor.HandleResponse(msg)
		} else {
			log.Printf("[WebSocket服务器] 工作流执行器为空，无法转发消息")
		}

	case "screenshot":
		log.Printf("[WebSocket服务器] 收到screenshot消息")
		// 转发给工作流执行器
		if ws.app.workflowExecutor != nil {
			ws.app.workflowExecutor.HandleResponse(msg)
		}

	case "get_page_info":
		log.Printf("[WebSocket服务器] 收到get_page_info消息")
		// 转发给工作流执行器
		if ws.app.workflowExecutor != nil {
			ws.app.workflowExecutor.HandleResponse(msg)
		}

	case "get_page_dom":
		log.Printf("[WebSocket服务器] 收到get_page_dom消息")
		// 转发给工作流执行器
		if ws.app.workflowExecutor != nil {
			ws.app.workflowExecutor.HandleResponse(msg)
		}

	default:
		log.Printf("[WebSocket] 未知消息类型: %s", msg.Type)
	}
}

func (ws *WebSocketServer) sendToClient(conn *websocket.Conn, msg WSMessage) error {
	return conn.WriteJSON(msg)
}

func (ws *WebSocketServer) Broadcast(msg WSMessage) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	log.Printf("[WebSocket] 广播消息给 %d 个客户端: %s", len(ws.clients), msg.Type)

	for client := range ws.clients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Printf("[WebSocket] ❌ 广播失败: %v", err)
		} else {
			log.Printf("[WebSocket] ✅ 消息已发送给客户端")
		}
	}
}

func (ws *WebSocketServer) GetWSPort() int {
	return ws.wsPort
}

func (ws *WebSocketServer) IsRunning() bool {
	return ws.running
}

func (ws *WebSocketServer) HasClients() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return len(ws.clients) > 0
}

func (ws *WebSocketServer) GetClientCount() int {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return len(ws.clients)
}
