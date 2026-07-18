-- Multiple independent entitlements may coexist for one user and group.
DROP INDEX CONCURRENTLY IF EXISTS user_subscriptions_user_group_unique_active;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_subscriptions_candidate_order
    ON user_subscriptions (user_id, group_id, status, expires_at, id)
    WHERE deleted_at IS NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payment_orders_subscription_id
    ON payment_orders (subscription_id)
    WHERE subscription_id IS NOT NULL;
