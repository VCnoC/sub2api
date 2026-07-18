-- Link each new subscription order to the exact entitlement it issued.
ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS subscription_id BIGINT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'payment_orders_subscription_id_fkey'
    ) THEN
        ALTER TABLE payment_orders
            ADD CONSTRAINT payment_orders_subscription_id_fkey
            FOREIGN KEY (subscription_id)
            REFERENCES user_subscriptions(id)
            ON DELETE SET NULL
            NOT VALID;
    END IF;
END $$;

ALTER TABLE payment_orders
    VALIDATE CONSTRAINT payment_orders_subscription_id_fkey;
