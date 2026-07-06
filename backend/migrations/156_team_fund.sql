-- 团队资金迁移脚本
-- teams 表新增 balance 字段（团队资金池，与 owner 个人余额隔离）

ALTER TABLE teams
    ADD COLUMN IF NOT EXISTS balance DECIMAL(20,8) NOT NULL DEFAULT 0;
