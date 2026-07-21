package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

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
	AddBalance(ctx context.Context, teamID int64, amount float64) error
	DeductBalance(ctx context.Context, teamID int64, amount float64) error
}

// TeamService defines the business operations for teams.
type TeamService interface {
	CreateTeam(ctx context.Context, ownerID int64, name, reason, additionalInfo string) (*TeamApplication, error)
	GetMyCreateApplication(ctx context.Context, userID int64) (*TeamApplication, error)
	GetCreationEligibility(ctx context.Context, userID int64) (*TeamCreationEligibility, error)
	GetMyTeam(ctx context.Context, userID int64) (*Team, string, error)
	RefreshInviteCode(ctx context.Context, ownerID int64) (string, error)
	JoinTeamByCode(ctx context.Context, userID int64, code, message string) (*TeamJoinRequest, error)
	ListJoinRequests(ctx context.Context, ownerID int64, status string) ([]TeamJoinRequest, error)
	ReviewJoinRequest(ctx context.Context, ownerID, requestID int64, approve bool, reason string) (*TeamJoinRequest, error)
	GetGovernanceState(ctx context.Context, userID int64) (*TeamGovernanceState, error)
	UpgradeTeam(ctx context.Context, ownerID int64) (*TeamGovernanceState, error)
	SubmitExpandApplication(ctx context.Context, ownerID int64, targetLimit int, reason string) (*TeamApplication, error)
	LeaveTeam(ctx context.Context, userID int64) error
	RemoveMember(ctx context.Context, ownerID, memberID int64) error
	ListMembers(ctx context.Context, userID int64, page, pageSize int) ([]TeamMember, int64, error)
	TransferBalance(ctx context.Context, ownerID, memberID int64, amount float64, password string) error
	ListMemberUsage(ctx context.Context, requesterID, memberID int64, startTime, endTime time.Time, page, pageSize int) ([]UsageLog, *pagination.PaginationResult, error)
	DepositToFund(ctx context.Context, userID int64, amount float64, password string) error
	AllocateFund(ctx context.Context, ownerID, memberID int64, amount float64, password string) error
	GetSettings(ctx context.Context) (*TeamGovernanceSettings, error)
	UpdateSettings(ctx context.Context, adminID int64, settings TeamGovernanceSettings) (*TeamGovernanceSettings, error)
	GetAdminStats(ctx context.Context) (*TeamAdminStats, error)
	ListAdminTeams(ctx context.Context, search, status string, page, pageSize int) ([]AdminTeamSummary, int64, error)
	GetAdminTeam(ctx context.Context, teamID int64) (*AdminTeamDetail, error)
	ListApplications(ctx context.Context, status string, page, pageSize int) ([]TeamApplication, int64, error)
	ReviewApplication(ctx context.Context, applicationID, adminID int64, input ReviewTeamApplicationInput) (*TeamApplication, error)
	SetTeamStatus(ctx context.Context, teamID int64, status string) error
	SetTeamMemberLimit(ctx context.Context, teamID int64, limit int) error
	MarkTeamReviewed(ctx context.Context, teamID int64) error
	AdminRemoveMember(ctx context.Context, teamID, memberID int64) error
}

// Team represents a team.
type Team struct {
	ID                  int64
	Name                string
	OwnerID             int64
	InviteCode          string
	Status              string
	Balance             float64
	MemberLimit         int
	Level               int
	ReviewRequired      bool
	MemberCount         int
	EffectiveRecharge   float64
	Spend7Days          float64
	TransferableBalance float64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// TeamMember extends User with team-specific usage information.
type TeamMember struct {
	User
	TotalUsage float64
	// BalanceVisible 余额是否对请求者可见：owner 可见全员，普通成员仅可见自己
	BalanceVisible bool
	// UsageVisible 累计用量是否对请求者可见：owner 可见全员，普通成员仅可见自己
	UsageVisible bool
}

type teamServiceImpl struct {
	teamRepo            TeamRepository
	governanceRepo      TeamGovernanceRepository
	userRepo            UserRepository
	usageLogRepo        UsageLogRepository
	authService         *AuthService
	billingCacheService *BillingCacheService
}

// NewTeamService creates a new TeamService.
func NewTeamService(
	teamRepo TeamRepository,
	governanceRepo TeamGovernanceRepository,
	userRepo UserRepository,
	usageLogRepo UsageLogRepository,
	authService *AuthService,
	billingCacheService *BillingCacheService,
) TeamService {
	return &teamServiceImpl{
		teamRepo:            teamRepo,
		governanceRepo:      governanceRepo,
		userRepo:            userRepo,
		usageLogRepo:        usageLogRepo,
		authService:         authService,
		billingCacheService: billingCacheService,
	}
}

func (s *teamServiceImpl) CreateTeam(ctx context.Context, ownerID int64, name, reason, additionalInfo string) (*TeamApplication, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 100 {
		return nil, ErrInvalidTeamName
	}
	owner, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if owner.TeamID != nil && *owner.TeamID != 0 {
		return nil, ErrAlreadyInTeam
	}

	return s.governanceRepo.SubmitCreateApplication(ctx, ownerID, name, strings.TrimSpace(reason), strings.TrimSpace(additionalInfo))
}

func (s *teamServiceImpl) GetMyCreateApplication(ctx context.Context, userID int64) (*TeamApplication, error) {
	return s.governanceRepo.GetLatestCreateApplication(ctx, userID)
}

func (s *teamServiceImpl) GetCreationEligibility(ctx context.Context, userID int64) (*TeamCreationEligibility, error) {
	days, recharge, err := s.governanceRepo.GetUserEligibility(ctx, userID)
	if err != nil {
		return nil, err
	}
	settings, err := s.governanceRepo.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	return &TeamCreationEligibility{
		RegistrationDays: days, EffectiveRecharge: recharge,
		Eligible: settings.Configured && days >= settings.MinRegistrationDays && recharge >= settings.MinTotalRecharge,
		Settings: *settings,
	}, nil
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
	state, err := s.governanceRepo.GetGovernanceState(ctx, userID, team.ID)
	if err != nil {
		return nil, "", err
	}
	team.MemberLimit = state.MemberLimit
	team.Level = state.Level
	team.ReviewRequired = state.ReviewRequired
	team.MemberCount = state.MemberCount
	team.EffectiveRecharge = state.EffectiveRecharge
	team.Spend7Days = state.Spend7Days
	team.TransferableBalance = state.TransferableBalance
	return team, user.TeamRole, nil
}

func (s *teamServiceImpl) RefreshInviteCode(ctx context.Context, ownerID int64) (string, error) {
	team, err := s.teamRepo.GetByOwnerID(ctx, ownerID)
	if err != nil {
		return "", err
	}
	if team.Status != StatusActive {
		return "", ErrTeamFrozen
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

func (s *teamServiceImpl) JoinTeamByCode(ctx context.Context, userID int64, code, message string) (*TeamJoinRequest, error) {
	if code == "" {
		return nil, ErrInviteCodeInvalid
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.TeamID != nil && *user.TeamID != 0 {
		return nil, ErrAlreadyInTeam
	}
	return s.governanceRepo.SubmitJoinRequest(ctx, userID, strings.TrimSpace(code), strings.TrimSpace(message))
}

func (s *teamServiceImpl) ListJoinRequests(ctx context.Context, ownerID int64, status string) ([]TeamJoinRequest, error) {
	return s.governanceRepo.ListJoinRequests(ctx, ownerID, status)
}

func (s *teamServiceImpl) ReviewJoinRequest(ctx context.Context, ownerID, requestID int64, approve bool, reason string) (*TeamJoinRequest, error) {
	return s.governanceRepo.ReviewJoinRequest(ctx, ownerID, requestID, approve, strings.TrimSpace(reason))
}

func (s *teamServiceImpl) GetGovernanceState(ctx context.Context, userID int64) (*TeamGovernanceState, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.TeamID == nil {
		return nil, ErrNotInTeam
	}
	return s.governanceRepo.GetGovernanceState(ctx, userID, *user.TeamID)
}

func (s *teamServiceImpl) UpgradeTeam(ctx context.Context, ownerID int64) (*TeamGovernanceState, error) {
	owner, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if owner.TeamID == nil || owner.TeamRole != TeamRoleOwner {
		return nil, ErrNotTeamOwner
	}
	return s.governanceRepo.UpgradeTeam(ctx, ownerID, *owner.TeamID)
}

func (s *teamServiceImpl) SubmitExpandApplication(ctx context.Context, ownerID int64, targetLimit int, reason string) (*TeamApplication, error) {
	if targetLimit <= 40 || strings.TrimSpace(reason) == "" {
		return nil, ErrInvalidTeamLimit
	}
	owner, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if owner.TeamID == nil || owner.TeamRole != TeamRoleOwner {
		return nil, ErrNotTeamOwner
	}
	return s.governanceRepo.SubmitExpandApplication(ctx, ownerID, *owner.TeamID, targetLimit, strings.TrimSpace(reason))
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

	viewerIsOwner := user.TeamRole == TeamRoleOwner
	members := make([]TeamMember, 0, len(users))
	for i := range users {
		// 全员可见成员列表；余额/累计用量仅 owner 或自己可见
		visible := viewerIsOwner || users[i].ID == userID
		m := TeamMember{
			User:           users[i],
			TotalUsage:     usageByUser[users[i].ID],
			BalanceVisible: visible,
			UsageVisible:   visible,
		}
		if !visible {
			m.Balance = 0
			m.TotalUsage = 0
		}
		members = append(members, m)
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

	if err := s.governanceRepo.TransferTeamBalance(ctx, ownerID, memberID, *owner.TeamID, amount); err != nil {
		return err
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

// DepositToFund 成员将自己的余额存入团队资金池。
func (s *teamServiceImpl) DepositToFund(ctx context.Context, userID int64, amount float64, password string) error {
	if amount <= 0 {
		return ErrInvalidFundAmount
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.TeamID == nil || *user.TeamID == 0 {
		return ErrNotInTeam
	}

	if !s.authService.CheckPassword(password, user.PasswordHash) {
		return ErrTeamPasswordIncorrect
	}
	if err := s.governanceRepo.DepositTeamFund(ctx, userID, *user.TeamID, amount); err != nil {
		return err
	}

	if s.billingCacheService != nil {
		go func() {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.billingCacheService.InvalidateUserBalance(cacheCtx, userID)
		}()
	}

	return nil
}

// AllocateFund owner 将团队资金分配给团队成员（含 owner 自己）。
func (s *teamServiceImpl) AllocateFund(ctx context.Context, ownerID, memberID int64, amount float64, password string) error {
	if amount <= 0 {
		return ErrInvalidFundAmount
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

	if !s.authService.CheckPassword(password, owner.PasswordHash) {
		return ErrTeamPasswordIncorrect
	}

	if err := s.governanceRepo.AllocateTeamFund(ctx, ownerID, memberID, *owner.TeamID, amount); err != nil {
		return err
	}

	if s.billingCacheService != nil {
		go func() {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
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

func (s *teamServiceImpl) GetSettings(ctx context.Context) (*TeamGovernanceSettings, error) {
	return s.governanceRepo.GetSettings(ctx)
}

func (s *teamServiceImpl) UpdateSettings(ctx context.Context, adminID int64, settings TeamGovernanceSettings) (*TeamGovernanceSettings, error) {
	if settings.MinRegistrationDays < 0 || settings.MinTotalRecharge < 0 {
		return nil, ErrInvalidTeamSettings
	}
	for _, level := range settings.Levels {
		if (level.Limit != 5 && level.Limit != 15 && level.Limit != 40) || level.Recharge < 0 || level.Spend7Days < 0 || (level.Mode != "and" && level.Mode != "or") {
			return nil, ErrInvalidTeamSettings
		}
	}
	return s.governanceRepo.UpdateSettings(ctx, adminID, settings)
}

func (s *teamServiceImpl) GetAdminStats(ctx context.Context) (*TeamAdminStats, error) {
	return s.governanceRepo.GetAdminStats(ctx)
}

func (s *teamServiceImpl) ListAdminTeams(ctx context.Context, search, status string, page, pageSize int) ([]AdminTeamSummary, int64, error) {
	return s.governanceRepo.ListAdminTeams(ctx, strings.TrimSpace(search), strings.TrimSpace(status), page, pageSize)
}

func (s *teamServiceImpl) GetAdminTeam(ctx context.Context, teamID int64) (*AdminTeamDetail, error) {
	return s.governanceRepo.GetAdminTeam(ctx, teamID)
}

func (s *teamServiceImpl) ListApplications(ctx context.Context, status string, page, pageSize int) ([]TeamApplication, int64, error) {
	return s.governanceRepo.ListApplications(ctx, strings.TrimSpace(status), page, pageSize)
}

func (s *teamServiceImpl) ReviewApplication(ctx context.Context, applicationID, adminID int64, input ReviewTeamApplicationInput) (*TeamApplication, error) {
	code := ""
	if input.Approve {
		var err error
		code, err = generateTeamInviteCode()
		if err != nil {
			return nil, err
		}
	}
	input.ReviewReason = strings.TrimSpace(input.ReviewReason)
	return s.governanceRepo.ReviewApplication(ctx, applicationID, adminID, input, code)
}

func (s *teamServiceImpl) SetTeamStatus(ctx context.Context, teamID int64, status string) error {
	if status != StatusActive && status != StatusDisabled {
		return ErrTeamNotFound
	}
	return s.governanceRepo.SetTeamStatus(ctx, teamID, status)
}

func (s *teamServiceImpl) SetTeamMemberLimit(ctx context.Context, teamID int64, limit int) error {
	if limit <= 0 {
		return ErrInvalidTeamLimit
	}
	return s.governanceRepo.SetTeamMemberLimit(ctx, teamID, limit)
}

func (s *teamServiceImpl) MarkTeamReviewed(ctx context.Context, teamID int64) error {
	return s.governanceRepo.MarkTeamReviewed(ctx, teamID)
}

func (s *teamServiceImpl) AdminRemoveMember(ctx context.Context, teamID, memberID int64) error {
	return s.governanceRepo.AdminRemoveMember(ctx, teamID, memberID)
}

func generateTeamInviteCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:8], nil
}
