package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

// newPlaygroundConversationRepoSQLite 创建基于 SQLite 内存库的仓储实例（测试专用）。
func newPlaygroundConversationRepoSQLite(t *testing.T) *playgroundConversationRepository {
	t.Helper()

	db, err := sql.Open("sqlite", "file:playground_conversation_repo_"+t.Name()+"?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	return &playgroundConversationRepository{client: client}
}

// mustCreatePC 创建一条会话记录并返回（last_activity_at 可指定，便于排序/过期测试）。
func mustCreatePC(
	t *testing.T,
	ctx context.Context,
	repo *playgroundConversationRepository,
	userID int64,
	title string,
	lastActivityAt time.Time,
) *service.PlaygroundConversation {
	t.Helper()
	model := "gpt-5.5"
	groupName := "default"
	c := &service.PlaygroundConversation{
		UserID:         userID,
		Title:          title,
		Model:          &model,
		GroupName:      &groupName,
		Messages:       json.RawMessage(`[{"role":"user","content":"hi"}]`),
		LastActivityAt: lastActivityAt,
	}
	require.NoError(t, repo.Create(ctx, c))
	require.NotZero(t, c.ID, "Create 应回填 ID")
	return c
}

func TestPlaygroundConversationRepo_CreateAndGet(t *testing.T) {
	repo := newPlaygroundConversationRepoSQLite(t)
	ctx := context.Background()

	created := mustCreatePC(t, ctx, repo, 1, "测试会话", time.Now())

	// 归属用户可以读取完整数据（含 messages）
	got, err := repo.GetByID(ctx, created.ID, 1)
	require.NoError(t, err)
	require.Equal(t, "测试会话", got.Title)
	require.NotNil(t, got.Model)
	require.Equal(t, "gpt-5.5", *got.Model)
	require.JSONEq(t, `[{"role":"user","content":"hi"}]`, string(got.Messages))

	// 越权防护：其他用户读取同一 ID 必须返回 NotFound
	_, err = repo.GetByID(ctx, created.ID, 2)
	require.ErrorIs(t, err, service.ErrPlaygroundConversationNotFound)
}

func TestPlaygroundConversationRepo_ListByUser(t *testing.T) {
	repo := newPlaygroundConversationRepoSQLite(t)
	ctx := context.Background()

	base := time.Now().Add(-time.Hour)
	mustCreatePC(t, ctx, repo, 1, "最旧", base)
	mustCreatePC(t, ctx, repo, 1, "中间", base.Add(10*time.Minute))
	mustCreatePC(t, ctx, repo, 1, "最新", base.Add(20*time.Minute))
	// 其他用户的会话不应出现在列表中
	mustCreatePC(t, ctx, repo, 2, "别人的", base.Add(30*time.Minute))

	list, err := repo.ListByUser(ctx, 1)
	require.NoError(t, err)
	require.Len(t, list, 3)

	// 按 last_activity_at 倒序
	require.Equal(t, "最新", list[0].Title)
	require.Equal(t, "中间", list[1].Title)
	require.Equal(t, "最旧", list[2].Title)
}

func TestPlaygroundConversationRepo_UpdateOwnershipAndClear(t *testing.T) {
	repo := newPlaygroundConversationRepoSQLite(t)
	ctx := context.Background()

	created := mustCreatePC(t, ctx, repo, 1, "原标题", time.Now())

	// 越权更新：user 2 修改 user 1 的会话 → NotFound
	evil := *created
	evil.UserID = 2
	evil.Title = "黑掉你"
	require.ErrorIs(t, repo.Update(ctx, &evil), service.ErrPlaygroundConversationNotFound)

	// 归属用户正常更新：改标题 + 清空 model/group_name + 替换 messages
	created.Title = "新标题"
	created.Model = nil
	created.GroupName = nil
	created.Messages = json.RawMessage(`[{"role":"user","content":"updated"}]`)
	created.LastActivityAt = time.Now().Add(time.Minute)
	require.NoError(t, repo.Update(ctx, created))

	got, err := repo.GetByID(ctx, created.ID, 1)
	require.NoError(t, err)
	require.Equal(t, "新标题", got.Title)
	require.Nil(t, got.Model, "model 传 nil 应被清空")
	require.Nil(t, got.GroupName, "group_name 传 nil 应被清空")
	require.JSONEq(t, `[{"role":"user","content":"updated"}]`, string(got.Messages))

	// 确认越权更新没有产生实际效果
	require.NotEqual(t, "黑掉你", got.Title)
}

func TestPlaygroundConversationRepo_Delete(t *testing.T) {
	repo := newPlaygroundConversationRepoSQLite(t)
	ctx := context.Background()

	created := mustCreatePC(t, ctx, repo, 1, "待删除", time.Now())

	// 越权删除 → NotFound，且记录仍存在
	require.ErrorIs(t, repo.Delete(ctx, created.ID, 2), service.ErrPlaygroundConversationNotFound)
	_, err := repo.GetByID(ctx, created.ID, 1)
	require.NoError(t, err)

	// 归属用户删除成功，再次读取 → NotFound
	require.NoError(t, repo.Delete(ctx, created.ID, 1))
	_, err = repo.GetByID(ctx, created.ID, 1)
	require.ErrorIs(t, err, service.ErrPlaygroundConversationNotFound)

	// 删除不存在的 ID → NotFound
	require.ErrorIs(t, repo.Delete(ctx, 99999, 1), service.ErrPlaygroundConversationNotFound)
}

func TestPlaygroundConversationRepo_CountByUser(t *testing.T) {
	repo := newPlaygroundConversationRepoSQLite(t)
	ctx := context.Background()

	count, err := repo.CountByUser(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	mustCreatePC(t, ctx, repo, 1, "a", time.Now())
	mustCreatePC(t, ctx, repo, 1, "b", time.Now())
	mustCreatePC(t, ctx, repo, 2, "别人的", time.Now())

	count, err = repo.CountByUser(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

func TestPlaygroundConversationRepo_DeleteExpired(t *testing.T) {
	repo := newPlaygroundConversationRepoSQLite(t)
	ctx := context.Background()

	// 时间边界：retention 截止点
	before := time.Now().Add(-72 * time.Hour).Truncate(time.Second)

	expired1 := mustCreatePC(t, ctx, repo, 1, "过期1", before.Add(-time.Second))
	expired2 := mustCreatePC(t, ctx, repo, 1, "过期2", before.Add(-time.Hour))
	// 边界值：last_activity_at 恰好等于 before，LT 为严格小于 → 不应被删除
	boundary := mustCreatePC(t, ctx, repo, 1, "恰好边界", before)
	fresh := mustCreatePC(t, ctx, repo, 1, "新鲜", before.Add(time.Hour))

	// batchSize=1 验证分批语义：每次最多删 1 条
	n, err := repo.DeleteExpired(ctx, before, 1)
	require.NoError(t, err)
	require.Equal(t, 1, n)

	// 循环调用直至返回 0（清理服务的调用方式）
	total := n
	for {
		n, err = repo.DeleteExpired(ctx, before, 1)
		require.NoError(t, err)
		if n == 0 {
			break
		}
		total += n
	}
	require.Equal(t, 2, total, "应只删除 2 条过期会话")

	// 过期的已删除
	_, err = repo.GetByID(ctx, expired1.ID, 1)
	require.ErrorIs(t, err, service.ErrPlaygroundConversationNotFound)
	_, err = repo.GetByID(ctx, expired2.ID, 1)
	require.ErrorIs(t, err, service.ErrPlaygroundConversationNotFound)

	// 边界与新鲜的保留
	_, err = repo.GetByID(ctx, boundary.ID, 1)
	require.NoError(t, err, "last_activity_at 恰好等于截止点的会话不应被删除")
	_, err = repo.GetByID(ctx, fresh.ID, 1)
	require.NoError(t, err)

	// batchSize<=0 时使用默认兜底值，不应报错
	n, err = repo.DeleteExpired(ctx, before, 0)
	require.NoError(t, err)
	require.Equal(t, 0, n)
}
