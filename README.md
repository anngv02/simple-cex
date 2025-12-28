# Simple CEX - Cryptocurrency Exchange Platform

A simple cryptocurrency exchange platform built with Go (backend) and React (frontend), supporting BTC/USDT trading with basic exchange features.

## 📋 Product Overview

Simple CEX is a mini cryptocurrency trading platform with the following main features:

- **Order Matching Engine**: Automatic order matching system with price-time priority algorithm
- **Orderbook**: Real-time order book display with top 10 best prices per side (Bid/Ask)
- **Candlestick Chart**: Candlestick chart with multiple timeframes (1m, 5m, 15m, 1h) using TradingView Lightweight Charts
- **Real-time Updates**: Real-time data updates via WebSocket
- **Market Simulation**: Trading simulation tool with multiple trader types (Market Maker, Big Traders, Small Traders)
- **Balance Management**: Balance management with lock/unlock mechanism when placing orders

## 🛠️ System Requirements

- **Go**: >= 1.24
- **Node.js**: >= 18.x
- **PostgreSQL**: >= 12.x
- **npm** or **yarn**

## 📦 Installation

### 1. Install PostgreSQL

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
Download and install from [PostgreSQL Downloads](https://www.postgresql.org/download/windows/)

### 2. Install Go

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
Download and install from [Go Downloads](https://go.dev/dl/)

### 3. Install Node.js

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
Download and install from [Node.js Downloads](https://nodejs.org/)

## 🚀 How to Run

### Step 1: Setup Database

1. Create database and user:
```bash
sudo -u postgres psql
```

In PostgreSQL shell:
```sql
CREATE DATABASE cexdb;
CREATE USER cex WITH PASSWORD 'cexpass';
GRANT ALL PRIVILEGES ON DATABASE cexdb TO cex;
\q
```

2. Initialize schema and seed data:
```bash
psql -U cex -d cexdb -f db/init.sql
psql -U cex -d cexdb -f db/seed_simulation.sql
```

**Note**: If the database already exists and you need to update balances, run:
```bash
psql -U cex -d cexdb -f db/fix_user1.sql
```

### Step 2: Install Backend Dependencies

```bash
cd /home/annez02/simple-cex
go mod download
```

### Step 3: Run Backend Server

```bash
cd backend
go run main.go db.go
```

Backend will run at `http://localhost:8010`

**API Endpoints:**
- `POST /order` - Place buy/sell order
- `GET /orderbook/:symbol` - Get orderbook
- `GET /trades/:symbol?interval=1m&limit=100` - Get OHLCV data for chart
- `GET /ws` - WebSocket connection

### Step 4: Install and Run Frontend

Open a new terminal:
```bash
cd frontend
npm install
npm run dev
```

Frontend will run at `http://localhost:5173` (or another port if 5173 is already in use)

### Step 5: (Optional) Run Market Simulation

Open a new terminal to run simulation that generates mock trades:
```bash
cd simulation
go run main.go
```

Simulation will create:
- **Market Maker** (User 1): Places orders every 1.5-3 seconds to maintain orderbook
- **Big Traders** (User 2-6): Trade 5k-100k USD, every 1 minute
- **Small Traders** (User 7-10): Trade 500-20k USD, every 3 seconds

## 🐳 Docker Setup

### Using Docker Compose

1. Build and run all services:
```bash
docker compose build
docker compose up -d
```

2. View logs:
```bash
docker compose logs -f
```

3. Stop services:
```bash
docker compose down
```

### Access Points

- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8010
- **PostgreSQL**: localhost:5432

## 📁 Directory Structure

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

## 📊 Main Features

### Order Matching
- Price-time priority matching algorithm
- Support for limit orders (BUY/SELL)
- Automatic settlement after matching

### Real-time Updates
- WebSocket for orderbook updates
- WebSocket for trade updates
- Chart automatically updates every 1 second

### Chart Features
- Candlestick chart with TradingView Lightweight Charts
- Support for multiple timeframes: 1m, 5m, 15m, 1h
- OHLCV calculation from trade data

## 🧪 Testing

To test API with curl:
```bash
# Place buy order
curl -X POST http://localhost:8010/order \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "symbol": "BTC_USDT", "side": "BUY", "price": 50000, "amount": 0.1}'

# Get orderbook
curl http://localhost:8010/orderbook/BTC_USDT

# Get chart data
curl http://localhost:8010/trades/BTC_USDT?interval=1m&limit=100
```

## 📝 Notes

- This is a demo/educational project, should not be used in production
- Authentication/authorization needed for production
- Rate limiting and security measures needed
- Database connection string should be configured via environment variables

## 📄 License

MIT
