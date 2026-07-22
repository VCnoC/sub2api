# 数据模型

## 核心表

### accounts
保存平台、认证类型、JSON 凭证、代理、并发与调度状态；账号通过中间表关联分组。

### groups
保存平台、计费类型、倍率、媒体价格、路由限制和模型列表配置。

- `video_billing_mode`：`per_second` 或 `per_request`，默认 `per_second`。
- 订阅分组的 `subscription_billing_mode` 为 `usd` 或 `request_count`；次数模式通过 `request_limit_5h`、`request_limit_1d` 设置成功文本请求上限，0 表示该窗口不限。
- 视频平台只允许 `standard` 余额计费。
- `max_reasoning_effort` 保存 OpenAI/Codex 推理强度上限；`reasoning_effort_mappings` 以 JSONB 保存精确映射，执行顺序为先映射、后限制上限。

### api_keys
保存 Key 所有者、状态、配额、限速和 `group_id` 首组兼容镜像。新业务顺序以 `api_key_groups` 为准；旧的非空 `group_id` 在迁移 185 中回填为优先级 0。

### api_key_groups
保存 API Key 的有序候选分组：`api_key_id`、`group_id`、`priority`（0-4）和创建时间。

- `(api_key_id, group_id)` 唯一，禁止同一 Key 重复分组。
- `(api_key_id, priority)` 唯一，确保顺序位置不冲突。
- 删除 Key 时级联删除；删除分组前由仓储移除绑定、压紧优先级并同步 `api_keys.group_id`。

### usage_logs
保存请求模型、账号、API Key、费用、媒体尺寸、视频分辨率与时长快照。

### usage_billing_dedup
以请求 ID 和 API Key ID 幂等应用余额及配额变更。

### user_subscriptions
保存每次独立获得的订阅权益、有效期、状态、日/周/月 USD 窗口，以及 5 小时/24 小时成功请求次数窗口。同一用户和分组允许多条未删除记录；有效候选按 `(user_id, group_id, status, expires_at, id)` 查询并优先使用最早到期记录。

### subscription_request_reservations
保存次数订阅的请求前占位：逻辑请求 ID、API Key、用户、具体订阅、窗口快照、`pending/committed/released` 状态和过期时间。

- `(request_id, subscription_id)` 唯一，账号重试不会重复占位。
- pending 占位计入 `user_subscriptions.request_usage_*`，成功在统一计费事务中确认，失败按窗口快照释放。
- 过期 pending 在后续占位事务中回收；记录不保存提示词、请求正文、响应或密钥。

### payment_orders
订阅订单通过可空 `subscription_id` 关联实际发放的 `user_subscriptions` 记录。迁移前历史订单可能为空，新订单履约必须在同一事务内写入权益、订单关联和发放审计。

### audit_logs
追加保存管理面变更、敏感读取和认证事件的脱敏审计记录；按时间、操作者、动作和客户端 IP 建立查询索引。

### ops_ingress_reject_aggregates
按时间桶、客户端 IP、拒绝原因、路由族和协议聚合入口拒绝次数，并保存可选用户/API Key 归属；使用 migration 188 创建，避免记录高基数原始请求正文。

### auth_cache_invalidation_outbox
保存 API Key 与用户鉴权缓存失效事件、投递状态、重试次数和错误摘要；事务提交后由 worker 发布到 Redis，使用 migration 189 创建。

### prompt_audit_jobs / prompt_audit_events
分别保存提示词审计任务与不可变审计结果。任务表保存脱敏预览和调度状态，事件表保存风险等级、处置动作、扫描证据及受控的完整提示词内容。

### 异步图片任务
任务状态和紧凑结果按 24 小时生命周期保存于 Redis，生成图片写入 S3 兼容对象存储；未配置对象存储时不创建任务。

### video_tasks
保存异步视频任务的最小计费快照和轮询状态：上游任务 ID、计费请求 ID、用户/API Key/账号/分组、退款金额、状态、轮询时间、租约、轮询次数、终态时间、退款时间与错误摘要。

- 状态仅允许 `pending`、`completed`、`failed`。
- `(upstream_task_id, account_id)` 和 `(billing_request_id, api_key_id)` 唯一，防止重复建任务。
- 到期任务通过 `status + next_poll_at` 索引和 `locked_until` 租约并发领取。
- 任务表不保存提示词、请求正文或上游 API Key。

### user_platform_quotas
平台约束已包含 `video`，用于管理端配额展示；视频请求本身固定从用户余额扣费，不消费订阅额度。

### support_tickets
保存工单所有者、标题、分类、状态、优先级、可选负责人、关闭人、关闭时间和最后消息时间。

- 状态为 `pending_admin`、`pending_user` 或 `closed`。
- 分类为 `account`、`billing`、`api`、`model` 或 `other`；优先级为 `normal`、`high` 或 `urgent`。
- 以用户与创建时间、用户与状态、全局处理队列、负责人和分类建立索引。

### support_ticket_messages
保存不可变的公开回复、内部备注和系统事件。公开消息对用户和管理员可见，内部备注及管理员审计事件只对管理员可见；`metadata` 保存状态、负责人和优先级变化的最小 JSON 审计信息。

### support_ticket_attachments
保存附件名称、随机存储键、服务端探测 MIME、大小、清理计划和删除审计信息。文件本体位于私有目录；工单关闭时设置 `delete_after`，物理删除后保留元数据。

### support_ticket_reads
以 `(ticket_id, user_id)` 唯一记录每位用户或管理员的最后已读消息 ID。用户游标只推进到公开消息，管理员游标包含内部消息和管理员事件。

### lottery_pools
保存固定 `normal`、`luxury` 双奖池的名称、启停、每日/每周周期次数及可选活动时间；迁移默认创建两条停用配置。

### lottery_prizes
保存奖品所属奖池、余额或订阅权益、百万分比概率、可选库存、排序和 Data URL 卡面；软删除保留历史引用。

### lottery_rules
保存 `signup`、`redeem`、`recharge` 固定事件的受益人、双奖池机会数、单笔/累计充值门槛及重复模式；产生流水后的核心行为不可修改。

### lottery_user_chances
以 `(user_id, pool_id)` 唯一保存当前周期键、基础剩余次数和长期额外次数。基础次数惰性刷新且优先消耗，额外次数跨周期保留。

### lottery_chance_ledger
保存机会发放、退款冲正和抽奖扣减的不可变审计流水；`dedupe_key` 唯一，`balance_after` 和 `metadata` 保存当时快照。

### lottery_draws
保存用户、奖池、幂等键、中奖结果、次数来源、奖品/兑换码引用及奖品快照；`(user_id, pool_id, idempotency_key)` 唯一。

### teams
团队在原有名称、owner、邀请码、状态和资金池余额之外保存 `member_limit`、`level` 和 `review_required`。现有团队迁移时保留成员与余额，上限至少为当前成员数，并标记待管理员复审。

### team_governance_settings
单行保存最低注册天数、最低有效累计充值，以及 5/15/40 各档累计充值、近 7 天 `actual_cost` 和 `and/or` 条件。`configured=false` 时禁止 owner 自助升级，管理员首次保存后置为已配置。

### team_applications
保存 `create/expand` 申请、申请人、团队/名称、目标人数、理由、补充信息及 `pending/approved/rejected` 审核状态。部分唯一索引保证同一申请人只有一个待审创建申请、同一团队只有一个待审扩容申请。

### team_join_requests
保存邀请码加入申请及 owner 审核结果。同一用户只能有一个待审加入申请；批准事务锁定团队，重新统计成员并更新用户团队关系。

### team_transferable_balances
保存用户当前可转赠额度。迁移时按现有余额初始化；真实支付对应的余额兑换、普通余额兑换码和正向管理员余额调整增加额度，抽奖、邀请、注册奖励和团队转入不增加。任何余额下降都会同步扣减额度，避免已消费资金或团队转入资金循环归集。

### team_fund_ledger
保存 `deposit/allocate/transfer` 团队资金动作、用户、对手方、操作人、金额和时间。资金存入、直接转赠和资金池分配在单个 PostgreSQL 事务中按团队、用户、可转赠额度顺序锁定。
