# API 手册

## API Key API
- `POST /api/v1/keys`：创建 API Key；`group_ids` 为 1-5 个同平台分组 ID 的有序数组，`group_id` 保留为首组兼容字段。
- `PUT /api/v1/keys/:id`：整体替换候选分组顺序；提交 `group_ids` 时按数组顺序保存，仅提交 `group_id` 时按旧语义替换为单分组。
- `GET /api/v1/keys`、`GET /api/v1/keys/:id`：响应同时返回 `group_ids`/`groups` 和首组兼容字段 `group_id`/`group`。

候选分组必须去重、全部有效、当前用户可绑定且平台相同。每个新网关请求从首组开始；余额或订阅资格不可用、组内无账号或安全可重试故障耗尽时依次推进。参数/权限错误、客户端取消、响应已提交及异步任务提交后不跨组重放。

## 管理 API
- `POST /api/v1/admin/accounts`：创建上游账号并绑定分组。
- `POST /api/v1/admin/accounts/models/sync-upstream-preview`：使用未保存的凭证获取上游模型。
- `POST /api/v1/admin/accounts/:id/models/sync-upstream`：获取已保存账号的上游模型。
- `POST /api/v1/admin/groups`、`PUT /api/v1/admin/groups/:id`：维护分组与媒体价格；订阅分组支持 `subscription_billing_mode=usd|request_count`、`request_limit_5h` 和 `request_limit_1d`。
- OpenAI 分组可通过 `max_reasoning_effort` 设置 `minimal/low/medium/high/xhigh/max` 上限，并通过 `reasoning_effort_mappings` 配置精确映射；服务端先映射再限制上限。
- `GET /api/v1/admin/ops/ingress-rejections`、`GET /api/v1/admin/ops/ingress-rejections/health`：查询入口拒绝聚合与采集健康状态。
- `GET /api/v1/admin/ops/auth-cache-invalidation/health`：查询跨实例鉴权缓存失效 outbox 健康状态。

视频账号使用 `platform=video` 和 API Key 认证，凭证包含 `base_url` 与 `api_key`；只能绑定 `video` 分组。视频分组的 `video_billing_mode` 取值为 `per_second` 或 `per_request`，且只允许标准余额计费。

次数订阅仅统计成功的文本生成请求。`GET /api/v1/subscriptions` 返回 `request_usage_5h`、`request_usage_1d` 和对应窗口起点；`GET /api/v1/subscriptions/progress` 在次数模式返回 `request_5h`、`request_1d` 的上限、已用、剩余和重置信息。

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

该入口只允许选择 `platform=openai` 的用户可用分组，不向浏览器暴露上游 API Key。前端将 `gpt-image-*` 和 `image-2-*` 识别为生图模型；选择 `gpt-image-2-1k/-2k/-4k` 或 `image-2-1k/-2k/-4k` 时，清晰度与模型后缀同步，画面比例会转换为该档位的精确 `size`。`quality`、`response_format`、`background`、`style`、`watermark` 仍按启用开关发送；服务端复用现有 `/v1/images/generations` 的图片权限、调度、计费、限流和用量记录链路。

若上游以 `502/503` 包装明确的内容审核、内容策略或敏感词拒绝，接口会规范化为 `400` 并保留脱敏后的上游错误消息与错误码；真实服务故障仍执行原有换号容灾。

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

## 兑换码与公开设置 API

- `POST /api/v1/admin/redeem-codes/generate`：当 `type=lottery_chance` 时，`value` 必须为正整数，`pool_key` 必须为 `normal` 或 `luxury`；列表和详情响应返回 `pool_key`。
- `POST /api/v1/redeem`：兑换抽奖次数码时，在同一事务内向指定奖池发放长期额外次数；以兑换码 ID 保证幂等，且不触发首次兑换邀请奖励。
- `GET /api/v1/settings/public`：公开设置响应包含纯文本 `dashboard_notice`；空字符串表示不展示仪表盘公告。

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

## 团队治理 API

用户接口使用登录 JWT：

- `POST /api/v1/user/team`：提交创建团队申请，不直接创建团队；请求包含 `name`、`reason`、`additional_info`。
- `GET /api/v1/user/team/application`：读取本人最近一次创建申请。
- `GET /api/v1/user/team/eligibility`：读取注册天数、有效累计充值、管理员门槛及当前是否满足。
- `POST /api/v1/user/team/join`：按邀请码提交加入申请，不直接加入。
- `GET /api/v1/user/team/join-requests`、`POST /api/v1/user/team/join-requests/:id/review`：owner 查看并审批加入申请；批准时锁定团队并校验人数上限。
- `GET /api/v1/user/team/governance`：读取等级、人数、有效充值、近 7 天消费、可转赠余额与升级配置。
- `POST /api/v1/user/team/upgrade`：owner 主动升级到当前满足的最高 5/15/40 档；条件未配置或团队待复审时拒绝。
- `POST /api/v1/user/team/expand`：提交超过 40 人的扩容申请。
- 既有资金接口保持路径不变，但存入和 owner 直接转赠同时受可转赠额度约束。

管理员接口使用管理员 JWT：

- `GET /api/v1/admin/teams/stats`、`GET /api/v1/admin/teams`、`GET /api/v1/admin/teams/:id`：读取团队总数、列表、成员、指标、申请和资金流水。
- `GET|PUT /api/v1/admin/teams/settings`：维护创建门槛与 5/15/40 各档充值、近 7 天消费和 AND/OR 条件。
- `GET /api/v1/admin/teams/applications`、`POST /api/v1/admin/teams/applications/:id/review`：审核创建/扩容；创建门槛不足时只有填写原因并设置 `waive=true` 才可批准。
- `PUT /api/v1/admin/teams/:id/status`：冻结或恢复团队。
- `PUT /api/v1/admin/teams/:id/member-limit`：直接修改单团队人数上限，不受自动升级条件限制，但不能低于当前人数。
- `POST /api/v1/admin/teams/:id/review-complete`、`DELETE /api/v1/admin/teams/:id/members/:member_id`：完成现有团队复审或移除普通成员。
