# 项目技术约定

## 技术栈
- **官方基线:** Sub2API v0.1.163
- **后端:** Go 1.26、Gin、Ent、PostgreSQL、Redis
- **前端:** Vue 3、TypeScript、Vite、Pinia、Tailwind CSS
- **部署:** Docker Compose

## 开发约定
- 延续现有领域常量、仓储、服务与 Wire 依赖注入模式。
- 数据库结构同时维护 Ent schema 与顺序 SQL migration。
- 本地 183-187 已用于生产迁移历史；官方 v0.1.163 的三个新增迁移固定使用 188-190，禁止回退编号或修改既有 migration checksum。
- 用户上传文件必须写入非静态私有目录，以随机存储键落盘并通过鉴权 Handler 读取；原始文件名只能用于展示。
- 平台能力通过分组平台隔离，账号只能参与相同平台的调度。
- 金额计算使用现有 BillingService，账务写入必须幂等。

## 错误与日志
- 上游错误对客户端脱敏，详细信息写入结构化日志。
- API Key、访问令牌和服务器地址不得写入知识库或日志正文。

## 测试与流程
- Ent/Wire 源定义发生变化后必须运行项目现有生成命令，并以生成结果作为唯一来源。
- 后端运行受影响包测试及 migration 回归测试。
- 前端运行类型检查、相关 Vitest 与生产构建。
- PostgreSQL 并发或权限边界使用 `integration` build tag 的临时容器测试验证。
- 发布前检查干净工作树、数据库备份和容器健康状态。
