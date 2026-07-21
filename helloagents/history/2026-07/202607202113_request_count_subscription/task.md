# 任务清单: 纯次数订阅套餐

目录: `helloagents/plan/202607202113_request_count_subscription/`

---

## 1. 数据模型与迁移
- [√] 1.1 在 `backend/ent/schema/group.go` 和 `backend/ent/schema/user_subscription.go` 增加次数计费模式、上限、窗口和计数字段，验证 why.md#需求-创建纯次数订阅分组-场景-管理员创建套餐
- [√] 1.2 新建 `backend/ent/schema/subscription_request_reservation.go`，定义占位状态、关联、唯一约束和 pending 到期索引，验证 why.md#需求-独立次数权益-场景-首次成功文本调用
- [√] 1.3 新建 `backend/migrations/186_request_count_subscriptions.sql` 并更新迁移回归测试，依赖任务1.1和1.2
- [√] 1.4 运行 Ent 代码生成并审计生成结果，确认只包含 schema 对应机械变更，依赖任务1.1和1.2

## 2. 分组领域与管理 API
- [√] 2.1 在 `backend/internal/service/group.go` 和 `backend/internal/service/admin_service.go` 增加计费模式与次数上限领域字段和输入字段
- [√] 2.2 在 `backend/internal/service/admin_group.go` 和相关测试中实现模式归一化、非负整数校验、双 0 拒绝及 USD/次数字段互斥，依赖任务2.1
- [√] 2.3 在 `backend/internal/handler/admin/group_handler.go` 和 Handler 测试中扩展创建、更新及响应 DTO，依赖任务2.2
- [√] 2.4 在 `backend/internal/repository/group_repo.go` 和映射测试中持久化并回读新字段，依赖任务1.4和2.1

## 3. 占位仓储与多订阅选择
- [√] 3.1 在 `backend/internal/service/user_subscription.go` 和 `backend/internal/service/user_subscription_port.go` 定义次数窗口、占位模型、错误和仓储接口
- [√] 3.2 在 `backend/internal/repository/user_subscription_repo.go` 实现事务性 reserve/commit/release、窗口重置、过期占位回收和最早到期选卡，依赖任务1.3和3.1
- [√] 3.3 在 `backend/internal/repository/user_subscription_repo_integration_test.go` 覆盖并发上限、幂等、失败释放、跨窗口释放和多卡切换，依赖任务3.2
- [√] 3.4 在 `backend/internal/service/subscription_service.go` 和 Service 测试中接入次数候选、429错误归并及缓存失效，依赖任务3.2

## 4. 文本请求分类与生命周期
- [√] 4.1 在 `backend/internal/handler/endpoint.go` 和测试中提供统一的可计次文本请求分类，排除 count_tokens、媒体意图和只读端点
- [√] 4.2 在 `backend/internal/server/middleware/api_key_group_failover.go` 和测试中保存当前 reservation，并支持切组前幂等释放，依赖任务3.4
- [√] 4.3 在 `backend/internal/handler/failover_loop.go` 和测试中让账号重试复用占位、跨组 failover 释放并重新占位，依赖任务4.1和4.2
- [√] 4.4 在 Messages、Chat Completions、Responses 与 Gemini 顶层 Handler 中调用统一占位入口，单任务最多修改两个 Handler 文件并分别提交测试，依赖任务4.3

## 5. 成功确认与纯次数结算
- [√] 5.1 在 `backend/internal/service/usage_billing.go` 和 `backend/internal/service/gateway_usage_billing.go` 增加 reservation ID，并让次数模式不生成 `SubscriptionCost`，验证 why.md#需求-独立次数权益-场景-首次成功文本调用
- [√] 5.2 在 `backend/internal/repository/usage_billing_repo.go` 和集成测试中把占位确认并入现有幂等计费事务，依赖任务5.1
- [√] 5.3 在 `backend/internal/service/openai_gateway_usage.go` 和通用 Gateway 用量路径中传递占位并覆盖零 USD 成本仍确认次数，依赖任务5.2
- [√] 5.4 更新流式响应测试，验证正常结束和上游已响应后的客户端中断确认次数，明确失败释放次数，依赖任务4.4和5.3

## 6. 管理端与用户展示
- [√] 6.1 在 `frontend/src/types/index.ts` 和管理 API 类型中增加次数计费字段
- [√] 6.2 在 `frontend/src/views/admin/GroupsView.vue` 中增加计费模式选择、5小时/24小时上限输入、列表展示和互斥表单校验，依赖任务6.1
- [√] 6.3 在 `frontend/src/views/admin/orders/PlanEditDialog.vue` 和 `frontend/src/components/payment/SubscriptionPlanCard.vue` 展示绑定分组的次数权益，依赖任务6.1
- [√] 6.4 在 `backend/internal/service/subscription_service.go` 和 `frontend/src/types/index.ts` 扩展订阅进度结构，返回已用、剩余和重置时间，依赖任务3.4
- [√] 6.5 更新用户订阅进度页面及中英文文案，并增加相关 Vitest，依赖任务6.4

## 7. 安全与一致性检查
- [√] 7.1 执行安全检查：占位所有权、请求 ID 冲突、输入上限、错误信息脱敏、占位表不保存提示词或密钥
- [√] 7.2 审计所有成功、失败、取消和跨组退出路径，确认每个 pending 占位最终 committed、released 或可过期回收
- [√] 7.3 验证退款、撤销、过期和删除订阅不会把新请求分配到失效权益；历史 USD 订阅默认行为不变

## 8. 知识库与 SDD 同步
- [√] 8.1 更新 `helloagents/wiki/data.md` 和 `helloagents/wiki/modules/groups-billing.md`，记录次数分组、窗口与占位规则
- [√] 8.2 更新 `helloagents/wiki/arch.md` 和 `helloagents/CHANGELOG.md`，登记 ADR 与新增功能
- [√] 8.3 执行完成后同步勾选 `sdd/specs/request-count-subscription/spec.md` 与 `tasks.md`，并更新 `sdd/project.md` 状态

## 9. 验证与发布
- [√] 9.1 运行分组、订阅、鉴权、failover、Gateway 和 usage billing 相关 Go 单元测试
- [√] 9.2 运行 PostgreSQL integration 与 migration 回归测试，验证并发占位、状态转换、索引和约束
- [√] 9.3 运行前端相关 Vitest、TypeScript 类型检查和生产构建
- [√] 9.4 执行 `git diff --check`、Ent/schema/migration/API DTO 一致性审计
- [-] 9.5 上线前备份数据库，滚动发布并观察次数耗尽 429、pending 占位数量和事务延迟
  > 备注: 本次仅完成本地实现与验证，未执行生产数据库备份、部署和线上观测。

## 执行总结
- 核心功能、数据库迁移、后端与前端验证均已完成。
- 仓库既有 `-tags=unit` 全包测试桩缺少接口方法，未作为本功能阻断项；本功能相关默认测试和 integration 测试均已通过。
