# 网关与视频任务

## 目的
对客户端暴露兼容 API，完成分组调度、上游转发、任务跟踪和用量记录。

## 规范

### 需求: 视频任务路由
**模块:** 网关与任务

#### 场景: 创建和查询
- 创建任务后保存上游任务 ID 与选中账号，查询继续使用同一账号。
- 视频平台不参与文本和图片路由。
- 后台自动查询终态，进程重启后继续处理。
- 支持 `/v1/videos`、`/v1/video/create`、`/v1/videos/generations`、`/v1/videos/edits`、`/v1/videos/extensions`、`/v1/videos/:id` 和 `/v1/video/query`。
- 使用账号 Base URL 和 Bearer API Key 透明转发请求与响应。

### 需求: 失败退款
**模块:** 网关与任务

#### 场景: 上游返回失败
- 任务状态进入 failed。
- 已扣用户余额原额退回，重复终态通知不得重复退款。
- 只有 `failed`、`error`、`expired`、`cancelled`/`canceled` 等明确失败状态触发退款；传输错误和未知状态按退避策略继续轮询。

### 需求: 对话广场视频生成
**模块:** 网关与任务

#### 场景: 文生视频与单图生视频
- 对话广场选择 `video` 分组后使用 JWT 视频路由，不调用 Chat Completions。
- `grok-imagine-video` 支持文生视频；`grok-imagine-video-1.5-preview` 必须提供一张参考图。
- `grok-imagine-video*` 在对话广场显示 1-15 秒时长和常用画面比例选择，并将 `seconds`、`aspect_ratio` 透传到创建请求；其他视频模型不显示这组控件。
- 参考图通过 `input_reference.image_url` 透传，文档和多图不进入视频请求。
- 前端展示创建、排队、生成、完成或失败状态，完成时固定显示 100% 并使用原生播放器。
- 状态查询仍校验登录用户和分组权限，但不因创建扣费后余额归零而拒绝。
