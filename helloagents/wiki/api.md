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
