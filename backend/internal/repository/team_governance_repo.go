// Package repository 中的本文件实现团队审批、指标和资金事务。
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type teamGovernanceRepository struct {
	db *sql.DB
}

func NewTeamGovernanceRepository(db *sql.DB) service.TeamGovernanceRepository {
	return &teamGovernanceRepository{db: db}
}

func (r *teamGovernanceRepository) GetSettings(ctx context.Context) (*service.TeamGovernanceSettings, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT configured, min_registration_days, min_total_recharge,
		       level_5_recharge, level_5_spend_7d, level_5_mode,
		       level_15_recharge, level_15_spend_7d, level_15_mode,
		       level_40_recharge, level_40_spend_7d, level_40_mode,
		       updated_by, updated_at
		FROM team_governance_settings WHERE id = 1`)
	return scanTeamSettings(row)
}

func (r *teamGovernanceRepository) UpdateSettings(ctx context.Context, adminID int64, settings service.TeamGovernanceSettings) (*service.TeamGovernanceSettings, error) {
	levels := normalizeLevelRequirements(settings.Levels)
	row := r.db.QueryRowContext(ctx, `
		UPDATE team_governance_settings SET
			configured = TRUE, min_registration_days = $1, min_total_recharge = $2,
			level_5_recharge = $3, level_5_spend_7d = $4, level_5_mode = $5,
			level_15_recharge = $6, level_15_spend_7d = $7, level_15_mode = $8,
			level_40_recharge = $9, level_40_spend_7d = $10, level_40_mode = $11,
			updated_by = $12, updated_at = NOW()
		WHERE id = 1
		RETURNING configured, min_registration_days, min_total_recharge,
			level_5_recharge, level_5_spend_7d, level_5_mode,
			level_15_recharge, level_15_spend_7d, level_15_mode,
			level_40_recharge, level_40_spend_7d, level_40_mode,
			updated_by, updated_at`,
		settings.MinRegistrationDays, settings.MinTotalRecharge,
		levels[5].Recharge, levels[5].Spend7Days, levels[5].Mode,
		levels[15].Recharge, levels[15].Spend7Days, levels[15].Mode,
		levels[40].Recharge, levels[40].Spend7Days, levels[40].Mode, adminID)
	return scanTeamSettings(row)
}

type teamRowScanner interface {
	Scan(dest ...any) error
}

func scanTeamSettings(row teamRowScanner) (*service.TeamGovernanceSettings, error) {
	settings := &service.TeamGovernanceSettings{}
	levels := []service.TeamLevelRequirement{{Limit: 5}, {Limit: 15}, {Limit: 40}}
	err := row.Scan(
		&settings.Configured, &settings.MinRegistrationDays, &settings.MinTotalRecharge,
		&levels[0].Recharge, &levels[0].Spend7Days, &levels[0].Mode,
		&levels[1].Recharge, &levels[1].Spend7Days, &levels[1].Mode,
		&levels[2].Recharge, &levels[2].Spend7Days, &levels[2].Mode,
		&settings.UpdatedBy, &settings.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	settings.Levels = levels
	return settings, nil
}

func normalizeLevelRequirements(levels []service.TeamLevelRequirement) map[int]service.TeamLevelRequirement {
	out := map[int]service.TeamLevelRequirement{
		5:  {Limit: 5, Mode: "and"},
		15: {Limit: 15, Mode: "and"},
		40: {Limit: 40, Mode: "and"},
	}
	for _, level := range levels {
		if _, ok := out[level.Limit]; !ok {
			continue
		}
		if level.Mode != "or" {
			level.Mode = "and"
		}
		out[level.Limit] = level
	}
	return out
}

func (r *teamGovernanceRepository) GetUserEligibility(ctx context.Context, userID int64) (int, float64, error) {
	var days int
	var recharge float64
	err := r.db.QueryRowContext(ctx, `
		SELECT GREATEST(0, FLOOR(EXTRACT(EPOCH FROM (NOW() - u.created_at)) / 86400)::INTEGER),
		       COALESCE((
		           SELECT SUM(GREATEST(po.amount - po.refund_amount, 0))
		           FROM payment_orders po
		           WHERE po.user_id = u.id AND po.order_type = 'balance'
		             AND po.status IN ('PAID', 'RECHARGING', 'COMPLETED', 'PARTIALLY_REFUNDED')
		       ), 0) + COALESCE((
		           SELECT SUM(rc.value)
		           FROM redeem_codes rc
		           WHERE rc.used_by = u.id AND rc.status = 'used' AND rc.type = 'balance' AND rc.value > 0
		             AND COALESCE(rc.notes, '') NOT LIKE '[lottery] %'
		             AND NOT EXISTS (SELECT 1 FROM payment_orders po WHERE po.recharge_code = rc.code)
		       ), 0)
		FROM users u WHERE u.id = $1 AND u.deleted_at IS NULL`, userID).Scan(&days, &recharge)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, service.ErrUserNotFound
	}
	return days, recharge, err
}

func (r *teamGovernanceRepository) SubmitCreateApplication(ctx context.Context, userID int64, teamName, reason, additionalInfo string) (*service.TeamApplication, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO team_applications (application_type, applicant_id, team_name, reason, additional_info)
		SELECT 'create', u.id, $2, $3, $4 FROM users u
		WHERE u.id = $1 AND u.deleted_at IS NULL AND u.team_id IS NULL
		RETURNING id, application_type, applicant_id, team_id, team_name, target_limit, reason,
		          additional_info, status, review_reason, waived, reviewer_id, reviewed_at,
		          created_team_id, created_at, updated_at`, userID, teamName, reason, additionalInfo)
	application, err := scanTeamApplication(row)
	if isTeamUniqueViolation(err) {
		return nil, service.ErrTeamApplicationPending
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrAlreadyInTeam
	}
	return application, err
}

func (r *teamGovernanceRepository) GetLatestCreateApplication(ctx context.Context, userID int64) (*service.TeamApplication, error) {
	row := r.db.QueryRowContext(ctx, applicationSelect+`
		WHERE a.application_type = 'create' AND a.applicant_id = $1
		ORDER BY a.created_at DESC LIMIT 1`, userID)
	application, err := scanTeamApplicationWithEmail(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return application, err
}

func (r *teamGovernanceRepository) SubmitExpandApplication(ctx context.Context, ownerID, teamID int64, targetLimit int, reason string) (*service.TeamApplication, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO team_applications (application_type, applicant_id, team_id, target_limit, reason)
		SELECT 'expand', $1, t.id, $3, $4 FROM teams t
		WHERE t.id = $2 AND t.owner_id = $1 AND t.status = 'active' AND NOT t.review_required
		RETURNING id, application_type, applicant_id, team_id, team_name, target_limit, reason,
		          additional_info, status, review_reason, waived, reviewer_id, reviewed_at,
		          created_team_id, created_at, updated_at`, ownerID, teamID, targetLimit, reason)
	application, err := scanTeamApplication(row)
	if isTeamUniqueViolation(err) {
		return nil, service.ErrTeamApplicationPending
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrTeamReviewRequired
	}
	return application, err
}

var applicationSelect = `
	SELECT a.id, a.application_type, a.applicant_id, u.email,
	       GREATEST(0, FLOOR(EXTRACT(EPOCH FROM (NOW() - u.created_at)) / 86400)::INTEGER),
	       ` + userRechargeSQL("u.id") + `,
	       a.team_id, a.team_name,
	       a.target_limit, a.reason, a.additional_info, a.status, a.review_reason, a.waived,
	       a.reviewer_id, a.reviewed_at, a.created_team_id, a.created_at, a.updated_at
	FROM team_applications a JOIN users u ON u.id = a.applicant_id `

func (r *teamGovernanceRepository) ListApplications(ctx context.Context, status string, page, pageSize int) ([]service.TeamApplication, int64, error) {
	where := "WHERE ($1 = '' OR a.status = $1)"
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM team_applications a `+where, status).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx, applicationSelect+where+` ORDER BY a.created_at DESC LIMIT $2 OFFSET $3`, status, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	applications := make([]service.TeamApplication, 0)
	for rows.Next() {
		application, scanErr := scanTeamApplicationWithEmail(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		applications = append(applications, *application)
	}
	return applications, total, rows.Err()
}

func (r *teamGovernanceRepository) ReviewApplication(ctx context.Context, applicationID, adminID int64, input service.ReviewTeamApplicationInput, inviteCode string) (*service.TeamApplication, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, applicationSelect+` WHERE a.id = $1 FOR UPDATE OF a`, applicationID)
	application, err := scanTeamApplicationWithEmail(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrTeamApplicationNotFound
	}
	if err != nil {
		return nil, err
	}
	if application.Status != service.TeamRequestPending {
		return nil, service.ErrTeamApplicationResolved
	}
	if input.Approve && application.ApplicationType == service.TeamApplicationCreate {
		days, recharge, eligibilityErr := r.GetUserEligibility(ctx, application.ApplicantID)
		if eligibilityErr != nil {
			return nil, eligibilityErr
		}
		settings, settingsErr := r.GetSettings(ctx)
		if settingsErr != nil {
			return nil, settingsErr
		}
		if (!settings.Configured || days < settings.MinRegistrationDays || recharge < settings.MinTotalRecharge) && !input.Waive {
			return nil, service.ErrTeamThresholdNotMet
		}
	}

	status := service.TeamRequestRejected
	if input.Approve {
		status = service.TeamRequestApproved
		if application.ApplicationType == service.TeamApplicationCreate {
			if err := r.approveCreateApplication(ctx, tx, application, inviteCode); err != nil {
				return nil, err
			}
		} else if err := r.approveExpandApplication(ctx, tx, application, input.TargetLimit); err != nil {
			return nil, err
		}
	}
	if input.Waive && strings.TrimSpace(input.ReviewReason) == "" {
		return nil, service.ErrTeamWaiverReasonRequired
	}
	if err := tx.QueryRowContext(ctx, `
		UPDATE team_applications SET status = $2, review_reason = $3, waived = $4,
		       reviewer_id = $5, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING reviewed_at, updated_at`, applicationID, status, input.ReviewReason, input.Waive, adminID).
		Scan(&application.ReviewedAt, &application.UpdatedAt); err != nil {
		return nil, err
	}
	application.Status = status
	application.ReviewReason = input.ReviewReason
	application.Waived = input.Waive
	application.ReviewerID = &adminID
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return application, nil
}

func (r *teamGovernanceRepository) approveCreateApplication(ctx context.Context, tx *sql.Tx, application *service.TeamApplication, inviteCode string) error {
	var teamID int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO teams (name, owner_id, invite_code, status, balance, member_limit, level, review_required)
		SELECT $2, u.id, $3, 'active', 0, 5, 5, FALSE FROM users u
		WHERE u.id = $1 AND u.deleted_at IS NULL AND u.team_id IS NULL
		RETURNING id`, application.ApplicantID, application.TeamName, inviteCode).Scan(&teamID)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrAlreadyInTeam
	}
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE users SET team_id = $2, team_role = 'owner', updated_at = NOW() WHERE id = $1 AND team_id IS NULL`, application.ApplicantID, teamID)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return service.ErrAlreadyInTeam
	}
	if _, err := tx.ExecContext(ctx, `UPDATE team_applications SET created_team_id = $2 WHERE id = $1`, application.ID, teamID); err != nil {
		return err
	}
	application.CreatedTeamID = &teamID
	return nil
}

func (r *teamGovernanceRepository) approveExpandApplication(ctx context.Context, tx *sql.Tx, application *service.TeamApplication, override *int) error {
	target := 0
	if application.TargetLimit != nil {
		target = *application.TargetLimit
	}
	if override != nil {
		target = *override
	}
	if target <= 40 {
		return service.ErrInvalidTeamLimit
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE teams SET member_limit = $2, level = GREATEST(level, 40), review_required = FALSE, updated_at = NOW()
		WHERE id = $1`, *application.TeamID, target)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return service.ErrTeamNotFound
	}
	application.TargetLimit = &target
	_, err = tx.ExecContext(ctx, `UPDATE team_applications SET target_limit = $2 WHERE id = $1`, application.ID, target)
	return err
}

func scanTeamApplication(row teamRowScanner) (*service.TeamApplication, error) {
	a := &service.TeamApplication{}
	err := row.Scan(&a.ID, &a.ApplicationType, &a.ApplicantID, &a.TeamID, &a.TeamName, &a.TargetLimit,
		&a.Reason, &a.AdditionalInfo, &a.Status, &a.ReviewReason, &a.Waived, &a.ReviewerID,
		&a.ReviewedAt, &a.CreatedTeamID, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

func scanTeamApplicationWithEmail(row teamRowScanner) (*service.TeamApplication, error) {
	a := &service.TeamApplication{}
	err := row.Scan(&a.ID, &a.ApplicationType, &a.ApplicantID, &a.ApplicantEmail, &a.RegistrationDays, &a.EffectiveRecharge, &a.TeamID, &a.TeamName,
		&a.TargetLimit, &a.Reason, &a.AdditionalInfo, &a.Status, &a.ReviewReason, &a.Waived,
		&a.ReviewerID, &a.ReviewedAt, &a.CreatedTeamID, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

func (r *teamGovernanceRepository) SubmitJoinRequest(ctx context.Context, userID int64, inviteCode, message string) (*service.TeamJoinRequest, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO team_join_requests (team_id, applicant_id, message)
		SELECT t.id, u.id, $3 FROM teams t CROSS JOIN users u
		WHERE t.invite_code = $2 AND t.status = 'active' AND u.id = $1 AND u.deleted_at IS NULL AND u.team_id IS NULL
		RETURNING id, team_id, applicant_id, message, status, review_reason, reviewed_by, reviewed_at, created_at, updated_at`, userID, inviteCode, message)
	request := &service.TeamJoinRequest{}
	err := row.Scan(&request.ID, &request.TeamID, &request.ApplicantID, &request.Message, &request.Status,
		&request.ReviewReason, &request.ReviewedBy, &request.ReviewedAt, &request.CreatedAt, &request.UpdatedAt)
	if isTeamUniqueViolation(err) {
		return nil, service.ErrTeamJoinPending
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrInviteCodeInvalid
	}
	return request, err
}

const joinRequestSelect = `
	SELECT r.id, r.team_id, t.name, r.applicant_id, u.email, r.message, r.status,
	       r.review_reason, r.reviewed_by, r.reviewed_at, r.created_at, r.updated_at
	FROM team_join_requests r
	JOIN teams t ON t.id = r.team_id
	JOIN users u ON u.id = r.applicant_id `

func (r *teamGovernanceRepository) ListJoinRequests(ctx context.Context, ownerID int64, status string) ([]service.TeamJoinRequest, error) {
	rows, err := r.db.QueryContext(ctx, joinRequestSelect+`
		WHERE t.owner_id = $1 AND ($2 = '' OR r.status = $2)
		ORDER BY r.created_at DESC`, ownerID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	requests := make([]service.TeamJoinRequest, 0)
	for rows.Next() {
		request, scanErr := scanJoinRequest(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		requests = append(requests, *request)
	}
	return requests, rows.Err()
}

func (r *teamGovernanceRepository) ReviewJoinRequest(ctx context.Context, ownerID, requestID int64, approve bool, reason string) (*service.TeamJoinRequest, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	request, err := scanJoinRequest(tx.QueryRowContext(ctx, joinRequestSelect+`
		WHERE r.id = $1 AND t.owner_id = $2 FOR UPDATE OF r, t`, requestID, ownerID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrTeamJoinNotFound
	}
	if err != nil {
		return nil, err
	}
	if request.Status != service.TeamRequestPending {
		return nil, service.ErrTeamJoinResolved
	}
	status := service.TeamRequestRejected
	if approve {
		var teamStatus string
		var memberLimit, memberCount int
		if err := tx.QueryRowContext(ctx, `SELECT status, member_limit FROM teams WHERE id = $1 FOR UPDATE`, request.TeamID).Scan(&teamStatus, &memberLimit); err != nil {
			return nil, err
		}
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*)::INTEGER FROM users WHERE team_id = $1 AND deleted_at IS NULL`, request.TeamID).Scan(&memberCount); err != nil {
			return nil, err
		}
		if teamStatus != service.StatusActive {
			return nil, service.ErrTeamFrozen
		}
		if memberCount >= memberLimit {
			return nil, service.ErrTeamFull
		}
		result, err := tx.ExecContext(ctx, `UPDATE users SET team_id = $2, team_role = 'member', updated_at = NOW() WHERE id = $1 AND team_id IS NULL AND deleted_at IS NULL`, request.ApplicantID, request.TeamID)
		if err != nil {
			return nil, err
		}
		if n, _ := result.RowsAffected(); n != 1 {
			return nil, service.ErrAlreadyInTeam
		}
		status = service.TeamRequestApproved
	}
	if err := tx.QueryRowContext(ctx, `
		UPDATE team_join_requests SET status = $2, review_reason = $3, reviewed_by = $4,
		       reviewed_at = NOW(), updated_at = NOW() WHERE id = $1
		RETURNING reviewed_at, updated_at`, requestID, status, reason, ownerID).Scan(&request.ReviewedAt, &request.UpdatedAt); err != nil {
		return nil, err
	}
	request.Status = status
	request.ReviewReason = reason
	request.ReviewedBy = &ownerID
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return request, nil
}

func scanJoinRequest(row teamRowScanner) (*service.TeamJoinRequest, error) {
	r := &service.TeamJoinRequest{}
	err := row.Scan(&r.ID, &r.TeamID, &r.TeamName, &r.ApplicantID, &r.ApplicantEmail, &r.Message,
		&r.Status, &r.ReviewReason, &r.ReviewedBy, &r.ReviewedAt, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

func (r *teamGovernanceRepository) GetGovernanceState(ctx context.Context, userID, teamID int64) (*service.TeamGovernanceState, error) {
	settings, err := r.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	state := &service.TeamGovernanceState{TeamID: teamID, Settings: *settings}
	err = r.db.QueryRowContext(ctx, `
		SELECT t.member_limit, t.level, t.review_required,
		       (SELECT COUNT(*)::INTEGER FROM users m WHERE m.team_id = t.id AND m.deleted_at IS NULL),
		       `+teamRechargeSQL("t.id")+`, `+teamSpendSQL("t.id")+`,
		       COALESCE((SELECT amount FROM team_transferable_balances WHERE user_id = $1), 0)
		FROM teams t WHERE t.id = $2`, userID, teamID).Scan(
		&state.MemberLimit, &state.Level, &state.ReviewRequired, &state.MemberCount,
		&state.EffectiveRecharge, &state.Spend7Days, &state.TransferableBalance)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrTeamNotFound
	}
	return state, err
}

func (r *teamGovernanceRepository) UpgradeTeam(ctx context.Context, ownerID, teamID int64) (*service.TeamGovernanceState, error) {
	state, err := r.GetGovernanceState(ctx, ownerID, teamID)
	if err != nil {
		return nil, err
	}
	if state.ReviewRequired {
		return nil, service.ErrTeamReviewRequired
	}
	if !state.Settings.Configured {
		return nil, service.ErrTeamUpgradeUnavailable
	}
	var ownerOK bool
	var teamStatus string
	if err := r.db.QueryRowContext(ctx, `SELECT owner_id = $2, status FROM teams WHERE id = $1`, teamID, ownerID).Scan(&ownerOK, &teamStatus); err != nil {
		return nil, err
	}
	if !ownerOK {
		return nil, service.ErrNotTeamOwner
	}
	if teamStatus != service.StatusActive {
		return nil, service.ErrTeamFrozen
	}
	target := state.Level
	for _, requirement := range state.Settings.Levels {
		if requirement.Limit > target && requirementSatisfiedForRepo(state.EffectiveRecharge, state.Spend7Days, requirement) {
			target = requirement.Limit
		}
	}
	if target <= state.Level {
		return nil, service.ErrTeamUpgradeUnavailable
	}
	if _, err := r.db.ExecContext(ctx, `UPDATE teams SET level = $2, member_limit = GREATEST(member_limit, $2), updated_at = NOW() WHERE id = $1 AND owner_id = $3`, teamID, target, ownerID); err != nil {
		return nil, err
	}
	return r.GetGovernanceState(ctx, ownerID, teamID)
}

func requirementSatisfiedForRepo(recharge, spend float64, requirement service.TeamLevelRequirement) bool {
	if requirement.Mode == "or" {
		return recharge >= requirement.Recharge || spend >= requirement.Spend7Days
	}
	return recharge >= requirement.Recharge && spend >= requirement.Spend7Days
}

func (r *teamGovernanceRepository) GetTransferableBalance(ctx context.Context, userID int64) (float64, error) {
	var amount float64
	err := r.db.QueryRowContext(ctx, `SELECT LEAST(tb.amount, u.balance) FROM users u LEFT JOIN team_transferable_balances tb ON tb.user_id = u.id WHERE u.id = $1`, userID).Scan(&amount)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, service.ErrUserNotFound
	}
	return amount, err
}

func (r *teamGovernanceRepository) DepositTeamFund(ctx context.Context, userID, teamID int64, amount float64) error {
	return r.withFundTx(ctx, func(tx *sql.Tx) error {
		if err := lockActiveTeam(ctx, tx, teamID); err != nil {
			return err
		}
		if err := lockSpendableBalance(ctx, tx, userID, amount); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `UPDATE users SET balance = balance - $2, updated_at = NOW() WHERE id = $1 AND team_id = $3 AND balance >= $2`, userID, amount, teamID)
		if err != nil {
			return err
		}
		if n, _ := result.RowsAffected(); n != 1 {
			return service.ErrInsufficientTeamBalance
		}
		result, err = tx.ExecContext(ctx, `UPDATE teams SET balance = balance + $2, updated_at = NOW() WHERE id = $1 AND status = 'active'`, teamID, amount)
		if err != nil {
			return err
		}
		if n, _ := result.RowsAffected(); n != 1 {
			return service.ErrTeamFrozen
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO team_fund_ledger (team_id, user_id, action, amount, operator_id) VALUES ($1, $2, 'deposit', $3, $2)`, teamID, userID, amount)
		return err
	})
}

func (r *teamGovernanceRepository) TransferTeamBalance(ctx context.Context, ownerID, memberID, teamID int64, amount float64) error {
	return r.withFundTx(ctx, func(tx *sql.Tx) error {
		if err := lockActiveTeamOwner(ctx, tx, teamID, ownerID); err != nil {
			return err
		}
		if err := lockSpendableBalance(ctx, tx, ownerID, amount); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `UPDATE users SET balance = balance - $2, updated_at = NOW() WHERE id = $1 AND balance >= $2`, ownerID, amount)
		if err != nil {
			return err
		}
		if n, _ := result.RowsAffected(); n != 1 {
			return service.ErrInsufficientTeamBalance
		}
		result, err = tx.ExecContext(ctx, `UPDATE users SET balance = balance + $3, updated_at = NOW() WHERE id = $1 AND team_id = $2 AND team_role = 'member'`, memberID, teamID, amount)
		if err != nil {
			return err
		}
		if n, _ := result.RowsAffected(); n != 1 {
			return service.ErrTeamMemberNotFound
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO team_fund_ledger (team_id, user_id, counterparty_user_id, action, amount, operator_id) VALUES ($1, $2, $3, 'transfer', $4, $2)`, teamID, ownerID, memberID, amount)
		return err
	})
}

func (r *teamGovernanceRepository) AllocateTeamFund(ctx context.Context, ownerID, memberID, teamID int64, amount float64) error {
	return r.withFundTx(ctx, func(tx *sql.Tx) error {
		if err := lockActiveTeamOwner(ctx, tx, teamID, ownerID); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `UPDATE teams SET balance = balance - $2, updated_at = NOW() WHERE id = $1 AND balance >= $2`, teamID, amount)
		if err != nil {
			return err
		}
		if n, _ := result.RowsAffected(); n != 1 {
			return service.ErrInsufficientTeamFund
		}
		result, err = tx.ExecContext(ctx, `UPDATE users SET balance = balance + $3, updated_at = NOW() WHERE id = $1 AND team_id = $2`, memberID, teamID, amount)
		if err != nil {
			return err
		}
		if n, _ := result.RowsAffected(); n != 1 {
			return service.ErrTeamMemberNotFound
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO team_fund_ledger (team_id, user_id, action, amount, operator_id) VALUES ($1, $2, 'allocate', $3, $4)`, teamID, memberID, amount, ownerID)
		return err
	})
}

func (r *teamGovernanceRepository) withFundTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func lockSpendableBalance(ctx context.Context, tx *sql.Tx, userID int64, amount float64) error {
	var balance, transferable float64
	if err := tx.QueryRowContext(ctx, `SELECT balance FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.ErrUserNotFound
		}
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT amount FROM team_transferable_balances WHERE user_id = $1 FOR UPDATE`, userID).Scan(&transferable); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.ErrInsufficientTransferableBalance
		}
		return err
	}
	if balance < amount {
		return service.ErrInsufficientTeamBalance
	}
	if transferable < amount {
		return service.ErrInsufficientTransferableBalance
	}
	return nil
}

func lockActiveTeam(ctx context.Context, tx *sql.Tx, teamID int64) error {
	var status string
	err := tx.QueryRowContext(ctx, `SELECT status FROM teams WHERE id = $1 FOR UPDATE`, teamID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrTeamNotFound
	}
	if err != nil {
		return err
	}
	if status != service.StatusActive {
		return service.ErrTeamFrozen
	}
	return nil
}

func lockActiveTeamOwner(ctx context.Context, tx *sql.Tx, teamID, ownerID int64) error {
	var status string
	err := tx.QueryRowContext(ctx, `SELECT status FROM teams WHERE id = $1 AND owner_id = $2 FOR UPDATE`, teamID, ownerID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrNotTeamOwner
	}
	if err != nil {
		return err
	}
	if status != service.StatusActive {
		return service.ErrTeamFrozen
	}
	return nil
}

func (r *teamGovernanceRepository) GetAdminStats(ctx context.Context) (*service.TeamAdminStats, error) {
	stats := &service.TeamAdminStats{}
	err := r.db.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM teams), (SELECT COUNT(*) FROM team_applications WHERE status = 'pending')`).Scan(&stats.TotalTeams, &stats.PendingApplications)
	return stats, err
}

const adminTeamSelect = `
	SELECT t.id, t.name, t.owner_id, owner.email, t.status, t.level,
	       (SELECT COUNT(*)::INTEGER FROM users m WHERE m.team_id = t.id AND m.deleted_at IS NULL),
	       t.member_limit, t.review_required, t.balance,
	       ` + `TEAM_RECHARGE` + `, ` + `TEAM_SPEND` + `, t.created_at
	FROM teams t JOIN users owner ON owner.id = t.owner_id `

func (r *teamGovernanceRepository) ListAdminTeams(ctx context.Context, search, status string, page, pageSize int) ([]service.AdminTeamSummary, int64, error) {
	where := `WHERE ($1 = '' OR t.name ILIKE '%' || $1 || '%' OR owner.email ILIKE '%' || $1 || '%') AND ($2 = '' OR t.status = $2)`
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM teams t JOIN users owner ON owner.id = t.owner_id `+where, search, status).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := strings.ReplaceAll(adminTeamSelect, "TEAM_RECHARGE", teamRechargeSQL("t.id"))
	query = strings.ReplaceAll(query, "TEAM_SPEND", teamSpendSQL("t.id"))
	rows, err := r.db.QueryContext(ctx, query+where+` ORDER BY t.created_at DESC LIMIT $3 OFFSET $4`, search, status, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	teams := make([]service.AdminTeamSummary, 0)
	for rows.Next() {
		team, scanErr := scanAdminTeam(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		teams = append(teams, *team)
	}
	return teams, total, rows.Err()
}

func (r *teamGovernanceRepository) GetAdminTeam(ctx context.Context, teamID int64) (*service.AdminTeamDetail, error) {
	query := strings.ReplaceAll(adminTeamSelect, "TEAM_RECHARGE", teamRechargeSQL("t.id"))
	query = strings.ReplaceAll(query, "TEAM_SPEND", teamSpendSQL("t.id"))
	team, err := scanAdminTeam(r.db.QueryRowContext(ctx, query+` WHERE t.id = $1`, teamID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrTeamNotFound
	}
	if err != nil {
		return nil, err
	}
	detail := &service.AdminTeamDetail{Team: *team}
	if detail.Members, err = r.listAdminMembers(ctx, teamID); err != nil {
		return nil, err
	}
	if detail.Applications, err = r.listTeamApplications(ctx, teamID); err != nil {
		return nil, err
	}
	if detail.FundLedger, err = r.listFundLedger(ctx, teamID); err != nil {
		return nil, err
	}
	return detail, nil
}

func scanAdminTeam(row teamRowScanner) (*service.AdminTeamSummary, error) {
	t := &service.AdminTeamSummary{}
	err := row.Scan(&t.ID, &t.Name, &t.OwnerID, &t.OwnerEmail, &t.Status, &t.Level, &t.MemberCount,
		&t.MemberLimit, &t.ReviewRequired, &t.Balance, &t.EffectiveRecharge, &t.Spend7Days, &t.CreatedAt)
	return t, err
}

func (r *teamGovernanceRepository) listAdminMembers(ctx context.Context, teamID int64) ([]service.TeamAdminMember, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.email, u.username, u.team_role, u.balance,
		       LEAST(COALESCE(tb.amount, 0), u.balance),
		       `+userRechargeSQL("u.id")+`,
		       COALESCE((SELECT SUM(actual_cost) FROM usage_logs l WHERE l.user_id = u.id AND l.created_at >= NOW() - INTERVAL '7 days'), 0)
		FROM users u LEFT JOIN team_transferable_balances tb ON tb.user_id = u.id
		WHERE u.team_id = $1 AND u.deleted_at IS NULL ORDER BY u.team_role DESC, u.created_at`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	members := make([]service.TeamAdminMember, 0)
	for rows.Next() {
		var member service.TeamAdminMember
		if err := rows.Scan(&member.ID, &member.Email, &member.Username, &member.Role, &member.Balance,
			&member.TransferableBalance, &member.EffectiveRecharge, &member.Spend7Days); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

func (r *teamGovernanceRepository) listTeamApplications(ctx context.Context, teamID int64) ([]service.TeamApplication, error) {
	rows, err := r.db.QueryContext(ctx, applicationSelect+` WHERE a.team_id = $1 OR a.created_team_id = $1 ORDER BY a.created_at DESC`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	applications := make([]service.TeamApplication, 0)
	for rows.Next() {
		a, scanErr := scanTeamApplicationWithEmail(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		applications = append(applications, *a)
	}
	return applications, rows.Err()
}

func (r *teamGovernanceRepository) listFundLedger(ctx context.Context, teamID int64) ([]service.TeamFundLedgerEntry, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, team_id, user_id, counterparty_user_id, action, amount, transferable, operator_id, note, created_at FROM team_fund_ledger WHERE team_id = $1 ORDER BY created_at DESC LIMIT 200`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	entries := make([]service.TeamFundLedgerEntry, 0)
	for rows.Next() {
		var entry service.TeamFundLedgerEntry
		if err := rows.Scan(&entry.ID, &entry.TeamID, &entry.UserID, &entry.CounterpartyUserID, &entry.Action, &entry.Amount, &entry.Transferable, &entry.OperatorID, &entry.Note, &entry.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (r *teamGovernanceRepository) SetTeamStatus(ctx context.Context, teamID int64, status string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE teams SET status = $2, updated_at = NOW() WHERE id = $1`, teamID, status)
	return expectOneTeam(result, err)
}

func (r *teamGovernanceRepository) SetTeamMemberLimit(ctx context.Context, teamID int64, limit int) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE teams SET member_limit = $2, level = CASE WHEN $2 >= 40 THEN GREATEST(level, 40) WHEN $2 >= 15 THEN GREATEST(level, 15) ELSE level END, updated_at = NOW()
		WHERE id = $1 AND $2 >= (SELECT COUNT(*) FROM users WHERE team_id = $1 AND deleted_at IS NULL)`, teamID, limit)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return service.ErrInvalidTeamLimit
	}
	return nil
}

func (r *teamGovernanceRepository) MarkTeamReviewed(ctx context.Context, teamID int64) error {
	result, err := r.db.ExecContext(ctx, `UPDATE teams SET review_required = FALSE, updated_at = NOW() WHERE id = $1`, teamID)
	return expectOneTeam(result, err)
}

func (r *teamGovernanceRepository) AdminRemoveMember(ctx context.Context, teamID, memberID int64) error {
	result, err := r.db.ExecContext(ctx, `UPDATE users SET team_id = NULL, team_role = '', updated_at = NOW() WHERE id = $2 AND team_id = $1 AND team_role <> 'owner'`, teamID, memberID)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return service.ErrTeamMemberNotFound
	}
	return nil
}

func expectOneTeam(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return service.ErrTeamNotFound
	}
	return nil
}

func userRechargeSQL(userExpr string) string {
	return fmt.Sprintf(`COALESCE((
		SELECT SUM(GREATEST(po.amount - po.refund_amount, 0)) FROM payment_orders po
		WHERE po.user_id = %s AND po.order_type = 'balance'
		  AND po.status IN ('PAID', 'RECHARGING', 'COMPLETED', 'PARTIALLY_REFUNDED')
	), 0) + COALESCE((
		SELECT SUM(rc.value) FROM redeem_codes rc
		WHERE rc.used_by = %s AND rc.status = 'used' AND rc.type = 'balance' AND rc.value > 0
		  AND COALESCE(rc.notes, '') NOT LIKE '[lottery] %%'
		  AND NOT EXISTS (SELECT 1 FROM payment_orders po WHERE po.recharge_code = rc.code)
	), 0)`, userExpr, userExpr)
}

func teamRechargeSQL(teamExpr string) string {
	return fmt.Sprintf(`COALESCE((SELECT SUM(%s) FROM users member WHERE member.team_id = %s AND member.deleted_at IS NULL), 0)`, userRechargeSQL("member.id"), teamExpr)
}

func teamSpendSQL(teamExpr string) string {
	return fmt.Sprintf(`COALESCE((
		SELECT SUM(l.actual_cost) FROM usage_logs l JOIN users member ON member.id = l.user_id
		WHERE member.team_id = %s AND member.deleted_at IS NULL AND l.created_at >= NOW() - INTERVAL '7 days'
	), 0)`, teamExpr)
}

func isTeamUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

var _ service.TeamGovernanceRepository = (*teamGovernanceRepository)(nil)
