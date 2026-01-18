#!/bin/bash
# Seed test data for trading engine database
# Usage: ./seed-test-data.sh

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Default configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-trading_engine}"

echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║            Trading Engine - Seed Test Data                        ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════╝${NC}"
echo ""

PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME <<'EOSQL'

-- ============================================================================
-- Seed Users
-- ============================================================================

INSERT INTO users (id, email, username, password_hash, first_name, last_name, kyc_status, is_active, is_verified)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'admin@tradingengine.com', 'admin', '$2a$10$YourHashedPasswordHere', 'Admin', 'User', 'approved', true, true),
    ('22222222-2222-2222-2222-222222222222', 'trader1@example.com', 'trader1', '$2a$10$YourHashedPasswordHere', 'John', 'Trader', 'approved', true, true),
    ('33333333-3333-3333-3333-333333333333', 'trader2@example.com', 'trader2', '$2a$10$YourHashedPasswordHere', 'Jane', 'Trader', 'approved', true, true),
    ('44444444-4444-4444-4444-444444444444', 'demo@example.com', 'demo', '$2a$10$YourHashedPasswordHere', 'Demo', 'User', 'pending', true, false)
ON CONFLICT (id) DO NOTHING;

-- User roles
INSERT INTO user_roles (user_id, role)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'super_admin'),
    ('11111111-1111-1111-1111-111111111111', 'risk_manager'),
    ('22222222-2222-2222-2222-222222222222', 'trader'),
    ('33333333-3333-3333-3333-333333333333', 'trader'),
    ('44444444-4444-4444-4444-444444444444', 'user')
ON CONFLICT (user_id, role) DO NOTHING;

-- ============================================================================
-- Seed Instruments
-- ============================================================================

INSERT INTO instruments (symbol, display_name, instrument_type, base_currency, quote_currency, tick_size, min_quantity, max_quantity, contract_size, leverage_max, is_tradable, is_active)
VALUES
    -- Forex
    ('EURUSD', 'EUR/USD', 'forex', 'EUR', 'USD', 0.00001, 0.01, 100.0, 100000, 500, true, true),
    ('GBPUSD', 'GBP/USD', 'forex', 'GBP', 'USD', 0.00001, 0.01, 100.0, 100000, 500, true, true),
    ('USDJPY', 'USD/JPY', 'forex', 'USD', 'JPY', 0.001, 0.01, 100.0, 100000, 500, true, true),
    ('AUDUSD', 'AUD/USD', 'forex', 'AUD', 'USD', 0.00001, 0.01, 100.0, 100000, 500, true, true),

    -- Crypto
    ('BTCUSD', 'Bitcoin/USD', 'crypto', 'BTC', 'USD', 0.01, 0.0001, 10.0, 1, 100, true, true),
    ('ETHUSD', 'Ethereum/USD', 'crypto', 'ETH', 'USD', 0.01, 0.001, 100.0, 1, 100, true, true),
    ('BNBUSD', 'Binance Coin/USD', 'crypto', 'BNB', 'USD', 0.01, 0.01, 1000.0, 1, 100, true, true),
    ('XRPUSD', 'Ripple/USD', 'crypto', 'XRP', 'USD', 0.0001, 1.0, 100000.0, 1, 100, true, true),
    ('SOLUSD', 'Solana/USD', 'crypto', 'SOL', 'USD', 0.01, 0.1, 1000.0, 1, 100, true, true),

    -- Commodities
    ('XAUUSD', 'Gold/USD', 'commodity', 'XAU', 'USD', 0.01, 0.01, 100.0, 100, 200, true, true),
    ('XAGUSD', 'Silver/USD', 'commodity', 'XAG', 'USD', 0.001, 0.1, 1000.0, 5000, 200, true, true),
    ('WTICOUSD', 'WTI Crude Oil', 'commodity', 'WTI', 'USD', 0.01, 0.01, 100.0, 1000, 100, true, true),

    -- Indices
    ('SPX500USD', 'S&P 500', 'index', 'SPX', 'USD', 0.01, 0.01, 100.0, 50, 100, true, true),
    ('NAS100USD', 'Nasdaq 100', 'index', 'NAS', 'USD', 0.01, 0.01, 100.0, 20, 100, true, true),
    ('US30USD', 'Dow Jones', 'index', 'DJI', 'USD', 0.01, 0.01, 100.0, 5, 100, true, true)
ON CONFLICT (symbol) DO NOTHING;

-- ============================================================================
-- Seed Accounts
-- ============================================================================

INSERT INTO accounts (id, user_id, account_number, account_type, currency, balance, leverage, is_active)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '22222222-2222-2222-2222-222222222222', 'ACC-000001', 'live', 'USD', 10000.00, 100, true),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '33333333-3333-3333-3333-333333333333', 'ACC-000002', 'live', 'USD', 25000.00, 200, true),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', '44444444-4444-4444-4444-444444444444', 'ACC-DEMO-001', 'demo', 'USD', 100000.00, 500, true),
    ('dddddddd-dddd-dddd-dddd-dddddddddddd', '22222222-2222-2222-2222-222222222222', 'ACC-000003', 'margin', 'USD', 50000.00, 50, true)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- Seed Liquidity Providers
-- ============================================================================

INSERT INTO liquidity_providers (id, name, provider_type, protocol, is_active, priority, connection_status)
VALUES
    ('e1e1e1e1-e1e1-e1e1-e1e1-e1e1e1e1e1e1', 'Binance', 'exchange', 'REST', true, 90, 'connected'),
    ('e2e2e2e2-e2e2-e2e2-e2e2-e2e2e2e2e2e2', 'OANDA', 'market_maker', 'REST', true, 85, 'connected'),
    ('e3e3e3e3-e3e3-e3e3-e3e3-e3e3e3e3e3e3', 'Prime Broker A', 'prime_broker', 'FIX', true, 100, 'connected'),
    ('e4e4e4e4-e4e4-e4e4-e4e4-e4e4e4e4e4e4', 'Aggregator XYZ', 'aggregator', 'WEBSOCKET', true, 80, 'connected')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- Seed SOR Rules
-- ============================================================================

INSERT INTO sor_rules (rule_name, rule_type, priority, is_active, conditions, actions)
VALUES
    ('High Value Crypto to A-Book', 'abook', 100, true,
     '{"instrument_type": "crypto", "order_value_min": 10000}',
     '{"route_to": "abook", "lp_priority": ["Binance", "Prime Broker A"]}'),

    ('Small Forex to B-Book', 'bbook', 90, true,
     '{"instrument_type": "forex", "order_value_max": 1000}',
     '{"route_to": "bbook"}'),

    ('Demo Accounts to B-Book', 'bbook', 95, true,
     '{"account_type": "demo"}',
     '{"route_to": "bbook"}'),

    ('VIP Clients to A-Book', 'abook', 100, true,
     '{"profile_type": "vip"}',
     '{"route_to": "abook", "priority_execution": true}')
ON CONFLICT (rule_name) DO NOTHING;

-- ============================================================================
-- Seed Risk Limits
-- ============================================================================

INSERT INTO risk_limits (limit_type, entity_type, entity_id, limit_value, warning_threshold, is_active)
VALUES
    ('daily_loss', 'global', NULL, 100000.00, 80000.00, true),
    ('max_drawdown', 'global', NULL, 50000.00, 40000.00, true),
    ('position_size', 'global', NULL, 1000000.00, 800000.00, true),
    ('account_leverage', 'global', NULL, 500.00, 400.00, true),
    ('open_positions', 'global', NULL, 1000.00, 800.00, true)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Seed Circuit Breakers
-- ============================================================================

INSERT INTO circuit_breakers (breaker_name, breaker_type, scope, trigger_condition, threshold_value, status, action)
VALUES
    ('Global Volatility Breaker', 'volatility', 'global', '{"volatility_spike": 0.1}', 0.1, 'armed', 'halt_trading'),
    ('Daily Loss Limit', 'loss_limit', 'global', '{"daily_loss": 100000}', 100000.00, 'armed', 'reject_orders'),
    ('Price Movement Crypto', 'price_movement', 'symbol', '{"price_change_pct": 0.15}', 0.15, 'armed', 'reduce_positions')
ON CONFLICT (breaker_name) DO NOTHING;

EOSQL

echo -e "${GREEN}✓ Test data seeded successfully!${NC}"
echo ""
echo -e "${YELLOW}Seeded data summary:${NC}"
echo "  • 4 Users (admin, trader1, trader2, demo)"
echo "  • 15 Instruments (Forex, Crypto, Commodities, Indices)"
echo "  • 4 Accounts (2 live, 1 demo, 1 margin)"
echo "  • 4 Liquidity Providers"
echo "  • 4 Smart Order Routing rules"
echo "  • 5 Global risk limits"
echo "  • 3 Circuit breakers"
echo ""
echo -e "${YELLOW}Test credentials:${NC}"
echo "  Admin: admin@tradingengine.com / (set password)"
echo "  Trader 1: trader1@example.com / (set password)"
echo "  Trader 2: trader2@example.com / (set password)"
echo "  Demo: demo@example.com / (set password)"
echo ""
