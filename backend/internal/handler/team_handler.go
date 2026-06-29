package handler

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// TeamHandler handles team-related requests
type TeamHandler struct {
	teamService service.TeamService
}

// NewTeamHandler creates a new TeamHandler
func NewTeamHandler(teamService service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

// CreateTeamRequest represents the create team request payload
type CreateTeamRequest struct {
	Name string `json:"name" binding:"required,max=100"`
}

// JoinTeamRequest represents the join team request payload
type JoinTeamRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

// TransferBalanceRequest represents the transfer balance request payload
type TransferBalanceRequest struct {
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Password string  `json:"password" binding:"required"`
}

// TeamResponse represents the team response
type TeamResponse struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	OwnerID    int64  `json:"owner_id"`
	InviteCode string `json:"invite_code,omitempty"`
	Status     string `json:"status"`
	Role       string `json:"role,omitempty"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

// TeamMemberResponse represents a team member response
type TeamMemberResponse struct {
	ID         int64   `json:"id"`
	Email      string  `json:"email"`
	Username   string  `json:"username"`
	Role       string  `json:"role"`
	Balance    float64 `json:"balance"`
	TotalUsage float64 `json:"total_usage"`
	CreatedAt  int64   `json:"created_at"`
}

// CreateTeam creates a new team
// POST /api/v1/team
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	team, err := h.teamService.CreateTeam(c.Request.Context(), subject.UserID, req.Name)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, teamToResponse(team, service.TeamRoleOwner))
}

// GetMyTeam returns the current user's team information
// GET /api/v1/team
func (h *TeamHandler) GetMyTeam(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	team, role, err := h.teamService.GetMyTeam(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if team == nil {
		response.Success(c, nil)
		return
	}

	response.Success(c, teamToResponse(team, role))
}

// RefreshInviteCode refreshes the team invite code
// POST /api/v1/team/invite-code
func (h *TeamHandler) RefreshInviteCode(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	code, err := h.teamService.RefreshInviteCode(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"invite_code": code})
}

// JoinTeam joins a team by invite code
// POST /api/v1/team/join
func (h *TeamHandler) JoinTeam(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req JoinTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.teamService.JoinTeamByCode(c.Request.Context(), subject.UserID, req.InviteCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "joined team successfully"})
}

// LeaveTeam allows a member to leave their team
// POST /api/v1/team/leave
func (h *TeamHandler) LeaveTeam(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	if err := h.teamService.LeaveTeam(c.Request.Context(), subject.UserID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "left team successfully"})
}

// RemoveMember removes a member from the team
// DELETE /api/v1/team/members/:id
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	memberID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || memberID <= 0 {
		response.BadRequest(c, "Invalid member ID")
		return
	}

	if err := h.teamService.RemoveMember(c.Request.Context(), subject.UserID, memberID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "member removed successfully"})
}

// ListMembers returns the team members with balance and usage
// GET /api/v1/team/members
func (h *TeamHandler) ListMembers(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)

	members, total, err := h.teamService.ListMembers(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]TeamMemberResponse, 0, len(members))
	for i := range members {
		out = append(out, teamMemberToResponse(&members[i]))
	}

	response.Paginated(c, out, total, page, pageSize)
}

// GetMemberUsage returns usage logs for a team member within a date range.
// GET /api/v1/user/team/members/:id/usage
func (h *TeamHandler) GetMemberUsage(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	memberID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || memberID <= 0 {
		response.BadRequest(c, "Invalid member ID")
		return
	}

	startDate := c.DefaultQuery("start_date", "")
	endDate := c.DefaultQuery("end_date", "")
	if startDate == "" || endDate == "" {
		response.BadRequest(c, "start_date and end_date are required")
		return
	}

	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		response.BadRequest(c, "Invalid start_date format, use YYYY-MM-DD")
		return
	}
	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		response.BadRequest(c, "Invalid end_date format, use YYYY-MM-DD")
		return
	}
	endTime = endTime.AddDate(0, 0, 1)

	page, pageSize := response.ParsePagination(c)

	logs, result, err := h.teamService.ListMemberUsage(c.Request.Context(), subject.UserID, memberID, startTime, endTime, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]*dto.UsageLog, 0, len(logs))
	for i := range logs {
		out = append(out, dto.UsageLogFromService(&logs[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

// TransferBalance transfers balance from owner to a member
// POST /api/v1/team/members/:id/transfer
func (h *TeamHandler) TransferBalance(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	memberID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || memberID <= 0 {
		response.BadRequest(c, "Invalid member ID")
		return
	}

	var req TransferBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.teamService.TransferBalance(c.Request.Context(), subject.UserID, memberID, req.Amount, req.Password); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "balance transferred successfully"})
}

func teamToResponse(team *service.Team, role string) *TeamResponse {
	if team == nil {
		return nil
	}
	return &TeamResponse{
		ID:         team.ID,
		Name:       team.Name,
		OwnerID:    team.OwnerID,
		InviteCode: team.InviteCode,
		Status:     team.Status,
		Role:       role,
		CreatedAt:  team.CreatedAt.Unix(),
		UpdatedAt:  team.UpdatedAt.Unix(),
	}
}

func teamMemberToResponse(member *service.TeamMember) TeamMemberResponse {
	return TeamMemberResponse{
		ID:         member.ID,
		Email:      member.Email,
		Username:   member.Username,
		Role:       member.TeamRole,
		Balance:    member.Balance,
		TotalUsage: member.TotalUsage,
		CreatedAt:  member.CreatedAt.Unix(),
	}
}
