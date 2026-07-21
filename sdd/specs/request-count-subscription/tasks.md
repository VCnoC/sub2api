# Task Breakdown & Execution Board: 纯次数订阅套餐 (Tasks)

## Phase 1: 数据模型与迁移
- [x] Task 1.1: 扩展 Group 和 UserSubscription Ent schema
- [x] Task 1.2: 新增 SubscriptionRequestReservation Ent schema
- [x] Task 1.3: 增加 186 migration、索引、约束和迁移测试
- [x] Task 1.4: 运行 Ent 生成并审计机械变更

## Phase 2: 分组与订阅核心逻辑
- [x] Task 2.1: 扩展分组领域模型、管理 API 和 Repository 映射
- [x] Task 2.2: 实现 USD/次数模式校验与兼容默认值
- [x] Task 2.3: 实现事务性 reserve/commit/release 和过期占位回收
- [x] Task 2.4: 接入最早到期多订阅选择与次数耗尽错误

## Phase 3: Gateway 与统一计费
- [x] Task 3.1: 建立统一文本请求分类并排除媒体/只读端点
- [x] Task 3.2: 账号重试复用占位，跨组 failover 释放并重新占位
- [x] Task 3.3: Messages、Chat Completions、Responses、Gemini 接入占位生命周期
- [x] Task 3.4: usage billing 幂等事务确认成功占位，次数模式跳过 USD 扣减
- [x] Task 3.5: 覆盖流式正常结束和客户端中断语义

## Phase 4: UI & Interaction
- [x] Task 4.1: 管理端分组表单增加计费模式和次数上限
- [x] Task 4.2: 套餐预览和购买卡片展示次数权益
- [x] Task 4.3: 用户订阅进度展示已用、剩余和重置时间
- [x] Task 4.4: 补充中英文文案、类型和组件测试

## Phase 5: Integration & Refinement
- [x] Task 5.1: 完成并发、幂等、失败释放、多卡和跨窗口 PostgreSQL 集成测试
- [x] Task 5.2: 完成 Gateway、failover、支付订阅与 USD 套餐回归测试
- [x] Task 5.3: 完成前端测试、类型检查和生产构建
- [x] Task 5.4: 更新知识库、CHANGELOG、SDD 状态并迁移 HelloAGENTS 方案包
