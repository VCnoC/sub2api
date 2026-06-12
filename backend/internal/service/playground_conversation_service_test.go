package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// pcStubRepo 是 PlaygroundConversationRepository 的测试桩。
// 通过字段注入返回值/错误，并记录调用入参供断言。
type pcStubRepo struct {
	count    int
	countErr error

	getResult *PlaygroundConversation
	getErr    error

	createErr     error
	createdRecord *PlaygroundConversation

	updateErr     error
	updatedRecord *PlaygroundConversation

	deleteErr error
}

func (r *pcStubRepo) ListByUser(_ context.Context, _ int64) ([]PlaygroundConversationSummary, error) {
	return nil, nil
}

func (r *pcStubRepo) GetByID(_ context.Context, _, _ int64) (*PlaygroundConversation, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if r.getResult == nil {
		return nil, ErrPlaygroundConversationNotFound
	}
	// 返回副本，避免测试断言被 service 层的原地修改干扰
	cp := *r.getResult
	return &cp, nil
}

func (r *pcStubRepo) Create(_ context.Context, c *PlaygroundConversation) error {
	if r.createErr != nil {
		return r.createErr
	}
	c.ID = 1
	r.createdRecord = c
	return nil
}

func (r *pcStubRepo) Update(_ context.Context, c *PlaygroundConversation) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.updatedRecord = c
	return nil
}

func (r *pcStubRepo) Delete(_ context.Context, _, _ int64) error {
	return r.deleteErr
}

func (r *pcStubRepo) CountByUser(_ context.Context, _ int64) (int, error) {
	return r.count, r.countErr
}

func (r *pcStubRepo) DeleteExpired(_ context.Context, _ time.Time, _ int) (int, error) {
	return 0, nil
}

func pcStrPtr(s string) *string { return &s }

func TestPlaygroundConversationService_Create_LimitExceeded(t *testing.T) {
	repo := &pcStubRepo{count: PlaygroundConversationMaxPerUser}
	svc := NewPlaygroundConversationService(repo)

	_, err := svc.Create(context.Background(), 1, "标题", nil, nil, nil)
	require.ErrorIs(t, err, ErrPlaygroundConversationLimitExceeded)
	require.Nil(t, repo.createdRecord, "超限时不应触达 repo.Create")
}

func TestPlaygroundConversationService_Create_TooLarge(t *testing.T) {
	repo := &pcStubRepo{count: 0}
	svc := NewPlaygroundConversationService(repo)

	oversized := json.RawMessage(make([]byte, PlaygroundConversationMaxMessagesBytes+1))
	_, err := svc.Create(context.Background(), 1, "标题", nil, nil, oversized)
	require.ErrorIs(t, err, ErrPlaygroundConversationTooLarge)
	require.Nil(t, repo.createdRecord, "超体积时不应触达 repo.Create")
}

func TestPlaygroundConversationService_Create_Success(t *testing.T) {
	repo := &pcStubRepo{count: PlaygroundConversationMaxPerUser - 1}
	svc := NewPlaygroundConversationService(repo)

	// 300 个中文字符的标题应被截断至 255 个 rune
	longTitle := strings.Repeat("题", 300)
	msgs := json.RawMessage(`[{"role":"user","content":"hi"}]`)

	got, err := svc.Create(context.Background(), 42, longTitle, pcStrPtr("gpt-5.5"), pcStrPtr("default"), msgs)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.EqualValues(t, 1, got.ID, "Create 应回填 ID")
	require.EqualValues(t, 42, got.UserID)
	require.Equal(t, 255, len([]rune(got.Title)), "标题应按 rune 截断至 255")
	require.Equal(t, "gpt-5.5", *got.Model)
	require.Equal(t, "default", *got.GroupName)
	require.False(t, got.LastActivityAt.IsZero(), "LastActivityAt 应被设置")
}

func TestPlaygroundConversationService_Update_NotFound(t *testing.T) {
	repo := &pcStubRepo{getErr: ErrPlaygroundConversationNotFound}
	svc := NewPlaygroundConversationService(repo)

	err := svc.Update(context.Background(), 1, 1, pcStrPtr("新标题"), nil, nil, nil)
	require.ErrorIs(t, err, ErrPlaygroundConversationNotFound)
	require.Nil(t, repo.updatedRecord)
}

func TestPlaygroundConversationService_Update_TooLarge(t *testing.T) {
	repo := &pcStubRepo{
		getResult: &PlaygroundConversation{ID: 1, UserID: 1, Title: "旧标题"},
	}
	svc := NewPlaygroundConversationService(repo)

	oversized := json.RawMessage(make([]byte, PlaygroundConversationMaxMessagesBytes+1))
	err := svc.Update(context.Background(), 1, 1, nil, nil, nil, oversized)
	require.ErrorIs(t, err, ErrPlaygroundConversationTooLarge)
	require.Nil(t, repo.updatedRecord, "超体积时不应触达 repo.Update")
}

// TestPlaygroundConversationService_Update_PartialSemantics 钉死 Update 的部分更新语义：
//   - title 传 nil  → 保持原值
//   - messages 传 nil → 保持原值
//   - model/groupName 传 nil → 清空（注意：调用方每次保存都必须带上 model/group_name！）
//   - 任何更新都刷新 LastActivityAt
func TestPlaygroundConversationService_Update_PartialSemantics(t *testing.T) {
	oldActivity := time.Now().Add(-time.Hour)
	repo := &pcStubRepo{
		getResult: &PlaygroundConversation{
			ID:             1,
			UserID:         1,
			Title:          "旧标题",
			Model:          pcStrPtr("gpt-5.4"),
			GroupName:      pcStrPtr("default"),
			Messages:       json.RawMessage(`[{"role":"user","content":"old"}]`),
			LastActivityAt: oldActivity,
		},
	}
	svc := NewPlaygroundConversationService(repo)

	// 只传 model/groupName，title/messages 均为 nil
	err := svc.Update(context.Background(), 1, 1, nil, pcStrPtr("gpt-5.5"), pcStrPtr("pro"), nil)
	require.NoError(t, err)
	require.NotNil(t, repo.updatedRecord)

	got := repo.updatedRecord
	require.Equal(t, "旧标题", got.Title, "title 传 nil 应保持原值")
	require.JSONEq(t, `[{"role":"user","content":"old"}]`, string(got.Messages), "messages 传 nil 应保持原值")
	require.Equal(t, "gpt-5.5", *got.Model)
	require.Equal(t, "pro", *got.GroupName)
	require.True(t, got.LastActivityAt.After(oldActivity), "LastActivityAt 应被刷新")

	// model/groupName 传 nil → 清空语义
	err = svc.Update(context.Background(), 1, 1, pcStrPtr("新标题"), nil, nil, nil)
	require.NoError(t, err)
	got = repo.updatedRecord
	require.Equal(t, "新标题", got.Title)
	require.Nil(t, got.Model, "model 传 nil 应清空")
	require.Nil(t, got.GroupName, "group_name 传 nil 应清空")
}

func TestPlaygroundConversationService_Update_TitleTruncate(t *testing.T) {
	repo := &pcStubRepo{
		getResult: &PlaygroundConversation{ID: 1, UserID: 1, Title: "旧标题"},
	}
	svc := NewPlaygroundConversationService(repo)

	longTitle := strings.Repeat("长", 300)
	err := svc.Update(context.Background(), 1, 1, &longTitle, nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 255, len([]rune(repo.updatedRecord.Title)))
}

func TestPlaygroundConversationService_Delete_Passthrough(t *testing.T) {
	repo := &pcStubRepo{deleteErr: ErrPlaygroundConversationNotFound}
	svc := NewPlaygroundConversationService(repo)

	err := svc.Delete(context.Background(), 1, 1)
	require.ErrorIs(t, err, ErrPlaygroundConversationNotFound)

	repo.deleteErr = nil
	require.NoError(t, svc.Delete(context.Background(), 1, 1))
}
