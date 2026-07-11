---
id: "video-platform"
status: in_progress
impact_radius:
  - "backend/internal/domain"
  - "backend/internal/service"
  - "backend/internal/repository"
  - "backend/internal/handler"
  - "backend/migrations"
  - "frontend/src/components/account"
  - "frontend/src/views/admin"
dependencies:
  - "cpa-fan OpenAI-compatible Grok Videos API"
---

# Specification: 独立视频平台与失败退款 (Specification)

## 1. Scope
- **In Scope**: 独立 `video` 平台、Base URL/API Key、模型同步、完整视频接口、分组按秒/按次价格、后台跟踪、余额失败退款、管理端配置。
- **Out of Scope**: 文本/图片调度、Grok OAuth 行为变更、订阅额度退款、支付渠道退款。

## 2. Functional Requirements

### 2.1 账号与模型
- **Trigger**: 管理员在添加账号中选择“视频”。
- **UI/UX**: 输入 Base URL、API Key，选择同平台分组，并可从 `/v1/models` 同步模型。
- **Logic**: 账号类型固定为 API Key，只参与 `video` 平台调度。

### 2.2 视频接口
- **Trigger**: 客户端调用 `/v1/videos`、`/v1/video/create`、`/v1/videos/generations`、`/v1/videos/edits`、`/v1/videos/extensions`、`/v1/videos/:id` 或 `/v1/video/query`。
- **UI/UX**: 保持上游兼容 JSON 与 HTTP 状态。
- **Logic**: 创建类接口计费并建任务；查询类接口不重复计费。

### 2.3 计费与退款
- **Trigger**: 创建任务返回有效任务 ID。
- **UI/UX**: 分组选择按秒或按次，分别填写 480p、720p 价格；创建账号时可批量写入所选分组。
- **Logic**: 按秒为价格乘时长，按次忽略时长；上游明确失败时只退用户余额，幂等执行。

## 3. Acceptance Checklist
- [ ] 添加账号中出现独立视频平台
- [ ] Base URL、API Key 和模型同步可用
- [ ] 视频平台不参与文本与图片调度
- [ ] 六类视频操作及统一查询兼容 cpa-fan
- [ ] 480p、720p 支持按秒和按次价格
- [ ] 后台重启后仍可继续跟踪未完成任务
- [ ] 失败任务余额自动且只退款一次
- [ ] 前端类型检查、后端测试和生产构建通过
- [ ] 生产迁移、健康检查和接口冒烟验证通过
