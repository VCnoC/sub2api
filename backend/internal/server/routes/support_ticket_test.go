// Package routes 验证工单 API 路由契约与鉴权边界。
package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	adminhandler "github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSupportTicketRoutesMatchContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	handlers := supportTicketRouteHandlers()
	registerUserSupportTicketRoutes(v1.Group(""), handlers)
	registerSupportTicketRoutes(v1.Group("/admin"), handlers)

	routes := make(map[string]bool)
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, route := range []string{
		"GET /api/v1/tickets",
		"POST /api/v1/tickets",
		"GET /api/v1/tickets/unread-count",
		"GET /api/v1/tickets/:id",
		"POST /api/v1/tickets/:id/replies",
		"GET /api/v1/ticket-attachments/:id",
		"GET /api/v1/admin/tickets",
		"GET /api/v1/admin/tickets/unread-count",
		"GET /api/v1/admin/tickets/:id",
		"POST /api/v1/admin/tickets/:id/replies",
		"PATCH /api/v1/admin/tickets/:id",
		"DELETE /api/v1/admin/ticket-attachments/:id",
	} {
		require.Truef(t, routes[route], "missing route %s", route)
	}
}

func TestSupportTicketRoutesRemainBehindAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handlers := supportTicketRouteHandlers()

	tests := []struct {
		name       string
		path       string
		statusCode int
		register   func(*gin.RouterGroup)
	}{
		{
			name: "user", path: "/api/v1/tickets", statusCode: http.StatusUnauthorized,
			register: func(group *gin.RouterGroup) {
				group.Use(abortTicketRequest(http.StatusUnauthorized))
				registerUserSupportTicketRoutes(group, handlers)
			},
		},
		{
			name: "admin", path: "/api/v1/admin/tickets", statusCode: http.StatusForbidden,
			register: func(group *gin.RouterGroup) {
				admin := group.Group("/admin")
				admin.Use(abortTicketRequest(http.StatusForbidden))
				registerSupportTicketRoutes(admin, handlers)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			test.register(router.Group("/api/v1"))
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.path, nil)

			router.ServeHTTP(recorder, request)

			require.Equal(t, test.statusCode, recorder.Code)
		})
	}
}

func supportTicketRouteHandlers() *handler.Handlers {
	return &handler.Handlers{
		SupportTicket: &handler.SupportTicketHandler{},
		Admin:         &handler.AdminHandlers{SupportTicket: &adminhandler.SupportTicketHandler{}},
	}
}

func abortTicketRequest(status int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatus(status)
	}
}
