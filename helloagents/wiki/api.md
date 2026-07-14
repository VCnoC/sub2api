# API 手册

## 管理 API
- `POST /api/v1/admin/accounts`：创建上游账号并绑定分组。
- `POST /api/v1/admin/accounts/models/sync-upstream-preview`：使用未保存的凭证获取上游模型。
- `POST /api/v1/admin/accounts/:id/models/sync-upstream`：获取已保存账号的上游模型。
- `POST /api/v1/admin/groups`、`PUT /api/v1/admin/groups/:id`：维护分组与媒体价格。

视频账号使用 `platform=video` 和 API Key 认证，凭证包含 `base_url` 与 `api_key`；只能绑定 `video` 分组。视频分组的 `video_billing_mode` 取值为 `per_second` 或 `per_request`，且只允许标准余额计费。

## 网关 API
- `GET /v1/models`：返回当前 API Key 分组可用模型。
- `POST /v1/videos`：OpenAI 风格视频创建；未提供时按 4 秒、720p 计费。
- `POST /v1/video/create`：`/v1/videos` 的兼容创建入口；未提供时按 4 秒、720p 计费。
- `POST /v1/videos/generations`：xAI 视频生成入口。
- `POST /v1/videos/edits`：基于已有视频或素材编辑。
- `POST /v1/videos/extensions`：延长已有视频。
- `GET /v1/videos/:video_id`：按路径参数查询任务。
- `GET /v1/video/query?id=:video_id`：按查询参数查询任务。

所有视频路由沿用客户端 API Key 鉴权，并使用所选视频账号的 `Authorization: Bearer <api_key>` 透明转发。创建类请求支持 `seconds`/`duration` 和 `size`/`resolution` 计费字段；查询类请求不重复计费。

## 对话广场视频 API
- `POST /api/v1/playground/videos`：JWT 用户创建视频任务，请求为 `{ model, group, prompt, seconds?, aspect_ratio?, input_reference?: { image_url } }`。
- `GET /api/v1/playground/videos/:request_id?group=:group`：JWT 用户查询已创建任务，响应透传 `status`、`progress`、`video_url` 和上游错误。

两个入口只允许选择 `platform=video` 的用户可用分组，不向浏览器暴露上游 API Key。`grok-imagine-video*` 可传 1-15 秒的 `seconds` 和常用预设 `aspect_ratio`；基础模型可文生视频，`grok-imagine-video-1.5-preview` 在对话广场必须携带一张 `input_reference.image_url` 参考图。创建请求执行余额资格检查并计费，状态查询不重复计费，且在余额被本任务扣至零后仍可读取终态。

## 对话广场生图 API
- `POST /api/v1/playground/images/generations`：JWT 用户提交 OpenAI Images 兼容生图请求，请求为 `{ model, group, prompt, n?, size, image?, quality?, response_format?, style?, background?, watermark? }`。

该入口只允许选择 `platform=openai` 的用户可用分组，不向浏览器暴露上游 API Key。当前前端在选择 `gpt-image-2-vip` 时显示 `size` 尺寸预设，并按启用开关发送 `quality`、`response_format`、`background`、`style`、`watermark`；服务端复用现有 `/v1/images/generations` 的图片权限、调度、计费、限流和用量记录链路。

## 站内工单 API

用户接口使用登录 JWT，且只能访问本人资源：

- `GET /api/v1/tickets`：按 `page`、`page_size`、`status`、`category` 分页查询本人工单。
- `POST /api/v1/tickets`：以 `multipart/form-data` 提交 `subject`、`category`、`body` 和最多 5 个 `files`。
- `GET /api/v1/tickets/:id`：读取本人工单详情并推进个人已读游标。
- `POST /api/v1/tickets/:id/replies`：提交公开回复和附件；回复已关闭工单会原子重开。
- `GET /api/v1/tickets/unread-count`：返回 `{ count }`。
- `GET /api/v1/ticket-attachments/:id`：鉴权打开图片或下载文本附件。

管理员接口使用管理员 JWT：

- `GET /api/v1/admin/tickets`：支持状态、分类、优先级、`mine`/`unassigned` 负责人和关键词筛选。
- `GET /api/v1/admin/tickets/:id`：读取公开消息、内部备注和系统审计事件。
- `POST /api/v1/admin/tickets/:id/replies`：`multipart/form-data` 公开回复；`internal=true` 时添加无附件的内部备注。
- `PATCH /api/v1/admin/tickets/:id`：更新 `priority`、`assignee_id` 或 `closed`。
- `GET /api/v1/admin/tickets/unread-count`：返回当前管理员个人未读工单数。
- `DELETE /api/v1/admin/ticket-attachments/:id`：以 `{ "reason": "..." }` 删除附件并保留审计元数据。

主要业务错误为 `TICKET_NOT_FOUND`、`TICKET_INVALID_INPUT`、`TICKET_ATTACHMENT_INVALID`、`TICKET_OPEN_LIMIT_REACHED`、`TICKET_DAILY_LIMIT_REACHED`、`TICKET_ASSIGNEE_INVALID` 和 `TICKET_ATTACHMENT_NOT_FOUND`。用户越权读取统一使用不存在错误，避免泄露资源是否存在。

## 邀请抽奖 API

用户接口使用登录 JWT：

- `GET /api/v1/lottery`：返回普通、豪华奖池的活动状态、基础/额外剩余次数和可见奖品。
- `POST /api/v1/lottery/pools/:key/draw`：使用 `Idempotency-Key` 抽取 `normal` 或 `luxury` 奖池；响应包含服务端结果、奖品快照和剩余次数。
- `GET /api/v1/lottery/history`：按 `page`、`page_size` 查询本人不可变抽奖历史。

管理员接口使用管理员 JWT：

- `GET /api/v1/admin/lottery/pools`、`PATCH /api/v1/admin/lottery/pools/:key`：读取和维护固定双奖池。
- `GET|POST /api/v1/admin/lottery/prizes`、`PATCH|DELETE /api/v1/admin/lottery/prizes/:id`：维护余额或订阅奖品。
- `GET|POST /api/v1/admin/lottery/rules`、`PATCH|DELETE /api/v1/admin/lottery/rules/:id`：维护注册、首次兑换和充值机会规则。
- `GET /api/v1/admin/lottery/draws`：按用户、奖池、结果和时间分页查询抽奖记录。
- `GET /api/v1/admin/lottery/chance-ledger`：按用户、奖池、动作和时间分页查询次数流水。

主要业务错误为 `LOTTERY_POOL_NOT_FOUND`、`LOTTERY_PRIZE_NOT_FOUND`、`LOTTERY_RULE_NOT_FOUND`、`LOTTERY_DRAW_NOT_FOUND`、`LOTTERY_INACTIVE`、`LOTTERY_NO_CHANCE`、`LOTTERY_INVALID_INPUT`、`LOTTERY_PROBABILITY_INVALID`、`LOTTERY_IMAGE_INVALID`、`LOTTERY_FULFILL_FAILED` 和 `LOTTERY_RULE_IMMUTABLE`。
