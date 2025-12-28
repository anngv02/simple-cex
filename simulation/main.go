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
	BIG_TRADE_MIN_USD  = 5000.0          // 5,000 USD (giảm min để random hơn)
	BIG_TRADE_MAX_USD  = 100000.0        // 100,000 USD (tăng max để random hơn)
	BIG_TRADE_INTERVAL = 1 * time.Minute // 1 phút 1 lần

	// Small traders: User 7-10 (4 users)
	SMALL_TRADER_START   = 7
	SMALL_TRADER_END     = 10
	SMALL_TRADE_MIN_USD  = 500.0           // 500 USD (giảm min để random hơn)
	SMALL_TRADE_MAX_USD  = 20000.0         // 20,000 USD (tăng max để random hơn)
	SMALL_TRADE_INTERVAL = 3 * time.Second // 3 giây 1 lần

	// Market Maker
	MARKET_MAKER_MIN_AMOUNT = 0.1 // 0.1 BTC
	MARKET_MAKER_MAX_AMOUNT = 0.8 // 0.8 BTC
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
		// Rải lệnh BÁN (Giá cao hơn 50k) - số lượng lệnh random từ 2-4
		numSellOrders := 2 + rand.Intn(3) // 2, 3, hoặc 4 lệnh
		for i := 1; i <= numSellOrders; i++ {
			price := BASE_PRICE + float64(i*50) + rand.Float64()*10 // Ví dụ: 50050, 50100...
			// Random amount từ 0.1 đến 0.8 BTC
			amount := MARKET_MAKER_MIN_AMOUNT + rand.Float64()*(MARKET_MAKER_MAX_AMOUNT-MARKET_MAKER_MIN_AMOUNT)
			placeOrder(1, "SELL", price, amount)
		}

		// Rải lệnh MUA (Giá thấp hơn 50k) - số lượng lệnh random từ 2-4
		numBuyOrders := 2 + rand.Intn(3) // 2, 3, hoặc 4 lệnh
		for i := 1; i <= numBuyOrders; i++ {
			price := BASE_PRICE - float64(i*50) - rand.Float64()*10 // Ví dụ: 49950, 49900...
			// Random amount từ 0.1 đến 0.8 BTC
			amount := MARKET_MAKER_MIN_AMOUNT + rand.Float64()*(MARKET_MAKER_MAX_AMOUNT-MARKET_MAKER_MIN_AMOUNT)
			placeOrder(1, "BUY", price, amount)
		}

		// Random sleep từ 1.5 đến 3 giây để tạo sự đa dạng
		sleepDuration := 1500 + rand.Intn(1500) // 1.5s đến 3s
		time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
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

		// Tính amount dựa trên giá trị giao dịch (5k - 100k USD) - random hơn
		// Sử dụng exponential distribution để có nhiều lệnh nhỏ hơn và ít lệnh lớn hơn (giống thực tế)
		randomFactor := rand.Float64() * rand.Float64() // Tạo distribution lệch về phía nhỏ hơn
		tradeValueUSD := BIG_TRADE_MIN_USD + randomFactor*(BIG_TRADE_MAX_USD-BIG_TRADE_MIN_USD)
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

		// Tính amount dựa trên giá trị giao dịch (500 - 20k USD) - random hơn
		// Sử dụng exponential distribution để có nhiều lệnh nhỏ hơn và ít lệnh lớn hơn (giống thực tế)
		randomFactor := rand.Float64() * rand.Float64() // Tạo distribution lệch về phía nhỏ hơn
		tradeValueUSD := SMALL_TRADE_MIN_USD + randomFactor*(SMALL_TRADE_MAX_USD-SMALL_TRADE_MIN_USD)
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
