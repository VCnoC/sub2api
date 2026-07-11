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
