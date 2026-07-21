---
id: "request-count-subscription"
status: completed
impact_radius:
  - "backend/ent/schema"
  - "backend/migrations"
  - "backend/internal/repository"
  - "backend/internal/service"
  - "backend/internal/server/middleware"
  - "backend/internal/handler"
  - "frontend/src/views/admin/GroupsView.vue"
  - "frontend/src/components/payment"
  - "frontend/src/types"
dependencies:
  - "multi-subscription-consumption"
  - "api-key-group-failover"
  - "usage-billing-dedup"
---

# Specification: 纯次数订阅套餐 (Specification)

## 1. Scope
- **In Scope**: 分组级纯次数计费模式、5小时/24小时首次成功调用起算窗口、仅文本生成请求计次、数据库精确占位、成功确认、失败释放、多订阅和多分组兼容、管理端配置与用户进度。
- **Out of Scope**: 任意自定义窗口、不同模型扣不同次数、按 Token 与次数混合扣费、Redis 计数、独立次卡商品系统、人工占位管理页面。

## 2. Functional Requirements

### 2.1 分组与套餐配置
- **Trigger**: 管理员创建或编辑订阅分组。
- **UI/UX**: 可选择 USD 或次数计费；次数模式填写 5 小时和 24 小时上限，套餐创建仍选择该分组。
- **Logic**: 上限为非负整数，0 表示该窗口不限，两个窗口不能同时为 0；次数模式不使用 USD 限额。

### 2.2 文本请求范围
- **Trigger**: Messages、Chat Completions、Responses 或 Gemini 文本生成请求。
- **UI/UX**: 无额外交互。
- **Logic**: 同步、流式均参与；count_tokens、图片、视频、批量任务、模型列表、用量和任务查询不参与。

### 2.3 请求占位与窗口
- **Trigger**: 次数分组的文本请求准备发送上游。
- **UI/UX**: 次数耗尽返回 429 和重置时间。
- **Logic**: 首次占位启动连续 5 小时和 24 小时窗口；占位事务选择最早到期且可用的具体订阅；任一启用窗口无容量时跳过该订阅。

### 2.4 成功确认与失败释放
- **Trigger**: 上游请求结束或进入跨组 failover。
- **UI/UX**: 用户最终只为成功调用消耗次数。
- **Logic**: 成功响应或已开始成功响应后的客户端中断确认 1 次；校验失败、上游超时、4xx/5xx 和未成功发送释放；账号重试不重复占位，跨组释放后重新占位。

### 2.5 纯次数结算
- **Trigger**: 成功请求进入统一 usage billing。
- **UI/UX**: 订阅进度显示次数而不是 USD。
- **Logic**: 占位确认与 billing 幂等键同事务；不增加 `daily_usage_usd` 等订阅 USD 用量；使用日志仍保留成本统计。

### 2.6 多订阅与恢复
- **Trigger**: 用户有多张同组次数订阅，或服务进程异常。
- **UI/UX**: 用户逐张查看权益和重置时间。
- **Logic**: 最早到期可用订阅优先；耗尽后切下一张；过期 pending 在下一次占位时回收，释放只作用于相同窗口代次。

## 3. Acceptance Checklist
- [x] 管理员可创建和编辑次数订阅分组，并通过现有套餐页面绑定该分组
- [x] USD 与次数模式字段互斥且后端严格校验
- [x] 5小时和24小时窗口从首次成功文本调用起算
- [x] 同步和流式文本成功请求各扣 1 次
- [x] 失败、超时和上游4xx/5xx最终不占次数
- [x] 上游已响应后的客户端主动中断仍扣 1 次
- [x] 同账号/换账号重试不重复扣次，跨组 failover 正确释放和重新占位
- [x] 并发请求不会超过任一窗口上限
- [x] 多张订阅按最早到期且有次数的记录消费
- [x] count_tokens、图片、视频、批量任务和只读接口不扣次数
- [x] 次数模式不消耗订阅 USD 额度，现有 USD 套餐行为不变
- [x] 用户可查看已用、剩余和重置时间
- [x] migration、后端测试、前端测试、类型检查和生产构建通过
