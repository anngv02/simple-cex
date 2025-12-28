package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Cấu hình Upgrader: Chuyển từ HTTP thường sang WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Quan trọng: Cho phép mọi nguồn kết nối (tránh lỗi CORS khi dev)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSManager quản lý các kết nối
type WSManager struct {
	clients    map[*websocket.Conn]bool // Danh sách user đang kết nối
	broadcast  chan interface{}         // Kênh nhận dữ liệu để bắn đi
	register   chan *websocket.Conn     // Kênh đăng ký user mới
	unregister chan *websocket.Conn     // Kênh hủy đăng ký user
	mutex      sync.Mutex
}

func NewWSManager() *WSManager {
	return &WSManager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan interface{}),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Run: Vòng lặp chính để xử lý tin nhắn
func (manager *WSManager) Run() {
	for {
		select {
		case conn := <-manager.register:
			manager.mutex.Lock()
			manager.clients[conn] = true
			manager.mutex.Unlock()
			log.Println("New client connected")

		case conn := <-manager.unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[conn]; ok {
				delete(manager.clients, conn)
				conn.Close()
			}
			manager.mutex.Unlock()
			log.Println("Client disconnected")

		case message := <-manager.broadcast:
			// Nhận được tin nhắn -> Gửi cho TOÀN BỘ clients
			manager.mutex.Lock()
			for conn := range manager.clients {
				err := conn.WriteJSON(message)
				if err != nil {
					log.Printf("WS Error: %v", err)
					conn.Close()
					delete(manager.clients, conn)
				}
			}
			manager.mutex.Unlock()
		}
	}
}

// Handler để Gin gọi vào khi có request /ws
func (manager *WSManager) ServeWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade failed:", err)
		return
	}
	manager.register <- conn
}