package engine

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Engine struct {
	DB         *pgxpool.Pool
	OrderBooks map[string]*OrderBook
}

func NewEngine(db *pgxpool.Pool) *Engine {
	books := make(map[string]*OrderBook)
	books["BTC_USDT"] = NewOrderBook("BTC_USDT")
	return &Engine{
		DB:         db,
		OrderBooks: books,
	}
}

// PlaceOrder: Hàm Entrypoint
func (e *Engine) PlaceOrder(userID int, symbol string, side string, price, amount float64) error {
	// 1. Validate & Lock tiền (Gọi hàm từ file accounting.go cùng package)
	// Lưu ý: Bạn cần sửa accounting.go để hàm trả về orderID thay vì chỉ error
    // Giả sử hàm CreateOrder trả về (int, error)
    var orderID int
    var err error
    
    if side == "BUY" {
        orderID, err = CreateBuyOrder(e.DB, userID, symbol, price, amount)
    } else {
        orderID, err = CreateSellOrder(e.DB, userID, symbol, price, amount)
    }
    
	if err != nil {
		return fmt.Errorf("accounting error: %v", err)
	}

	// 2. Khớp lệnh trên RAM
	ob, ok := e.OrderBooks[symbol]
	if !ok {
		return fmt.Errorf("symbol not found")
	}

	order := &Order{
		ID:        orderID,
		UserID:    userID,
		Side:      side,
		Price:     price,
		Amount:    amount,
		Filled:    0,
		Timestamp: time.Now().UnixNano(),
	}

	trades, _ := ob.Process(order)

	// 3. Settlement (Nếu có khớp)
	if len(trades) > 0 {
		err := e.Settlement(trades)
		if err != nil {
			log.Printf("CRITICAL: Settlement failed for trades %v: %v", trades, err)
            // Trong thực tế, cần log vào bảng 'errors' để admin xử lý bằng tay
		} else {
            log.Printf("Matched %d trades", len(trades))
        }
	}
	return nil
}

// Settlement: Xử lý tiền sau khớp lệnh
func (e *Engine) Settlement(trades []Trade) error {
	ctx := context.Background()
	tx, err := e.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, t := range trades {
		// A. Lưu Trade History
		_, err := tx.Exec(ctx,
			`INSERT INTO trades (maker_order_id, taker_order_id, price, amount) 
             VALUES ($1, $2, $3, $4)`,
			t.MakerOrderID, t.TakerOrderID, t.Price, t.Amount)
		if err != nil { return err }

		// B. Update Trạng thái Orders (Maker & Taker)
		// Update Maker
		_, err = tx.Exec(ctx,
			`UPDATE orders SET filled = filled + $1,
			 status = CASE WHEN filled + $1 >= amount THEN 'FILLED' ELSE 'PARTIAL' END
			 WHERE id = $2`, t.Amount, t.MakerOrderID)
		if err != nil { return err }

		// Update Taker
		_, err = tx.Exec(ctx,
			`UPDATE orders SET filled = filled + $1,
			 status = CASE WHEN filled + $1 >= amount THEN 'FILLED' ELSE 'PARTIAL' END
			 WHERE id = $2`, t.Amount, t.TakerOrderID)
		if err != nil { return err }

		// C. CHUYỂN TIỀN (Phần quan trọng nhất)
        // Cần lấy UserID của Maker và Taker để cộng tiền
        var makerID, takerID int
        var makerSide string
        
        // Lấy thông tin Maker (để biết ai mua ai bán)
        err = tx.QueryRow(ctx, "SELECT user_id, side FROM orders WHERE id=$1", t.MakerOrderID).Scan(&makerID, &makerSide)
        if err != nil { return err }
        
        // Lấy thông tin Taker
        err = tx.QueryRow(ctx, "SELECT user_id FROM orders WHERE id=$1", t.TakerOrderID).Scan(&takerID)
        if err != nil { return err }

        // Logic chuyển tiền:
        // Tiền bị LOCK (Locked) đã bị trừ khỏi Available lúc đặt lệnh rồi.
        // Giờ ta chỉ cần: Trừ Locked của người bán -> Cộng Available người mua.
        
        costUSDT := t.Price * t.Amount
        amountBTC := t.Amount

        if makerSide == "BUY" {
            // Maker là người MUA (đã lock USDT), Taker là người BÁN (đã lock BTC)
            
            // 1. Maker (Mua): Trừ USDT Locked (đã dùng) -> Nhận BTC Available
            _, err = tx.Exec(ctx, `UPDATE balances SET locked = locked - $1 WHERE user_id=$2 AND asset_symbol='USDT'`, costUSDT, makerID)
            if err != nil { return err }
            _, err = tx.Exec(ctx, `UPDATE balances SET available = available + $1 WHERE user_id=$2 AND asset_symbol='BTC'`, amountBTC, makerID)
            if err != nil { return err }

            // 2. Taker (Bán): Trừ BTC Locked (đã bán) -> Nhận USDT Available
            _, err = tx.Exec(ctx, `UPDATE balances SET locked = locked - $1 WHERE user_id=$2 AND asset_symbol='BTC'`, amountBTC, takerID)
            if err != nil { return err }
            _, err = tx.Exec(ctx, `UPDATE balances SET available = available + $1 WHERE user_id=$2 AND asset_symbol='USDT'`, costUSDT, takerID)
            if err != nil { return err }

        } else { // makerSide == "SELL"
            // Maker là người BÁN (đã lock BTC), Taker là người MUA (đã lock USDT)

            // 1. Maker (Bán): Trừ BTC Locked -> Nhận USDT Available
            _, err = tx.Exec(ctx, `UPDATE balances SET locked = locked - $1 WHERE user_id=$2 AND asset_symbol='BTC'`, amountBTC, makerID)
            if err != nil { return err }
             _, err = tx.Exec(ctx, `UPDATE balances SET available = available + $1 WHERE user_id=$2 AND asset_symbol='USDT'`, costUSDT, makerID)
            if err != nil { return err }

            // 2. Taker (Mua): Trừ USDT Locked -> Nhận BTC Available
            _, err = tx.Exec(ctx, `UPDATE balances SET locked = locked - $1 WHERE user_id=$2 AND asset_symbol='USDT'`, costUSDT, takerID)
            if err != nil { return err }
            _, err = tx.Exec(ctx, `UPDATE balances SET available = available + $1 WHERE user_id=$2 AND asset_symbol='BTC'`, amountBTC, takerID)
            if err != nil { return err }
        }
	}

	return tx.Commit(ctx)
}