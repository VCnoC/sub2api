// Package admin 中的本文件提供团队治理管理接口。
package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	service service.TeamService
}

func NewTeamHandler(teamService service.TeamService) *TeamHandler {
	return &TeamHandler{service: teamService}
}

type reviewTeamApplicationRequest struct {
	Approve      bool   `json:"approve"`
	ReviewReason string `json:"review_reason" binding:"max=2000"`
	Waive        bool   `json:"waive"`
	TargetLimit  *int   `json:"target_limit"`
}

type teamStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active frozen disabled"`
}

type teamLimitRequest struct {
	MemberLimit int `json:"member_limit" binding:"required,gt=0"`
}

func (h *TeamHandler) Stats(c *gin.Context) {
	stats, err := h.service.GetAdminStats(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *TeamHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.ListAdminTeams(c.Request.Context(), c.Query("search"), c.Query("status"), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *TeamHandler) Get(c *gin.Context) {
	teamID, ok := positiveTeamID(c)
	if !ok {
		return
	}
	item, err := h.service.GetAdminTeam(c.Request.Context(), teamID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *TeamHandler) ListApplications(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.ListApplications(c.Request.Context(), c.Query("status"), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *TeamHandler) ReviewApplication(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	applicationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || applicationID <= 0 {
		response.BadRequest(c, "Invalid application ID")
		return
	}
	var req reviewTeamApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.service.ReviewApplication(c.Request.Context(), applicationID, subject.UserID, service.ReviewTeamApplicationInput{
		Approve: req.Approve, ReviewReason: req.ReviewReason, Waive: req.Waive, TargetLimit: req.TargetLimit,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *TeamHandler) GetSettings(c *gin.Context) {
	settings, err := h.service.GetSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, settings)
}

func (h *TeamHandler) UpdateSettings(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	var req service.TeamGovernanceSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	settings, err := h.service.UpdateSettings(c.Request.Context(), subject.UserID, req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, settings)
}

func (h *TeamHandler) SetStatus(c *gin.Context) {
	teamID, ok := positiveTeamID(c)
	if !ok {
		return
	}
	var req teamStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.Status == "frozen" {
		req.Status = service.StatusDisabled
	}
	if err := h.service.SetTeamStatus(c.Request.Context(), teamID, req.Status); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"status": req.Status})
}

func (h *TeamHandler) SetLimit(c *gin.Context) {
	teamID, ok := positiveTeamID(c)
	if !ok {
		return
	}
	var req teamLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.service.SetTeamMemberLimit(c.Request.Context(), teamID, req.MemberLimit); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"member_limit": req.MemberLimit})
}

func (h *TeamHandler) MarkReviewed(c *gin.Context) {
	teamID, ok := positiveTeamID(c)
	if !ok {
		return
	}
	if err := h.service.MarkTeamReviewed(c.Request.Context(), teamID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"review_required": false})
}

func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID, ok := positiveTeamID(c)
	if !ok {
		return
	}
	memberID, err := strconv.ParseInt(c.Param("member_id"), 10, 64)
	if err != nil || memberID <= 0 {
		response.BadRequest(c, "Invalid member ID")
		return
	}
	if err := h.service.AdminRemoveMember(c.Request.Context(), teamID, memberID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "member removed"})
}

func positiveTeamID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid team ID")
		return 0, false
	}
	return id, true
}
