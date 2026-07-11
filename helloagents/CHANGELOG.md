# Changelog

本文件记录项目的重要变更，格式遵循 Keep a Changelog。

## [Unreleased]

### Added
- 新增独立 `video` 平台，支持使用 Base URL、API Key、模型同步和模型白名单管理 cpa-fan 视频上游账号。
- 新增 cpa-fan 兼容的视频创建、生成、编辑、延长和查询路由。
- 新增 `video_tasks` 持久化任务与数据库租约 worker，进程重启后可继续跟踪异步终态。
- 对话广场新增视频分组生成能力，支持文生视频、1.5 preview 单图生视频、0-100% 进度和原生播放器。
- 新增 JWT 鉴权的 `POST /api/v1/playground/videos` 与 `GET /api/v1/playground/videos/:request_id`。

### Changed
- 视频分组支持 `per_second` 和 `per_request` 两种计费模式；独立视频平台固定使用余额计费。
- 视频任务扣费与任务写入在同一幂等事务中完成。
- 默认 CSP 允许加载 HTTPS 视频媒体；视频状态查询作为已扣费任务的只读操作，不重复要求正余额。
- 对话广场切换到 `grok-imagine-video*` 时显示 1-15 秒时长和画面比例选择，并随创建请求发送。

### Fixed
- 上游明确返回失败终态时自动退回本次余额，并通过任务锁和退款时间保证重复处理不重复退款。
- 修复对话广场视频模式仍显示文档入口、参考图无法移除，以及余额恰好扣至零后无法继续查询任务的问题。
