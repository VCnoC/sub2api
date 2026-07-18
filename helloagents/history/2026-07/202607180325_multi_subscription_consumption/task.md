# 任务清单: 多订阅独立消费

目录: `helloagents/history/2026-07/202607180325_multi_subscription_consumption/`

---

## 1. 数据模型与迁移
- [√] 1.1 在 `backend/migrations/183_add_payment_order_subscription_link.sql` 和 `backend/migrations/184_multi_subscription_candidate_indexes_notx.sql` 中增加支付订单订阅外键、移除同组唯一索引并增加候选排序索引，验证 why.md#需求-独立发放订阅-场景-重复获得同一套餐
- [√] 1.2 在 `backend/ent/schema/payment_order.go` 中增加可空 `subscription_id` 字段，在 `backend/ent/schema/user_subscription.go` 中更新多订阅索引说明，依赖任务1.1
- [√] 1.3 运行 Ent 代码生成并审计生成结果，确认只包含 schema 对应机械变更，依赖任务1.2
- [√] 1.4 在 migration 回归测试中验证唯一索引移除、候选索引与外键存在，依赖任务1.1

## 2. 仓储与独立权益发放
- [√] 2.1 在 `backend/internal/service/user_subscription_port.go` 和 `backend/internal/repository/user_subscription_repo.go` 中增加同组有效候选列表查询并固定 `expires_at ASC, id ASC`，验证 why.md#需求-最早到期优先消费-场景-第一张仍有额度
- [√] 2.2 在 `backend/internal/repository/user_subscription_repo_integration_test.go` 中覆盖同组多行创建、排序、软删除和过期过滤，依赖任务2.1
- [√] 2.3 在 `backend/internal/service/subscription_service.go` 中新增独立发放入口，并保留注册默认权益的幂等入口，验证 why.md#需求-独立发放订阅-场景-重复获得同一套餐
- [√] 2.4 在订阅 Service 测试中验证重复发放创建新记录且不重置原记录，依赖任务2.3

## 3. 最早到期选卡
- [√] 3.1 在 `backend/internal/service/subscription_service.go` 中实现候选遍历、窗口维护、额度错误跳过和全部耗尽错误归并，验证 why.md#需求-最早到期优先消费-场景-第一张当前额度耗尽
- [√] 3.2 在订阅 Service 测试中覆盖最早到期选择、耗尽切换、多日窗口重置回归和1日卡一次性额度，依赖任务3.1
- [√] 3.3 在 `backend/internal/server/middleware/api_key_auth.go` 和 `backend/internal/server/middleware/api_key_auth_google.go` 中统一使用已选择的具体订阅并保持403/429错误语义，依赖任务3.1
- [√] 3.4 更新主鉴权和 Google 鉴权测试，验证上下文传递的订阅 ID 与候选选择一致，依赖任务3.3

## 4. 计费与缓存一致性
- [√] 4.1 在 `backend/internal/service/billing_cache_service.go` 中让订阅资格检查使用传入的具体订阅快照，停止聚合订阅缓存参与放行，验证 why.md#需求-最早到期优先消费-场景-第一张仍有额度
- [√] 4.2 在 `backend/internal/service/gateway_usage_billing.go` 中停止更新无订阅 ID 维度的聚合用量缓存，确认现有数据库幂等扣费继续按选中 `subscription_id` 写入，依赖任务4.1
- [√] 4.3 更新计费缓存和 Gateway 用量测试，验证同组多卡用量互不污染，依赖任务4.1和4.2

## 5. 发放入口与精确退款
- [√] 5.1 在 `backend/internal/service/payment_fulfillment.go` 中事务性创建独立订阅、写回订单 `subscription_id` 并保留审计幂等，验证 why.md#需求-独立发放订阅-场景-重复获得同一套餐
- [√] 5.2 在 `backend/internal/service/payment_fulfillment_test.go` 中验证并发或重试履约只创建一条权益且订单关联正确，依赖任务5.1
- [√] 5.3 在 `backend/internal/service/payment_refund.go` 中优先按订单 `subscription_id` 调整和回滚同一权益，并保留历史空关联的受控兼容路径，验证 why.md#需求-精确退款-场景-新订单退款
- [√] 5.4 更新支付退款测试，覆盖精确扣减、失败回滚、已耗尽权益和历史订单警告，依赖任务5.3
- [√] 5.5 在 `backend/internal/service/redeem_service.go` 中让正数订阅兑换创建独立记录，并让负数扣减按最早到期顺序处理，验证 why.md#需求-独立发放订阅-场景-重复获得同一套餐
- [√] 5.6 更新兑换码与抽奖订阅测试，验证每次成功事件只发一张独立订阅，依赖任务5.5
- [√] 5.7 调整管理员单次和批量订阅发放入口为创建独立权益，同时保留注册默认分配幂等语义，依赖任务2.3

## 6. 用户端展示
- [√] 6.1 在 `frontend/src/components/payment/SubscriptionPlanCard.vue` 和支付类型中支持同组订阅“再次购买”语义及可选订单订阅 ID
- [√] 6.2 更新中英文支付文案并验证同组多张订阅在现有订阅页面逐条显示，依赖任务6.1
- [√] 6.3 更新相关 Vitest，覆盖同组多订阅列表和再次购买按钮，依赖任务6.1和6.2

## 7. 安全与一致性检查
- [√] 7.1 执行安全检查：订阅所有权边界、支付回调幂等、退款目标精确性、软删除过滤、错误信息脱敏
- [√] 7.2 审计所有 `GetByUserIDAndGroupID` 和 `Only()` 生产调用，消除多行假设，依赖任务2.1
- [√] 7.3 检查并发边界：同一订单重复履约、并发请求选中同一卡、窗口重置竞争和退款回滚

## 8. 知识库同步
- [√] 8.1 更新 `helloagents/wiki/data.md` 和 `helloagents/wiki/modules/groups-billing.md`，记录独立订阅和选卡规则
- [√] 8.2 更新 `helloagents/wiki/arch.md` 的 ADR 索引及 `helloagents/CHANGELOG.md`，依赖任务8.1

## 9. 验证与发布
- [√] 9.1 运行订阅、计费、支付、退款、兑换码、抽奖和鉴权相关 Go 单元测试
- [X] 9.2 运行 PostgreSQL integration 测试及 migration 回归测试，验证多行排序与约束
  > 备注: migration 回归测试通过；PostgreSQL integration 环境启动超过 180 秒后超时，未产生断言失败输出。
- [√] 9.3 运行前端相关 Vitest、TypeScript 类型检查和生产构建
- [√] 9.4 执行一致性审计，确认代码、Ent schema、SQL migration、API DTO 和知识库规则一致
- [?] 9.5 上线前备份数据库并检查候选查询 `EXPLAIN` 命中复合索引
  > 备注: 需要在获授权的目标部署数据库执行，当前未连接生产服务。

## 执行备注

- 默认后端受影响包测试通过；`go test -tags=unit` 仍受工作区既有账号/团队测试桩缺少 `GrantSignupBonusBalances`、`ClearTeamMembership` 方法影响，属于本方案外问题。
- 前端订阅 Store Vitest 13 项、订阅卡片组件 Vitest 4 项、TypeScript 类型检查和 Vite 生产构建通过。
- `git diff --check` 与 migration 单元测试通过。
