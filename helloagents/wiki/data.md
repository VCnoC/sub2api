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
