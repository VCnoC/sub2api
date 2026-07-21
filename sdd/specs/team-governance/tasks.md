# Task Breakdown & Execution Board: 团队治理与资金防滥用 (Tasks)

## Phase 1: 数据与仓储
- [x] Task 1.1: 增加团队治理、申请、加入审批、可转赠额度和资金流水 migration
- [x] Task 1.2: 使用原生 SQL 实现治理设置、申请、团队列表、指标、审批和额度事务仓储
- [x] Task 1.3: 增加 migration 与仓储关键约束测试

## Phase 2: 核心业务
- [x] Task 2.1: 创建团队改为申请和管理员审批，接入注册天数与有效充值门槛
- [x] Task 2.2: 邀请码加入改为 owner 审批，并发校验人数上限
- [x] Task 2.3: 实现 5/15/40 条件计算、owner 主动升级和超 40 扩容申请
- [x] Task 2.4: 实现管理员团队查询、冻结、移除成员、复审和直接修改上限
- [x] Task 2.5: 限制存入/转赠可用来源，并阻止团队转入资金再次转赠

## Phase 3: API 与依赖注入
- [x] Task 3.1: 扩展用户 Team Handler 与路由，提供申请、加入审批、升级和扩容接口
- [x] Task 3.2: 新增 Admin Team Handler 与路由，提供完整管理接口
- [x] Task 3.3: 接入 Repository、Service 和 Handler Wire provider

## Phase 4: 用户端与管理端
- [x] Task 4.1: 扩展团队 API 类型和用户 TeamView 的申请、审批、升级及可转赠余额交互
- [x] Task 4.2: 新增独立管理端团队页面、详情组件、配置与审核交互
- [x] Task 4.3: 注册管理端路由、侧边栏及中英文文案

## Phase 5: Integration & Refinement
- [x] Task 5.1: 完成业务、权限、并发、资金来源与 migration 回归测试
- [x] Task 5.2: 完成前端类型检查、后端测试和生产构建
- [x] Task 5.3: 更新知识库、CHANGELOG、SDD 状态并迁移 HelloAGENTS 方案包
