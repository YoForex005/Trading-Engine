-- Migration 006: Payment Gateway Tables (PostgreSQL)
-- Description: Create comprehensive payment gateway tables for deposits, withdrawals, and reconciliation
-- Author: Backend API Developer Agent (Converted to PostgreSQL)
-- Date: 2026-01-18

-- =====================================================
-- Payment Transactions Table
-- =====================================================
CREATE TABLE IF NOT EXISTS payment_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('deposit', 'withdrawal', 'refund', 'chargeback')),
    method VARCHAR(30) NOT NULL CHECK (method IN (
        'card', 'bank_transfer', 'ach', 'sepa', 'wire',
        'paypal', 'skrill', 'neteller',
        'bitcoin', 'ethereum', 'usdt',
        'local_payment'
    )),
    provider VARCHAR(30) NOT NULL CHECK (provider IN (
        'stripe', 'braintree', 'paypal', 'coinbase', 'bitpay', 'circle', 'wise', 'internal'
    )),
    status VARCHAR(20) NOT NULL CHECK (status IN (
        'pending', 'processing', 'completed', 'failed', 'cancelled', 'refunded', 'disputed'
    )),
    amount DECIMAL(20, 8) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    fee DECIMAL(20, 8) NOT NULL DEFAULT 0,
    net_amount DECIMAL(20, 8) NOT NULL,
    exchange_rate DECIMAL(20, 8),
    provider_tx_id VARCHAR(128),
    payment_details JSONB,
    metadata JSONB,
    ip_address INET NOT NULL,
    device_id VARCHAR(128),
    country VARCHAR(2),
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    confirmations_required INTEGER DEFAULT 0,
    confirmations_received INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_payment_transactions_user_id ON payment_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_status ON payment_transactions(status);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_type ON payment_transactions(type);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_provider ON payment_transactions(provider);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_provider_tx_id ON payment_transactions(provider_tx_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_created_at ON payment_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_completed_at ON payment_transactions(completed_at);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_user_status ON payment_transactions(user_id, status);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_user_type ON payment_transactions(user_id, type);

-- =====================================================
-- User Balances Table
-- =====================================================
CREATE TABLE IF NOT EXISTS user_balances (
    user_id UUID NOT NULL,
    currency VARCHAR(10) NOT NULL,
    available_balance DECIMAL(20, 8) NOT NULL DEFAULT 0,
    reserved_balance DECIMAL(20, 8) NOT NULL DEFAULT 0,
    total_balance DECIMAL(20, 8) GENERATED ALWAYS AS (available_balance + reserved_balance) STORED,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, currency),
    CHECK (available_balance >= 0),
    CHECK (reserved_balance >= 0)
);

CREATE INDEX IF NOT EXISTS idx_user_balances_user_id ON user_balances(user_id);

-- =====================================================
-- Balance Ledger Table (Audit Trail)
-- =====================================================
CREATE TABLE IF NOT EXISTS balance_ledger (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    currency VARCHAR(10) NOT NULL,
    transaction_id UUID NOT NULL,
    operation VARCHAR(20) NOT NULL CHECK (operation IN ('credit', 'debit', 'reserve', 'unreserve', 'debit_reserved')),
    amount DECIMAL(20, 8) NOT NULL,
    balance_before DECIMAL(20, 8) NOT NULL,
    balance_after DECIMAL(20, 8) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (user_id, currency) REFERENCES user_balances(user_id, currency) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_balance_ledger_user_id ON balance_ledger(user_id);
CREATE INDEX IF NOT EXISTS idx_balance_ledger_transaction_id ON balance_ledger(transaction_id);
CREATE INDEX IF NOT EXISTS idx_balance_ledger_created_at ON balance_ledger(created_at);

-- =====================================================
-- Payment Limits Table
-- =====================================================
CREATE TABLE IF NOT EXISTS payment_limits (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID,
    verification_level INTEGER NOT NULL DEFAULT 0,
    method VARCHAR(30) NOT NULL,
    min_amount DECIMAL(20, 8) NOT NULL,
    max_amount DECIMAL(20, 8) NOT NULL,
    daily_limit DECIMAL(20, 8) NOT NULL,
    weekly_limit DECIMAL(20, 8) NOT NULL,
    monthly_limit DECIMAL(20, 8) NOT NULL,
    requires_verification BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (user_id, method)
);

CREATE INDEX IF NOT EXISTS idx_payment_limits_user_id ON payment_limits(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_limits_method ON payment_limits(method);

-- =====================================================
-- Fraud Checks Table
-- =====================================================
CREATE TABLE IF NOT EXISTS fraud_checks (
    id BIGSERIAL PRIMARY KEY,
    transaction_id UUID NOT NULL,
    risk_score DECIMAL(5, 2) NOT NULL,
    risk_level VARCHAR(20) NOT NULL CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    blocked BOOLEAN NOT NULL DEFAULT FALSE,
    reason TEXT,
    flags JSONB,
    checks JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (transaction_id) REFERENCES payment_transactions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_fraud_checks_transaction_id ON fraud_checks(transaction_id);
CREATE INDEX IF NOT EXISTS idx_fraud_checks_risk_level ON fraud_checks(risk_level);
CREATE INDEX IF NOT EXISTS idx_fraud_checks_blocked ON fraud_checks(blocked);

-- =====================================================
-- Device Tracking Table
-- =====================================================
CREATE TABLE IF NOT EXISTS device_tracking (
    device_id VARCHAR(128) PRIMARY KEY,
    user_id UUID NOT NULL,
    first_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    transaction_count INTEGER NOT NULL DEFAULT 0,
    failed_count INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    device_fingerprint JSONB
);

CREATE INDEX IF NOT EXISTS idx_device_tracking_user_id ON device_tracking(user_id);
CREATE INDEX IF NOT EXISTS idx_device_tracking_failed_count ON device_tracking(failed_count);

-- =====================================================
-- IP Tracking Table
-- =====================================================
CREATE TABLE IF NOT EXISTS ip_tracking (
    ip_address INET PRIMARY KEY,
    user_id UUID NOT NULL,
    country VARCHAR(2),
    city VARCHAR(100),
    is_tor BOOLEAN NOT NULL DEFAULT FALSE,
    is_proxy BOOLEAN NOT NULL DEFAULT FALSE,
    is_vpn BOOLEAN NOT NULL DEFAULT FALSE,
    is_datacenter BOOLEAN NOT NULL DEFAULT FALSE,
    reputation_score DECIMAL(5, 2),
    first_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    transaction_count INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_ip_tracking_user_id ON ip_tracking(user_id);
CREATE INDEX IF NOT EXISTS idx_ip_tracking_country ON ip_tracking(country);
CREATE INDEX IF NOT EXISTS idx_ip_tracking_reputation_score ON ip_tracking(reputation_score);

-- =====================================================
-- Exchange Rates Table
-- =====================================================
CREATE TABLE IF NOT EXISTS exchange_rates (
    from_currency VARCHAR(10) NOT NULL,
    to_currency VARCHAR(10) NOT NULL,
    rate DECIMAL(20, 8) NOT NULL,
    source VARCHAR(50) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (from_currency, to_currency)
);

CREATE INDEX IF NOT EXISTS idx_exchange_rates_updated_at ON exchange_rates(updated_at);

-- =====================================================
-- Webhook Events Table
-- =====================================================
CREATE TABLE IF NOT EXISTS webhook_events (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(30) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    transaction_id UUID,
    provider_tx_id VARCHAR(128),
    status VARCHAR(20),
    amount DECIMAL(20, 8),
    currency VARCHAR(10),
    payload JSONB NOT NULL,
    signature VARCHAR(256),
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    processed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_webhook_events_provider ON webhook_events(provider);
CREATE INDEX IF NOT EXISTS idx_webhook_events_transaction_id ON webhook_events(transaction_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_provider_tx_id ON webhook_events(provider_tx_id);
CREATE INDEX IF NOT EXISTS idx_webhook_events_processed ON webhook_events(processed);
CREATE INDEX IF NOT EXISTS idx_webhook_events_created_at ON webhook_events(created_at);

-- =====================================================
-- Reconciliation Results Table
-- =====================================================
CREATE TABLE IF NOT EXISTS reconciliation_results (
    id BIGSERIAL PRIMARY KEY,
    transaction_id UUID NOT NULL,
    provider_tx_id VARCHAR(128),
    our_status VARCHAR(20) NOT NULL,
    provider_status VARCHAR(20),
    matched BOOLEAN NOT NULL,
    discrepancy TEXT,
    reconciled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (transaction_id) REFERENCES payment_transactions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reconciliation_results_transaction_id ON reconciliation_results(transaction_id);
CREATE INDEX IF NOT EXISTS idx_reconciliation_results_matched ON reconciliation_results(matched);
CREATE INDEX IF NOT EXISTS idx_reconciliation_results_reconciled_at ON reconciliation_results(reconciled_at);

-- =====================================================
-- Settlement Reports Table
-- =====================================================
CREATE TABLE IF NOT EXISTS settlement_reports (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(30) NOT NULL,
    from_date DATE NOT NULL,
    to_date DATE NOT NULL,
    total_deposits DECIMAL(20, 8) NOT NULL DEFAULT 0,
    total_withdrawals DECIMAL(20, 8) NOT NULL DEFAULT 0,
    total_fees DECIMAL(20, 8) NOT NULL DEFAULT 0,
    net_settlement DECIMAL(20, 8) NOT NULL DEFAULT 0,
    deposit_count INTEGER NOT NULL DEFAULT 0,
    withdrawal_count INTEGER NOT NULL DEFAULT 0,
    report_data JSONB,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (provider, from_date, to_date)
);

CREATE INDEX IF NOT EXISTS idx_settlement_reports_provider ON settlement_reports(provider);
CREATE INDEX IF NOT EXISTS idx_settlement_reports_from_to ON settlement_reports(from_date, to_date);
CREATE INDEX IF NOT EXISTS idx_settlement_reports_generated_at ON settlement_reports(generated_at);

-- =====================================================
-- User Verification Table
-- =====================================================
CREATE TABLE IF NOT EXISTS user_verification (
    user_id UUID PRIMARY KEY,
    verification_level INTEGER NOT NULL DEFAULT 0,
    kyc_status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (kyc_status IN ('pending', 'verified', 'rejected')),
    identity_verified BOOLEAN NOT NULL DEFAULT FALSE,
    address_verified BOOLEAN NOT NULL DEFAULT FALSE,
    phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_user_verification_level ON user_verification(verification_level);
CREATE INDEX IF NOT EXISTS idx_user_verification_kyc_status ON user_verification(kyc_status);

-- =====================================================
-- PostgreSQL Functions (replacing MySQL stored procedures)
-- =====================================================

-- Credit user balance function
CREATE OR REPLACE FUNCTION credit_user_balance(
    p_user_id UUID,
    p_currency VARCHAR(10),
    p_amount DECIMAL(20, 8),
    p_transaction_id UUID
) RETURNS VOID AS $$
DECLARE
    v_balance_before DECIMAL(20, 8);
BEGIN
    -- Get current balance with row lock
    SELECT available_balance INTO v_balance_before
    FROM user_balances
    WHERE user_id = p_user_id AND currency = p_currency
    FOR UPDATE;

    -- Insert or update balance
    INSERT INTO user_balances (user_id, currency, available_balance, last_updated)
    VALUES (p_user_id, p_currency, p_amount, NOW())
    ON CONFLICT (user_id, currency)
    DO UPDATE SET
        available_balance = user_balances.available_balance + p_amount,
        last_updated = NOW();

    -- Insert ledger entry
    INSERT INTO balance_ledger (
        user_id, currency, transaction_id, operation,
        amount, balance_before, balance_after, created_at
    ) VALUES (
        p_user_id, p_currency, p_transaction_id, 'credit',
        p_amount, COALESCE(v_balance_before, 0), COALESCE(v_balance_before, 0) + p_amount, NOW()
    );
END;
$$ LANGUAGE plpgsql;

-- Reserve user balance function
CREATE OR REPLACE FUNCTION reserve_user_balance(
    p_user_id UUID,
    p_currency VARCHAR(10),
    p_amount DECIMAL(20, 8),
    p_transaction_id UUID
) RETURNS VOID AS $$
DECLARE
    v_available DECIMAL(20, 8);
BEGIN
    -- Get available balance with row lock
    SELECT available_balance INTO v_available
    FROM user_balances
    WHERE user_id = p_user_id AND currency = p_currency
    FOR UPDATE;

    -- Check sufficient funds
    IF v_available IS NULL OR v_available < p_amount THEN
        RAISE EXCEPTION 'Insufficient funds';
    END IF;

    -- Move from available to reserved
    UPDATE user_balances
    SET available_balance = available_balance - p_amount,
        reserved_balance = reserved_balance + p_amount,
        last_updated = NOW()
    WHERE user_id = p_user_id AND currency = p_currency;

    -- Insert ledger entry
    INSERT INTO balance_ledger (
        user_id, currency, transaction_id, operation,
        amount, balance_before, balance_after, created_at
    ) VALUES (
        p_user_id, p_currency, p_transaction_id, 'reserve',
        p_amount, v_available, v_available - p_amount, NOW()
    );
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- Insert Default Payment Limits
-- =====================================================
INSERT INTO payment_limits (verification_level, method, min_amount, max_amount, daily_limit, weekly_limit, monthly_limit, requires_verification) VALUES
-- Level 0 (Unverified)
(0, 'card', 10, 500, 1000, 3000, 10000, FALSE),
(0, 'paypal', 10, 300, 500, 1500, 5000, FALSE),
-- Level 1 (Email Verified)
(1, 'card', 10, 2000, 5000, 20000, 50000, FALSE),
(1, 'bank_transfer', 50, 5000, 10000, 40000, 100000, FALSE),
(1, 'paypal', 10, 1000, 2000, 8000, 20000, FALSE),
-- Level 2 (KYC Verified)
(2, 'card', 10, 10000, 50000, 200000, 500000, FALSE),
(2, 'bank_transfer', 50, 50000, 100000, 400000, 1000000, FALSE),
(2, 'wire', 100, 100000, 500000, 2000000, 5000000, FALSE),
(2, 'bitcoin', 100, 50000, 100000, 400000, 1000000, FALSE),
(2, 'ethereum', 100, 50000, 100000, 400000, 1000000, FALSE),
(2, 'usdt', 100, 50000, 100000, 400000, 1000000, FALSE)
ON CONFLICT (user_id, method) DO NOTHING;

-- =====================================================
-- Table Comments
-- =====================================================
COMMENT ON TABLE payment_transactions IS 'Stores all payment transactions (deposits, withdrawals, refunds, chargebacks)';
COMMENT ON TABLE user_balances IS 'User account balances with available and reserved amounts';
COMMENT ON TABLE balance_ledger IS 'Audit trail for all balance changes';
COMMENT ON TABLE payment_limits IS 'Transaction limits by verification level and payment method';
COMMENT ON TABLE fraud_checks IS 'Fraud detection results for each transaction';
COMMENT ON TABLE device_tracking IS 'Device fingerprinting and tracking';
COMMENT ON TABLE ip_tracking IS 'IP address reputation and geo-location tracking';
COMMENT ON TABLE exchange_rates IS 'Currency exchange rates';
COMMENT ON TABLE webhook_events IS 'Payment provider webhook events';
COMMENT ON TABLE reconciliation_results IS 'Payment reconciliation results';
COMMENT ON TABLE settlement_reports IS 'Provider settlement reports';
COMMENT ON TABLE user_verification IS 'User KYC verification status';
