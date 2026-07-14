-- 双奖池抽奖、邀请奖励规则、次数账户和不可变审计流水。
CREATE TABLE IF NOT EXISTS lottery_pools (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(16) NOT NULL UNIQUE,
    name VARCHAR(80) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    cycle_type VARCHAR(16) NOT NULL DEFAULT 'daily',
    cycle_chances INTEGER NOT NULL DEFAULT 1,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT lottery_pools_key_check CHECK (key IN ('normal', 'luxury')),
    CONSTRAINT lottery_pools_cycle_type_check CHECK (cycle_type IN ('daily', 'weekly')),
    CONSTRAINT lottery_pools_cycle_chances_check CHECK (cycle_chances >= 0 AND cycle_chances <= 100),
    CONSTRAINT lottery_pools_time_check CHECK (starts_at IS NULL OR ends_at IS NULL OR ends_at > starts_at)
);

INSERT INTO lottery_pools (key, name, enabled, cycle_type, cycle_chances)
VALUES
    ('normal', '普通抽奖', FALSE, 'daily', 1),
    ('luxury', '豪华抽奖', FALSE, 'weekly', 1)
ON CONFLICT (key) DO NOTHING;

CREATE TABLE IF NOT EXISTS lottery_prizes (
    id BIGSERIAL PRIMARY KEY,
    pool_id BIGINT NOT NULL REFERENCES lottery_pools(id) ON DELETE RESTRICT,
    name VARCHAR(120) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    image_data TEXT NOT NULL DEFAULT '',
    prize_type VARCHAR(16) NOT NULL,
    balance_amount NUMERIC(20, 8),
    group_id BIGINT REFERENCES groups(id) ON DELETE RESTRICT,
    validity_days INTEGER,
    probability_ppm INTEGER NOT NULL DEFAULT 0,
    stock_total BIGINT,
    stock_used BIGINT NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT lottery_prizes_type_check CHECK (prize_type IN ('balance', 'subscription')),
    CONSTRAINT lottery_prizes_probability_check CHECK (probability_ppm >= 0 AND probability_ppm <= 1000000),
    CONSTRAINT lottery_prizes_stock_check CHECK (stock_total IS NULL OR (stock_total >= 0 AND stock_used >= 0 AND stock_used <= stock_total)),
    CONSTRAINT lottery_prizes_value_check CHECK (
        (prize_type = 'balance' AND balance_amount > 0 AND group_id IS NULL AND validity_days IS NULL)
        OR
        (prize_type = 'subscription' AND balance_amount IS NULL AND group_id IS NOT NULL AND validity_days > 0)
    )
);

CREATE INDEX IF NOT EXISTS idx_lottery_prizes_pool_active ON lottery_prizes(pool_id, enabled, sort_order, id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS lottery_rules (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    event_type VARCHAR(16) NOT NULL,
    beneficiary VARCHAR(16) NOT NULL DEFAULT 'inviter',
    normal_chances INTEGER NOT NULL DEFAULT 0,
    luxury_chances INTEGER NOT NULL DEFAULT 0,
    recharge_mode VARCHAR(16),
    recharge_threshold NUMERIC(20, 8),
    repeatable BOOLEAN NOT NULL DEFAULT FALSE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT lottery_rules_event_check CHECK (event_type IN ('signup', 'redeem', 'recharge')),
    CONSTRAINT lottery_rules_beneficiary_check CHECK (beneficiary IN ('inviter', 'invitee')),
    CONSTRAINT lottery_rules_chances_check CHECK (
        normal_chances BETWEEN 0 AND 100000
        AND luxury_chances BETWEEN 0 AND 100000
        AND normal_chances + luxury_chances > 0
    ),
    CONSTRAINT lottery_rules_event_fields_check CHECK (
        (event_type = 'recharge' AND beneficiary = 'inviter' AND recharge_mode IN ('single', 'cumulative') AND recharge_threshold > 0)
        OR
        (event_type <> 'recharge' AND recharge_mode IS NULL AND recharge_threshold IS NULL AND repeatable = FALSE)
    ),
    CONSTRAINT lottery_rules_redeem_beneficiary_check CHECK (event_type <> 'redeem' OR beneficiary = 'inviter')
);

CREATE INDEX IF NOT EXISTS idx_lottery_rules_event_active ON lottery_rules(event_type, enabled, id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS lottery_user_chances (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pool_id BIGINT NOT NULL REFERENCES lottery_pools(id) ON DELETE CASCADE,
    period_key VARCHAR(32) NOT NULL DEFAULT '',
    base_remaining INTEGER NOT NULL DEFAULT 0,
    extra_remaining BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, pool_id),
    CONSTRAINT lottery_user_chances_nonnegative_check CHECK (base_remaining >= 0 AND extra_remaining >= 0)
);

CREATE TABLE IF NOT EXISTS lottery_chance_ledger (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    pool_id BIGINT NOT NULL REFERENCES lottery_pools(id) ON DELETE RESTRICT,
    action VARCHAR(24) NOT NULL,
    base_delta INTEGER NOT NULL DEFAULT 0,
    extra_delta BIGINT NOT NULL DEFAULT 0,
    rule_id BIGINT REFERENCES lottery_rules(id) ON DELETE SET NULL,
    source_type VARCHAR(32) NOT NULL,
    source_id VARCHAR(128) NOT NULL,
    source_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    tier_no INTEGER NOT NULL DEFAULT 0,
    dedupe_key VARCHAR(255) NOT NULL UNIQUE,
    balance_after JSONB NOT NULL DEFAULT '{}'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT lottery_chance_ledger_action_check CHECK (action IN ('grant', 'refund_reversal', 'draw'))
);

CREATE INDEX IF NOT EXISTS idx_lottery_chance_ledger_user_time ON lottery_chance_ledger(user_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_lottery_chance_ledger_rule_source ON lottery_chance_ledger(rule_id, source_type, source_id, tier_no);

CREATE TABLE IF NOT EXISTS lottery_draws (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    pool_id BIGINT NOT NULL REFERENCES lottery_pools(id) ON DELETE RESTRICT,
    idempotency_key VARCHAR(128) NOT NULL,
    outcome VARCHAR(16) NOT NULL,
    chance_source VARCHAR(16) NOT NULL,
    prize_id BIGINT REFERENCES lottery_prizes(id) ON DELETE SET NULL,
    redeem_code_id BIGINT REFERENCES redeem_codes(id) ON DELETE SET NULL,
    random_roll INTEGER NOT NULL,
    prize_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT lottery_draws_outcome_check CHECK (outcome IN ('win', 'none')),
    CONSTRAINT lottery_draws_chance_source_check CHECK (chance_source IN ('base', 'extra')),
    CONSTRAINT lottery_draws_roll_check CHECK (random_roll >= 0 AND random_roll < 1000000),
    CONSTRAINT lottery_draws_user_pool_key_unique UNIQUE (user_id, pool_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_lottery_draws_user_time ON lottery_draws(user_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_lottery_draws_pool_time ON lottery_draws(pool_id, created_at DESC, id DESC);
