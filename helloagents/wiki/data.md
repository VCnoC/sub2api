# 数据模型

## 核心表

### accounts
保存平台、认证类型、JSON 凭证、代理、并发与调度状态；账号通过中间表关联分组。

### groups
保存平台、计费类型、倍率、媒体价格、路由限制和模型列表配置。

- `video_billing_mode`：`per_second` 或 `per_request`，默认 `per_second`。
- 视频平台只允许 `standard` 余额计费。

### usage_logs
保存请求模型、账号、API Key、费用、媒体尺寸、视频分辨率与时长快照。

### usage_billing_dedup
以请求 ID 和 API Key ID 幂等应用余额及配额变更。

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
