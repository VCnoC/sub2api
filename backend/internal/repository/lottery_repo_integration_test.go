//go:build integration

// Package repository 的抽奖集成测试验证 PostgreSQL 行锁、幂等和库存约束。
package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestLotteryRepositoryChanceIdempotencyAndStock(t *testing.T) {
	ctx := context.Background()
	repo := NewLotteryRepository(integrationEntClient)
	userA := createLotteryTestUser(t)
	userB := createLotteryTestUser(t)

	pools, err := repo.ListPools(ctx)
	require.NoError(t, err)
	require.Len(t, pools, 2)
	normal := pools[0]

	rule, err := repo.CreateRule(ctx, service.LotteryRuleInput{
		Name: "integration signup", EventType: service.LotteryEventSignup,
		Beneficiary: service.LotteryBeneficiaryInviter, NormalChances: 2, Enabled: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(ctx, `DELETE FROM lottery_chance_ledger WHERE rule_id=$1`, rule.ID)
		_, _ = integrationDB.ExecContext(ctx, `DELETE FROM lottery_rules WHERE id=$1`, rule.ID)
	})

	grant := service.LotteryChanceGrant{
		UserID: userA, PoolKey: service.LotteryPoolNormal, Chances: 2, RuleID: rule.ID,
		SourceType: service.LotteryEventSignup, SourceID: fmt.Sprint(userB), SourceUserID: userB,
		DedupeKey: fmt.Sprintf("integration-grant-%d", time.Now().UnixNano()),
	}
	applied, err := repo.GrantExtraChance(ctx, grant)
	require.NoError(t, err)
	require.True(t, applied)
	applied, err = repo.GrantExtraChance(ctx, grant)
	require.NoError(t, err)
	require.False(t, applied)

	account := consumeLotteryChance(t, repo, userA, normal, "shared-key")
	require.Equal(t, normal.CycleChances-1, account.BaseRemaining)
	require.EqualValues(t, 2, account.ExtraRemaining)
	consumeLotteryChance(t, repo, userB, normal, "shared-key")

	amount := 1.0
	stock := int64(1)
	prize, err := repo.CreatePrize(ctx, service.LotteryPrizeInput{
		PoolID: normal.ID, Name: "last stock", PrizeType: service.LotteryPrizeBalance,
		BalanceAmount: &amount, ProbabilityPPM: 1, StockTotal: &stock, Enabled: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = integrationDB.ExecContext(ctx, `DELETE FROM lottery_prizes WHERE id=$1`, prize.ID) })

	start := make(chan struct{})
	results := make(chan bool, 2)
	errorsCh := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			claimed, claimErr := repo.ClaimPrizeStock(ctx, prize.ID)
			results <- claimed
			errorsCh <- claimErr
		}()
	}
	close(start)
	group.Wait()
	close(results)
	close(errorsCh)

	successes := 0
	for claimErr := range errorsCh {
		require.NoError(t, claimErr)
	}
	for claimed := range results {
		if claimed {
			successes++
		}
	}
	require.Equal(t, 1, successes)
}

func consumeLotteryChance(t *testing.T, repo service.LotteryRepository, userID int64, pool service.LotteryPool, key string) *service.LotteryChanceAccount {
	t.Helper()
	ctx := context.Background()
	tx, err := integrationEntClient.Tx(ctx)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	account, err := repo.LockChanceAccount(txCtx, userID, pool, "d:2099-01-01")
	require.NoError(t, err)
	_, account, err = repo.ConsumeChance(txCtx, *account, pool.ID, key)
	require.NoError(t, err, "不同用户复用同一客户端键不应冲突")
	require.NoError(t, tx.Commit())
	return account
}

func createLotteryTestUser(t *testing.T) int64 {
	t.Helper()
	ctx := context.Background()
	var id int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO users (email,password_hash,role,status,balance,concurrency)
		VALUES ($1,'test-hash','user','active',0,5) RETURNING id`,
		fmt.Sprintf("lottery-%d@example.com", time.Now().UnixNano()),
	).Scan(&id))
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(ctx, `DELETE FROM lottery_chance_ledger WHERE user_id=$1 OR source_user_id=$1`, id)
		_, _ = integrationDB.ExecContext(ctx, `DELETE FROM lottery_draws WHERE user_id=$1`, id)
		_, _ = integrationDB.ExecContext(ctx, `DELETE FROM lottery_user_chances WHERE user_id=$1`, id)
		_, _ = integrationDB.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, id)
	})
	return id
}
