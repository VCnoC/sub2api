CREATE TABLE IF NOT EXISTS api_key_groups (
    id BIGSERIAL PRIMARY KEY,
    api_key_id BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups(id),
    priority SMALLINT NOT NULL CHECK (priority >= 0 AND priority < 5),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT api_key_groups_api_key_group_key UNIQUE (api_key_id, group_id),
    CONSTRAINT api_key_groups_api_key_priority_key UNIQUE (api_key_id, priority)
);

CREATE INDEX IF NOT EXISTS idx_api_key_groups_group_id
    ON api_key_groups (group_id);

INSERT INTO api_key_groups (api_key_id, group_id, priority)
SELECT id, group_id, 0
FROM api_keys
WHERE group_id IS NOT NULL AND deleted_at IS NULL
ON CONFLICT (api_key_id, group_id) DO NOTHING;
