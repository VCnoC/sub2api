# 架构设计

## 总体架构

```mermaid
flowchart LR
    Client[API 客户端] --> Gateway[Gin 网关]
    Playground[对话广场 JWT] --> Gateway
    Admin[Vue 管理端] --> AdminAPI[管理 API]
    Gateway --> Scheduler[分组调度]
    Scheduler --> Account[上游账号]
    Gateway --> Billing[计费服务]
    Gateway --> PromptAudit[提示词审计]
    AdminAPI --> AuditLog[管理操作审计]
    Billing --> PG[(PostgreSQL)]
    PromptAudit --> PG
    AuditLog --> PG
    Gateway --> VideoTask[视频任务]
    VideoTask --> PG
    Worker[视频任务 Worker] --> VideoTask
    Worker --> Account
    Scheduler --> Redis[(Redis)]
    UserUI[用户工单页] --> TicketAPI[工单 API]
    Admin --> TicketAPI
    TicketAPI --> TicketService[工单服务]
    TicketService --> PG
    TicketService --> TicketFiles[私有附件目录]
    TicketService --> Mail[通知邮件]
    UserUI --> LotteryAPI[抽奖 API]
    Admin --> LotteryAPI
    LotteryAPI --> LotteryService[抽奖服务]
    LotteryService --> PG
    LotteryService --> Billing
```

## 核心约束
- API Key 绑定分组，分组平台决定协议与账号池。
- Redis 负责并发、短期缓存和粘性会话；PostgreSQL 保存长期事实。
- 用量扣费通过 `usage_billing_dedup` 保证同一请求最多应用一次。
- 管理操作审计与提示词审计使用独立存储和权限边界；敏感管理员操作通过 step-up 2FA 再验证。
- 异步图片任务只在对象存储配置完整时启用，Redis 保存紧凑任务状态，图片结果写入 S3 兼容存储。
- 同组多张订阅独立计时和记账，鉴权按最早到期顺序选择当前仍有额度的具体 `subscription_id`。
- 视频扣余额与 `video_tasks` 写入在同一数据库事务中提交。
- Worker 使用数据库租约领取任务；未知状态和传输错误继续重试，只有上游明确失败终态才退款。
- 失败退款在数据库事务内锁定任务并更新余额、任务和用量记录，重复执行不重复退款。
- 对话广场通过持久化内部 API Key 复用同一视频调度、计费和退款链路；前端只负责创建、轮询和展示。
- 视频状态查询是已扣费任务的只读操作，保留身份与分组权限校验，不重复执行余额资格检查。
- 工单用户接口始终按登录用户附加所有权边界，越权的工单和附件统一返回不存在；管理员负责人只用于协作和邮件收件人选择，不形成独占权限。
- 工单正文和系统事件不可变，附件保存在非静态目录并通过鉴权接口访问；关闭 30 天后只物理清理附件，文字和删除元数据长期保留。
- 抽奖使用固定普通/豪华双奖池；服务端以百万分比安全随机决定结果，前端轮带只展示已确定结果。
- 抽奖在单个数据库事务内锁定次数与库存并完成余额或订阅发奖；邀请、兑换、充值及退款通过幂等流水发放或冲正额外次数。

## 重大架构决策

| adr_id | title | date | status | affected_modules | details |
|--------|-------|------|--------|------------------|---------|
| ADR-20260718-UPSTREAM-160-001 | 使用快照分支执行三方合并 | 2026-07-18 | ✅已实施 | 全局架构、版本管理 | [方案](../history/2026-07/202607181652_upstream_0_1_160_merge/how.md#adr-20260718-upstream-160-001-使用快照分支执行三方合并) |
| ADR-20260718-UPSTREAM-160-002 | 生成代码统一重建 | 2026-07-18 | ✅已实施 | Ent、Wire、依赖注入 | [方案](../history/2026-07/202607181652_upstream_0_1_160_merge/how.md#adr-20260718-upstream-160-002-生成代码统一重建) |
| ADR-20260718-MULTI-SUB-001 | 独立权益记录并由数据库选择候选 | 2026-07-18 | ✅已实施 | 订阅、鉴权、计费 | [方案](../history/2026-07/202607180325_multi_subscription_consumption/how.md#adr-20260718-multi-sub-001-独立权益记录并由数据库选择候选) |
| ADR-20260718-MULTI-SUB-002 | 每张订阅获得后立即计时 | 2026-07-18 | ✅已实施 | 订阅、用户端 | [方案](../history/2026-07/202607180325_multi_subscription_consumption/how.md#adr-20260718-multi-sub-002-每张订阅获得后立即计时) |
| ADR-20260718-MULTI-SUB-003 | 支付订单关联精确权益 | 2026-07-18 | ✅已实施 | 支付、退款、订阅 | [方案](../history/2026-07/202607180325_multi_subscription_consumption/how.md#adr-20260718-multi-sub-003-支付订单关联精确权益) |
| ADR-20260714-UPSTREAM-MERGE | 按平台保留 Grok/Video 双路由 | 2026-07-14 | ✅已实施 | 网关媒体、账号管理、模型同步 | [方案](../history/2026-07/202607141328_upstream_0_1_153_merge/how.md#adr-20260714-upstream-merge-按平台保留双路由) |
| ADR-004 | 在现有单体内建立工单模块 | 2026-07-12 | ✅已实施 | 工单、权限、邮件、私有附件 | [方案](../history/2026-07/202607120533_support_tickets/how.md#adr-004-在现有单体内建立工单模块) |
| ADR-VIDEO-001 | 视频任务持久化与余额补偿 | 2026-07-11 | ✅已实施 | 账号、分组、网关、计费 | [方案](../plan/202607110153_video_platform/how.md#adr-video-001-视频任务持久化与余额补偿) |
| ADR-20260711-PLAYGROUND-VIDEO | 对话广场复用视频网关 | 2026-07-11 | ✅已实施 | 对话广场、视频网关 | [方案](../history/2026-07/202607111841_playground_video/how.md#adr-20260711-playground-video-复用视频网关而非新建-playground-视频服务) |
| ADR-LOTTERY-001 | 使用领域专用表和固定事件类型 | 2026-07-12 | ✅已实施 | 抽奖、邀请、充值、兑换码 | [方案](../history/2026-07/202607121617_lottery_system/how.md#adr-lottery-001-使用领域专用表和固定事件类型) |
| ADR-LOTTERY-002 | 抽奖结果由服务端安全随机确定 | 2026-07-12 | ✅已实施 | 抽奖、前端 | [方案](../history/2026-07/202607121617_lottery_system/how.md#adr-lottery-002-抽奖结果由服务端安全随机确定) |
| ADR-LOTTERY-003 | 兑换码核心支持复用外层事务 | 2026-07-12 | ✅已实施 | 抽奖、兑换码、订阅 | [方案](../history/2026-07/202607121617_lottery_system/how.md#adr-lottery-003-兑换码核心支持复用外层事务) |
