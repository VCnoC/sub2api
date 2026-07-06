package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

// TeamRepository defines the persistence operations for teams.
type TeamRepository interface {
	Create(ctx context.Context, team *Team) error
	GetByID(ctx context.Context, id int64) (*Team, error)
	GetByOwnerID(ctx context.Context, ownerID int64) (*Team, error)
	GetByInviteCode(ctx context.Context, code string) (*Team, error)
	UpdateInviteCode(ctx context.Context, teamID int64, code string) error
	Update(ctx context.Context, team *Team) error
	Delete(ctx context.Context, teamID int64) error
}

// TeamService defines the business operations for teams.
type TeamService interface {
	CreateTeam(ctx context.Context, ownerID int64, name string) (*Team, error)
	GetMyTeam(ctx context.Context, userID int64) (*Team, string, error)
	RefreshInviteCode(ctx context.Context, ownerID int64) (string, error)
	JoinTeamByCode(ctx context.Context, userID int64, code string) error
	LeaveTeam(ctx context.Context, userID int64) error
	RemoveMember(ctx context.Context, ownerID, memberID int64) error
	ListMembers(ctx context.Context, userID int64, page, pageSize int) ([]TeamMember, int64, error)
	TransferBalance(ctx context.Context, ownerID, memberID int64, amount float64, password string) error
	ListMemberUsage(ctx context.Context, requesterID, memberID int64, startTime, endTime time.Time, page, pageSize int) ([]UsageLog, *pagination.PaginationResult, error)
}

// Team represents a team.
type Team struct {
	ID         int64
	Name       string
	OwnerID    int64
	InviteCode string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TeamMember extends User with team-specific usage information.
type TeamMember struct {
	User
	TotalUsage float64
}

type teamServiceImpl struct {
	teamRepo            TeamRepository
	userRepo            UserRepository
	usageLogRepo        UsageLogRepository
	redeemCodeRepo      RedeemCodeRepository
	authService         *AuthService
	billingCacheService *BillingCacheService
}

// NewTeamService creates a new TeamService.
func NewTeamService(
	teamRepo TeamRepository,
	userRepo UserRepository,
	usageLogRepo UsageLogRepository,
	redeemCodeRepo RedeemCodeRepository,
	authService *AuthService,
	billingCacheService *BillingCacheService,
) TeamService {
	return &teamServiceImpl{
		teamRepo:            teamRepo,
		userRepo:            userRepo,
		usageLogRepo:        usageLogRepo,
		redeemCodeRepo:      redeemCodeRepo,
		authService:         authService,
		billingCacheService: billingCacheService,
	}
}

func (s *teamServiceImpl) CreateTeam(ctx context.Context, ownerID int64, name string) (*Team, error) {
	owner, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if owner.TeamID != nil && *owner.TeamID != 0 {
		return nil, ErrAlreadyInTeam
	}

	code, err := generateTeamInviteCode()
	if err != nil {
		return nil, err
	}

	team := &Team{
		Name:       name,
		OwnerID:    ownerID,
		InviteCode: code,
		Status:     StatusActive,
	}
	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	// Bind owner to the team
	if err := s.userRepo.UpdateTeamMembership(ctx, ownerID, team.ID, TeamRoleOwner); err != nil {
		return nil, err
	}

	return team, nil
}

func (s *teamServiceImpl) GetMyTeam(ctx context.Context, userID int64) (*Team, string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if user.TeamID == nil || *user.TeamID == 0 {
		return nil, "", nil
	}

	team, err := s.teamRepo.GetByID(ctx, *user.TeamID)
	if err != nil {
		return nil, "", err
	}
	return team, user.TeamRole, nil
}

func (s *teamServiceImpl) RefreshInviteCode(ctx context.Context, ownerID int64) (string, error) {
	team, err := s.teamRepo.GetByOwnerID(ctx, ownerID)
	if err != nil {
		return "", err
	}

	code, err := generateTeamInviteCode()
	if err != nil {
		return "", err
	}

	if err := s.teamRepo.UpdateInviteCode(ctx, team.ID, code); err != nil {
		return "", err
	}
	return code, nil
}

func (s *teamServiceImpl) JoinTeamByCode(ctx context.Context, userID int64, code string) error {
	if code == "" {
		return ErrInviteCodeInvalid
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.TeamID != nil && *user.TeamID != 0 {
		return ErrAlreadyInTeam
	}

	team, err := s.teamRepo.GetByInviteCode(ctx, code)
	if err != nil {
		return ErrInviteCodeInvalid
	}

	return s.userRepo.UpdateTeamMembership(ctx, userID, team.ID, TeamRoleMember)
}

func (s *teamServiceImpl) LeaveTeam(ctx context.Context, userID int64) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.TeamID == nil || *user.TeamID == 0 {
		return nil
	}
	if user.TeamRole == TeamRoleOwner {
		return ErrCannotLeaveAsOwner
	}
	return s.userRepo.ClearTeamMembership(ctx, userID)
}

func (s *teamServiceImpl) RemoveMember(ctx context.Context, ownerID, memberID int64) error {
	owner, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return err
	}
	if owner.TeamRole != TeamRoleOwner || owner.TeamID == nil {
		return ErrNotTeamOwner
	}

	member, err := s.userRepo.GetByID(ctx, memberID)
	if err != nil {
		return err
	}
	if member.TeamID == nil || *member.TeamID != *owner.TeamID {
		return ErrTeamMemberNotFound
	}
	if member.TeamRole == TeamRoleOwner {
		return ErrCannotRemoveOwner
	}

	return s.userRepo.ClearTeamMembership(ctx, memberID)
}

func (s *teamServiceImpl) ListMembers(ctx context.Context, userID int64, page, pageSize int) ([]TeamMember, int64, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	// Any team member can view the member list; owner-only actions
	// (remove/transfer/usage) keep their own checks.
	if user.TeamID == nil || *user.TeamID == 0 {
		return nil, 0, ErrNotInTeam
	}

	params := pagination.PaginationParams{Page: page, PageSize: pageSize, SortBy: "created_at", SortOrder: "desc"}
	users, result, err := s.userRepo.ListByTeamID(ctx, *user.TeamID, params)
	if err != nil {
		return nil, 0, err
	}

	memberIDs := make([]int64, 0, len(users))
	for i := range users {
		memberIDs = append(memberIDs, users[i].ID)
	}

	// Aggregate usage by user IDs. Usage logs are append-only; sum(actual_cost) gives total usage.
	usageByUser, err := s.aggregateUsageByUserIDs(ctx, memberIDs)
	if err != nil {
		return nil, 0, err
	}

	members := make([]TeamMember, 0, len(users))
	for i := range users {
		members = append(members, TeamMember{
			User:       users[i],
			TotalUsage: usageByUser[users[i].ID],
		})
	}
	return members, result.Total, nil
}

func (s *teamServiceImpl) TransferBalance(ctx context.Context, ownerID, memberID int64, amount float64, password string) error {
	if amount <= 0 {
		return ErrInsufficientTeamBalance
	}

	owner, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return err
	}
	if owner.TeamRole != TeamRoleOwner || owner.TeamID == nil {
		return ErrNotTeamOwner
	}

	member, err := s.userRepo.GetByID(ctx, memberID)
	if err != nil {
		return err
	}
	if member.TeamID == nil || *member.TeamID != *owner.TeamID {
		return ErrTeamMemberNotFound
	}
	if member.TeamRole == TeamRoleOwner {
		return ErrTeamMemberNotFound
	}

	// Verify owner password
	if !s.authService.CheckPassword(password, owner.PasswordHash) {
		return ErrTeamPasswordIncorrect
	}

	if owner.Balance < amount {
		return ErrInsufficientTeamBalance
	}

	// Deduct from owner and add to member atomically via repository operations.
	// Both UpdateBalance calls support negative amounts; we run them sequentially
	// inside the service. The repository UpdateBalance uses an UPDATE statement
	// which is atomic at the row level. For stricter atomicity, this can later be
	// wrapped in a DB transaction.
	if err := s.userRepo.UpdateBalance(ctx, ownerID, -amount); err != nil {
		return err
	}
	if err := s.userRepo.UpdateBalance(ctx, memberID, amount); err != nil {
		// Best-effort rollback: restore owner balance.
		_ = s.userRepo.UpdateBalance(ctx, ownerID, amount)
		return err
	}

	// Record audit redeem codes
	outCode, _ := GenerateRedeemCode()
	inCode, _ := GenerateRedeemCode()
	now := time.Now()

	outRecord := &RedeemCode{
		Code:   outCode,
		Type:   RedeemTypeTeamTransferOut,
		Value:  -amount,
		Status: StatusUsed,
		UsedBy: &ownerID,
		Notes:  fmt.Sprintf("transfer to team member %d", memberID),
		UsedAt: &now,
	}
	inRecord := &RedeemCode{
		Code:   inCode,
		Type:   RedeemTypeTeamTransferIn,
		Value:  amount,
		Status: StatusUsed,
		UsedBy: &memberID,
		Notes:  fmt.Sprintf("transfer from team owner %d", ownerID),
		UsedAt: &now,
	}

	if err := s.redeemCodeRepo.Create(ctx, outRecord); err != nil {
		logger.LegacyPrintf("service.team", "failed to create team transfer out record: %v", err)
	}
	if err := s.redeemCodeRepo.Create(ctx, inRecord); err != nil {
		logger.LegacyPrintf("service.team", "failed to create team transfer in record: %v", err)
	}

	// Invalidate billing cache for both users
	if s.billingCacheService != nil {
		go func() {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.billingCacheService.InvalidateUserBalance(cacheCtx, ownerID)
			_ = s.billingCacheService.InvalidateUserBalance(cacheCtx, memberID)
		}()
	}

	return nil
}

func (s *teamServiceImpl) aggregateUsageByUserIDs(ctx context.Context, userIDs []int64) (map[int64]float64, error) {
	result := make(map[int64]float64, len(userIDs))
	if len(userIDs) == 0 || s.usageLogRepo == nil {
		return result, nil
	}

	stats, err := s.usageLogRepo.GetBatchUserUsageStats(
		ctx,
		userIDs,
		time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Now().UTC(),
	)
	if err != nil {
		return nil, err
	}
	for userID, st := range stats {
		result[userID] = st.TotalActualCost
	}
	return result, nil
}

func (s *teamServiceImpl) ListMemberUsage(ctx context.Context, requesterID, memberID int64, startTime, endTime time.Time, page, pageSize int) ([]UsageLog, *pagination.PaginationResult, error) {
	requester, err := s.userRepo.GetByID(ctx, requesterID)
	if err != nil {
		return nil, nil, err
	}
	if requester.TeamID == nil || *requester.TeamID == 0 {
		return nil, nil, ErrNotInTeam
	}
	// Owner can view any member's usage; regular members can only view their own.
	if requester.TeamRole != TeamRoleOwner && requesterID != memberID {
		return nil, nil, ErrNotTeamOwner
	}

	member, err := s.userRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, nil, err
	}
	if member.TeamID == nil || *member.TeamID != *requester.TeamID {
		return nil, nil, ErrTeamMemberNotFound
	}

	params := pagination.PaginationParams{Page: page, PageSize: pageSize, SortBy: "created_at", SortOrder: "desc"}
	filters := usagestats.UsageLogFilters{
		UserID:    memberID,
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	return s.usageLogRepo.ListWithFilters(ctx, params, filters)
}

func generateTeamInviteCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:8], nil
}
