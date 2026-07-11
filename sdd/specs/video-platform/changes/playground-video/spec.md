---
id: "video-platform/playground-video"
status: completed
impact_radius:
  - "backend/internal/server/middleware/playground_context.go"
  - "backend/internal/server/routes/playground.go"
  - "frontend/src/components/playground"
  - "frontend/src/composables/playground"
dependencies:
  - "video-platform"
---

# Specification: 对话广场视频生成 (Specification)

## 1. Scope
- **In Scope**: video 分组模型选择、文生视频、单图图生视频、任务轮询、进度、播放、错误显示、会话保存。
- **Out of Scope**: 视频编辑/延长 UI、多参考图、视频专属时长和分辨率控件、新计费或退款实现、自动部署。

## 2. Functional Requirements
### 2.1 视频发送
- **Trigger**: 当前分组 platform 为 `video`，用户提交输入。
- **UI/UX**: 沿用现有输入和消息列表；1.5 模型必须有一张图片。
- **Logic**: 文本传 prompt，首张图片传 `input_reference.image_url`，忽略文档和多余图片。

### 2.2 状态与结果
- **Trigger**: 创建响应包含任务 ID。
- **UI/UX**: 生成中显示进度，成功后显示原生视频播放器，失败显示错误。
- **Logic**: 每 2 秒轮询，最长 10 分钟；停止只取消前端等待。

## 3. Acceptance Checklist
- [x] video 分组不会调用 Chat Completions
- [x] 普通模型可文生视频
- [x] 1.5 模型无图时给出错误，有图时传正确字段
- [x] 任务进度与成功视频可在消息中恢复和展示
- [x] 明确失败显示上游错误且不在前端操作余额
- [x] 后端测试、前端类型检查和生产构建通过
