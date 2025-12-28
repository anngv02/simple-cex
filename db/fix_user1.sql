-- Fix balance cho tất cả users (1-10): Tăng BTC và USDT lên nhiều hơn
-- User 1 (Market Maker): 5000k USDT, 500 BTC
-- User 2-6 (Big Traders): 1000k USDT, 250 BTC
-- User 7-10 (Small Traders): 500k USDT, 50 BTC

-- Update USDT balance cho User 1 (Market Maker)
UPDATE balances 
SET available = 5000000.0 
WHERE user_id = 1 AND asset_symbol = 'USDT';

-- Update USDT balance cho User 2-6 (Big Traders) - 1 triệu USD
UPDATE balances 
SET available = 1000000.0  
WHERE user_id BETWEEN 2 AND 6 AND asset_symbol = 'USDT';

-- Update USDT balance cho User 7-10 (Small Traders)
UPDATE balances 
SET available = 500000.0 
WHERE user_id BETWEEN 7 AND 10 AND asset_symbol = 'USDT';

-- Update BTC balance cho User 1 (Market Maker)
UPDATE balances 
SET available = 500.0 
WHERE user_id = 1 AND asset_symbol = 'BTC';

-- Update BTC balance cho User 2-6 (Big Traders) - 25 BTC (~1.25 triệu USD)
UPDATE balances 
SET available = 25.0 
WHERE user_id BETWEEN 2 AND 6 AND asset_symbol = 'BTC';

-- Update BTC balance cho User 7-10 (Small Traders)
UPDATE balances 
SET available = 50.0 
WHERE user_id BETWEEN 7 AND 10 AND asset_symbol = 'BTC';

-- Nếu chưa có balance, tạo mới cho tất cả users (1-10)
INSERT INTO balances(user_id, asset_symbol, available)
SELECT 
    generate_series,
    'USDT',
    CASE 
        WHEN generate_series = 1 THEN 5000000.0
        WHEN generate_series BETWEEN 2 AND 6 THEN 1000000.0
        ELSE 500000.0
    END
FROM generate_series(1, 10)
ON CONFLICT (user_id, asset_symbol) DO UPDATE SET available = EXCLUDED.available;

INSERT INTO balances(user_id, asset_symbol, available)
SELECT 
    generate_series,
    'BTC',
    CASE 
        WHEN generate_series = 1 THEN 500.0
        WHEN generate_series BETWEEN 2 AND 6 THEN 25.0
        ELSE 50.0
    END
FROM generate_series(1, 10)
ON CONFLICT (user_id, asset_symbol) DO UPDATE SET available = EXCLUDED.available;

