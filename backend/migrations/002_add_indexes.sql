-- Migration: 002_add_indexes
-- Description: Performance indexes for trading engine
-- Author: Trading Engine Team
-- Date: 2026-01-18

-- ============================================================================
-- UP Migration
-- ============================================================================

-- Users indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_is_active ON users(is_active) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_kyc_status ON users(kyc_status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

-- User roles indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_roles_role ON user_roles(role);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_roles_expires_at ON user_roles(expires_at) WHERE expires_at IS NOT NULL;

-- Accounts indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounts_user_id ON accounts(user_id) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounts_account_number ON accounts(account_number) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounts_account_type ON accounts(account_type) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounts_is_active ON accounts(is_active) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounts_broker_id ON accounts(broker_id) WHERE broker_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounts_created_at ON accounts(created_at DESC);

-- Account transactions indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_account_transactions_account_id ON account_transactions(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_account_transactions_type ON account_transactions(transaction_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_account_transactions_status ON account_transactions(status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_account_transactions_created_at ON account_transactions(created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_account_transactions_reference ON account_transactions(reference_id, reference_type) WHERE reference_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_account_transactions_processed_at ON account_transactions(processed_at DESC) WHERE processed_at IS NOT NULL;

-- Orders indexes (critical for performance)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_account_id ON orders(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_symbol ON orders(symbol);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_client_order_id ON orders(client_order_id) WHERE client_order_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_lp_order_id ON orders(lp_order_id) WHERE lp_order_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_parent_order_id ON orders(parent_order_id) WHERE parent_order_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_submitted_at ON orders(submitted_at DESC) WHERE submitted_at IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_filled_at ON orders(filled_at DESC) WHERE filled_at IS NOT NULL;

-- Composite indexes for common queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_account_symbol_status ON orders(account_id, symbol, status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_symbol_status_created ON orders(symbol, status, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_user_status_created ON orders(user_id, status, created_at DESC);

-- Order fills indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_order_fills_order_id ON order_fills(order_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_order_fills_created_at ON order_fills(created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_order_fills_execution_id ON order_fills(execution_id) WHERE execution_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_order_fills_venue ON order_fills(venue) WHERE venue IS NOT NULL;

-- Positions indexes (critical for real-time queries)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_account_id ON positions(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_user_id ON positions(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_symbol ON positions(symbol);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_is_open ON positions(is_open);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_open_time ON positions(open_time DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_close_time ON positions(close_time DESC) WHERE close_time IS NOT NULL;

-- Composite indexes for positions
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_account_symbol_open ON positions(account_id, symbol, is_open);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_user_is_open ON positions(user_id, is_open);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_symbol_is_open ON positions(symbol, is_open);

-- Trades indexes (for reporting and analytics)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_account_id ON trades(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_user_id ON trades(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_position_id ON trades(position_id) WHERE position_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_order_id ON trades(order_id) WHERE order_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_symbol ON trades(symbol);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_execution_id ON trades(execution_id) WHERE execution_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_executed_at ON trades(executed_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_created_at ON trades(created_at DESC);

-- Composite indexes for trades
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_account_symbol_executed ON trades(account_id, symbol, executed_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_user_executed ON trades(user_id, executed_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_symbol_executed ON trades(symbol, executed_at DESC);

-- Instruments indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_instruments_symbol ON instruments(symbol) WHERE is_active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_instruments_type ON instruments(instrument_type) WHERE is_active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_instruments_is_tradable ON instruments(is_tradable);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_instruments_base_currency ON instruments(base_currency) WHERE is_active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_instruments_quote_currency ON instruments(quote_currency) WHERE is_active = true;

-- JSONB indexes for metadata queries (GIN indexes)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_metadata ON orders USING GIN(metadata) WHERE metadata IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_metadata ON positions USING GIN(metadata) WHERE metadata IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_metadata ON trades USING GIN(metadata) WHERE metadata IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_instruments_metadata ON instruments USING GIN(metadata) WHERE metadata IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_account_transactions_metadata ON account_transactions USING GIN(metadata) WHERE metadata IS NOT NULL;

-- Text search indexes for common queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email_trgm ON users USING GIN(email gin_trgm_ops);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_instruments_display_name_trgm ON instruments USING GIN(display_name gin_trgm_ops);

-- BRIN indexes for time-series data (very efficient for large tables)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trades_executed_at_brin ON trades USING BRIN(executed_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_created_at_brin ON orders USING BRIN(created_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_account_transactions_created_at_brin ON account_transactions USING BRIN(created_at);

-- ============================================================================
-- Statistics and Query Optimization
-- ============================================================================

-- Increase statistics for better query planning on high-cardinality columns
ALTER TABLE orders ALTER COLUMN symbol SET STATISTICS 1000;
ALTER TABLE orders ALTER COLUMN status SET STATISTICS 1000;
ALTER TABLE positions ALTER COLUMN symbol SET STATISTICS 1000;
ALTER TABLE trades ALTER COLUMN symbol SET STATISTICS 1000;

-- Analyze tables to update statistics
ANALYZE users;
ANALYZE user_roles;
ANALYZE accounts;
ANALYZE account_transactions;
ANALYZE orders;
ANALYZE order_fills;
ANALYZE positions;
ANALYZE trades;
ANALYZE instruments;

-- ============================================================================
-- DOWN Migration
-- ============================================================================

-- To rollback, uncomment and run the following:
/*
-- Reset statistics
ALTER TABLE orders ALTER COLUMN symbol SET STATISTICS -1;
ALTER TABLE orders ALTER COLUMN status SET STATISTICS -1;
ALTER TABLE positions ALTER COLUMN symbol SET STATISTICS -1;
ALTER TABLE trades ALTER COLUMN symbol SET STATISTICS -1;

-- Drop BRIN indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_account_transactions_created_at_brin;
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_created_at_brin;
DROP INDEX CONCURRENTLY IF EXISTS idx_trades_executed_at_brin;

-- Drop text search indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_instruments_display_name_trgm;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_email_trgm;

-- Drop JSONB indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_account_transactions_metadata;
DROP INDEX CONCURRENTLY IF EXISTS idx_instruments_metadata;
DROP INDEX CONCURRENTLY IF EXISTS idx_trades_metadata;
DROP INDEX CONCURRENTLY IF EXISTS idx_positions_metadata;
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_metadata;

-- Drop instruments indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_instruments_quote_currency;
DROP INDEX CONCURRENTLY IF EXISTS idx_instruments_base_currency;
DROP INDEX CONCURRENTLY IF EXISTS idx_instruments_is_tradable;
DROP INDEX CONCURRENTLY IF EXISTS idx_instruments_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_instruments_symbol;

-- Drop trades indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_trades_symbol_executed;
DROP INDEX CONCURRENTLY IF EXISTS idx_trades_user_executed;
DROP INDEX CONCURRENTLY IF EXISTS idx_trades_account_symbol_executed;
DROP INDEX CONCURRENTLY IF EXISTS idx_trades_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_trades_executed_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_trades_execution_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_trades_symbol;
DROP INDEX CONCURRENTLY IF EXISTS idx_trades_order_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_trades_position_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_trades_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_trades_account_id;

-- Drop positions indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_positions_symbol_is_open;
DROP INDEX CONCURRENTLY IF EXISTS idx_positions_user_is_open;
DROP INDEX CONCURRENTLY IF EXISTS idx_positions_account_symbol_open;
DROP INDEX CONCURRENTLY IF EXISTS idx_positions_close_time;
DROP INDEX CONCURRENTLY IF EXISTS idx_positions_open_time;
DROP INDEX CONCURRENTLY IF EXISTS idx_positions_is_open;
DROP INDEX CONCURRENTLY IF EXISTS idx_positions_symbol;
DROP INDEX CONCURRENTLY IF EXISTS idx_positions_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_positions_account_id;

-- Drop order fills indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_order_fills_venue;
DROP INDEX CONCURRENTLY IF EXISTS idx_order_fills_execution_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_order_fills_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_order_fills_order_id;

-- Drop orders indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_user_status_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_symbol_status_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_account_symbol_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_filled_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_submitted_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_parent_order_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_lp_order_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_client_order_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_symbol;
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_account_id;

-- Drop account transactions indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_account_transactions_processed_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_account_transactions_reference;
DROP INDEX CONCURRENTLY IF EXISTS idx_account_transactions_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_account_transactions_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_account_transactions_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_account_transactions_account_id;

-- Drop accounts indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_accounts_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_accounts_broker_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_accounts_is_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_accounts_account_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_accounts_account_number;
DROP INDEX CONCURRENTLY IF EXISTS idx_accounts_user_id;

-- Drop user roles indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_user_roles_expires_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_user_roles_role;
DROP INDEX CONCURRENTLY IF EXISTS idx_user_roles_user_id;

-- Drop users indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_users_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_kyc_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_is_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_username;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_email;
*/
