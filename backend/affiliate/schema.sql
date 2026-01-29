-- Affiliate and Referral Program Database Schema

-- 1. AFFILIATE PROGRAMS
CREATE TABLE affiliate_programs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    commission_model VARCHAR(20) NOT NULL, -- CPA, REVSHARE, HYBRID
    cpa_amount DECIMAL(18, 2) DEFAULT 0,
    revshare_percent DECIMAL(5, 2) DEFAULT 0,
    min_payout DECIMAL(18, 2) DEFAULT 50.00,
    payout_schedule VARCHAR(20) DEFAULT 'MONTHLY', -- MONTHLY, BIWEEKLY, ON_DEMAND
    cookie_duration INT DEFAULT 30, -- Days
    sub_affiliate_enabled BOOLEAN DEFAULT FALSE,
    sub_affiliate_percent DECIMAL(5, 2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'ACTIVE', -- ACTIVE, PAUSED, CLOSED
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_affiliate_programs_status ON affiliate_programs(status);

-- 2. AFFILIATES
CREATE TABLE affiliates (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(user_id),
    affiliate_code VARCHAR(20) UNIQUE NOT NULL,
    company_name VARCHAR(200),
    contact_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(50),
    country VARCHAR(100),
    website VARCHAR(255),
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, ACTIVE, SUSPENDED, BANNED
    tier INT DEFAULT 1, -- 1-5 commission tiers
    parent_affiliate_id BIGINT REFERENCES affiliates(id),
    commission_model VARCHAR(20), -- Override program default
    custom_cpa DECIMAL(18, 2),
    custom_revshare DECIMAL(5, 2),
    payout_method VARCHAR(20), -- BANK, PAYPAL, CRYPTO, WIRE
    bank_details TEXT,
    crypto_address VARCHAR(255),
    tax_id VARCHAR(50),
    total_earnings DECIMAL(18, 2) DEFAULT 0,
    total_paid DECIMAL(18, 2) DEFAULT 0,
    pending_balance DECIMAL(18, 2) DEFAULT 0,
    lifetime_clicks BIGINT DEFAULT 0,
    lifetime_signups BIGINT DEFAULT 0,
    lifetime_deposits BIGINT DEFAULT 0,
    conversion_rate DECIMAL(5, 2) DEFAULT 0,
    fraud_score DECIMAL(5, 2) DEFAULT 0, -- 0-100
    last_activity_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_affiliates_code ON affiliates(affiliate_code);
CREATE INDEX idx_affiliates_email ON affiliates(email);
CREATE INDEX idx_affiliates_status ON affiliates(status);
CREATE INDEX idx_affiliates_parent ON affiliates(parent_affiliate_id);

-- 3. AFFILIATE LINKS
CREATE TABLE affiliate_links (
    id BIGSERIAL PRIMARY KEY,
    affiliate_id BIGINT REFERENCES affiliates(id) ON DELETE CASCADE,
    link_code VARCHAR(20) UNIQUE NOT NULL,
    full_url TEXT NOT NULL,
    landing_page VARCHAR(500),
    campaign VARCHAR(100),
    source VARCHAR(100), -- facebook, google, email
    medium VARCHAR(100), -- cpc, banner, social
    content VARCHAR(100),
    total_clicks BIGINT DEFAULT 0,
    unique_clicks BIGINT DEFAULT 0,
    conversions BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_affiliate_links_code ON affiliate_links(link_code);
CREATE INDEX idx_affiliate_links_affiliate ON affiliate_links(affiliate_id);

-- 4. CLICK TRACKING
CREATE TABLE affiliate_clicks (
    id BIGSERIAL PRIMARY KEY,
    affiliate_id BIGINT REFERENCES affiliates(id) ON DELETE CASCADE,
    link_id BIGINT REFERENCES affiliate_links(id),
    click_id UUID UNIQUE NOT NULL,
    ip_address INET NOT NULL,
    user_agent TEXT,
    country VARCHAR(100),
    city VARCHAR(100),
    device VARCHAR(20), -- DESKTOP, MOBILE, TABLET
    browser VARCHAR(50),
    os VARCHAR(50),
    referrer TEXT,
    landing_page TEXT,
    is_unique BOOLEAN DEFAULT TRUE,
    is_fraudulent BOOLEAN DEFAULT FALSE,
    fraud_reason VARCHAR(100),
    converted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_affiliate_clicks_affiliate ON affiliate_clicks(affiliate_id, created_at DESC);
CREATE INDEX idx_affiliate_clicks_click_id ON affiliate_clicks(click_id);
CREATE INDEX idx_affiliate_clicks_ip ON affiliate_clicks(ip_address, created_at DESC);

-- Create hypertable for time-series optimization (TimescaleDB)
-- SELECT create_hypertable('affiliate_clicks', 'created_at', if_not_exists => TRUE);

-- 5. CONVERSIONS
CREATE TABLE affiliate_conversions (
    id BIGSERIAL PRIMARY KEY,
    affiliate_id BIGINT REFERENCES affiliates(id) ON DELETE CASCADE,
    click_id UUID REFERENCES affiliate_clicks(click_id),
    user_id UUID REFERENCES users(user_id),
    account_id BIGINT REFERENCES rtx_accounts(id),
    conversion_type VARCHAR(20) NOT NULL, -- SIGNUP, DEPOSIT, FIRST_TRADE
    attribution_model VARCHAR(20) DEFAULT 'LAST_CLICK', -- FIRST_CLICK, LAST_CLICK, LINEAR
    value DECIMAL(18, 2) DEFAULT 0, -- Deposit amount
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, APPROVED, REJECTED
    created_at TIMESTAMPTZ DEFAULT NOW(),
    approved_at TIMESTAMPTZ
);

CREATE INDEX idx_affiliate_conversions_affiliate ON affiliate_conversions(affiliate_id, created_at DESC);
CREATE INDEX idx_affiliate_conversions_user ON affiliate_conversions(user_id);
CREATE INDEX idx_affiliate_conversions_click ON affiliate_conversions(click_id);

-- 6. COMMISSIONS
CREATE TABLE affiliate_commissions (
    id BIGSERIAL PRIMARY KEY,
    affiliate_id BIGINT REFERENCES affiliates(id) ON DELETE CASCADE,
    conversion_id BIGINT REFERENCES affiliate_conversions(id),
    account_id BIGINT REFERENCES rtx_accounts(id),
    commission_type VARCHAR(20) NOT NULL, -- CPA, REVSHARE, SUB_AFFILIATE
    amount DECIMAL(18, 2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    description TEXT,
    period VARCHAR(20), -- For RevShare: 2024-01
    trading_volume DECIMAL(18, 2),
    trading_fees DECIMAL(18, 2),
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, APPROVED, PAID, REVERSED
    payout_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    paid_at TIMESTAMPTZ
);

CREATE INDEX idx_affiliate_commissions_affiliate ON affiliate_commissions(affiliate_id, status);
CREATE INDEX idx_affiliate_commissions_status ON affiliate_commissions(status, created_at DESC);

-- 7. PAYOUTS
CREATE TABLE affiliate_payouts (
    id BIGSERIAL PRIMARY KEY,
    affiliate_id BIGINT REFERENCES affiliates(id) ON DELETE CASCADE,
    amount DECIMAL(18, 2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    method VARCHAR(20) NOT NULL, -- BANK, PAYPAL, CRYPTO, WIRE
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, PROCESSING, COMPLETED, FAILED
    transaction_ref VARCHAR(100),
    bank_details TEXT,
    crypto_address VARCHAR(255),
    notes TEXT,
    processed_by VARCHAR(100),
    processed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_affiliate_payouts_affiliate ON affiliate_payouts(affiliate_id, created_at DESC);
CREATE INDEX idx_affiliate_payouts_status ON affiliate_payouts(status);

-- 8. REFERRAL CODES (User-to-User Referrals)
CREATE TABLE referral_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(user_id),
    code VARCHAR(20) UNIQUE NOT NULL,
    referrer_bonus DECIMAL(18, 2) DEFAULT 0,
    referee_bonus DECIMAL(18, 2) DEFAULT 0,
    total_uses BIGINT DEFAULT 0,
    max_uses INT DEFAULT 0, -- 0 = unlimited
    expires_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_referral_codes_code ON referral_codes(code);
CREATE INDEX idx_referral_codes_user ON referral_codes(user_id);

-- 9. REFERRAL REWARDS
CREATE TABLE referral_rewards (
    id BIGSERIAL PRIMARY KEY,
    referral_code_id BIGINT REFERENCES referral_codes(id),
    referrer_user_id UUID REFERENCES users(user_id),
    referee_user_id UUID REFERENCES users(user_id),
    referrer_reward DECIMAL(18, 2) NOT NULL,
    referee_reward DECIMAL(18, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, CREDITED, EXPIRED
    credited_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_referral_rewards_referrer ON referral_rewards(referrer_user_id);
CREATE INDEX idx_referral_rewards_referee ON referral_rewards(referee_user_id);

-- 10. MARKETING MATERIALS
CREATE TABLE marketing_materials (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- BANNER, EMAIL, VIDEO, LANDING_PAGE, SOCIAL
    format VARCHAR(20), -- JPG, PNG, HTML, MP4
    file_url TEXT NOT NULL,
    preview_url TEXT,
    dimensions VARCHAR(50), -- 728x90, 300x250
    language VARCHAR(10) DEFAULT 'en',
    tags TEXT, -- Comma-separated
    downloads BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_marketing_materials_type ON marketing_materials(type, is_active);

-- 11. FRAUD DETECTION RULES
CREATE TABLE fraud_detection_rules (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, -- IP_FRAUD, CLICK_FRAUD, DUPLICATE_ACCOUNT, VELOCITY
    threshold DECIMAL(10, 2),
    action VARCHAR(20) DEFAULT 'FLAG', -- FLAG, BLOCK, SUSPEND
    is_active BOOLEAN DEFAULT TRUE,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 12. FRAUD INCIDENTS
CREATE TABLE fraud_incidents (
    id BIGSERIAL PRIMARY KEY,
    affiliate_id BIGINT REFERENCES affiliates(id),
    rule_id BIGINT REFERENCES fraud_detection_rules(id),
    incident_type VARCHAR(50),
    severity VARCHAR(20), -- LOW, MEDIUM, HIGH, CRITICAL
    description TEXT,
    evidence JSONB,
    ip_address INET,
    user_agent TEXT,
    status VARCHAR(20) DEFAULT 'OPEN', -- OPEN, INVESTIGATING, RESOLVED, FALSE_POSITIVE
    resolved_by VARCHAR(100),
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_fraud_incidents_affiliate ON fraud_incidents(affiliate_id, status);
CREATE INDEX idx_fraud_incidents_status ON fraud_incidents(status, created_at DESC);

-- Insert default affiliate program
INSERT INTO affiliate_programs (name, commission_model, cpa_amount, revshare_percent, min_payout, payout_schedule, cookie_duration, sub_affiliate_enabled, sub_affiliate_percent, status)
VALUES ('Standard Affiliate Program', 'HYBRID', 100.00, 25.00, 100.00, 'MONTHLY', 30, TRUE, 10.00, 'ACTIVE')
ON CONFLICT DO NOTHING;

-- Insert default fraud detection rules
INSERT INTO fraud_detection_rules (name, type, threshold, action, description) VALUES
('IP Click Velocity', 'VELOCITY', 100, 'FLAG', 'Flag IPs with more than 100 clicks per hour'),
('Private IP Detection', 'IP_FRAUD', 0, 'BLOCK', 'Block clicks from private IP ranges'),
('Bot Detection', 'CLICK_FRAUD', 0, 'BLOCK', 'Block known bot user agents'),
('Duplicate Account', 'DUPLICATE_ACCOUNT', 0, 'FLAG', 'Flag multiple accounts from same device')
ON CONFLICT DO NOTHING;

-- Insert sample marketing materials
INSERT INTO marketing_materials (title, description, type, format, file_url, preview_url, dimensions, language) VALUES
('Banner 728x90', 'Leaderboard banner for header placement', 'BANNER', 'PNG', '/materials/banner-728x90.png', '/materials/preview-728x90.png', '728x90', 'en'),
('Email Template - Welcome', 'Welcome email template for new referrals', 'EMAIL', 'HTML', '/materials/email-welcome.html', '/materials/preview-email.png', NULL, 'en'),
('Social Post - Facebook', 'Facebook post template', 'SOCIAL', 'JPG', '/materials/social-facebook.jpg', '/materials/preview-social.jpg', '1200x628', 'en'),
('Landing Page - Trading', 'High-converting trading landing page', 'LANDING_PAGE', 'HTML', '/materials/landing-trading.html', '/materials/preview-landing.png', NULL, 'en')
ON CONFLICT DO NOTHING;
