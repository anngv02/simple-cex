package engine

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateBuyOrder(db *pgxpool.Pool, userID int, symbol string, price, amount float64) (int, error) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Printf("CreateBuyOrder: Failed to begin transaction for user %d: %v", userID, err)
		return 0, err
	}
	defer tx.Rollback(ctx)

	cost := price * amount
	log.Printf("CreateBuyOrder: User %d, BUY %f %s @ %f, cost: %f", userID, amount, symbol, price, cost)

	// 1. Check balance
	var available float64
	err = tx.QueryRow(ctx,
		`SELECT available FROM balances
		 WHERE user_id=$1 AND asset_symbol='USDT' FOR UPDATE`,
		userID).Scan(&available)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("CreateBuyOrder: User %d has no USDT balance", userID)
			return 0, errors.New("user balance not found - user may not exist or have no USDT balance")
		}
		log.Printf("CreateBuyOrder: Error checking balance for user %d: %v", userID, err)
		return 0, err
	}

	log.Printf("CreateBuyOrder: User %d has %f USDT available, need %f", userID, available, cost)

	if available < cost {
		log.Printf("CreateBuyOrder: User %d insufficient balance: %f < %f", userID, available, cost)
		return 0, errors.New("insufficient balance")
	}

	// 2. Update balances
	_, err = tx.Exec(ctx,
		`UPDATE balances
		 SET available = available - $1,
		     locked = locked + $1
		 WHERE user_id=$2 AND asset_symbol='USDT'`,
		cost, userID)
	if err != nil {
		return 0, err
	}

	// 3. Create order
	var orderID int
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (user_id, symbol, side, price, amount)
		 VALUES ($1,$2,'BUY',$3,$4)
		 RETURNING id`,
		userID, symbol, price, amount).Scan(&orderID)

	if err != nil {
		return 0, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

func CreateSellOrder(db *pgxpool.Pool, userID int, symbol string, price, amount float64) (int, error) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Printf("CreateSellOrder: Failed to begin transaction for user %d: %v", userID, err)
		return 0, err
	}
	defer tx.Rollback(ctx)

	// Với lệnh SELL (Bán BTC), ta chỉ cần khóa số lượng BTC (amount)
	// Không quan tâm giá (price) khi tính toán số dư cần khóa
	cost := amount
	assetToLock := "BTC"
	log.Printf("CreateSellOrder: User %d, SELL %f %s @ %f, need %f BTC", userID, amount, symbol, price, cost)

	// 1. Check balance BTC
	var available float64
	err = tx.QueryRow(ctx,
		`SELECT available FROM balances
		 WHERE user_id=$1 AND asset_symbol=$2 FOR UPDATE`,
		userID, assetToLock).Scan(&available)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("CreateSellOrder: User %d has no BTC balance", userID)
			return 0, errors.New("user balance not found - user may not exist or have no BTC balance")
		}
		log.Printf("CreateSellOrder: Error checking balance for user %d: %v", userID, err)
		return 0, err
	}

	log.Printf("CreateSellOrder: User %d has %f BTC available, need %f", userID, available, cost)

	if available < cost {
		log.Printf("CreateSellOrder: User %d insufficient balance: %f < %f", userID, available, cost)
		return 0, errors.New("insufficient balance")
	}

	// 2. Update balances (Trừ BTC available, cộng BTC locked)
	_, err = tx.Exec(ctx,
		`UPDATE balances
		 SET available = available - $1,
		     locked = locked + $1
		 WHERE user_id=$2 AND asset_symbol=$3`,
		cost, userID, assetToLock)
	if err != nil {
		return 0, err
	}

	// 3. Insert Order (Side = 'SELL')
	var orderID int
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (user_id, symbol, side, price, amount)
		 VALUES ($1,$2,'SELL',$3,$4)
		 RETURNING id`,
		userID, symbol, price, amount).Scan(&orderID)

	if err != nil {
		return 0, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

func CancelOrder(db *pgxpool.Pool, orderID int, userID int) error {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status, side, symbol string
	var price, amount, filled float64

	// Lấy thêm 'side' và 'symbol' để biết trả lại tiền gì
	err = tx.QueryRow(ctx,
		`SELECT status, side, symbol, price, amount, filled
		 FROM orders
		 WHERE id=$1 AND user_id=$2
		 FOR UPDATE`,
		orderID, userID).
		Scan(&status, &side, &symbol, &price, &amount, &filled)

	if err != nil {
		return err 
	}

	if status != "OPEN" && status != "PARTIAL" {
		return errors.New("order cannot be cancelled")
	}

	var assetToRefund string
	var amountToRefund float64

	// Giả sử symbol là "BTC_USDT" -> Cần tách chuỗi để lấy Base (BTC) và Quote (USDT)
	// Để đơn giản cho demo, ta hardcode logic cắt chuỗi hoặc quy ước
	// Ở đây giả định symbol format chuẩn là "BASE_QUOTE"

	if side == "BUY" {
		// Mua BTC bằng USDT -> Trả lại USDT
		assetToRefund = "USDT"
		amountToRefund = (amount - filled) * price
	} else { // SELL
		// Bán BTC lấy USDT -> Trả lại BTC
		assetToRefund = "BTC"
		amountToRefund = amount - filled
	}

	// Unlock funds (Cộng lại Available, Trừ Locked)
	_, err = tx.Exec(ctx,
		`UPDATE balances
		 SET available = available + $1,
		     locked = locked - $1
		 WHERE user_id=$2 AND asset_symbol=$3`,
		amountToRefund, userID, assetToRefund)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`UPDATE orders SET status='CANCELLED'
		 WHERE id=$1`,
		orderID)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
