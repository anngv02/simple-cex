# Simple CEX - Cryptocurrency Exchange Platform

Một sàn giao dịch tiền điện tử đơn giản được xây dựng với Go (backend) và React (frontend), hỗ trợ giao dịch BTC/USDT với các tính năng cơ bản của một sàn giao dịch.

## 📋 Tổng quan sản phẩm

Simple CEX là một nền tảng giao dịch tiền điện tử mini với các tính năng chính:

- **Order Matching Engine**: Hệ thống khớp lệnh tự động với thuật toán price-time priority
- **Orderbook**: Hiển thị sổ lệnh real-time với 10 giá tốt nhất mỗi bên (Bid/Ask)
- **Candlestick Chart**: Biểu đồ nến với nhiều khung thời gian (1m, 5m, 15m, 1h) sử dụng TradingView Lightweight Charts
- **Real-time Updates**: Cập nhật dữ liệu real-time qua WebSocket
- **Market Simulation**: Tool giả lập giao dịch với nhiều loại traders (Market Maker, Big Traders, Small Traders)
- **Balance Management**: Quản lý số dư với cơ chế lock/unlock khi đặt lệnh

## 🛠️ Yêu cầu hệ thống

- **Go**: >= 1.24
- **Node.js**: >= 18.x
- **PostgreSQL**: >= 12.x
- **npm** hoặc **yarn**

## 📦 Cài đặt

### 1. Cài đặt PostgreSQL

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

**macOS:**
```bash
brew install postgresql
brew services start postgresql
```

**Windows:**
Tải và cài đặt từ [PostgreSQL Downloads](https://www.postgresql.org/download/windows/)

### 2. Cài đặt Go

**Ubuntu/Debian:**
```bash
wget https://go.dev/dl/go1.24.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

**macOS:**
```bash
brew install go
```

**Windows:**
Tải và cài đặt từ [Go Downloads](https://go.dev/dl/)

### 3. Cài đặt Node.js

**Ubuntu/Debian:**
```bash
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
```

**macOS:**
```bash
brew install node
```

**Windows:**
Tải và cài đặt từ [Node.js Downloads](https://nodejs.org/)

## 🚀 Cách chạy

### Bước 1: Thiết lập Database

1. Tạo database và user:
```bash
sudo -u postgres psql
```

Trong PostgreSQL shell:
```sql
CREATE DATABASE cexdb;
CREATE USER cex WITH PASSWORD 'cexpass';
GRANT ALL PRIVILEGES ON DATABASE cexdb TO cex;
\q
```

2. Khởi tạo schema và seed data:
```bash
psql -U cex -d cexdb -f db/init.sql
psql -U cex -d cexdb -f db/seed_simulation.sql
```

**Lưu ý**: Nếu database đã tồn tại và cần cập nhật balance, chạy:
```bash
psql -U cex -d cexdb -f db/fix_user1.sql
```

### Bước 2: Cài đặt dependencies Backend

```bash
cd /home/annez02/simple-cex
go mod download
```

### Bước 3: Chạy Backend Server

```bash
cd backend
go run main.go db.go
```

Backend sẽ chạy tại `http://localhost:8010`

**API Endpoints:**
- `POST /order` - Đặt lệnh mua/bán
- `GET /orderbook/:symbol` - Lấy orderbook
- `GET /trades/:symbol?interval=1m&limit=100` - Lấy dữ liệu OHLCV cho chart
- `GET /ws` - WebSocket connection

### Bước 4: Cài đặt và chạy Frontend

Mở terminal mới:
```bash
cd frontend
npm install
npm run dev
```

Frontend sẽ chạy tại `http://localhost:5173` (hoặc port khác nếu 5173 đã được sử dụng)

### Bước 5: (Tùy chọn) Chạy Market Simulation

Mở terminal mới để chạy simulation tạo giao dịch giả lập:
```bash
cd simulation
go run main.go
```

Simulation sẽ tạo:
- **Market Maker** (User 1): Rải lệnh mỗi 2 giây để duy trì orderbook
- **Big Traders** (User 2-6): Giao dịch 10k-50k USD, mỗi 1 phút
- **Small Traders** (User 7-10): Giao dịch 1k-10k USD, mỗi 3 giây

## 📁 Cấu trúc thư mục

```
simple-cex/
├── api/              # API server (Gin framework)
├── backend/          # Backend entry point
├── engine/           # Core matching engine logic
│   ├── manager.go    # Order processing & settlement
│   ├── orderbook.go  # Orderbook data structure
│   └── accouting.go  # Balance management
├── db/               # Database scripts
│   ├── init.sql      # Schema & initial data
│   ├── seed_simulation.sql  # Seed data for simulation
│   └── fix_user1.sql # Fix balance script
├── frontend/         # React frontend
│   └── src/
│       ├── App.tsx
│       └── components/
│           ├── CandlestickChart.tsx
│           ├── Orderbook.tsx
│           └── OrderForm.tsx
└── simulation/       # Market simulation tool
```

## 📊 Tính năng chính

### Order Matching
- Thuật toán khớp lệnh price-time priority
- Hỗ trợ limit orders (BUY/SELL)
- Tự động settlement sau khi khớp

### Real-time Updates
- WebSocket cho orderbook updates
- WebSocket cho trade updates
- Chart tự động cập nhật mỗi 1 giây

### Chart Features
- Candlestick chart với TradingView Lightweight Charts
- Hỗ trợ nhiều timeframe: 1m, 5m, 15m, 1h
- Tính toán OHLCV từ dữ liệu trades

## 🧪 Testing

Để test API bằng curl:
```bash
# Đặt lệnh mua
curl -X POST http://localhost:8010/order \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "symbol": "BTC_USDT", "side": "BUY", "price": 50000, "amount": 0.1}'

# Lấy orderbook
curl http://localhost:8010/orderbook/BTC_USDT

# Lấy dữ liệu chart
curl http://localhost:8010/trades/BTC_USDT?interval=1m&limit=100
```

## 📝 Lưu ý

- Đây là một project demo/educational, không nên sử dụng trong production
- Cần thêm authentication/authorization cho production
- Cần thêm rate limiting và security measures
- Database connection string nên được config qua environment variables

## 📄 License

MIT
