-- Add independent video billing mode and durable terminal-state tracking.

ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS video_billing_mode VARCHAR(20) NOT NULL DEFAULT 'per_second';

COMMENT ON COLUMN groups.video_billing_mode IS '视频生成计费模式：per_second 或 per_request';

ALTER TABLE user_platform_quotas
    DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check;

ALTER TABLE user_platform_quotas
    ADD CONSTRAINT user_platform_quotas_platform_check
    CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'video'));

CREATE TABLE IF NOT EXISTS video_tasks (
    id BIGSERIAL PRIMARY KEY,
    upstream_task_id VARCHAR(255) NOT NULL,
    billing_request_id VARCHAR(128) NOT NULL,
    user_id BIGINT NOT NULL,
    api_key_id BIGINT NOT NULL,
    account_id BIGINT NOT NULL,
    group_id BIGINT NOT NULL,
    refund_amount DECIMAL(20,10) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    next_poll_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_until TIMESTAMPTZ,
    poll_attempts INTEGER NOT NULL DEFAULT 0,
    terminal_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT video_tasks_status_check CHECK (status IN ('pending', 'completed', 'failed')),
    CONSTRAINT video_tasks_refund_amount_check CHECK (refund_amount >= 0),
    CONSTRAINT video_tasks_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT video_tasks_api_key_fk FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE RESTRICT,
    CONSTRAINT video_tasks_account_fk FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE RESTRICT,
    CONSTRAINT video_tasks_group_fk FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_video_tasks_upstream_account
    ON video_tasks (upstream_task_id, account_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_video_tasks_billing_request
    ON video_tasks (billing_request_id, api_key_id);
CREATE INDEX IF NOT EXISTS idx_video_tasks_due
    ON video_tasks (status, next_poll_at)
    WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_video_tasks_locked_until
    ON video_tasks (locked_until)
    WHERE locked_until IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_video_tasks_user_id
    ON video_tasks (user_id);
