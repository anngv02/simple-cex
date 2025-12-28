-- Seed data cho simulation
-- Tạo thêm users và balances cho user 1-10

-- Tạo users (nếu chưa có)
INSERT INTO users(email, password_hash)
SELECT 
    'user' || generate_series || '@test.com',
    'hash'
FROM generate_series(2, 10)
ON CONFLICT (email) DO NOTHING;

-- Tạo balances cho tất cả users (1-10) với số lượng lớn
-- User 1: Market Maker - cần nhiều balance hơn
-- User 2-6: Big Traders - cần 1 triệu USD để giao dịch lớn
-- User 7-10: Small Traders - cần balance vừa phải

-- Tạo USDT balance cho tất cả users (1-10)
INSERT INTO balances(user_id, asset_symbol, available)
SELECT 
    generate_series,
    'USDT',
    CASE 
        WHEN generate_series = 1 THEN 500000.0      -- User 1 (Market Maker): 500k USDT
        WHEN generate_series BETWEEN 2 AND 6 THEN 1000000.0  -- User 2-6 (Big Traders): 1 triệu USDT
        ELSE 500000.0                               -- User 7-10 (Small Traders): 500k USDT
    END
FROM generate_series(1, 10)
ON CONFLICT (user_id, asset_symbol) DO UPDATE SET available = EXCLUDED.available;

-- Tạo BTC balance cho tất cả users (1-10)
-- Với giá BTC ~50k: 1 triệu USD = ~20 BTC, nhưng cho thêm để có thể SELL
INSERT INTO balances(user_id, asset_symbol, available)
SELECT 
    generate_series,
    'BTC',
    CASE 
        WHEN generate_series = 1 THEN 50.0          -- User 1 (Market Maker): 50 BTC
        WHEN generate_series BETWEEN 2 AND 6 THEN 25.0  -- User 2-6 (Big Traders): 25 BTC (~1.25 triệu USD)
        ELSE 50.0                                   -- User 7-10 (Small Traders): 50 BTC
    END
FROM generate_series(1, 10)
ON CONFLICT (user_id, asset_symbol) DO UPDATE SET available = EXCLUDED.available;

