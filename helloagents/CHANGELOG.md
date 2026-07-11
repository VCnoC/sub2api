# Changelog

本文件记录项目的重要变更，格式遵循 Keep a Changelog。

## [Unreleased]

### Added
- 新增独立 `video` 平台，支持使用 Base URL、API Key、模型同步和模型白名单管理 cpa-fan 视频上游账号。
- 新增 cpa-fan 兼容的视频创建、生成、编辑、延长和查询路由。
- 新增 `video_tasks` 持久化任务与数据库租约 worker，进程重启后可继续跟踪异步终态。

### Changed
- 视频分组支持 `per_second` 和 `per_request` 两种计费模式；独立视频平台固定使用余额计费。
- 视频任务扣费与任务写入在同一幂等事务中完成。

### Fixed
- 上游明确返回失败终态时自动退回本次余额，并通过任务锁和退款时间保证重复处理不重复退款。
