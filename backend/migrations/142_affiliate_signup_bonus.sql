-- 邀请返利：注册奖励（双向）增强
-- 1) aff_inviter_bonus_override: 用户作为邀请人时的专属注册奖励金额（NULL 表示沿用全局默认）
--    例：KOL 邀请人享受更高的注册奖励（$20），普通用户用全局默认（$5）

ALTER TABLE user_affiliates
    ADD COLUMN IF NOT EXISTS aff_inviter_bonus_override DECIMAL(20, 10);

COMMENT ON COLUMN user_affiliates.aff_inviter_bonus_override IS '邀请人专属注册奖励金额（USD，NULL 表示沿用全局 affiliate_inviter_bonus_usd）';
