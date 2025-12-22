-- BrainSait AI Platform - Database Schema
-- PostgreSQL 15+

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- Users & Authentication
-- ============================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    api_key VARCHAR(64) UNIQUE NOT NULL DEFAULT encode(gen_random_bytes(32), 'hex'),
    tier VARCHAR(50) DEFAULT 'free' CHECK (tier IN ('free', 'pro', 'enterprise')),
    organization VARCHAR(255),
    domain VARCHAR(50) DEFAULT 'general' CHECK (domain IN ('general', 'arabic', 'healthcare', 'developer')),
    language VARCHAR(10) DEFAULT 'en',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for API key lookups (critical for performance)
CREATE INDEX idx_users_api_key ON users(api_key) WHERE is_active = true;
CREATE INDEX idx_users_email ON users(email);

-- ============================================
-- Usage Tracking
-- ============================================
CREATE TABLE usage_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    endpoint VARCHAR(100) NOT NULL,
    method VARCHAR(10) DEFAULT 'POST',
    model_used VARCHAR(100),
    domain VARCHAR(50),
    tokens_input INT DEFAULT 0,
    tokens_output INT DEFAULT 0,
    credits_used INT DEFAULT 1,
    latency_ms INT,
    status_code INT,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Partitioning by month for better performance
CREATE INDEX idx_usage_user_time ON usage_logs(user_id, created_at DESC);
CREATE INDEX idx_usage_domain ON usage_logs(domain, created_at DESC);

-- ============================================
-- Billing & Subscriptions
-- ============================================
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    stripe_customer_id VARCHAR(100),
    stripe_subscription_id VARCHAR(100),
    plan VARCHAR(50) NOT NULL DEFAULT 'free',
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'canceled', 'past_due', 'trialing')),
    monthly_credits INT DEFAULT 100,
    credits_used INT DEFAULT 0,
    current_period_start DATE,
    current_period_end DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_stripe ON subscriptions(stripe_customer_id);

-- ============================================
-- Invoices
-- ============================================
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    stripe_invoice_id VARCHAR(100),
    amount_cents INT NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(50) DEFAULT 'pending',
    period_start DATE,
    period_end DATE,
    paid_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- Prompts Library
-- ============================================
CREATE TABLE prompts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    domain VARCHAR(50) DEFAULT 'general',
    language VARCHAR(10) DEFAULT 'en',
    content JSONB NOT NULL, -- Store .prompt.yml as JSON
    is_public BOOLEAN DEFAULT false,
    is_featured BOOLEAN DEFAULT false,
    usage_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_prompts_domain ON prompts(domain) WHERE is_public = true;
CREATE INDEX idx_prompts_language ON prompts(language);

-- ============================================
-- API Rate Limiting
-- ============================================
CREATE TABLE rate_limits (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    endpoint VARCHAR(100),
    requests_count INT DEFAULT 0,
    window_start TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    window_size_seconds INT DEFAULT 60,
    max_requests INT DEFAULT 60
);

CREATE UNIQUE INDEX idx_rate_limits_user_endpoint ON rate_limits(user_id, endpoint);

-- ============================================
-- Audit Log (HIPAA Compliance)
-- ============================================
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_user ON audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_action ON audit_logs(action, created_at DESC);

-- ============================================
-- Functions & Triggers
-- ============================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_prompts_updated_at
    BEFORE UPDATE ON prompts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Views
-- ============================================

-- User usage summary
CREATE VIEW user_usage_summary AS
SELECT 
    u.id,
    u.email,
    u.tier,
    u.domain,
    s.monthly_credits,
    s.credits_used,
    (s.monthly_credits - s.credits_used) as credits_remaining,
    COUNT(ul.id) as total_requests,
    SUM(ul.tokens_input + ul.tokens_output) as total_tokens
FROM users u
LEFT JOIN subscriptions s ON u.id = s.user_id
LEFT JOIN usage_logs ul ON u.id = ul.user_id 
    AND ul.created_at >= s.current_period_start
GROUP BY u.id, u.email, u.tier, u.domain, s.monthly_credits, s.credits_used;

-- Daily usage analytics
CREATE VIEW daily_usage_analytics AS
SELECT 
    DATE(created_at) as date,
    domain,
    model_used,
    COUNT(*) as request_count,
    SUM(tokens_input) as total_input_tokens,
    SUM(tokens_output) as total_output_tokens,
    AVG(latency_ms) as avg_latency_ms,
    COUNT(CASE WHEN status_code >= 400 THEN 1 END) as error_count
FROM usage_logs
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE(created_at), domain, model_used
ORDER BY date DESC;

-- ============================================
-- Seed Data: Default Tiers
-- ============================================
INSERT INTO users (email, name, tier, domain, api_key) VALUES
    ('system@brainsait.ai', 'System', 'enterprise', 'general', 'system_internal_key_do_not_use'),
    ('demo@brainsait.ai', 'Demo User', 'pro', 'developer', 'demo_api_key_for_testing_only');

-- Add default subscription for demo user
INSERT INTO subscriptions (user_id, plan, monthly_credits, current_period_start, current_period_end)
SELECT id, 'pro', 10000, DATE_TRUNC('month', NOW()), DATE_TRUNC('month', NOW()) + INTERVAL '1 month'
FROM users WHERE email = 'demo@brainsait.ai';
