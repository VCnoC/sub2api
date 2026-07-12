# Task Breakdown & Execution Board: 站内工单系统 (Tasks)

> 开发、自动化验证、生产备份、迁移和健康检查已完成；真实用户流程、SMTP 和清理任务冒烟待验收，因此规格保持 `in_progress`。

## Phase 1: Domain, Schema & Migration
- [x] Task 1.1: 在 `backend/internal/domain/` 定义工单状态、分类、优先级、消息类型和附件限制常量
- [x] Task 1.2: 在 `backend/ent/schema/` 增加 Ticket、TicketMessage、TicketAttachment、TicketRead schema 及用户/管理员关联
- [x] Task 1.3: 生成 Ent 代码并在 `backend/migrations/` 增加仅向前的新表、外键和查询索引迁移
- [x] Task 1.4: 增加 schema 与迁移测试，验证唯一约束、级联策略、时间类型和索引

## Phase 2: Repository & Lifecycle
- [x] Task 2.1: 在 `backend/internal/service/` 定义工单实体、过滤器、分页结果和 Repository 接口
- [x] Task 2.2: 在 `backend/internal/repository/` 实现用户隔离列表、管理员筛选、详情和消息分页查询
- [x] Task 2.3: 实现创建事务，并通过用户行锁原子检查 5 个未关闭工单和每日 10 个新工单限制
- [x] Task 2.4: 实现用户回复、管理员公开回复、内部备注、关闭、自动重开和不可变系统事件
- [x] Task 2.5: 实现负责人、优先级、个人已读游标和未读汇总逻辑
- [x] Task 2.6: 增加生命周期、并发限流、权限隔离、负责人失效和未读计算测试

## Phase 3: Private Attachments & Cleanup
- [x] Task 3.1: 实现 multipart 限制、扩展名/MIME/内容校验、随机存储键和失败回滚清理
- [x] Task 3.2: 在 `/app/data/ticket-attachments` 实现私有文件存储及鉴权下载，设置安全响应头和目录权限
- [x] Task 3.3: 实现管理员附件删除、关闭后 30 天 `delete_after` 和重开取消未到期清理
- [x] Task 3.4: 按现有后台服务模式实现分批、幂等、可重试的 `TicketAttachmentCleanupService`
- [x] Task 3.5: 增加超限、伪造类型、路径穿越、越权下载、删除重试和重开保留测试

## Phase 4: API, Email & Wiring
- [x] Task 4.1: 在 `backend/internal/handler/` 实现用户工单 DTO/Handler 和统一错误码
- [x] Task 4.2: 在 `backend/internal/handler/admin/` 实现管理员筛选、回复、备注、认领、优先级和关闭 Handler
- [x] Task 4.3: 在 `backend/internal/server/routes/` 注册用户与管理员路由，并在 `wire.go` 注入仓储、服务和清理任务
- [x] Task 4.4: 扩展 `NotificationEmailService` 工单事件与中英文模板，复用投递幂等键
- [x] Task 4.5: 实现新工单通知全部管理员、认领后通知负责人以及用户回复/关闭通知，邮件失败只记录日志
- [x] Task 4.6: 增加 API 契约、鉴权、邮件收件人回退、邮件失败降级和重复投递测试

## Phase 5: User UI
- [x] Task 5.1: 在 `frontend/src/types/` 和 `frontend/src/api/` 增加工单类型、分页、multipart 和附件下载 API
- [x] Task 5.2: 在 `frontend/src/views/user/` 实现工单列表、新建表单和双向对话详情
- [x] Task 5.3: 实现附件选择、数量/大小前置校验、上传状态、过期附件元数据和错误状态
- [x] Task 5.4: 增加工单未读 store，在用户侧边栏接入稳定尺寸角标、页面可见刷新和焦点刷新
- [x] Task 5.5: 增加中文/英文文案、响应式布局、键盘操作和组件测试

## Phase 6: Admin UI
- [x] Task 6.1: 在 `frontend/src/views/admin/` 实现紧凑工单表格、分页及状态/分类/优先级/负责人/关键词筛选
- [x] Task 6.2: 实现管理员详情、公开回复、内部备注、认领/转交、优先级和关闭操作
- [x] Task 6.3: 实现敏感附件删除确认、原因输入和不可变审计事件展示
- [x] Task 6.4: 在管理员侧边栏接入个人未读角标，并增加中英文文案和交互测试

## Phase 7: Verification, Documentation & Delivery
- [x] Task 7.1: 完成后端单元/集成测试、前端组件测试、TypeScript、ESLint 和生产构建
- [x] Task 7.2: 执行安全检查：IDOR、存储型 XSS、上传绕过、路径穿越、敏感日志、并发限流和越权附件访问
- [x] Task 7.3: 更新 `helloagents/wiki/api.md`、`helloagents/wiki/data.md`、模块文档和 `helloagents/CHANGELOG.md`
- [x] Task 7.4: 备份数据库和 `/app/data`，执行迁移，验证持久化目录权限和生产健康状态
- [ ] Task 7.5: 部署后验证用户/管理员完整流程、SMTP 和清理任务，并将 `spec.md` 与 `sdd/project.md` 状态更新为 `completed`
