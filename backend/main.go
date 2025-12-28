package main

import (
	"log"
	
    // Import package engine của bạn (tên module/engine)
    // Nếu go.mod của bạn là "simple-cex", thì import là:
	"simple-cex/engine" 
)


func main() {
	if err := InitDB(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("DB connected")

	// 1. Khởi tạo Engine
	tradeEngine := engine.NewEngine(db)

	// 2. Tạo kịch bản Test
	log.Println("--- START TESTING ---")

	// User 1 (đã có 100k USDT từ bước seed data): Đặt mua 1 BTC giá 50,000
    // Cần đảm bảo User 1 có tiền trong DB
	log.Println("User 1 places BUY order...")
	err := tradeEngine.PlaceOrder(1, "BTC_USDT", "BUY", 50000, 1.0)
	if err != nil {
		log.Printf("Error placing buy order: %v", err)
	} else {
        log.Println("User 1 BUY order placed successfully")
    }

    // User 2 (Tạo thêm user này trong DB nếu chưa có): Đặt bán 1 BTC giá 49,000 (Rẻ hơn -> Khớp ngay)
    // Bạn cần insert manual user 2 có BTC vào DB trước khi chạy dòng này
    // INSERT INTO users...; INSERT INTO balances (user 2, BTC, 10)...
	log.Println("User 2 places SELL order...")
	err = tradeEngine.PlaceOrder(2, "BTC_USDT", "SELL", 49000, 1.0) // Bán rẻ
	if err != nil {
		log.Printf("Error placing sell order: %v", err)
	} else {
        log.Println("User 2 SELL order placed successfully")
    }

    // Nếu code đúng:
    // - Console sẽ báo "Matched 1 trades"
    // - Check DB bảng trades sẽ thấy record
    // - Check DB balances: User 1 trừ USDT, cộng BTC. User 2 trừ BTC, cộng USDT.
}