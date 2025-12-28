package api

import (
	"net/http"
	"simple-cex/engine" // Import package engine của bạn

	"github.com/gin-gonic/gin"
)

// Server struct chứa Engine và Router
type Server struct {
	engine *engine.Engine
	router *gin.Engine
}

// Khởi tạo Server
func NewServer(eng *engine.Engine) *Server {
	server := &Server{
		engine: eng,
		router: gin.Default(),
	}
	server.setupRoutes()
	return server
}

// Định nghĩa các đường dẫn (Router)
func (s *Server) setupRoutes() {
	// Middleware CORS (Để Frontend gọi được mà không bị chặn)
	s.router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API Đặt lệnh
	s.router.POST("/order", s.handlePlaceOrder)
	
	// API Lấy Orderbook (để vẽ chart/bảng giá)
	s.router.GET("/orderbook/:symbol", s.handleGetOrderBook)
}

// Start server
func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

// --- HANDLERS ---

// Request Body cho đặt lệnh
type placeOrderRequest struct {
	UserID int     `json:"user_id"`
	Symbol string  `json:"symbol"`
	Side   string  `json:"side"` // BUY hoặc SELL
	Price  float64 `json:"price"`
	Amount float64 `json:"amount"`
}

func (s *Server) handlePlaceOrder(c *gin.Context) {
	var req placeOrderRequest
	// 1. Validate JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Gọi Matching Engine
	err := s.engine.PlaceOrder(req.UserID, req.Symbol, req.Side, req.Price, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully"})
}

func (s *Server) handleGetOrderBook(c *gin.Context) {
	symbol := c.Param("symbol")
	
	// Lấy Orderbook từ RAM
	ob, ok := s.engine.OrderBooks[symbol]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Symbol not found"})
		return
	}

	// Trả về JSON của Orderbook (Gồm Bids và Asks)
	// Lưu ý: Trong môi trường production high-load, cần RWMutex ở đây để tránh race condition
	c.JSON(http.StatusOK, gin.H{
		"symbol": ob.Symbol,
		"asks":   ob.Asks, // Danh sách người bán
		"bids":   ob.Bids, // Danh sách người mua
	})
}