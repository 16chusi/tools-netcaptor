package network

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"netcaptor/internal/types"
	"netcaptor/internal/utils"
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
		utils.AppLog.Info(fmt.Sprintf("[WebSocket] 服务器启动在端口 %d", ws.wsPort))
		if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.AppLog.Info(fmt.Sprintf("[WebSocket] 服务器错误: %v", err))
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
	utils.AppLog.Info(fmt.Sprintf("[WebSocket] 服务器已停止"))
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
		utils.AppLog.Info(fmt.Sprintf("[WebSocket] 升级失败: %v", err))
		return
	}

	ws.mu.Lock()
	ws.clients[conn] = true
	ws.mu.Unlock()

	utils.AppLog.Info(fmt.Sprintf("[WebSocket] 客户端已连接,当前连接数: %d", len(ws.clients)))

	defer func() {
		ws.mu.Lock()
		delete(ws.clients, conn)
		ws.mu.Unlock()
		conn.Close()
		utils.AppLog.Info(fmt.Sprintf("[WebSocket] 客户端已断开,当前连接数: %d", len(ws.clients)))
	}()

	for {
		var msg types.WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				utils.AppLog.Info(fmt.Sprintf("[WebSocket] 读取错误: %v", err))
			}
			break
		}

		ws.handleMessage(conn, msg)
	}
}

func (ws *WebSocketServer) handleMessage(conn *websocket.Conn, msg types.WSMessage) {
	utils.AppLog.Debug(fmt.Sprintf("[WebSocket] 收到消息: %s", msg.Type))

	switch msg.Type {
	case "connection":
		ws.sendToClient(conn, types.WSMessage{
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
		utils.AppLog.Info(fmt.Sprintf("[WebSocket服务器] 收到action_result消息: %+v", msg.Data))
		// 转发给工作流执行器
		if ws.app.workflowExecutor != nil {
			utils.AppLog.Info(fmt.Sprintf("[WebSocket服务器] 转发消息给工作流执行器"))
			ws.app.workflowExecutor.HandleResponse(msg)
		} else {
			utils.AppLog.Info(fmt.Sprintf("[WebSocket服务器] 工作流执行器为空，无法转发消息"))
		}

	case "screenshot":
		utils.AppLog.Info(fmt.Sprintf("[WebSocket服务器] 收到screenshot消息"))
		// 转发给工作流执行器
		if ws.app.workflowExecutor != nil {
			ws.app.workflowExecutor.HandleResponse(msg)
		}

	case "get_page_info":
		utils.AppLog.Info(fmt.Sprintf("[WebSocket服务器] 收到get_page_info消息"))
		// 转发给工作流执行器
		if ws.app.workflowExecutor != nil {
			ws.app.workflowExecutor.HandleResponse(msg)
		}

	case "get_page_dom":
		utils.AppLog.Info(fmt.Sprintf("[WebSocket服务器] 收到get_page_dom消息"))
		// 转发给工作流执行器
		if ws.app.workflowExecutor != nil {
			ws.app.workflowExecutor.HandleResponse(msg)
		}

	default:
		utils.AppLog.Info(fmt.Sprintf("[WebSocket] 未知消息类型: %s", msg.Type))
	}
}

func (ws *WebSocketServer) sendToClient(conn *websocket.Conn, msg types.WSMessage) error {
	return conn.WriteJSON(msg)
}

func (ws *WebSocketServer) Broadcast(msg types.WSMessage) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	utils.AppLog.Info(fmt.Sprintf("[WebSocket] 广播消息给 %d 个客户端: %s", len(ws.clients), msg.Type))

	for client := range ws.clients {
		err := client.WriteJSON(msg)
		if err != nil {
			utils.AppLog.Info(fmt.Sprintf("[WebSocket] ❌ 广播失败: %v", err))
		} else {
			utils.AppLog.Info(fmt.Sprintf("[WebSocket] ✅ 消息已发送给客户端"))
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

// BroadcastMessage 实现 types.WebSocketHandler 接口
func (ws *WebSocketServer) BroadcastMessage(msg types.WSMessage) {
	ws.Broadcast(msg)
}

// SendMessage 实现 types.WebSocketHandler 接口
func (ws *WebSocketServer) SendMessage(msg types.WSMessage) {
	ws.Broadcast(msg)
}
