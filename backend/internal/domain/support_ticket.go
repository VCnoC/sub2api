// Package domain 定义跨层共享的业务常量。
package domain

const (
	TicketStatusPendingAdmin = "pending_admin"
	TicketStatusPendingUser  = "pending_user"
	TicketStatusClosed       = "closed"

	TicketCategoryAccount = "account"
	TicketCategoryBilling = "billing"
	TicketCategoryAPI     = "api"
	TicketCategoryModel   = "model"
	TicketCategoryOther   = "other"

	TicketPriorityNormal = "normal"
	TicketPriorityHigh   = "high"
	TicketPriorityUrgent = "urgent"

	TicketMessageKindPublic   = "public"
	TicketMessageKindInternal = "internal"
	TicketMessageKindSystem   = "system"

	TicketVisibilityUser  = "user"
	TicketVisibilityAdmin = "admin"
)

const (
	TicketMaxOpenPerUser    = 5
	TicketMaxCreatedPerDay  = 10
	TicketMaxFilesPerReply  = 5
	TicketMaxFileBytes      = 5 << 20
	TicketMaxReplyFileBytes = 20 << 20
)
