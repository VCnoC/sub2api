// Package repository 为抽奖模块提供 PostgreSQL 持久化、行锁和幂等流水。
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type lotteryRepository struct {
	client *dbent.Client
}

func NewLotteryRepository(client *dbent.Client) service.LotteryRepository {
	return &lotteryRepository{client: client}
}

func (r *lotteryRepository) ListPools(ctx context.Context) ([]service.LotteryPool, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
SELECT id, key, name, enabled, cycle_type, cycle_chances, starts_at, ends_at, created_at, updated_at
FROM lottery_pools ORDER BY CASE key WHEN 'normal' THEN 0 ELSE 1 END`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	pools := make([]service.LotteryPool, 0, 2)
	for rows.Next() {
		pool, scanErr := scanLotteryPool(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		pools = append(pools, *pool)
	}
	return pools, rows.Err()
}

func (r *lotteryRepository) GetPoolByKey(ctx context.Context, key string) (*service.LotteryPool, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
SELECT id, key, name, enabled, cycle_type, cycle_chances, starts_at, ends_at, created_at, updated_at
FROM lottery_pools WHERE key = $1 LIMIT 1`, key)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, service.ErrLotteryPoolNotFound
	}
	return scanLotteryPool(rows)
}

func (r *lotteryRepository) UpdatePool(ctx context.Context, key string, input service.LotteryPoolUpdate) (*service.LotteryPool, error) {
	result, err := clientFromContext(ctx, r.client).ExecContext(ctx, `
UPDATE lottery_pools
SET name = $2, enabled = $3, cycle_type = $4, cycle_chances = $5,
    starts_at = $6, ends_at = $7, updated_at = NOW()
WHERE key = $1`, key, strings.TrimSpace(input.Name), input.Enabled, input.CycleType, input.CycleChances, input.StartsAt, input.EndsAt)
	if err != nil {
		return nil, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return nil, service.ErrLotteryPoolNotFound
	}
	return r.GetPoolByKey(ctx, key)
}

func scanLotteryPool(scanner interface{ Scan(...any) error }) (*service.LotteryPool, error) {
	var pool service.LotteryPool
	if err := scanner.Scan(&pool.ID, &pool.Key, &pool.Name, &pool.Enabled, &pool.CycleType, &pool.CycleChances, &pool.StartsAt, &pool.EndsAt, &pool.CreatedAt, &pool.UpdatedAt); err != nil {
		return nil, err
	}
	return &pool, nil
}

func (r *lotteryRepository) ListPrizes(ctx context.Context, poolID int64, includeDisabled bool) ([]service.LotteryPrize, error) {
	query := `
SELECT id, pool_id, name, description, image_data, prize_type,
       balance_amount::double precision, group_id, validity_days, probability_ppm,
       stock_total, stock_used, enabled, sort_order, created_at, updated_at
FROM lottery_prizes WHERE pool_id = $1 AND deleted_at IS NULL`
	if !includeDisabled {
		query += ` AND enabled = TRUE`
	}
	query += ` ORDER BY sort_order, id`
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, query, poolID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.LotteryPrize, 0)
	for rows.Next() {
		item, scanErr := scanLotteryPrize(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *lotteryRepository) GetPrize(ctx context.Context, id int64, forUpdate bool) (*service.LotteryPrize, error) {
	query := `
SELECT id, pool_id, name, description, image_data, prize_type,
       balance_amount::double precision, group_id, validity_days, probability_ppm,
       stock_total, stock_used, enabled, sort_order, created_at, updated_at
FROM lottery_prizes WHERE id = $1 AND deleted_at IS NULL`
	if forUpdate {
		query += ` FOR UPDATE`
	}
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, service.ErrLotteryPrizeNotFound
	}
	return scanLotteryPrize(rows)
}

func scanLotteryPrize(scanner interface{ Scan(...any) error }) (*service.LotteryPrize, error) {
	var item service.LotteryPrize
	var balance sql.NullFloat64
	var groupID sql.NullInt64
	var validity sql.NullInt64
	var stockTotal sql.NullInt64
	if err := scanner.Scan(
		&item.ID, &item.PoolID, &item.Name, &item.Description, &item.ImageData, &item.PrizeType,
		&balance, &groupID, &validity, &item.ProbabilityPPM, &stockTotal, &item.StockUsed,
		&item.Enabled, &item.SortOrder, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if balance.Valid {
		value := balance.Float64
		item.BalanceAmount = &value
	}
	if groupID.Valid {
		value := groupID.Int64
		item.GroupID = &value
	}
	if validity.Valid {
		value := int(validity.Int64)
		item.ValidityDays = &value
	}
	if stockTotal.Valid {
		value := stockTotal.Int64
		item.StockTotal = &value
	}
	return &item, nil
}

func (r *lotteryRepository) CreatePrize(ctx context.Context, input service.LotteryPrizeInput) (*service.LotteryPrize, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
INSERT INTO lottery_prizes (
    pool_id, name, description, image_data, prize_type, balance_amount, group_id,
    validity_days, probability_ppm, stock_total, enabled, sort_order
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
RETURNING id`, input.PoolID, strings.TrimSpace(input.Name), strings.TrimSpace(input.Description), input.ImageData,
		input.PrizeType, input.BalanceAmount, input.GroupID, input.ValidityDays, input.ProbabilityPPM,
		input.StockTotal, input.Enabled, input.SortOrder)
	if err != nil {
		return nil, err
	}
	var id int64
	if !rows.Next() {
		rowsErr := rows.Err()
		_ = rows.Close()
		if rowsErr != nil {
			return nil, rowsErr
		}
		return nil, errors.New("create lottery prize returned no id")
	}
	if err := rows.Scan(&id); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("scan created lottery prize id: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return r.GetPrize(ctx, id, false)
}

func (r *lotteryRepository) UpdatePrize(ctx context.Context, id int64, input service.LotteryPrizeInput) (*service.LotteryPrize, error) {
	result, err := clientFromContext(ctx, r.client).ExecContext(ctx, `
UPDATE lottery_prizes SET
    pool_id=$2, name=$3, description=$4, image_data=$5, prize_type=$6,
    balance_amount=$7, group_id=$8, validity_days=$9, probability_ppm=$10,
    stock_total=$11, enabled=$12, sort_order=$13, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL AND ($11::bigint IS NULL OR stock_used <= $11)`,
		id, input.PoolID, strings.TrimSpace(input.Name), strings.TrimSpace(input.Description), input.ImageData,
		input.PrizeType, input.BalanceAmount, input.GroupID, input.ValidityDays, input.ProbabilityPPM,
		input.StockTotal, input.Enabled, input.SortOrder)
	if err != nil {
		return nil, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return nil, service.ErrLotteryPrizeNotFound
	}
	return r.GetPrize(ctx, id, false)
}

func (r *lotteryRepository) DisablePrize(ctx context.Context, id int64) error {
	result, err := clientFromContext(ctx, r.client).ExecContext(ctx, `
UPDATE lottery_prizes SET enabled=FALSE, deleted_at=NOW(), updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return service.ErrLotteryPrizeNotFound
	}
	return nil
}

func (r *lotteryRepository) EnabledProbabilityTotal(ctx context.Context, poolID, excludePrizeID int64) (int, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
SELECT COALESCE(SUM(probability_ppm), 0)::bigint
FROM lottery_prizes
WHERE pool_id=$1 AND enabled=TRUE AND deleted_at IS NULL AND id<>$2`, poolID, excludePrizeID)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	var total int64
	if rows.Next() {
		err = rows.Scan(&total)
	}
	return int(total), err
}

func (r *lotteryRepository) ClaimPrizeStock(ctx context.Context, prizeID int64) (bool, error) {
	result, err := clientFromContext(ctx, r.client).ExecContext(ctx, `
UPDATE lottery_prizes SET stock_used=stock_used+1, updated_at=NOW()
WHERE id=$1 AND enabled=TRUE AND deleted_at IS NULL
  AND (stock_total IS NULL OR stock_used < stock_total)`, prizeID)
	if err != nil {
		return false, err
	}
	affected, _ := result.RowsAffected()
	return affected == 1, nil
}

func (r *lotteryRepository) CreditBalance(ctx context.Context, userID int64, amount float64) error {
	result, err := clientFromContext(ctx, r.client).ExecContext(ctx, `
UPDATE users SET balance=balance+$2, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, userID, amount)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return service.ErrUserNotFound
	}
	return nil
}

func (r *lotteryRepository) ListRules(ctx context.Context, eventType string, includeDisabled bool) ([]service.LotteryRule, error) {
	query := `
SELECT id, name, event_type, beneficiary, normal_chances, luxury_chances,
       recharge_mode, recharge_threshold::double precision, repeatable, enabled, created_at, updated_at
FROM lottery_rules WHERE deleted_at IS NULL`
	args := make([]any, 0, 1)
	if eventType != "" {
		query += ` AND event_type=$1`
		args = append(args, eventType)
	}
	if !includeDisabled {
		query += ` AND enabled=TRUE`
	}
	query += ` ORDER BY id`
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.LotteryRule, 0)
	for rows.Next() {
		item, scanErr := scanLotteryRule(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *lotteryRepository) GetRule(ctx context.Context, id int64) (*service.LotteryRule, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
SELECT id, name, event_type, beneficiary, normal_chances, luxury_chances,
       recharge_mode, recharge_threshold::double precision, repeatable, enabled, created_at, updated_at
FROM lottery_rules WHERE id=$1 AND deleted_at IS NULL LIMIT 1`, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, service.ErrLotteryRuleNotFound
	}
	return scanLotteryRule(rows)
}

func scanLotteryRule(scanner interface{ Scan(...any) error }) (*service.LotteryRule, error) {
	var item service.LotteryRule
	var mode sql.NullString
	var threshold sql.NullFloat64
	if err := scanner.Scan(&item.ID, &item.Name, &item.EventType, &item.Beneficiary, &item.NormalChances,
		&item.LuxuryChances, &mode, &threshold, &item.Repeatable, &item.Enabled, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	if mode.Valid {
		value := mode.String
		item.RechargeMode = &value
	}
	if threshold.Valid {
		value := threshold.Float64
		item.RechargeThreshold = &value
	}
	return &item, nil
}

func (r *lotteryRepository) CreateRule(ctx context.Context, input service.LotteryRuleInput) (*service.LotteryRule, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
INSERT INTO lottery_rules (
    name, event_type, beneficiary, normal_chances, luxury_chances,
    recharge_mode, recharge_threshold, repeatable, enabled
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, strings.TrimSpace(input.Name), input.EventType,
		input.Beneficiary, input.NormalChances, input.LuxuryChances, input.RechargeMode,
		input.RechargeThreshold, input.Repeatable, input.Enabled)
	if err != nil {
		return nil, err
	}
	var id int64
	if !rows.Next() {
		rowsErr := rows.Err()
		_ = rows.Close()
		if rowsErr != nil {
			return nil, rowsErr
		}
		return nil, errors.New("create lottery rule returned no id")
	}
	if err := rows.Scan(&id); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("scan created lottery rule id: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return r.GetRule(ctx, id)
}

func (r *lotteryRepository) UpdateRule(ctx context.Context, id int64, input service.LotteryRuleInput) (*service.LotteryRule, error) {
	result, err := clientFromContext(ctx, r.client).ExecContext(ctx, `
UPDATE lottery_rules SET name=$2, event_type=$3, beneficiary=$4, normal_chances=$5,
    luxury_chances=$6, recharge_mode=$7, recharge_threshold=$8, repeatable=$9,
    enabled=$10, updated_at=NOW()
WHERE id=$1 AND deleted_at IS NULL`, id, strings.TrimSpace(input.Name), input.EventType,
		input.Beneficiary, input.NormalChances, input.LuxuryChances, input.RechargeMode,
		input.RechargeThreshold, input.Repeatable, input.Enabled)
	if err != nil {
		return nil, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return nil, service.ErrLotteryRuleNotFound
	}
	return r.GetRule(ctx, id)
}

func (r *lotteryRepository) DisableRule(ctx context.Context, id int64) error {
	result, err := clientFromContext(ctx, r.client).ExecContext(ctx, `
	UPDATE lottery_rules SET enabled=FALSE, updated_at=NOW()
	WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return service.ErrLotteryRuleNotFound
	}
	return nil
}

func (r *lotteryRepository) LockChanceAccount(ctx context.Context, userID int64, pool service.LotteryPool, periodKey string) (*service.LotteryChanceAccount, error) {
	client := clientFromContext(ctx, r.client)
	if _, err := client.ExecContext(ctx, `
INSERT INTO lottery_user_chances (user_id, pool_id, period_key, base_remaining, extra_remaining)
VALUES ($1,$2,$3,$4,0) ON CONFLICT (user_id,pool_id) DO NOTHING`, userID, pool.ID, periodKey, pool.CycleChances); err != nil {
		return nil, err
	}
	rows, err := client.QueryContext(ctx, `
SELECT user_id, pool_id, period_key, base_remaining, extra_remaining, updated_at
FROM lottery_user_chances WHERE user_id=$1 AND pool_id=$2 FOR UPDATE`, userID, pool.ID)
	if err != nil {
		return nil, err
	}
	var account service.LotteryChanceAccount
	if !rows.Next() {
		rowsErr := rows.Err()
		_ = rows.Close()
		if rowsErr != nil {
			return nil, rowsErr
		}
		return nil, errors.New("lottery chance account missing after upsert")
	}
	if err := rows.Scan(&account.UserID, &account.PoolID, &account.PeriodKey, &account.BaseRemaining, &account.ExtraRemaining, &account.UpdatedAt); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if account.PeriodKey != periodKey {
		if _, err := client.ExecContext(ctx, `
UPDATE lottery_user_chances SET period_key=$3, base_remaining=$4, updated_at=NOW()
WHERE user_id=$1 AND pool_id=$2`, userID, pool.ID, periodKey, pool.CycleChances); err != nil {
			return nil, err
		}
		account.PeriodKey = periodKey
		account.BaseRemaining = pool.CycleChances
		account.UpdatedAt = time.Now()
	}
	return &account, nil
}

func (r *lotteryRepository) GrantExtraChance(ctx context.Context, input service.LotteryChanceGrant) (bool, error) {
	if input.Chances <= 0 {
		return false, nil
	}
	var applied bool
	err := r.withTx(ctx, func(txCtx context.Context) error {
		pool, err := r.GetPoolByKey(txCtx, input.PoolKey)
		if err != nil {
			return err
		}
		client := clientFromContext(txCtx, r.client)
		if _, err := client.ExecContext(txCtx, `
INSERT INTO lottery_user_chances (user_id,pool_id,period_key,base_remaining,extra_remaining)
VALUES ($1,$2,'',0,0) ON CONFLICT (user_id,pool_id) DO NOTHING`, input.UserID, pool.ID); err != nil {
			return err
		}
		rows, err := client.QueryContext(txCtx, `
SELECT extra_remaining FROM lottery_user_chances WHERE user_id=$1 AND pool_id=$2 FOR UPDATE`, input.UserID, pool.ID)
		if err != nil {
			return err
		}
		var before int64
		if !rows.Next() {
			_ = rows.Close()
			if err := rows.Err(); err != nil {
				return err
			}
			return errors.New("lottery chance account unavailable")
		}
		if err := rows.Scan(&before); err != nil {
			_ = rows.Close()
			return err
		}
		_ = rows.Close()
		exists, err := ledgerDedupeExists(txCtx, client, input.DedupeKey)
		if err != nil || exists {
			return err
		}
		after := before + input.Chances
		if _, err := client.ExecContext(txCtx, `
UPDATE lottery_user_chances SET extra_remaining=$3, updated_at=NOW() WHERE user_id=$1 AND pool_id=$2`, input.UserID, pool.ID, after); err != nil {
			return err
		}
		balance, _ := json.Marshal(map[string]any{"extra_remaining": after})
		metadata, _ := json.Marshal(input.Metadata)
		var sourceUser any
		if input.SourceUserID > 0 {
			sourceUser = input.SourceUserID
		}
		if _, err := client.ExecContext(txCtx, `
INSERT INTO lottery_chance_ledger (
    user_id,pool_id,action,extra_delta,rule_id,source_type,source_id,
    source_user_id,tier_no,dedupe_key,balance_after,metadata
) VALUES ($1,$2,'grant',$3,$4,$5,$6,$7,$8,$9,$10,$11)`, input.UserID, pool.ID,
			input.Chances, input.RuleID, input.SourceType, input.SourceID, sourceUser, input.TierNo,
			input.DedupeKey, balance, metadata); err != nil {
			return err
		}
		applied = true
		return nil
	})
	return applied, err
}

func ledgerDedupeExists(ctx context.Context, client *dbent.Client, key string) (bool, error) {
	rows, err := client.QueryContext(ctx, `SELECT 1 FROM lottery_chance_ledger WHERE dedupe_key=$1 LIMIT 1`, key)
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	return rows.Next(), nil
}

func (r *lotteryRepository) ReverseExtraChance(ctx context.Context, grantDedupeKey, reversalDedupeKey string, metadata map[string]any) (int64, bool, error) {
	var recovered int64
	var applied bool
	err := r.withTx(ctx, func(txCtx context.Context) error {
		client := clientFromContext(txCtx, r.client)
		if exists, err := ledgerDedupeExists(txCtx, client, reversalDedupeKey); err != nil || exists {
			return err
		}
		rows, err := client.QueryContext(txCtx, `
SELECT user_id,pool_id,extra_delta,rule_id,source_user_id,tier_no
FROM lottery_chance_ledger WHERE dedupe_key=$1 AND action='grant' LIMIT 1`, grantDedupeKey)
		if err != nil {
			return err
		}
		var userID, poolID, amount, ruleID int64
		var sourceUser sql.NullInt64
		var tier int
		if !rows.Next() {
			_ = rows.Close()
			return nil
		}
		if err := rows.Scan(&userID, &poolID, &amount, &ruleID, &sourceUser, &tier); err != nil {
			_ = rows.Close()
			return err
		}
		_ = rows.Close()
		accountRows, err := client.QueryContext(txCtx, `
SELECT extra_remaining FROM lottery_user_chances WHERE user_id=$1 AND pool_id=$2 FOR UPDATE`, userID, poolID)
		if err != nil {
			return err
		}
		var available int64
		if !accountRows.Next() {
			_ = accountRows.Close()
			if err := accountRows.Err(); err != nil {
				return err
			}
			return errors.New("lottery chance account unavailable")
		}
		if err := accountRows.Scan(&available); err != nil {
			_ = accountRows.Close()
			return err
		}
		_ = accountRows.Close()
		recovered = amount
		if recovered > available {
			recovered = available
		}
		after := available - recovered
		if _, err := client.ExecContext(txCtx, `
UPDATE lottery_user_chances SET extra_remaining=$3, updated_at=NOW() WHERE user_id=$1 AND pool_id=$2`, userID, poolID, after); err != nil {
			return err
		}
		if metadata == nil {
			metadata = make(map[string]any)
		}
		metadata["grant_dedupe_key"] = grantDedupeKey
		metadata["unrecovered"] = amount - recovered
		metaJSON, _ := json.Marshal(metadata)
		balanceJSON, _ := json.Marshal(map[string]any{"extra_remaining": after})
		var sourceUserValue any
		if sourceUser.Valid {
			sourceUserValue = sourceUser.Int64
		}
		if _, err := client.ExecContext(txCtx, `
INSERT INTO lottery_chance_ledger (
    user_id,pool_id,action,extra_delta,rule_id,source_type,source_id,
    source_user_id,tier_no,dedupe_key,balance_after,metadata
) VALUES ($1,$2,'refund_reversal',$3,$4,'refund',$5,$6,$7,$8,$9,$10)`, userID, poolID,
			-recovered, ruleID, grantDedupeKey, sourceUserValue, tier, reversalDedupeKey, balanceJSON, metaJSON); err != nil {
			return err
		}
		applied = true
		return nil
	})
	return recovered, applied, err
}

func (r *lotteryRepository) ConsumeChance(ctx context.Context, account service.LotteryChanceAccount, poolID int64, drawDedupeKey string) (string, *service.LotteryChanceAccount, error) {
	client := clientFromContext(ctx, r.client)
	source := ""
	baseDelta := 0
	extraDelta := int64(0)
	if account.BaseRemaining > 0 {
		source = "base"
		account.BaseRemaining--
		baseDelta = -1
	} else if account.ExtraRemaining > 0 {
		source = "extra"
		account.ExtraRemaining--
		extraDelta = -1
	} else {
		return "", nil, service.ErrLotteryNoChance
	}
	if _, err := client.ExecContext(ctx, `
UPDATE lottery_user_chances SET base_remaining=$3, extra_remaining=$4, updated_at=NOW()
WHERE user_id=$1 AND pool_id=$2`, account.UserID, poolID, account.BaseRemaining, account.ExtraRemaining); err != nil {
		return "", nil, err
	}
	balance, _ := json.Marshal(map[string]any{"base_remaining": account.BaseRemaining, "extra_remaining": account.ExtraRemaining})
	ledgerDedupeKey := fmt.Sprintf("draw:%d:%d:%s", account.UserID, poolID, drawDedupeKey)
	if _, err := client.ExecContext(ctx, `
	INSERT INTO lottery_chance_ledger (
	    user_id,pool_id,action,base_delta,extra_delta,source_type,source_id,dedupe_key,balance_after
	) VALUES ($1,$2,'draw',$3,$4,'draw',$5,$6,$7)`, account.UserID, poolID, baseDelta, extraDelta,
		drawDedupeKey, ledgerDedupeKey, balance); err != nil {
		return "", nil, err
	}
	return source, &account, nil
}

func (r *lotteryRepository) CreateDraw(ctx context.Context, userID int64, poolID int64, idempotencyKey string, outcome, chanceSource string, prizeID, redeemCodeID *int64, randomRoll int, snapshot map[string]any) (*service.LotteryDraw, error) {
	snapshotJSON, _ := json.Marshal(snapshot)
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
INSERT INTO lottery_draws (
    user_id,pool_id,idempotency_key,outcome,chance_source,prize_id,redeem_code_id,random_roll,prize_snapshot
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (user_id,pool_id,idempotency_key) DO NOTHING
RETURNING id,created_at`, userID, poolID, idempotencyKey, outcome, chanceSource, prizeID, redeemCodeID, randomRoll, snapshotJSON)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var draw service.LotteryDraw
	if !rows.Next() {
		return nil, service.ErrLotteryAlreadyExists
	}
	if err := rows.Scan(&draw.ID, &draw.CreatedAt); err != nil {
		return nil, err
	}
	draw.UserID, draw.PoolID, draw.Outcome, draw.ChanceSource = userID, poolID, outcome, chanceSource
	draw.PrizeID, draw.RedeemCodeID, draw.RandomRoll, draw.PrizeSnapshot = prizeID, redeemCodeID, randomRoll, snapshot
	return &draw, nil
}

func (r *lotteryRepository) GetDrawByIdempotencyKey(ctx context.Context, userID, poolID int64, key string) (*service.LotteryDraw, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
SELECT d.id,d.user_id,d.pool_id,p.key,d.outcome,d.chance_source,d.prize_id,d.redeem_code_id,
       d.random_roll,d.prize_snapshot,d.created_at
FROM lottery_draws d JOIN lottery_pools p ON p.id=d.pool_id
WHERE d.user_id=$1 AND d.pool_id=$2 AND d.idempotency_key=$3 LIMIT 1`, userID, poolID, key)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, service.ErrLotteryDrawNotFound
	}
	return scanLotteryDraw(rows)
}

func scanLotteryDraw(scanner interface{ Scan(...any) error }) (*service.LotteryDraw, error) {
	var item service.LotteryDraw
	var prizeID, codeID sql.NullInt64
	var snapshot []byte
	if err := scanner.Scan(&item.ID, &item.UserID, &item.PoolID, &item.PoolKey, &item.Outcome, &item.ChanceSource,
		&prizeID, &codeID, &item.RandomRoll, &snapshot, &item.CreatedAt); err != nil {
		return nil, err
	}
	if prizeID.Valid {
		value := prizeID.Int64
		item.PrizeID = &value
	}
	if codeID.Valid {
		value := codeID.Int64
		item.RedeemCodeID = &value
	}
	_ = json.Unmarshal(snapshot, &item.PrizeSnapshot)
	return &item, nil
}

func (r *lotteryRepository) ListUserDraws(ctx context.Context, userID int64, params pagination.PaginationParams, poolKey string) ([]service.LotteryDraw, *pagination.PaginationResult, error) {
	return r.listDraws(ctx, params, &userID, poolKey, "")
}

func (r *lotteryRepository) ListDraws(ctx context.Context, params pagination.PaginationParams, userID *int64, poolKey, outcome string) ([]service.LotteryDraw, *pagination.PaginationResult, error) {
	return r.listDraws(ctx, params, userID, poolKey, outcome)
}

func (r *lotteryRepository) listDraws(ctx context.Context, params pagination.PaginationParams, userID *int64, poolKey, outcome string) ([]service.LotteryDraw, *pagination.PaginationResult, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 3)
	if userID != nil {
		args = append(args, *userID)
		where = append(where, fmt.Sprintf("d.user_id=$%d", len(args)))
	}
	if poolKey != "" {
		args = append(args, poolKey)
		where = append(where, fmt.Sprintf("p.key=$%d", len(args)))
	}
	if outcome != "" {
		args = append(args, outcome)
		where = append(where, fmt.Sprintf("d.outcome=$%d", len(args)))
	}
	clause := strings.Join(where, " AND ")
	client := clientFromContext(ctx, r.client)
	countRows, err := client.QueryContext(ctx, `SELECT COUNT(*) FROM lottery_draws d JOIN lottery_pools p ON p.id=d.pool_id WHERE `+clause, args...)
	if err != nil {
		return nil, nil, err
	}
	var total int64
	if countRows.Next() {
		err = countRows.Scan(&total)
	}
	_ = countRows.Close()
	if err != nil {
		return nil, nil, err
	}
	args = append(args, params.Limit(), params.Offset())
	query := fmt.Sprintf(`
SELECT d.id,d.user_id,d.pool_id,p.key,d.outcome,d.chance_source,d.prize_id,d.redeem_code_id,
       d.random_roll,d.prize_snapshot,d.created_at
FROM lottery_draws d JOIN lottery_pools p ON p.id=d.pool_id
WHERE %s ORDER BY d.created_at DESC,d.id DESC LIMIT $%d OFFSET $%d`, clause, len(args)-1, len(args))
	rows, err := client.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.LotteryDraw, 0)
	for rows.Next() {
		item, scanErr := scanLotteryDraw(rows)
		if scanErr != nil {
			return nil, nil, scanErr
		}
		items = append(items, *item)
	}
	return items, paginationResultFromTotal(total, params), rows.Err()
}

func (r *lotteryRepository) ListChanceLedger(ctx context.Context, params pagination.PaginationParams, userID *int64, poolKey, action string) ([]service.LotteryChanceLedgerEntry, *pagination.PaginationResult, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 3)
	if userID != nil {
		args = append(args, *userID)
		where = append(where, fmt.Sprintf("l.user_id=$%d", len(args)))
	}
	if poolKey != "" {
		args = append(args, poolKey)
		where = append(where, fmt.Sprintf("p.key=$%d", len(args)))
	}
	if action != "" {
		args = append(args, action)
		where = append(where, fmt.Sprintf("l.action=$%d", len(args)))
	}
	clause := strings.Join(where, " AND ")
	client := clientFromContext(ctx, r.client)
	countRows, err := client.QueryContext(ctx, `SELECT COUNT(*) FROM lottery_chance_ledger l JOIN lottery_pools p ON p.id=l.pool_id WHERE `+clause, args...)
	if err != nil {
		return nil, nil, err
	}
	var total int64
	if countRows.Next() {
		err = countRows.Scan(&total)
	}
	_ = countRows.Close()
	if err != nil {
		return nil, nil, err
	}
	args = append(args, params.Limit(), params.Offset())
	query := fmt.Sprintf(`
SELECT l.id,l.user_id,l.pool_id,p.key,l.action,l.base_delta,l.extra_delta,l.rule_id,
       l.source_type,l.source_id,l.source_user_id,l.tier_no,l.metadata,l.created_at
FROM lottery_chance_ledger l JOIN lottery_pools p ON p.id=l.pool_id
WHERE %s ORDER BY l.created_at DESC,l.id DESC LIMIT $%d OFFSET $%d`, clause, len(args)-1, len(args))
	rows, err := client.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.LotteryChanceLedgerEntry, 0)
	for rows.Next() {
		var item service.LotteryChanceLedgerEntry
		var ruleID, sourceUser sql.NullInt64
		var metadata []byte
		if err := rows.Scan(&item.ID, &item.UserID, &item.PoolID, &item.PoolKey, &item.Action, &item.BaseDelta,
			&item.ExtraDelta, &ruleID, &item.SourceType, &item.SourceID, &sourceUser, &item.TierNo, &metadata, &item.CreatedAt); err != nil {
			return nil, nil, err
		}
		if ruleID.Valid {
			value := ruleID.Int64
			item.RuleID = &value
		}
		if sourceUser.Valid {
			value := sourceUser.Int64
			item.SourceUserID = &value
		}
		_ = json.Unmarshal(metadata, &item.Metadata)
		items = append(items, item)
	}
	return items, paginationResultFromTotal(total, params), rows.Err()
}

func (r *lotteryRepository) GetInviterID(ctx context.Context, userID int64) (*int64, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `SELECT inviter_id FROM user_affiliates WHERE user_id=$1 LIMIT 1`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var inviter sql.NullInt64
	if !rows.Next() {
		return nil, nil
	}
	if err := rows.Scan(&inviter); err != nil {
		return nil, err
	}
	if !inviter.Valid {
		return nil, nil
	}
	value := inviter.Int64
	return &value, nil
}

func (r *lotteryRepository) HasPriorRedeem(ctx context.Context, userID, excludingCodeID int64) (bool, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
	SELECT EXISTS (
	    SELECT 1 FROM redeem_codes
	    WHERE used_by=$1 AND id<>$2 AND status='used'
	      AND COALESCE(notes, '') NOT LIKE $3
	)`, userID, excludingCodeID, service.LotterySystemRedeemNotePrefix+"%")
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return false, rows.Err()
	}
	var exists bool
	if err := rows.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *lotteryRepository) NetCompletedRecharge(ctx context.Context, userID int64, excludingOrderID int64) (float64, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
SELECT COALESCE(SUM(GREATEST(amount - COALESCE(refund_amount,0),0)),0)::double precision
FROM payment_orders
WHERE user_id=$1 AND id<>$2 AND status IN ('COMPLETED','PARTIALLY_REFUNDED')`, userID, excludingOrderID)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	var total float64
	if rows.Next() {
		err = rows.Scan(&total)
	}
	return total, err
}

func (r *lotteryRepository) GrantMatchesSource(ctx context.Context, dedupeKey, sourceType, sourceID string) (bool, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
	SELECT EXISTS (
	    SELECT 1 FROM lottery_chance_ledger g
	    WHERE g.dedupe_key=$1 AND g.action='grant' AND g.source_type=$2 AND g.source_id=$3
	      AND NOT EXISTS (
	          SELECT 1 FROM lottery_chance_ledger r WHERE r.dedupe_key='reverse:' || g.dedupe_key
	      )
	)`, dedupeKey, sourceType, sourceID)
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return false, rows.Err()
	}
	var exists bool
	if err := rows.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *lotteryRepository) ListActiveCumulativeGrantKeys(ctx context.Context, userID, ruleID, sourceUserID int64, aboveTier int) ([]string, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
SELECT g.dedupe_key
FROM lottery_chance_ledger g
WHERE g.user_id=$1 AND g.rule_id=$2 AND g.source_user_id=$3
  AND g.action='grant' AND g.source_type='recharge_cumulative' AND g.tier_no>$4
  AND NOT EXISTS (
      SELECT 1 FROM lottery_chance_ledger r WHERE r.dedupe_key='reverse:' || g.dedupe_key
  )
ORDER BY g.tier_no DESC`, userID, ruleID, sourceUserID, aboveTier)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	keys := make([]string, 0)
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (r *lotteryRepository) RuleHasLedger(ctx context.Context, ruleID int64) (bool, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `
	SELECT EXISTS (SELECT 1 FROM lottery_chance_ledger WHERE rule_id=$1)`, ruleID)
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return false, rows.Err()
	}
	var exists bool
	if err := rows.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *lotteryRepository) withTx(ctx context.Context, fn func(context.Context) error) error {
	if dbent.TxFromContext(ctx) != nil {
		return fn(ctx)
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	return tx.Commit()
}
