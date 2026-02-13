-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create accounts table
CREATE TABLE IF NOT EXISTS accounts (
    account_id VARCHAR(50) PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    balance DECIMAL(18, 2) DEFAULT 0.00,
    equity DECIMAL(18, 2) DEFAULT 0.00,
    margin DECIMAL(18, 2) DEFAULT 0.00,
    free_margin DECIMAL(18, 2) DEFAULT 0.00,
    margin_level DECIMAL(10, 2) DEFAULT 0.00,
    currency VARCHAR(10) DEFAULT 'USD',
    leverage INTEGER DEFAULT 100,
    account_type VARCHAR(20) DEFAULT 'demo',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create positions table
CREATE TABLE IF NOT EXISTS positions (
    position_id SERIAL PRIMARY KEY,
    account_id VARCHAR(50) REFERENCES accounts(account_id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    volume DECIMAL(18, 8) NOT NULL,
    open_price DECIMAL(18, 8) NOT NULL,
    current_price DECIMAL(18, 8),
    stop_loss DECIMAL(18, 8),
    take_profit DECIMAL(18, 8),
    profit DECIMAL(18, 2) DEFAULT 0.00,
    swap DECIMAL(18, 2) DEFAULT 0.00,
    commission DECIMAL(18, 2) DEFAULT 0.00,
    opened_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    order_id SERIAL PRIMARY KEY,
    account_id VARCHAR(50) REFERENCES accounts(account_id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    order_type VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    volume DECIMAL(18, 8) NOT NULL,
    price DECIMAL(18, 8),
    stop_loss DECIMAL(18, 8),
    take_profit DECIMAL(18, 8),
    status VARCHAR(20) DEFAULT 'pending',
    filled_volume DECIMAL(18, 8) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    executed_at TIMESTAMP
);

-- Create trades table (historical)
CREATE TABLE IF NOT EXISTS trades (
    trade_id SERIAL PRIMARY KEY,
    account_id VARCHAR(50) REFERENCES accounts(account_id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    volume DECIMAL(18, 8) NOT NULL,
    open_price DECIMAL(18, 8) NOT NULL,
    close_price DECIMAL(18, 8) NOT NULL,
    profit DECIMAL(18, 2) NOT NULL,
    swap DECIMAL(18, 2) DEFAULT 0.00,
    commission DECIMAL(18, 2) DEFAULT 0.00,
    opened_at TIMESTAMP NOT NULL,
    closed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create sessions table
CREATE TABLE IF NOT EXISTS sessions (
    session_id VARCHAR(255) PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_positions_account_id ON positions(account_id);
CREATE INDEX IF NOT EXISTS idx_orders_account_id ON orders(account_id);
CREATE INDEX IF NOT EXISTS idx_trades_account_id ON trades(account_id);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Insert admin user (password: Admin@123, bcrypt hash)
-- Hash generated for "Admin@123"
INSERT INTO users (username, password_hash, role) 
VALUES ('admin', '$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5kuWEGhVBxOEC', 'admin')
ON CONFLICT (username) DO NOTHING;

-- Insert demo user
INSERT INTO users (username, password_hash, role) 
VALUES ('demo', '$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5kuWEGhVBxOEC', 'user')
ON CONFLICT (username) DO NOTHING;

-- Create accounts for the users
INSERT INTO accounts (account_id, user_id, balance, equity, free_margin, currency, leverage, account_type)
SELECT 'ADMIN001', id, 100000.00, 100000.00, 100000.00, 'USD', 100, 'demo'
FROM users WHERE username = 'admin'
ON CONFLICT (account_id) DO NOTHING;

INSERT INTO accounts (account_id, user_id, balance, equity, free_margin, currency, leverage, account_type)
SELECT 'DEMO001', id, 10000.00, 10000.00, 10000.00, 'USD', 100, 'demo'
FROM users WHERE username = 'demo'
ON CONFLICT (account_id) DO NOTHING;

-- Display created users
SELECT 'Created Users:' as info;
SELECT id, username, role, created_at FROM users;

-- Display created accounts
SELECT 'Created Accounts:' as info;
SELECT account_id, user_id, balance, equity, currency, leverage, account_type FROM accounts;
