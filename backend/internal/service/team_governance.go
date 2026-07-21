// Package service 中的本文件定义团队治理领域模型与持久化边界。
package service

import (
	"context"
	"time"
)

const (
	TeamApplicationCreate = "create"
	TeamApplicationExpand = "expand"
	TeamRequestPending    = "pending"
	TeamRequestApproved   = "approved"
	TeamRequestRejected   = "rejected"
)

type TeamLevelRequirement struct {
	Limit      int     `json:"limit"`
	Recharge   float64 `json:"recharge"`
	Spend7Days float64 `json:"spend_7d"`
	Mode       string  `json:"mode"`
}

type TeamGovernanceSettings struct {
	Configured          bool                   `json:"configured"`
	MinRegistrationDays int                    `json:"min_registration_days"`
	MinTotalRecharge    float64                `json:"min_total_recharge"`
	Levels              []TeamLevelRequirement `json:"levels"`
	UpdatedBy           *int64                 `json:"updated_by,omitempty"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

type TeamApplication struct {
	ID                int64      `json:"id"`
	ApplicationType   string     `json:"application_type"`
	ApplicantID       int64      `json:"applicant_id"`
	ApplicantEmail    string     `json:"applicant_email,omitempty"`
	RegistrationDays  int        `json:"registration_days"`
	EffectiveRecharge float64    `json:"effective_recharge"`
	TeamID            *int64     `json:"team_id,omitempty"`
	TeamName          string     `json:"team_name"`
	TargetLimit       *int       `json:"target_limit,omitempty"`
	Reason            string     `json:"reason"`
	AdditionalInfo    string     `json:"additional_info"`
	Status            string     `json:"status"`
	ReviewReason      string     `json:"review_reason"`
	Waived            bool       `json:"waived"`
	ReviewerID        *int64     `json:"reviewer_id,omitempty"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	CreatedTeamID     *int64     `json:"created_team_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type TeamJoinRequest struct {
	ID             int64      `json:"id"`
	TeamID         int64      `json:"team_id"`
	TeamName       string     `json:"team_name"`
	ApplicantID    int64      `json:"applicant_id"`
	ApplicantEmail string     `json:"applicant_email"`
	Message        string     `json:"message"`
	Status         string     `json:"status"`
	ReviewReason   string     `json:"review_reason"`
	ReviewedBy     *int64     `json:"reviewed_by,omitempty"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type TeamGovernanceState struct {
	TeamID              int64                  `json:"team_id"`
	MemberLimit         int                    `json:"member_limit"`
	Level               int                    `json:"level"`
	ReviewRequired      bool                   `json:"review_required"`
	MemberCount         int                    `json:"member_count"`
	EffectiveRecharge   float64                `json:"effective_recharge"`
	Spend7Days          float64                `json:"spend_7d"`
	TransferableBalance float64                `json:"transferable_balance"`
	Settings            TeamGovernanceSettings `json:"settings"`
}

type TeamCreationEligibility struct {
	RegistrationDays  int                    `json:"registration_days"`
	EffectiveRecharge float64                `json:"effective_recharge"`
	Eligible          bool                   `json:"eligible"`
	Settings          TeamGovernanceSettings `json:"settings"`
}

type AdminTeamSummary struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	OwnerID           int64     `json:"owner_id"`
	OwnerEmail        string    `json:"owner_email"`
	Status            string    `json:"status"`
	Level             int       `json:"level"`
	MemberCount       int       `json:"member_count"`
	MemberLimit       int       `json:"member_limit"`
	ReviewRequired    bool      `json:"review_required"`
	Balance           float64   `json:"balance"`
	EffectiveRecharge float64   `json:"effective_recharge"`
	Spend7Days        float64   `json:"spend_7d"`
	CreatedAt         time.Time `json:"created_at"`
}

type TeamAdminMember struct {
	ID                  int64   `json:"id"`
	Email               string  `json:"email"`
	Username            string  `json:"username"`
	Role                string  `json:"role"`
	Balance             float64 `json:"balance"`
	TransferableBalance float64 `json:"transferable_balance"`
	EffectiveRecharge   float64 `json:"effective_recharge"`
	Spend7Days          float64 `json:"spend_7d"`
}

type TeamFundLedgerEntry struct {
	ID                 int64     `json:"id"`
	TeamID             int64     `json:"team_id"`
	UserID             *int64    `json:"user_id,omitempty"`
	CounterpartyUserID *int64    `json:"counterparty_user_id,omitempty"`
	Action             string    `json:"action"`
	Amount             float64   `json:"amount"`
	Transferable       bool      `json:"transferable"`
	OperatorID         *int64    `json:"operator_id,omitempty"`
	Note               string    `json:"note"`
	CreatedAt          time.Time `json:"created_at"`
}

type AdminTeamDetail struct {
	Team         AdminTeamSummary      `json:"team"`
	Members      []TeamAdminMember     `json:"members"`
	Applications []TeamApplication     `json:"applications"`
	FundLedger   []TeamFundLedgerEntry `json:"fund_ledger"`
}

type TeamAdminStats struct {
	TotalTeams          int64 `json:"total_teams"`
	PendingApplications int64 `json:"pending_applications"`
}

type ReviewTeamApplicationInput struct {
	Approve      bool
	ReviewReason string
	Waive        bool
	TargetLimit  *int
}

type TeamGovernanceRepository interface {
	GetSettings(ctx context.Context) (*TeamGovernanceSettings, error)
	UpdateSettings(ctx context.Context, adminID int64, settings TeamGovernanceSettings) (*TeamGovernanceSettings, error)
	GetUserEligibility(ctx context.Context, userID int64) (registrationDays int, effectiveRecharge float64, err error)
	SubmitCreateApplication(ctx context.Context, userID int64, teamName, reason, additionalInfo string) (*TeamApplication, error)
	GetLatestCreateApplication(ctx context.Context, userID int64) (*TeamApplication, error)
	SubmitExpandApplication(ctx context.Context, ownerID, teamID int64, targetLimit int, reason string) (*TeamApplication, error)
	ListApplications(ctx context.Context, status string, page, pageSize int) ([]TeamApplication, int64, error)
	ReviewApplication(ctx context.Context, applicationID, adminID int64, input ReviewTeamApplicationInput, inviteCode string) (*TeamApplication, error)
	SubmitJoinRequest(ctx context.Context, userID int64, inviteCode, message string) (*TeamJoinRequest, error)
	ListJoinRequests(ctx context.Context, ownerID int64, status string) ([]TeamJoinRequest, error)
	ReviewJoinRequest(ctx context.Context, ownerID, requestID int64, approve bool, reason string) (*TeamJoinRequest, error)
	GetGovernanceState(ctx context.Context, userID, teamID int64) (*TeamGovernanceState, error)
	UpgradeTeam(ctx context.Context, ownerID, teamID int64) (*TeamGovernanceState, error)
	GetTransferableBalance(ctx context.Context, userID int64) (float64, error)
	DepositTeamFund(ctx context.Context, userID, teamID int64, amount float64) error
	TransferTeamBalance(ctx context.Context, ownerID, memberID, teamID int64, amount float64) error
	AllocateTeamFund(ctx context.Context, ownerID, memberID, teamID int64, amount float64) error
	GetAdminStats(ctx context.Context) (*TeamAdminStats, error)
	ListAdminTeams(ctx context.Context, search, status string, page, pageSize int) ([]AdminTeamSummary, int64, error)
	GetAdminTeam(ctx context.Context, teamID int64) (*AdminTeamDetail, error)
	SetTeamStatus(ctx context.Context, teamID int64, status string) error
	SetTeamMemberLimit(ctx context.Context, teamID int64, limit int) error
	MarkTeamReviewed(ctx context.Context, teamID int64) error
	AdminRemoveMember(ctx context.Context, teamID, memberID int64) error
}

func requirementSatisfied(metricRecharge, metricSpend float64, requirement TeamLevelRequirement) bool {
	rechargeOK := metricRecharge >= requirement.Recharge
	spendOK := metricSpend >= requirement.Spend7Days
	if requirement.Mode == "or" {
		return rechargeOK || spendOK
	}
	return rechargeOK && spendOK
}
