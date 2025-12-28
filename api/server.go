package api

import (
	"context"
	"log"
	"net/http"
	"simple-cex/engine"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server struct chứa Engine và Router
type Server struct {
	engine    *engine.Engine
	router    *gin.Engine
	wsManager *WSManager
	db        *pgxpool.Pool
}

// Khởi tạo Server
func NewServer(eng *engine.Engine, db *pgxpool.Pool) *Server {
	server := &Server{
		engine:    eng,
		router:    gin.Default(),
		wsManager: NewWSManager(),
		db:        db,
	}
	go server.wsManager.Run()
	server.setupRoutes()
	return server
}

// Định nghĩa các đường dẫn (Router)
func (s *Server) setupRoutes() {
	// Middleware CORS
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

	// API Lấy Orderbook
	s.router.GET("/orderbook/:symbol", s.handleGetOrderBook)

	// API Lấy dữ liệu OHLCV cho chart nến
	s.router.GET("/trades/:symbol", s.handleGetTrades)

	// Route WebSocket
	s.router.GET("/ws", func(c *gin.Context) {
		s.wsManager.ServeWS(c)
	})
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
	Side   string  `json:"side"`
	Price  float64 `json:"price"`
	Amount float64 `json:"amount"`
}

func (s *Server) handlePlaceOrder(c *gin.Context) {
	var req placeOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Gọi Matching Engine
	err := s.engine.PlaceOrder(req.UserID, req.Symbol, req.Side, req.Price, req.Amount)
	if err != nil {
		log.Printf("handlePlaceOrder: Error placing order for user %d: %v", req.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. Sau khi đặt lệnh xong, gửi Orderbook mới nhất cho tất cả client
	// Lấy Orderbook hiện tại từ RAM
	if ob, ok := s.engine.OrderBooks[req.Symbol]; ok {
		// Giới hạn chỉ gửi 10 orders đầu tiên cho mỗi bên
		asks := ob.Asks
		bids := ob.Bids

		// ASKS: Lấy 10 giá thấp nhất (đầu mảng)
		if len(asks) > 10 {
			asks = asks[:10]
		}

		// BIDS: Lấy 10 giá cao nhất (đầu mảng)
		if len(bids) > 10 {
			bids = bids[:10]
		}

		// Tạo message update
		updateMsg := gin.H{
			"type":   "ORDERBOOK_UPDATE",
			"symbol": req.Symbol,
			"asks":   asks,
			"bids":   bids,
		}
		// Bắn vào kênh broadcast -> Client tự nhận được
		s.wsManager.broadcast <- updateMsg
	}

	// 3. Gửi TRADE_UPDATE để chart cập nhật real-time
	ctx := context.Background()
	var lastTradePrice, lastTradeAmount float64
	var lastTradeTime time.Time
	err = s.db.QueryRow(ctx,
		`SELECT price, amount, created_at 
		 FROM trades 
		 WHERE EXISTS (
			 SELECT 1 FROM orders WHERE orders.id = trades.maker_order_id AND orders.symbol = $1
		 )
		 ORDER BY created_at DESC 
		 LIMIT 1`,
		req.Symbol).Scan(&lastTradePrice, &lastTradeAmount, &lastTradeTime)

	if err == nil {
		// Gửi trade update qua WebSocket
		tradeUpdateMsg := gin.H{
			"type":   "TRADE_UPDATE",
			"symbol": req.Symbol,
			"price":  lastTradePrice,
			"amount": lastTradeAmount,
			"time":   lastTradeTime.Unix() * 1000, // milliseconds
		}
		s.wsManager.broadcast <- tradeUpdateMsg
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

	// Giới hạn chỉ trả về 10 orders đầu tiên cho mỗi bên
	asks := ob.Asks
	bids := ob.Bids

	if len(asks) > 10 {
		asks = asks[:10]
	}
	if len(bids) > 10 {
		bids = bids[:10]
	}

	// Trả về JSON của Orderbook (Gồm Bids và Asks)
	c.JSON(http.StatusOK, gin.H{
		"symbol": ob.Symbol,
		"asks":   asks,
		"bids":   bids,
	})
}

// Handler để lấy dữ liệu OHLCV từ trades
func (s *Server) handleGetTrades(c *gin.Context) {
	symbol := c.Param("symbol")
	interval := c.DefaultQuery("interval", "1m") // 1m, 5m, 15m, 1h, 4h, 1d
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	// Tính interval trong giây
	var intervalSeconds int
	switch interval {
	case "1m":
		intervalSeconds = 60
	case "5m":
		intervalSeconds = 300
	case "15m":
		intervalSeconds = 900
	case "1h":
		intervalSeconds = 3600
	case "4h":
		intervalSeconds = 14400
	case "1d":
		intervalSeconds = 86400
	default:
		intervalSeconds = 60
	}

	ctx := context.Background()

	// Lấy trades từ database - join với orders để lọc theo symbol
	rows, err := s.db.Query(ctx,
		`SELECT t.price, t.amount, t.created_at 
		 FROM trades t
		 INNER JOIN orders o ON o.id = t.maker_order_id
		 WHERE o.symbol = $1
		 ORDER BY t.created_at DESC 
		 LIMIT $2`,
		symbol, limit*10) // Lấy nhiều hơn để tính OHLCV

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type Trade struct {
		Price     float64
		Amount    float64
		CreatedAt time.Time
	}

	var trades []Trade
	for rows.Next() {
		var t Trade
		if err := rows.Scan(&t.Price, &t.Amount, &t.CreatedAt); err != nil {
			continue
		}
		trades = append(trades, t)
	}

	// Tính OHLCV theo interval
	ohlcvMap := make(map[int64]struct {
		open   float64
		high   float64
		low    float64
		close  float64
		volume float64
	})

	for _, trade := range trades {
		// Tính timestamp của nến (làm tròn xuống theo interval)
		candleTime := trade.CreatedAt.Unix() / int64(intervalSeconds) * int64(intervalSeconds)

		if candle, exists := ohlcvMap[candleTime]; exists {
			// Cập nhật nến hiện có
			if trade.Price > candle.high {
				candle.high = trade.Price
			}
			if trade.Price < candle.low {
				candle.low = trade.Price
			}
			candle.close = trade.Price
			candle.volume += trade.Amount
			ohlcvMap[candleTime] = candle
		} else {
			// Tạo nến mới
			ohlcvMap[candleTime] = struct {
				open   float64
				high   float64
				low    float64
				close  float64
				volume float64
			}{
				open:   trade.Price,
				high:   trade.Price,
				low:    trade.Price,
				close:  trade.Price,
				volume: trade.Amount,
			}
		}
	}

	// Chuyển đổi map sang slice và sắp xếp theo thời gian
	type OHLCV struct {
		Time   int64   `json:"time"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	}

	var result []OHLCV
	for time, candle := range ohlcvMap {
		result = append(result, OHLCV{
			Time:   time * 1000, // Chuyển sang milliseconds
			Open:   candle.open,
			High:   candle.high,
			Low:    candle.low,
			Close:  candle.close,
			Volume: candle.volume,
		})
	}

	// Sắp xếp theo thời gian tăng dần
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Time > result[j].Time {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	// Giới hạn số lượng
	if len(result) > limit {
		result = result[len(result)-limit:]
	}

	c.JSON(http.StatusOK, result)
}
