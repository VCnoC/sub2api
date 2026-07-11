package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// playgroundRequestEnvelope 中间件内部反序列化结构。
// 仅提取必要字段（model + group），其余 OpenAI 标准字段忽略并由下游
// chat completions handler 通过 gjson 在原始 body 上读取。
// 保持私有以避免 middleware 包反向依赖 handler/dto 子包。
type playgroundRequestEnvelope struct {
	Model string `json:"model"`
	Group string `json:"group"`
}

// ContextKeyPlaygroundRequest 标记请求来自对话广场的 ctx key，
// 供下游 handler / OpsService 区分流量来源做审计。
const ContextKeyPlaygroundRequest ContextKey = "playground_request"

// NewPlaygroundContextMiddleware 创建对话广场上下文中间件。
//
// 前置条件：必须在 JWTAuthMiddleware 之后使用（ctx 已含 AuthSubject{UserID}）。
//
// 中间件职责：
//  1. 读取请求体（同时重置 Body 让下游 chat completions handler 能再次读取）
//  2. 提取 model + group 字段并基础校验
//  3. 校验 group 属于用户可用分组（防越权访问）
//  4. 加载完整 User + Group 对象
//  5. 加载订阅（若 group 是订阅类型；订阅缺失则拒绝）
//  6. 非订阅创建请求执行余额检查；只读状态查询不重复检查余额
//  7. 构造虚拟 APIKey（含 User/Group 预加载）注入 ctx
//  8. 标记 ctx 为 playground 来源便于审计
//
// 失败模式（OpenAI 兼容错误格式）：
//   - body 解析失败 → 400 invalid_request_error
//   - group 越权 → 403 permission_error
//   - 用户不存在/未激活 → 401 authentication_error
//   - 订阅类型 group 无活跃订阅 → 403 permission_error
//   - 非订阅创建请求且余额 ≤0 → 403 insufficient_quota
func NewPlaygroundContextMiddleware(
	apiKeyService *service.APIKeyService,
	userService *service.UserService,
	subscriptionService *service.SubscriptionService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		subject, ok := GetAuthSubjectFromContext(c)
		if !ok {
			playgroundError(c, http.StatusUnauthorized, "authentication_error", "Authentication required")
			return
		}

		var req playgroundRequestEnvelope
		if c.Request.Method == http.MethodGet {
			req.Group = c.Query("group")
		} else {
			body, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
			if err != nil {
				playgroundError(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
				return
			}
			if len(body) == 0 {
				playgroundError(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
				return
			}
			// 下游媒体/聊天处理器仍需读取完整请求体。
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			c.Request.ContentLength = int64(len(body))
			if err := json.Unmarshal(body, &req); err != nil {
				playgroundError(c, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
				return
			}
		}
		req.Model = strings.TrimSpace(req.Model)
		req.Group = strings.TrimSpace(req.Group)
		if c.Request.Method != http.MethodGet && req.Model == "" {
			playgroundError(c, http.StatusBadRequest, "invalid_request_error", "model is required")
			return
		}
		if req.Group == "" {
			playgroundError(c, http.StatusBadRequest, "invalid_request_error", "group is required")
			return
		}

		ctx := c.Request.Context()

		// 校验 group 属于用户可用分组
		availableGroups, err := apiKeyService.GetAvailableGroups(ctx, subject.UserID)
		if err != nil {
			playgroundError(c, http.StatusInternalServerError, "api_error", "Failed to load available groups")
			return
		}

		var targetGroup *service.Group
		for i := range availableGroups {
			if availableGroups[i].Name == req.Group {
				targetGroup = &availableGroups[i]
				break
			}
		}
		if targetGroup == nil {
			playgroundError(c, http.StatusForbidden, "permission_error", "Selected group is not available for current user")
			return
		}

		// 加载完整用户对象（用于计费/订阅/余额检查）
		user, err := userService.GetByID(ctx, subject.UserID)
		if err != nil {
			playgroundError(c, http.StatusUnauthorized, "authentication_error", "User not found")
			return
		}
		if !user.IsActive() {
			playgroundError(c, http.StatusUnauthorized, "authentication_error", "User account is not active")
			return
		}

		// 加载订阅（若 group 是订阅类型）
		var subscription *service.UserSubscription
		if targetGroup.IsSubscriptionType() && subscriptionService != nil {
			sub, subErr := subscriptionService.GetActiveSubscription(ctx, user.ID, targetGroup.ID)
			if subErr != nil {
				playgroundError(c, http.StatusForbidden, "permission_error", "No active subscription for this group")
				return
			}
			subscription = sub
		}

		// 状态查询属于已扣费任务，即使本次扣费后余额归零也必须允许查询终态。
		if c.Request.Method != http.MethodGet && subscription == nil && user.Balance <= 0 {
			playgroundError(c, http.StatusForbidden, "insufficient_quota", "Insufficient account balance")
			return
		}

		// 获取或创建持久化的对话广场 APIKey。
		// 必须使用真实持久化 key 才能让 usage_logs.api_key_id 外键约束通过，
		// 进而让用量记录正确入库 + 在用户「使用记录」页面可见。
		playgroundKey, err := apiKeyService.GetOrCreatePlaygroundKey(ctx, user.ID, targetGroup.ID)
		if err != nil {
			playgroundError(c, http.StatusInternalServerError, "api_error", "Failed to acquire playground key")
			return
		}
		// 回灌完整的 User / Group 引用（防止 service 层未填充时下游 nil 解引用）
		playgroundKey.User = user
		playgroundKey.Group = targetGroup

		// 写入 ctx 模拟 APIKeyAuth 中间件的产物
		c.Set(string(ContextKeyAPIKey), playgroundKey)
		c.Set(string(ContextKeyUserRole), user.Role)
		if subscription != nil {
			c.Set(string(ContextKeySubscription), subscription)
		}
		setGroupContext(c, targetGroup)
		c.Set(string(ContextKeyPlaygroundRequest), true)

		// 标记 ops 来源，便于 ops 错误日志 / 用量统计区分对话广场流量
		service.SetOpsRequestSource(c, service.OpsRequestSourcePlayground)

		c.Next()
	}
}

// playgroundError 以 OpenAI 兼容格式返回错误响应。
// 与现有 chat completions handler 错误响应格式一致，便于前端统一处理。
func playgroundError(c *gin.Context, status int, errType, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	})
}

// IsPlaygroundRequest 判断当前请求是否来自对话广场。
// 供下游 handler / OpsService 区分流量来源（用于审计与统计）。
func IsPlaygroundRequest(c *gin.Context) bool {
	value, exists := c.Get(string(ContextKeyPlaygroundRequest))
	if !exists {
		return false
	}
	b, ok := value.(bool)
	return ok && b
}
