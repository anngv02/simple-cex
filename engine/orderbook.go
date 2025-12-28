package engine

import (
	"sort"
	"time"
)

// Order đại diện cho lệnh đang nằm trên RAM
type Order struct {
	ID        int
	UserID    int
	Side      string  // "BUY" or "SELL"
	Price     float64
	Amount    float64 // Số lượng ban đầu
	Filled    float64 // Số lượng đã khớp
	Timestamp int64   // Để ưu tiên ai đến trước (FIFO)
}

// OrderBook chứa 2 danh sách lệnh
type OrderBook struct {
	Symbol string
	Bids   []*Order // Mua: Giá cao xếp trước
	Asks   []*Order // Bán: Giá thấp xếp trước
}

// Trade ghi lại kết quả khớp lệnh để lưu xuống DB sau này
type Trade struct {
	MakerOrderID int     // Lệnh đang nằm chờ (bị khớp)
	TakerOrderID int     // Lệnh mới bay vào (chủ động khớp)
	Price        float64
	Amount       float64
	CreatedAt    time.Time
}

// Hàm tạo OrderBook mới
func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		Bids:   make([]*Order, 0),
		Asks:   make([]*Order, 0),
	}
}

// Logic thêm lệnh vào sổ (khi không khớp được ngay)
func (ob *OrderBook) AddOrder(o *Order) {
	if o.Side == "BUY" {
		ob.Bids = append(ob.Bids, o)
		// Sort: Giá cao nhất lên đầu. Nếu bằng giá, ai đến trước lên đầu.
		sort.Slice(ob.Bids, func(i, j int) bool {
			if ob.Bids[i].Price == ob.Bids[j].Price {
				return ob.Bids[i].Timestamp < ob.Bids[j].Timestamp
			}
			return ob.Bids[i].Price > ob.Bids[j].Price
		})
	} else {
		ob.Asks = append(ob.Asks, o)
		// Sort: Giá thấp nhất lên đầu.
		sort.Slice(ob.Asks, func(i, j int) bool {
			if ob.Asks[i].Price == ob.Asks[j].Price {
				return ob.Asks[i].Timestamp < ob.Asks[j].Timestamp
			}
			return ob.Asks[i].Price < ob.Asks[j].Price
		})
	}
}

// Process xử lý một lệnh mới bay vào
func (ob *OrderBook) Process(order *Order) ([]Trade, *Order) {
	var trades []Trade

	// Nếu là lệnh MUA, thì soi bên BÁN (Asks)
	if order.Side == "BUY" {
		for len(ob.Asks) > 0 {
			bestAsk := ob.Asks[0] // Lấy thằng bán rẻ nhất

			// Nếu giá mua thấp hơn giá bán rẻ nhất -> Không khớp được -> Dừng
			if order.Price < bestAsk.Price {
				break
			}

			// Tính số lượng khớp (min của 2 bên)
			qtyNeeded := order.Amount - order.Filled
			qtyAvailable := bestAsk.Amount - bestAsk.Filled
			tradeQty := qtyNeeded

			if qtyAvailable < qtyNeeded {
				tradeQty = qtyAvailable
			}

			// Ghi nhận trade
			trades = append(trades, Trade{
				MakerOrderID: bestAsk.ID,
				TakerOrderID: order.ID,
				Price:        bestAsk.Price, // Khớp theo giá của người treo lệnh (Maker)
				Amount:       tradeQty,
				CreatedAt:    time.Now(),
			})

			// Cập nhật số lượng đã khớp
			bestAsk.Filled += tradeQty
			order.Filled += tradeQty

			// Nếu lệnh treo (Maker) đã khớp hết -> Xóa khỏi sổ
			if bestAsk.Filled >= bestAsk.Amount {
				ob.Asks = ob.Asks[1:] // Xóa phần tử đầu tiên
			}

			// Nếu lệnh mới (Taker) đã khớp hết -> Xong
			if order.Filled >= order.Amount {
				return trades, nil // Nil nghĩa là không cần thêm vào sổ nữa
			}
		}
	} else {
		// --- PHẦN ELSE (LOGIC SELL) ---
		for len(ob.Bids) > 0 {
			bestBid := ob.Bids[0] // Lấy người mua giá cao nhất

			// Nếu mình bán đắt hơn giá họ mua -> Không khớp -> Dừng
			if order.Price > bestBid.Price {
				break
			}

			// Tính toán số lượng khớp
			qtyNeeded := order.Amount - order.Filled
			qtyAvailable := bestBid.Amount - bestBid.Filled
			tradeQty := qtyNeeded

			if qtyAvailable < qtyNeeded {
				tradeQty = qtyAvailable
			}

			// Ghi nhận Trade
			trades = append(trades, Trade{
				MakerOrderID: bestBid.ID,   // Người treo lệnh mua
				TakerOrderID: order.ID,     // Mình (người bán)
				Price:        bestBid.Price, // Khớp theo giá người treo (Maker)
				Amount:       tradeQty,
				CreatedAt:    time.Now(),
			})

			bestBid.Filled += tradeQty
			order.Filled += tradeQty

			// Xóa lệnh mua nếu đã khớp hết
			if bestBid.Filled >= bestBid.Amount {
				ob.Bids = ob.Bids[1:]
			}

			// Nếu lệnh bán của mình đã khớp hết -> Xong
			if order.Filled >= order.Amount {
				return trades, nil
			}
		}
	}

	// Nếu chạy hết vòng lặp mà lệnh vẫn chưa khớp hết -> Thêm phần dư vào sổ
	ob.AddOrder(order)
	return trades, order // order này sẽ được lưu vào RAM
}