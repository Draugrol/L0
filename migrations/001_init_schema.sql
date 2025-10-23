-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number VARCHAR(255) NOT NULL,
    entry VARCHAR(255) NOT NULL,
    locale VARCHAR(10),
    internal_signature VARCHAR(255),
    customer_id VARCHAR(255),
    delivery_service VARCHAR(255),
    shardkey VARCHAR(10),
    sm_id INTEGER DEFAULT 0,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    oof_shard VARCHAR(10)
);

-- Create delivery table
CREATE TABLE IF NOT EXISTS delivery (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) UNIQUE NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    zip VARCHAR(20),
    city VARCHAR(255),
    address TEXT,
    region VARCHAR(255),
    email VARCHAR(255)
);

-- Create payment table
CREATE TABLE IF NOT EXISTS payment (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) UNIQUE NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction VARCHAR(255) NOT NULL,
    request_id VARCHAR(255),
    currency VARCHAR(10),
    provider VARCHAR(255),
    amount INTEGER DEFAULT 0,
    payment_dt BIGINT,
    bank VARCHAR(255),
    delivery_cost INTEGER DEFAULT 0,
    goods_total INTEGER DEFAULT 0,
    custom_fee INTEGER DEFAULT 0
);

-- Create items table
CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id INTEGER,
    track_number VARCHAR(255),
    price INTEGER,
    rid VARCHAR(255),
    name VARCHAR(255),
    sale INTEGER,
    size VARCHAR(50),
    total_price INTEGER,
    nm_id INTEGER,
    brand VARCHAR(255),
    status INTEGER
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_orders_date ON orders(date_created DESC);
CREATE INDEX IF NOT EXISTS idx_items_order_uid ON items(order_uid);
