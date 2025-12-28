-- USERS
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ASSETS
CREATE TABLE assets (
    symbol VARCHAR(10) PRIMARY KEY,
    precision INT DEFAULT 8
);

-- BALANCES (CORE)
CREATE TABLE balances (
    user_id INT REFERENCES users(id),
    asset_symbol VARCHAR(10) REFERENCES assets(symbol),
    available DECIMAL(20, 8) DEFAULT 0,
    locked DECIMAL(20, 8) DEFAULT 0,
    PRIMARY KEY (user_id, asset_symbol),
    CHECK (available >= 0),
    CHECK (locked >= 0)
);

-- ORDERS
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(4) NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    filled DECIMAL(20, 8) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'OPEN',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- TRADES
CREATE TABLE trades (
    id SERIAL PRIMARY KEY,
    maker_order_id INT REFERENCES orders(id),
    taker_order_id INT REFERENCES orders(id),
    price DECIMAL(20, 8) NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- SEED DATA
INSERT INTO assets(symbol, precision) VALUES
('BTC', 8),
('USDT', 6);

INSERT INTO users(email, password_hash)
VALUES ('userA@test.com', 'hash');

INSERT INTO balances(user_id, asset_symbol, available)
VALUES
(1, 'USDT', 100000),
(1, 'BTC', 0);
