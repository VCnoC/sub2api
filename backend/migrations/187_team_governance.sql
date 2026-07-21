-- 团队治理、审批与可转赠余额来源控制。

ALTER TABLE teams
    ADD COLUMN IF NOT EXISTS member_limit INTEGER,
    ADD COLUMN IF NOT EXISTS level INTEGER,
    ADD COLUMN IF NOT EXISTS review_required BOOLEAN;

UPDATE teams t
SET member_limit = COALESCE(member_limit, GREATEST(5, (SELECT COUNT(*)::INTEGER FROM users u WHERE u.team_id = t.id AND u.deleted_at IS NULL))),
    level = COALESCE(level, CASE
        WHEN (SELECT COUNT(*) FROM users u WHERE u.team_id = t.id AND u.deleted_at IS NULL) > 15 THEN 40
        WHEN (SELECT COUNT(*) FROM users u WHERE u.team_id = t.id AND u.deleted_at IS NULL) > 5 THEN 15
        ELSE 5
    END),
    review_required = COALESCE(review_required, TRUE)
WHERE member_limit IS NULL OR level IS NULL OR review_required IS NULL;

ALTER TABLE teams
    ALTER COLUMN member_limit SET DEFAULT 5,
    ALTER COLUMN member_limit SET NOT NULL,
    ALTER COLUMN level SET DEFAULT 5,
    ALTER COLUMN level SET NOT NULL,
    ALTER COLUMN review_required SET DEFAULT FALSE,
    ALTER COLUMN review_required SET NOT NULL;

CREATE TABLE IF NOT EXISTS team_governance_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    configured BOOLEAN NOT NULL DEFAULT FALSE,
    min_registration_days INTEGER NOT NULL DEFAULT 0 CHECK (min_registration_days >= 0),
    min_total_recharge DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (min_total_recharge >= 0),
    level_5_recharge DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (level_5_recharge >= 0),
    level_5_spend_7d DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (level_5_spend_7d >= 0),
    level_5_mode VARCHAR(3) NOT NULL DEFAULT 'and' CHECK (level_5_mode IN ('and', 'or')),
    level_15_recharge DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (level_15_recharge >= 0),
    level_15_spend_7d DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (level_15_spend_7d >= 0),
    level_15_mode VARCHAR(3) NOT NULL DEFAULT 'and' CHECK (level_15_mode IN ('and', 'or')),
    level_40_recharge DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (level_40_recharge >= 0),
    level_40_spend_7d DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (level_40_spend_7d >= 0),
    level_40_mode VARCHAR(3) NOT NULL DEFAULT 'and' CHECK (level_40_mode IN ('and', 'or')),
    updated_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO team_governance_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS team_applications (
    id BIGSERIAL PRIMARY KEY,
    application_type VARCHAR(10) NOT NULL CHECK (application_type IN ('create', 'expand')),
    applicant_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id BIGINT REFERENCES teams(id) ON DELETE CASCADE,
    team_name VARCHAR(100) NOT NULL DEFAULT '',
    target_limit INTEGER CHECK (target_limit IS NULL OR target_limit > 0),
    reason TEXT NOT NULL DEFAULT '',
    additional_info TEXT NOT NULL DEFAULT '',
    status VARCHAR(10) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    review_reason TEXT NOT NULL DEFAULT '',
    waived BOOLEAN NOT NULL DEFAULT FALSE,
    reviewer_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_team_id BIGINT REFERENCES teams(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK ((application_type = 'create' AND team_id IS NULL AND team_name <> '') OR
           (application_type = 'expand' AND team_id IS NOT NULL AND target_limit IS NOT NULL))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_team_applications_pending_create
    ON team_applications(applicant_id) WHERE application_type = 'create' AND status = 'pending';
CREATE UNIQUE INDEX IF NOT EXISTS idx_team_applications_pending_expand
    ON team_applications(team_id) WHERE application_type = 'expand' AND status = 'pending';
CREATE INDEX IF NOT EXISTS idx_team_applications_status_created
    ON team_applications(status, created_at DESC);

CREATE TABLE IF NOT EXISTS team_join_requests (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    applicant_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL DEFAULT '',
    status VARCHAR(10) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    review_reason TEXT NOT NULL DEFAULT '',
    reviewed_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_team_join_requests_pending_user
    ON team_join_requests(applicant_id) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_team_join_requests_team_status
    ON team_join_requests(team_id, status, created_at DESC);

CREATE TABLE IF NOT EXISTS team_transferable_balances (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO team_transferable_balances (user_id, amount)
SELECT id, GREATEST(balance, 0) FROM users
ON CONFLICT (user_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS team_fund_ledger (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    counterparty_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(20) NOT NULL CHECK (action IN ('deposit', 'allocate', 'transfer')),
    amount DECIMAL(20,8) NOT NULL CHECK (amount > 0),
    transferable BOOLEAN NOT NULL DEFAULT FALSE,
    operator_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_team_fund_ledger_team_created
    ON team_fund_ledger(team_id, created_at DESC);

CREATE OR REPLACE FUNCTION sync_team_transferable_on_balance_decrease()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.balance < OLD.balance THEN
        INSERT INTO team_transferable_balances (user_id, amount, updated_at)
        VALUES (NEW.id, 0, NOW())
        ON CONFLICT (user_id) DO UPDATE
        SET amount = GREATEST(0, team_transferable_balances.amount - (OLD.balance - NEW.balance)),
            updated_at = NOW();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_users_team_transferable_decrease ON users;
CREATE TRIGGER trg_users_team_transferable_decrease
AFTER UPDATE OF balance ON users
FOR EACH ROW EXECUTE FUNCTION sync_team_transferable_on_balance_decrease();

CREATE OR REPLACE FUNCTION sync_team_transferable_on_redeem()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'used'
       AND NEW.used_by IS NOT NULL
       AND NEW.value > 0
       AND NEW.type IN ('balance', 'admin_balance')
       AND COALESCE(NEW.notes, '') NOT LIKE '[lottery] %'
       AND (TG_OP = 'INSERT' OR OLD.status IS DISTINCT FROM 'used') THEN
        INSERT INTO team_transferable_balances (user_id, amount, updated_at)
        VALUES (NEW.used_by, NEW.value, NOW())
        ON CONFLICT (user_id) DO UPDATE
        SET amount = team_transferable_balances.amount + EXCLUDED.amount,
            updated_at = NOW();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_redeem_team_transferable ON redeem_codes;
CREATE TRIGGER trg_redeem_team_transferable
AFTER INSERT OR UPDATE OF status ON redeem_codes
FOR EACH ROW EXECUTE FUNCTION sync_team_transferable_on_redeem();
