---
id: "api-key-group-failover"
status: completed
impact_radius:
  - "backend/ent/schema"
  - "backend/migrations"
  - "backend/internal/repository"
  - "backend/internal/service"
  - "backend/internal/server/middleware"
  - "backend/internal/handler"
  - "frontend/src/views/user/KeysView.vue"
  - "frontend/src/api"
dependencies:
  - "api-key-auth-cache"
  - "multi-subscription-consumption"
  - "gateway-account-failover"
---

# Specification: API Key 有序多分组容灾 (Specification)

## 1. Scope
- **In Scope**: 每 Key 1-5 个同平台有序分组、拖拽排序、余额/订阅资格跳组、组内账号耗尽后的安全跨组 failover、实际分组计费、旧字段兼容。
- **Out of Scope**: 跨平台协议转换、响应输出后的重放、异步任务发送后的跨组重试、修改全局 `fallback_group_id` 语义、自动学习或动态调整用户顺序。

## 2. Functional Requirements

### 2.1 候选分组配置
- **Trigger**: 用户创建或编辑 API Key。
- **UI/UX**: 从有权限分组中多选 1-5 个；拖拽排序；显示序号、平台和计费类型。
- **Logic**: 所有分组必须属于同一平台且去重；数组顺序就是严格优先级。

### 2.2 向后兼容
- **Trigger**: 迁移旧数据或旧客户端提交 `group_id`。
- **UI/UX**: 旧 Key 编辑时正常显示为单项候选链。
- **Logic**: 非空旧 `group_id` 回填到优先级 0；旧字段写入转为单分组链；响应继续返回首组 `group_id`。

### 2.3 计费资格推进
- **Trigger**: 新请求进入 API Key 鉴权。
- **UI/UX**: 无额外交互。
- **Logic**: 每次从第一组开始；余额不足、无有效订阅、订阅耗尽/过期或分组停用时尝试下一组。订阅组内部按最早到期且仍有额度的订阅消费。

### 2.4 上游故障推进
- **Trigger**: 当前组账号不可调度，或账号 failover 耗尽。
- **UI/UX**: 保持现有兼容错误响应。
- **Logic**: 连接错误、超时、429、5xx 可推进；参数、权限、客户端取消不可推进。当前组内部始终先执行既有账号/凭证 failover。

### 2.5 重放与计费边界
- **Trigger**: 流式响应、非流式响应或异步任务提交。
- **UI/UX**: 无重复内容或重复任务。
- **Logic**: 响应已提交或异步任务可能已创建后禁止跨组；最终成功组决定余额/订阅扣费和用量日志归属。

## 3. Acceptance Checklist
- [x] API Key 可选择并拖拽排序 1-5 个同平台分组
- [x] 重复、跨平台、无权限和超过 5 个分组被服务端拒绝
- [x] 旧 `group_id` 数据和客户端保持兼容
- [x] 每个新请求从第一优先级重新判断
- [x] 余额不足和订阅耗尽/过期会推进下一组
- [x] 订阅组内部仍按最早到期且有额度的订阅消费
- [x] 无账号、连接错误、超时、429、5xx 在组内耗尽后推进下一组
- [x] 参数错误、权限错误和客户端取消不推进
- [x] 响应已提交或异步任务可能已创建时不重放
- [x] 用量和计费记录实际成功的分组、订阅和账号
- [x] Chat Completions、Responses、Gemini 和通用网关回归通过
- [x] 前端测试、类型检查、后端测试和生产构建通过
