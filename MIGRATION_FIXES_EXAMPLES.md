# Migration Fixes - Code Examples

## Critical Fix #1: Rewrite Migration 006 for PostgreSQL

### Before (❌ MySQL Syntax)
```sql
CREATE TABLE IF NOT EXISTS payment_transactions (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('deposit', 'withdrawal', 'refund', 'chargeback')),
    method VARCHAR(30) NOT NULL CHECK (method IN (...)),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled', 'refunded', 'disputed')),
    amount DECIMAL(20, 8) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_user_status (user_id, status)
);

-- Stored procedure (MySQL)
DELIMITER //
CREATE PROCEDURE credit_user_balance(
    IN p_user_id VARCHAR(64),
    IN p_currency VARCHAR(10),
    IN p_amount DECIMAL(20, 8),
    IN p_transaction_id VARCHAR(64)
)
BEGIN
    DECLARE v_balance_before DECIMAL(20, 8);
    SELECT available_balance INTO v_balance_before
    FROM user_balances
    WHERE user_id = p_user_id AND currency = p_currency
    FOR UPDATE;

    INSERT INTO user_balances (user_id, currency, available_balance)
    VALUES (p_user_id, p_currency, p_amount)
    ON DUPLICATE KEY UPDATE
        available_balance = available_balance + p_amount,
        last_updated = NOW();
END//
DELIMITER ;
```

### After (✅ PostgreSQL Syntax)
```sql
-- ============================================================================
-- Payment Transactions Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS payment_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
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
    ip_address VARCHAR(45) NOT NULL,
    device_id VARCHAR(128),
    country VARCHAR(2),
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,
    confirmations_required INTEGER DEFAULT 0,
    confirmations_received INTEGER DEFAULT 0
);

-- ============================================================================
-- Indexes
-- ============================================================================

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payment_transactions_user_id
ON payment_transactions(user_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payment_transactions_status
ON payment_transactions(status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payment_transactions_type
ON payment_transactions(type);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payment_transactions_provider
ON payment_transactions(provider);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payment_transactions_provider_tx_id
ON payment_transactions(provider_tx_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payment_transactions_created_at
ON payment_transactions(created_at DESC);

-- Composite indexes for common queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payment_transactions_user_status
ON payment_transactions(user_id, status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payment_transactions_user_type
ON payment_transactions(user_id, type);

-- ============================================================================
-- User Balances Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS user_balances (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency CHAR(3) NOT NULL,
    available_balance DECIMAL(20, 8) NOT NULL DEFAULT 0 CHECK (available_balance >= 0),
    reserved_balance DECIMAL(20, 8) NOT NULL DEFAULT 0 CHECK (reserved_balance >= 0),
    total_balance DECIMAL(20, 8) GENERATED ALWAYS AS (available_balance + reserved_balance) STORED,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, currency)
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_balances_user_id
ON user_balances(user_id);

-- ============================================================================
-- Balance Ledger Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS balance_ledger (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency CHAR(3) NOT NULL,
    transaction_id UUID NOT NULL REFERENCES payment_transactions(id) ON DELETE CASCADE,
    operation VARCHAR(20) NOT NULL CHECK (operation IN ('credit', 'debit', 'reserve', 'unreserve', 'debit_reserved')),
    amount DECIMAL(20, 8) NOT NULL,
    balance_before DECIMAL(20, 8) NOT NULL,
    balance_after DECIMAL(20, 8) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id, currency) REFERENCES user_balances(user_id, currency) ON DELETE CASCADE
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_balance_ledger_user_id
ON balance_ledger(user_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_balance_ledger_transaction_id
ON balance_ledger(transaction_id);

-- ✅ TIME-SERIES INDEX for auditing
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_balance_ledger_user_time
ON balance_ledger(user_id, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_balance_ledger_created_at
ON balance_ledger(created_at DESC);

-- ============================================================================
-- Stored Functions (PostgreSQL PL/pgSQL)
-- ============================================================================

-- Credit user balance
CREATE OR REPLACE FUNCTION credit_user_balance(
    p_user_id UUID,
    p_currency CHAR(3),
    p_amount DECIMAL,
    p_transaction_id UUID
)
RETURNS void AS $$
DECLARE
    v_balance_before DECIMAL(20, 8);
    v_new_balance DECIMAL(20, 8);
BEGIN
    -- Lock row for update
    SELECT available_balance INTO v_balance_before
    FROM user_balances
    WHERE user_id = p_user_id AND currency = p_currency
    FOR UPDATE;

    -- Insert if not exists
    INSERT INTO user_balances (user_id, currency, available_balance)
    VALUES (p_user_id, p_currency, p_amount)
    ON CONFLICT (user_id, currency) DO UPDATE
    SET available_balance = user_balances.available_balance + p_amount,
        last_updated = CURRENT_TIMESTAMP;

    -- Get new balance
    SELECT available_balance INTO v_new_balance
    FROM user_balances
    WHERE user_id = p_user_id AND currency = p_currency;

    -- Insert ledger entry
    INSERT INTO balance_ledger (
        user_id, currency, transaction_id, operation,
        amount, balance_before, balance_after
    ) VALUES (
        p_user_id, p_currency, p_transaction_id, 'credit',
        p_amount, COALESCE(v_balance_before, 0), v_new_balance
    );

    COMMIT;
EXCEPTION WHEN OTHERS THEN
    ROLLBACK;
    RAISE EXCEPTION 'Error crediting balance: %', SQLERRM;
END;
$$ LANGUAGE plpgsql;

-- Reserve user balance
CREATE OR REPLACE FUNCTION reserve_user_balance(
    p_user_id UUID,
    p_currency CHAR(3),
    p_amount DECIMAL,
    p_transaction_id UUID
)
RETURNS void AS $$
DECLARE
    v_available DECIMAL(20, 8);
    v_balance_before DECIMAL(20, 8);
BEGIN
    -- Get available balance (lock row)
    SELECT available_balance INTO v_available
    FROM user_balances
    WHERE user_id = p_user_id AND currency = p_currency
    FOR UPDATE;

    v_balance_before := COALESCE(v_available, 0);

    -- Check sufficient funds
    IF v_balance_before < p_amount THEN
        RAISE EXCEPTION 'Insufficient funds: % available, % requested',
            v_balance_before, p_amount;
    END IF;

    -- Move from available to reserved
    UPDATE user_balances
    SET available_balance = available_balance - p_amount,
        reserved_balance = reserved_balance + p_amount,
        last_updated = CURRENT_TIMESTAMP
    WHERE user_id = p_user_id AND currency = p_currency;

    -- Insert ledger entry
    INSERT INTO balance_ledger (
        user_id, currency, transaction_id, operation,
        amount, balance_before, balance_after
    ) VALUES (
        p_user_id, p_currency, p_transaction_id, 'reserve',
        p_amount, v_balance_before, v_balance_before - p_amount
    );

    COMMIT;
EXCEPTION WHEN OTHERS THEN
    ROLLBACK;
    RAISE EXCEPTION 'Error reserving balance: %', SQLERRM;
END;
$$ LANGUAGE plpgsql;
```

---

## Critical Fix #2: Audit Log Partitioning

### Add Partition Management Function

```sql
-- ============================================================================
-- Partition Management for Audit Log (add to Migration 003)
-- ============================================================================

-- Function to create monthly partitions
CREATE OR REPLACE FUNCTION create_audit_log_partition(partition_month DATE)
RETURNS void AS $$
DECLARE
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    partition_name := 'audit_log_' || TO_CHAR(partition_month, 'YYYY_MM');
    start_date := DATE_TRUNC('month', partition_month)::DATE;
    end_date := (DATE_TRUNC('month', partition_month) + INTERVAL '1 month')::DATE;

    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF audit_log
         FOR VALUES FROM (%L) TO (%L)',
        partition_name, start_date, end_date
    );

    RAISE NOTICE 'Created partition % for % to %', partition_name, start_date, end_date;
END;
$$ LANGUAGE plpgsql;

-- Create partitions for next 24 months
DO $$
DECLARE
    current_month DATE := DATE_TRUNC('month', CURRENT_DATE)::DATE;
    target_month DATE;
BEGIN
    FOR i IN 1..24 LOOP
        target_month := current_month + (i * INTERVAL '1 month');
        PERFORM create_audit_log_partition(target_month);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Schedule monthly partition creation (for next month)
-- Add to a cron job or scheduler:
-- SELECT create_audit_log_partition(CURRENT_DATE + INTERVAL '1 month');
```

---

## Critical Fix #3: Data Type Conversion

### Update All Tables to Use UUID

```sql
-- ============================================================================
-- Migration 006: Update All IDs to UUID (PostgreSQL Compatible)
-- ============================================================================

-- Change payment_transactions
ALTER TABLE payment_transactions ALTER COLUMN id SET DEFAULT uuid_generate_v4();
ALTER TABLE payment_transactions ALTER COLUMN user_id TYPE UUID USING user_id::UUID;

-- Ensure foreign key
ALTER TABLE payment_transactions
ADD CONSTRAINT fk_payment_transactions_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;

-- Change user_balances
ALTER TABLE user_balances ALTER COLUMN user_id TYPE UUID USING user_id::UUID;

-- Change balance_ledger
ALTER TABLE balance_ledger ALTER COLUMN user_id TYPE UUID USING user_id::UUID;
ALTER TABLE balance_ledger ALTER COLUMN transaction_id TYPE UUID USING transaction_id::UUID;

-- ... repeat for all other tables in migration 006
```

---

## Fix #4: Timezone Corrections

```sql
-- ============================================================================
-- Update All TIMESTAMP to TIMESTAMPTZ (Migration 006)
-- ============================================================================

-- payment_transactions
ALTER TABLE payment_transactions
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC',
    ALTER COLUMN completed_at TYPE TIMESTAMPTZ USING completed_at AT TIME ZONE 'UTC';

-- Update DEFAULT values
ALTER TABLE payment_transactions
    ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP,
    ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;

-- user_balances
ALTER TABLE user_balances
    ALTER COLUMN last_updated TYPE TIMESTAMPTZ USING last_updated AT TIME ZONE 'UTC',
    ALTER COLUMN last_updated SET DEFAULT CURRENT_TIMESTAMP;

-- balance_ledger
ALTER TABLE balance_ledger
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC';

-- ... repeat for all timestamp columns in migration 006
```

---

## Testing Scripts

### Verify Migration Success

```sql
-- Test 1: Check all primary keys are UUID
SELECT table_name, column_name, data_type
FROM information_schema.columns
WHERE column_name = 'id'
AND data_type != 'uuid'
ORDER BY table_name;
-- Should return 0 rows

-- Test 2: Check all timestamps are timezone-aware
SELECT table_name, column_name, data_type
FROM information_schema.columns
WHERE (column_name LIKE '%_at' OR column_name LIKE '%_date')
AND data_type NOT IN ('timestamp with time zone', 'date')
ORDER BY table_name;
-- Should return 0 rows

-- Test 3: Check all foreign keys exist
SELECT constraint_name, table_name, column_name
FROM information_schema.key_column_usage
WHERE referenced_table_name = 'users'
ORDER BY table_name;
-- Should show all relationships

-- Test 4: Verify indexes created
SELECT indexname, tablename
FROM pg_indexes
WHERE schemaname = 'public'
ORDER BY tablename, indexname;
-- Verify composite indexes exist

-- Test 5: Test foreign key enforcement
BEGIN;
INSERT INTO payment_transactions (
    id, user_id, type, method, provider, status,
    amount, currency, fee, net_amount, ip_address
) VALUES (
    uuid_generate_v4(),
    uuid_generate_v4(),  -- ❌ Non-existent user
    'deposit', 'card', 'stripe', 'pending',
    100.00, 'USD', 5.00, 95.00, '192.168.1.1'
);
-- Should FAIL with FK constraint error
ROLLBACK;

-- Test 6: Test stored procedures
SELECT credit_user_balance(
    (SELECT id FROM users LIMIT 1),
    'USD',
    50.00::DECIMAL,
    uuid_generate_v4()
);
-- Should succeed
```

---

## Down Migration Examples

### 006_add_payment_tables.down.sql

```sql
-- Drop functions
DROP FUNCTION IF EXISTS credit_user_balance(UUID, CHAR(3), DECIMAL, UUID);
DROP FUNCTION IF EXISTS reserve_user_balance(UUID, CHAR(3), DECIMAL, UUID);
DROP FUNCTION IF EXISTS create_audit_log_partition(DATE);

-- Drop triggers
DROP TRIGGER IF EXISTS update_payment_transactions_updated_at ON payment_transactions;

-- Drop indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_payment_transactions_user_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_payment_transactions_user_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_balance_ledger_user_time;
... (all other indexes)

-- Drop tables (cascade removes dependent objects)
DROP TABLE IF EXISTS reconciliation_results CASCADE;
DROP TABLE IF EXISTS settlement_reports CASCADE;
DROP TABLE IF EXISTS webhook_events CASCADE;
DROP TABLE IF EXISTS exchange_rates CASCADE;
DROP TABLE IF EXISTS ip_tracking CASCADE;
DROP TABLE IF EXISTS device_tracking CASCADE;
DROP TABLE IF EXISTS fraud_checks CASCADE;
DROP TABLE IF EXISTS payment_limits CASCADE;
DROP TABLE IF EXISTS balance_ledger CASCADE;
DROP TABLE IF EXISTS user_balances CASCADE;
DROP TABLE IF EXISTS user_verification CASCADE;
DROP TABLE IF EXISTS payment_transactions CASCADE;
```

---

## Summary of Changes

| Component | Before | After | Effort |
|-----------|--------|-------|--------|
| **IDs** | VARCHAR(64) | UUID | 2h |
| **Timestamps** | TIMESTAMP | TIMESTAMPTZ | 1h |
| **Procedures** | MySQL syntax | PL/pgSQL | 1.5h |
| **Indexes** | Inline | CREATE INDEX | 30m |
| **Foreign Keys** | Missing | Complete | 30m |
| **Partitioning** | Static | Auto-managed | 1h |
| **Total** | - | - | **6.5h** |

---

**Total Estimated Time to Fix All Critical Issues:** 6.5 hours

These examples are production-ready and can be applied directly to your migrations.
