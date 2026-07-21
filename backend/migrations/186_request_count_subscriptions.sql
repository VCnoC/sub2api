ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS subscription_billing_mode VARCHAR(20) NOT NULL DEFAULT 'usd',
    ADD COLUMN IF NOT EXISTS request_limit_5h INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS request_limit_1d INTEGER NOT NULL DEFAULT 0;

ALTER TABLE groups
    DROP CONSTRAINT IF EXISTS groups_subscription_billing_mode_check,
    DROP CONSTRAINT IF EXISTS groups_request_limit_5h_check,
    DROP CONSTRAINT IF EXISTS groups_request_limit_1d_check;

ALTER TABLE groups
    ADD CONSTRAINT groups_subscription_billing_mode_check
        CHECK (subscription_billing_mode IN ('usd', 'request_count')),
    ADD CONSTRAINT groups_request_limit_5h_check CHECK (request_limit_5h >= 0),
    ADD CONSTRAINT groups_request_limit_1d_check CHECK (request_limit_1d >= 0);

ALTER TABLE user_subscriptions
    ADD COLUMN IF NOT EXISTS request_usage_5h INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS request_usage_1d INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS request_window_5h_start TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS request_window_1d_start TIMESTAMPTZ;

ALTER TABLE user_subscriptions
    DROP CONSTRAINT IF EXISTS user_subscriptions_request_usage_5h_check,
    DROP CONSTRAINT IF EXISTS user_subscriptions_request_usage_1d_check;

ALTER TABLE user_subscriptions
    ADD CONSTRAINT user_subscriptions_request_usage_5h_check CHECK (request_usage_5h >= 0),
    ADD CONSTRAINT user_subscriptions_request_usage_1d_check CHECK (request_usage_1d >= 0);

CREATE TABLE IF NOT EXISTS subscription_request_reservations (
    id BIGSERIAL PRIMARY KEY,
    request_id VARCHAR(128) NOT NULL,
    api_key_id BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id BIGINT NOT NULL REFERENCES user_subscriptions(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'committed', 'released')),
    window_5h_start TIMESTAMPTZ,
    window_1d_start TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT subscription_request_reservations_request_subscription_key
        UNIQUE (request_id, subscription_id)
);

CREATE INDEX IF NOT EXISTS idx_subscription_request_reservations_pending_expiry
    ON subscription_request_reservations (subscription_id, expires_at)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_subscription_request_reservations_api_key_id
    ON subscription_request_reservations (api_key_id);

CREATE INDEX IF NOT EXISTS idx_subscription_request_reservations_user_id
    ON subscription_request_reservations (user_id);
