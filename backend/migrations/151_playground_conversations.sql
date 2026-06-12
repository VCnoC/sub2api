-- 对话广场多会话持久化：新建 playground_conversations 表
-- 采用物理删除策略，由后台定时任务按 last_activity_at 清理超期会话
CREATE TABLE IF NOT EXISTS playground_conversations (
    id                BIGSERIAL    PRIMARY KEY,
    user_id           BIGINT       NOT NULL,
    title             VARCHAR(255) NOT NULL DEFAULT '',
    model             TEXT,
    group_name        TEXT,
    messages          JSONB,
    last_activity_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 用户会话列表查询：按活跃度倒序分页（NFR-003 P95 ≤ 200ms）
CREATE INDEX IF NOT EXISTS playground_conversations_user_id_last_activity_at_idx
    ON playground_conversations (user_id, last_activity_at);

-- 过期清理任务全表扫描：按 last_activity_at 范围删除
CREATE INDEX IF NOT EXISTS playground_conversations_last_activity_at_idx
    ON playground_conversations (last_activity_at);
