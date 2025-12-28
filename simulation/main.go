package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// Cấu hình
const (
	API_URL    = "http://localhost:8010/order"
	SYMBOL     = "BTC_USDT"
	BASE_PRICE = 50000.0 // Giá mốc Bitcoin
	NUM_USERS  = 10      // Số lượng user giả lập

	// Big traders: User 2-6 (5 users)
	BIG_TRADER_START   = 2
	BIG_TRADER_END     = 6
	BIG_TRADE_MIN_USD  = 10000.0         // 10,000 USD
	BIG_TRADE_MAX_USD  = 50000.0         // 50,000 USD
	BIG_TRADE_INTERVAL = 1 * time.Minute // 1 phút 1 lần

	// Small traders: User 7-10 (4 users)
	SMALL_TRADER_START   = 7
	SMALL_TRADER_END     = 10
	SMALL_TRADE_MIN_USD  = 1000.0          // 1,000 USD
	SMALL_TRADE_MAX_USD  = 10000.0         // 10,000 USD
	SMALL_TRADE_INTERVAL = 3 * time.Second // 3 giây 1 lần
)

type OrderRequest struct {
	UserID int     `json:"user_id"`
	Symbol string  `json:"symbol"`
	Side   string  `json:"side"`
	Price  float64 `json:"price"`
	Amount float64 `json:"amount"`
}

func main() {
	fmt.Println("🚀 STARTING MARKET SIMULATION...")
	fmt.Println("Press Ctrl+C to stop")

	var wg sync.WaitGroup

	// 1. Chạy Bot Market Maker (User 1 - Luôn giữ Orderbook dày)
	wg.Add(1)
	go runMarketMaker()

	// 2. Chạy Big Traders (User 2-6: Giao dịch lớn 10k-50k USD, mỗi 1 phút)
	for i := BIG_TRADER_START; i <= BIG_TRADER_END; i++ {
		wg.Add(1)
		go runBigTrader(i)
	}

	// 3. Chạy Small Traders (User 7-10: Giao dịch nhỏ 1k-10k USD, mỗi 3 giây)
	for i := SMALL_TRADER_START; i <= SMALL_TRADER_END; i++ {
		wg.Add(1)
		go runSmallTrader(i)
	}

	wg.Wait()
}

// Bot Market Maker: Cứ 2 giây lại rải lệnh Mua/Bán xung quanh giá 50k
// Để đảm bảo Orderbook luôn đẹp
func runMarketMaker() {
	for {
		// Rải lệnh BÁN (Giá cao hơn 50k) - giảm số lệnh để tránh lock hết BTC
		for i := 1; i <= 3; i++ {
			price := BASE_PRICE + float64(i*50) + rand.Float64()*10 // Ví dụ: 50050, 50100...
			placeOrder(1, "SELL", price, 0.3)                       // Giảm amount từ 0.5 xuống 0.3
		}

		// Rải lệnh MUA (Giá thấp hơn 50k) - giảm số lệnh để tránh lock hết USDT
		for i := 1; i <= 3; i++ {
			price := BASE_PRICE - float64(i*50) - rand.Float64()*10 // Ví dụ: 49950, 49900...
			placeOrder(1, "BUY", price, 0.3)                        // Giảm amount từ 0.5 xuống 0.3
		}

		time.Sleep(2 * time.Second)
	}
}

// Big Trader: Giao dịch lớn (10,000 - 50,000 USD), mỗi 1 phút
func runBigTrader(userID int) {
	// Tránh tất cả user vào lệnh cùng 1 tích tắc
	initialDelay := time.Duration(rand.Intn(10000)) * time.Millisecond
	time.Sleep(initialDelay)

	for {
		// Random hành động: Mua hoặc Bán
		side := "BUY"
		if rand.Intn(2) == 0 {
			side = "SELL"
		}

		// Giá dao động nhẹ quanh BASE_PRICE (±2%)
		fluctuation := (rand.Float64()*0.04 - 0.02) * BASE_PRICE
		price := BASE_PRICE + fluctuation

		// Tính amount dựa trên giá trị giao dịch (10k - 50k USD)
		tradeValueUSD := BIG_TRADE_MIN_USD + rand.Float64()*(BIG_TRADE_MAX_USD-BIG_TRADE_MIN_USD)
		amount := tradeValueUSD / price

		// Gửi lệnh
		placeOrder(userID, side, price, amount)

		// Đợi 1 phút trước khi đặt lệnh tiếp
		time.Sleep(BIG_TRADE_INTERVAL)
	}
}

// Small Trader: Giao dịch nhỏ (1,000 - 10,000 USD), mỗi 3 giây
func runSmallTrader(userID int) {
	// Tránh tất cả user vào lệnh cùng 1 tích tắc
	initialDelay := time.Duration(rand.Intn(3000)) * time.Millisecond
	time.Sleep(initialDelay)

	for {
		// Random hành động: Mua hoặc Bán
		side := "BUY"
		if rand.Intn(2) == 0 {
			side = "SELL"
		}

		// Giá dao động nhẹ quanh BASE_PRICE (±1%)
		fluctuation := (rand.Float64()*0.02 - 0.01) * BASE_PRICE
		price := BASE_PRICE + fluctuation

		// Tính amount dựa trên giá trị giao dịch (1k - 10k USD)
		tradeValueUSD := SMALL_TRADE_MIN_USD + rand.Float64()*(SMALL_TRADE_MAX_USD-SMALL_TRADE_MIN_USD)
		amount := tradeValueUSD / price

		// Gửi lệnh
		placeOrder(userID, side, price, amount)

		// Đợi 3 giây trước khi đặt lệnh tiếp
		time.Sleep(SMALL_TRADE_INTERVAL)
	}
}

func placeOrder(userID int, side string, price, amount float64) {
	reqBody, _ := json.Marshal(OrderRequest{
		UserID: userID,
		Symbol: SYMBOL,
		Side:   side,
		Price:  price,
		Amount: amount,
	})

	resp, err := http.Post(API_URL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("[User %d] Error: %v\n", userID, err)
		return
	}
	defer resp.Body.Close()

	// Chỉ in ra nếu có lỗi hoặc thỉnh thoảng in để đỡ rác màn hình
	if resp.StatusCode != 200 {
		fmt.Printf("[User %d] Failed: %s\n", userID, resp.Status)
	} else {
		// fmt.Printf("[User %d] %s %.2f @ %.2f\n", userID, side, amount, price)
	}
}
