-- 站内工单、不可变消息、私有附件元数据和个人已读游标。
CREATE TABLE IF NOT EXISTS support_tickets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    subject VARCHAR(200) NOT NULL,
    category VARCHAR(32) NOT NULL DEFAULT 'other',
    status VARCHAR(24) NOT NULL DEFAULT 'pending_admin',
    priority VARCHAR(16) NOT NULL DEFAULT 'normal',
    assignee_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    closed_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    closed_at TIMESTAMPTZ,
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT support_tickets_category_check CHECK (category IN ('account', 'billing', 'api', 'model', 'other')),
    CONSTRAINT support_tickets_status_check CHECK (status IN ('pending_admin', 'pending_user', 'closed')),
    CONSTRAINT support_tickets_priority_check CHECK (priority IN ('normal', 'high', 'urgent'))
);

CREATE INDEX IF NOT EXISTS idx_support_tickets_user_created ON support_tickets(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_support_tickets_user_status ON support_tickets(user_id, status);
CREATE INDEX IF NOT EXISTS idx_support_tickets_queue ON support_tickets(status, priority, last_message_at DESC);
CREATE INDEX IF NOT EXISTS idx_support_tickets_assignee_status ON support_tickets(assignee_id, status);
CREATE INDEX IF NOT EXISTS idx_support_tickets_category_status ON support_tickets(category, status);

CREATE TABLE IF NOT EXISTS support_ticket_messages (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    author_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    kind VARCHAR(16) NOT NULL DEFAULT 'public',
    visibility VARCHAR(16) NOT NULL DEFAULT 'user',
    body TEXT NOT NULL DEFAULT '',
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT support_ticket_messages_kind_check CHECK (kind IN ('public', 'internal', 'system')),
    CONSTRAINT support_ticket_messages_visibility_check CHECK (visibility IN ('user', 'admin'))
);

CREATE INDEX IF NOT EXISTS idx_support_ticket_messages_ticket_id ON support_ticket_messages(ticket_id, id);
CREATE INDEX IF NOT EXISTS idx_support_ticket_messages_visibility ON support_ticket_messages(ticket_id, visibility, id);
CREATE INDEX IF NOT EXISTS idx_support_ticket_messages_author ON support_ticket_messages(author_id);

CREATE TABLE IF NOT EXISTS support_ticket_attachments (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL REFERENCES support_ticket_messages(id) ON DELETE CASCADE,
    uploader_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    original_name VARCHAR(255) NOT NULL,
    storage_key VARCHAR(255) NOT NULL UNIQUE,
    media_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,
    delete_after TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    deleted_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    delete_reason VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_support_ticket_attachments_message ON support_ticket_attachments(message_id);
CREATE INDEX IF NOT EXISTS idx_support_ticket_attachments_cleanup ON support_ticket_attachments(delete_after, deleted_at);
CREATE INDEX IF NOT EXISTS idx_support_ticket_attachments_uploader ON support_ticket_attachments(uploader_id);

CREATE TABLE IF NOT EXISTS support_ticket_reads (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_message_id BIGINT NOT NULL DEFAULT 0,
    read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT support_ticket_reads_ticket_user_unique UNIQUE (ticket_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_support_ticket_reads_user_read ON support_ticket_reads(user_id, read_at);
