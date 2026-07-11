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
