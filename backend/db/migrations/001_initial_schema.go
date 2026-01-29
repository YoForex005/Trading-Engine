package migrations

import (
	"database/sql"
)

func init() {
	RegisterMigration(&Migration{
		Version: 1,
		Name:    "initial_schema",
		Up:      initialSchemaUp,
		Down:    initialSchemaDown,
	})
}

func initialSchemaUp(tx *sql.Tx) error {
	schema := `
	-- Users table
	CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		full_name VARCHAR(255),
		group_id BIGINT DEFAULT 1,
		status VARCHAR(50) DEFAULT 'ACTIVE',
		kyc_level INT DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_login TIMESTAMP
	);

	CREATE INDEX idx_users_email ON users(email);
	CREATE INDEX idx_users_group_id ON users(group_id);
	CREATE INDEX idx_users_status ON users(status);

	-- Trading accounts table
	CREATE TABLE IF NOT EXISTS accounts (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		account_number VARCHAR(50) UNIQUE NOT NULL,
		account_type VARCHAR(50) NOT NULL DEFAULT 'DEMO',
		currency VARCHAR(10) NOT NULL DEFAULT 'USD',
		balance DECIMAL(20, 5) NOT NULL DEFAULT 0,
		equity DECIMAL(20, 5) NOT NULL DEFAULT 0,
		margin DECIMAL(20, 5) NOT NULL DEFAULT 0,
		free_margin DECIMAL(20, 5) NOT NULL DEFAULT 0,
		margin_level DECIMAL(10, 2) DEFAULT 0,
		leverage INT NOT NULL DEFAULT 100,
		credit DECIMAL(20, 5) DEFAULT 0,
		peak_equity DECIMAL(20, 5) DEFAULT 0,
		status VARCHAR(50) DEFAULT 'ACTIVE',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX idx_accounts_user_id ON accounts(user_id);
	CREATE INDEX idx_accounts_account_number ON accounts(account_number);
	CREATE INDEX idx_accounts_status ON accounts(status);

	-- Trading groups table
	CREATE TABLE IF NOT EXISTS trading_groups (
		id BIGSERIAL PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		description TEXT,
		execution_mode VARCHAR(50) NOT NULL DEFAULT 'BBOOK',
		markup DECIMAL(10, 5) DEFAULT 0,
		commission DECIMAL(10, 2) DEFAULT 0,
		max_leverage INT DEFAULT 100,
		default_balance DECIMAL(20, 5) DEFAULT 10000,
		margin_mode VARCHAR(50) DEFAULT 'HEDGING',
		status VARCHAR(50) DEFAULT 'ACTIVE',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_by VARCHAR(255)
	);

	CREATE INDEX idx_trading_groups_status ON trading_groups(status);

	-- Positions table
	CREATE TABLE IF NOT EXISTS positions (
		id BIGSERIAL PRIMARY KEY,
		account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
		symbol VARCHAR(50) NOT NULL,
		side VARCHAR(10) NOT NULL CHECK (side IN ('BUY', 'SELL')),
		volume DECIMAL(20, 5) NOT NULL,
		open_price DECIMAL(20, 10) NOT NULL,
		current_price DECIMAL(20, 10),
		unrealized_pnl DECIMAL(20, 5) DEFAULT 0,
		realized_pnl DECIMAL(20, 5) DEFAULT 0,
		swap DECIMAL(20, 5) DEFAULT 0,
		commission DECIMAL(20, 5) DEFAULT 0,
		stop_loss DECIMAL(20, 10),
		take_profit DECIMAL(20, 10),
		open_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		close_time TIMESTAMP,
		close_price DECIMAL(20, 10),
		status VARCHAR(50) DEFAULT 'OPEN',
		magic_number BIGINT,
		comment TEXT
	);

	CREATE INDEX idx_positions_account_id ON positions(account_id);
	CREATE INDEX idx_positions_symbol ON positions(symbol);
	CREATE INDEX idx_positions_status ON positions(status);
	CREATE INDEX idx_positions_open_time ON positions(open_time);

	-- Orders table
	CREATE TABLE IF NOT EXISTS orders (
		id BIGSERIAL PRIMARY KEY,
		account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
		client_order_id VARCHAR(255) UNIQUE,
		lp_order_id VARCHAR(255),
		symbol VARCHAR(50) NOT NULL,
		side VARCHAR(10) NOT NULL CHECK (side IN ('BUY', 'SELL')),
		order_type VARCHAR(50) NOT NULL,
		volume DECIMAL(20, 5) NOT NULL,
		price DECIMAL(20, 10),
		stop_loss DECIMAL(20, 10),
		take_profit DECIMAL(20, 10),
		execution_mode VARCHAR(50) DEFAULT 'BBOOK',
		selected_lp VARCHAR(50),
		status VARCHAR(50) DEFAULT 'PENDING',
		reject_reason TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		executed_at TIMESTAMP,
		magic_number BIGINT,
		comment TEXT
	);

	CREATE INDEX idx_orders_account_id ON orders(account_id);
	CREATE INDEX idx_orders_client_order_id ON orders(client_order_id);
	CREATE INDEX idx_orders_symbol ON orders(symbol);
	CREATE INDEX idx_orders_status ON orders(status);
	CREATE INDEX idx_orders_created_at ON orders(created_at);

	-- Trades table (order fills)
	CREATE TABLE IF NOT EXISTS trades (
		id BIGSERIAL PRIMARY KEY,
		order_id BIGINT REFERENCES orders(id) ON DELETE SET NULL,
		position_id BIGINT REFERENCES positions(id) ON DELETE SET NULL,
		account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
		symbol VARCHAR(50) NOT NULL,
		side VARCHAR(10) NOT NULL CHECK (side IN ('BUY', 'SELL')),
		volume DECIMAL(20, 5) NOT NULL,
		price DECIMAL(20, 10) NOT NULL,
		commission DECIMAL(20, 5) DEFAULT 0,
		swap DECIMAL(20, 5) DEFAULT 0,
		pnl DECIMAL(20, 5) DEFAULT 0,
		trade_type VARCHAR(50) NOT NULL,
		execution_mode VARCHAR(50),
		lp_name VARCHAR(50),
		executed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		magic_number BIGINT,
		comment TEXT
	);

	CREATE INDEX idx_trades_account_id ON trades(account_id);
	CREATE INDEX idx_trades_order_id ON trades(order_id);
	CREATE INDEX idx_trades_position_id ON trades(position_id);
	CREATE INDEX idx_trades_symbol ON trades(symbol);
	CREATE INDEX idx_trades_executed_at ON trades(executed_at);

	-- Admin users table
	CREATE TABLE IF NOT EXISTS admin_users (
		id BIGSERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL CHECK (role IN ('super_admin', 'admin', 'support')),
		ip_whitelist TEXT[],
		status VARCHAR(50) DEFAULT 'ACTIVE',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_login TIMESTAMP,
		created_by VARCHAR(255)
	);

	CREATE INDEX idx_admin_users_username ON admin_users(username);
	CREATE INDEX idx_admin_users_role ON admin_users(role);
	CREATE INDEX idx_admin_users_status ON admin_users(status);

	-- Admin audit log table
	CREATE TABLE IF NOT EXISTS admin_audit_log (
		id BIGSERIAL PRIMARY KEY,
		admin_id BIGINT REFERENCES admin_users(id) ON DELETE SET NULL,
		admin_username VARCHAR(255) NOT NULL,
		action VARCHAR(255) NOT NULL,
		resource_type VARCHAR(100) NOT NULL,
		resource_id BIGINT,
		details JSONB,
		reason TEXT,
		ip_address VARCHAR(50),
		user_agent TEXT,
		status VARCHAR(50) DEFAULT 'SUCCESS',
		error_message TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX idx_admin_audit_log_admin_id ON admin_audit_log(admin_id);
	CREATE INDEX idx_admin_audit_log_action ON admin_audit_log(action);
	CREATE INDEX idx_admin_audit_log_resource_type ON admin_audit_log(resource_type);
	CREATE INDEX idx_admin_audit_log_created_at ON admin_audit_log(created_at);

	-- Risk alerts table
	CREATE TABLE IF NOT EXISTS risk_alerts (
		id VARCHAR(255) PRIMARY KEY,
		account_id BIGINT REFERENCES accounts(id) ON DELETE CASCADE,
		alert_type VARCHAR(100) NOT NULL,
		severity VARCHAR(50) NOT NULL,
		message TEXT NOT NULL,
		acknowledged BOOLEAN DEFAULT FALSE,
		acknowledged_at TIMESTAMP,
		acknowledged_by VARCHAR(255),
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX idx_risk_alerts_account_id ON risk_alerts(account_id);
	CREATE INDEX idx_risk_alerts_type ON risk_alerts(alert_type);
	CREATE INDEX idx_risk_alerts_severity ON risk_alerts(severity);
	CREATE INDEX idx_risk_alerts_acknowledged ON risk_alerts(acknowledged);
	CREATE INDEX idx_risk_alerts_created_at ON risk_alerts(created_at);

	-- Liquidation events table
	CREATE TABLE IF NOT EXISTS liquidation_events (
		id VARCHAR(255) PRIMARY KEY,
		account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
		trigger_reason VARCHAR(255) NOT NULL,
		margin_level_before DECIMAL(10, 2),
		positions_closed JSONB,
		total_pnl DECIMAL(20, 5),
		slippage DECIMAL(20, 5),
		executed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX idx_liquidation_events_account_id ON liquidation_events(account_id);
	CREATE INDEX idx_liquidation_events_executed_at ON liquidation_events(executed_at);

	-- FIX credentials table (encrypted)
	CREATE TABLE IF NOT EXISTS fix_credentials (
		user_id VARCHAR(255) PRIMARY KEY,
		sender_comp_id VARCHAR(255) NOT NULL,
		password_encrypted BYTEA NOT NULL,
		rate_limit_tier VARCHAR(50) NOT NULL DEFAULT 'basic',
		max_sessions INT DEFAULT 1,
		allowed_ips TEXT[],
		expires_at TIMESTAMP,
		status VARCHAR(50) DEFAULT 'ACTIVE',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_used TIMESTAMP
	);

	CREATE INDEX idx_fix_credentials_sender_comp_id ON fix_credentials(sender_comp_id);
	CREATE INDEX idx_fix_credentials_status ON fix_credentials(status);

	-- FIX sessions table
	CREATE TABLE IF NOT EXISTS fix_sessions (
		session_id VARCHAR(255) PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL REFERENCES fix_credentials(user_id) ON DELETE CASCADE,
		sender_comp_id VARCHAR(255) NOT NULL,
		ip_address VARCHAR(50),
		connected_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_activity TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		message_count BIGINT DEFAULT 0,
		order_count BIGINT DEFAULT 0,
		status VARCHAR(50) DEFAULT 'ACTIVE'
	);

	CREATE INDEX idx_fix_sessions_user_id ON fix_sessions(user_id);
	CREATE INDEX idx_fix_sessions_status ON fix_sessions(status);
	CREATE INDEX idx_fix_sessions_connected_at ON fix_sessions(connected_at);

	-- Symbol settings for trading groups
	CREATE TABLE IF NOT EXISTS group_symbol_settings (
		id BIGSERIAL PRIMARY KEY,
		group_id BIGINT NOT NULL REFERENCES trading_groups(id) ON DELETE CASCADE,
		symbol VARCHAR(50) NOT NULL,
		enabled BOOLEAN DEFAULT TRUE,
		markup DECIMAL(10, 5),
		commission DECIMAL(10, 2),
		min_volume DECIMAL(20, 5),
		max_volume DECIMAL(20, 5),
		step_volume DECIMAL(20, 5),
		UNIQUE(group_id, symbol)
	);

	CREATE INDEX idx_group_symbol_settings_group_id ON group_symbol_settings(group_id);
	CREATE INDEX idx_group_symbol_settings_symbol ON group_symbol_settings(symbol);

	-- Order history for analytics
	CREATE TABLE IF NOT EXISTS order_history (
		id BIGSERIAL PRIMARY KEY,
		account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
		symbol VARCHAR(50) NOT NULL,
		side VARCHAR(10) NOT NULL,
		volume DECIMAL(20, 5) NOT NULL,
		price DECIMAL(20, 10) NOT NULL,
		order_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX idx_order_history_account_id ON order_history(account_id);
	CREATE INDEX idx_order_history_symbol ON order_history(symbol);
	CREATE INDEX idx_order_history_order_time ON order_history(order_time);

	-- Correlation matrix
	CREATE TABLE IF NOT EXISTS symbol_correlations (
		symbol1 VARCHAR(50) NOT NULL,
		symbol2 VARCHAR(50) NOT NULL,
		correlation DECIMAL(5, 4) NOT NULL CHECK (correlation >= -1 AND correlation <= 1),
		calculated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (symbol1, symbol2)
	);

	CREATE INDEX idx_symbol_correlations_symbol1 ON symbol_correlations(symbol1);
	CREATE INDEX idx_symbol_correlations_symbol2 ON symbol_correlations(symbol2);
	`

	_, err := tx.Exec(schema)
	return err
}

func initialSchemaDown(tx *sql.Tx) error {
	dropTables := `
	DROP TABLE IF EXISTS symbol_correlations;
	DROP TABLE IF EXISTS order_history;
	DROP TABLE IF EXISTS group_symbol_settings;
	DROP TABLE IF EXISTS fix_sessions;
	DROP TABLE IF EXISTS fix_credentials;
	DROP TABLE IF EXISTS liquidation_events;
	DROP TABLE IF EXISTS risk_alerts;
	DROP TABLE IF EXISTS admin_audit_log;
	DROP TABLE IF EXISTS admin_users;
	DROP TABLE IF EXISTS trades;
	DROP TABLE IF EXISTS orders;
	DROP TABLE IF EXISTS positions;
	DROP TABLE IF EXISTS trading_groups;
	DROP TABLE IF EXISTS accounts;
	DROP TABLE IF EXISTS users;
	`

	_, err := tx.Exec(dropTables)
	return err
}
