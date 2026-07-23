-- Add optional pool_key for lottery_chance redeem codes.
-- Only lottery_chance rows should set this to 'normal' or 'luxury'.

ALTER TABLE redeem_codes
    ADD COLUMN IF NOT EXISTS pool_key VARCHAR(16);

ALTER TABLE redeem_codes
    DROP CONSTRAINT IF EXISTS redeem_codes_pool_key_check;

ALTER TABLE redeem_codes
    ADD CONSTRAINT redeem_codes_pool_key_check
    CHECK (pool_key IS NULL OR pool_key IN ('normal', 'luxury'));
