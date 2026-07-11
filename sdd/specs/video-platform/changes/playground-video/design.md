# Design: 对话广场视频生成 (Design)

## 1. Architecture
Playground JWT 请求经现有中间件转换为持久化虚拟 API Key，再交给既有 Grok 视频处理器。

## 2. Data Model & Interfaces
- 新增 Playground 视频创建与状态查询 API。
- `MessageVersion` 保存视频任务 ID、URL 与进度，无数据库迁移。

## 3. Data Flow & Interaction
1. 用户选择 video 分组、模型并输入提示词或附一张图。
2. 前端创建视频任务并每 2 秒查询。
3. 后端复用现有调度、扣费与任务记录。
4. 成功后消息展示播放器；失败显示上游错误，后台按既有流程退款。

## 4. Error Handling
- 1.5 模型没有参考图时前端拒绝提交。
- 查询网络错误显示为消息错误；用户停止会取消轮询。
- 只有上游明确失败才由后端退款，前端不操作余额。
- 状态查询不重复执行余额资格检查，避免创建扣费后余额归零导致轮询中断。
