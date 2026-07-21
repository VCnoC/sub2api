---
id: "team-governance"
status: completed
impact_radius:
  - "backend/migrations"
  - "backend/internal/repository"
  - "backend/internal/service"
  - "backend/internal/handler"
  - "backend/internal/server/routes"
  - "frontend/src/api"
  - "frontend/src/views/user"
  - "frontend/src/views/admin"
  - "frontend/src/components/layout"
dependencies:
  - "existing team and user authentication"
  - "payment_orders, redeem_codes and usage_logs"
---

# Specification: 团队治理与资金防滥用 (Specification)

## 1. Scope
- **In Scope**: 创建审核、加入审批、5/15/40 等级、超 40 人扩容申请、管理员完整团队管理、可转赠余额来源限制、现有团队兼容。
- **Out of Scope**: 自动定时升级、自动降级、团队 owner 转让、团队删除、外部风控服务、回溯拆分上线前余额来源。

## 2. Functional Requirements

### 2.1 创建团队
- **Trigger**: 未加入团队的用户填写团队名称和申请说明。
- **UI/UX**: 页面展示当前注册天数、有效累计充值和管理员配置门槛；提交后展示申请状态及审核原因，被拒绝后可立即补充说明重新申请。
- **Logic**: 创建必须由管理员批准。门槛为最低注册天数和最低累计充值；累计充值只包含已完成真实支付和已使用余额兑换码。管理员可豁免，但必须填写原因并记录审核人、时间。

### 2.2 加入团队
- **Trigger**: 用户输入有效邀请码提交申请。
- **UI/UX**: 申请人可查看待处理状态；owner 可查看申请人并批准或拒绝。
- **Logic**: 提交申请不立即加入。批准时再次校验申请人未加入其他团队、团队正常且成员数低于上限。

### 2.3 等级与自助升级
- **Trigger**: owner 点击“检查升级”。
- **UI/UX**: 展示当前等级、人数上限、全团队有效累计充值、近 7 天消费和下一档门槛。
- **Logic**: 固定 5、15、40 三档。每档累计充值、近 7 天消费及 AND/OR 由管理员配置；指标按团队全员汇总，消费使用倍率换算后的 `actual_cost`，订阅同样按 `actual_cost`。一次点击直接升到满足的最高档，只升级不自动降级。

### 2.4 超 40 人扩容
- **Trigger**: 当前上限至少 40 的 owner 填写目标人数和理由。
- **UI/UX**: 管理员可查看申请指标并调整目标人数后批准或拒绝。
- **Logic**: 超 40 人不能通过等级自动获得，必须管理员批准。管理员始终可直接修改任意团队人数上限，不受条件限制。

### 2.5 管理员团队管理
- **Trigger**: 管理员进入“团队管理”。
- **UI/UX**: 展示团队总数、待处理申请数、等级、成员数/上限、创建者、资金余额、有效充值和近 7 天消费；详情展示成员、资金流水、创建/扩容申请。
- **Logic**: 管理员可审核创建和扩容、冻结/恢复、移除成员、修改单团队上限、标记现有团队复审完成。团队实体状态仅正常/冻结，申请状态独立为 pending/approved/rejected。

### 2.6 团队资金来源
- **Trigger**: 成员存入团队资金、owner 直接转赠或从资金池分配。
- **UI/UX**: 存入和转赠表单展示可转赠余额，额度不足时显示明确错误。
- **Logic**: 真实支付、余额兑换码、管理员赠送可转赠；注册赠送、邀请奖励、抽奖奖励和团队转入不可转赠。团队分配到账不增加可转赠额度，不能再次转赠。上线时现有余额全部初始化为可转赠，之后按新增来源区分。

### 2.7 现有团队兼容
- **Trigger**: migration 执行后首次访问现有团队。
- **UI/UX**: 现有成员、owner、邀请码和资金保持可见。
- **Logic**: 现有团队人数上限不得低于当前成员数，并标记待复审；管理员复审前禁止 owner 自助升级或申请扩容。管理员直接修改上限不受影响。

## 3. Acceptance Checklist
- [x] 创建申请未批准前不会产生团队或 owner 关系
- [x] 创建门槛只统计真实支付和余额兑换码，豁免必须填写原因并审计
- [x] 邀请码加入改为 owner 审批，并发批准不会超过人数上限
- [x] 5/15/40 条件、AND/OR 与最高可达档计算正确
- [x] 团队消费按近 7 天 `actual_cost` 汇总，余额和订阅倍率口径一致
- [x] owner 主动升级且只升不降，复审中的现有团队不能自助扩容
- [x] 超 40 人申请可由管理员调整目标人数后批准
- [x] 管理员可查看全部团队、成员、申请、指标和资金流水
- [x] 管理员可冻结/恢复、移除成员并直接修改任意团队人数上限
- [x] 冻结团队无法进行加入、扩容和资金操作
- [x] 存入和直接转赠不能超过可转赠额度
- [x] 团队转入余额不会增加可转赠额度，不能循环归集
- [x] 现有余额初始化为可转赠，现有团队成员和余额保持不变
- [x] 用户端和管理端在桌面及移动端无溢出、重叠或不可操作控件
- [x] 后端测试、前端组件测试、TypeScript 和生产构建通过

## 4. Deployment Verification
- 2026-07-21 部署至海外 D 散户站 `sub.vcnovb.cn`，镜像 `sub2api-local:team-governance-20260721` (`sha256:1b2c29556bef8f92232fc9d82c3dbfc108b4ab606a9c00697f1b23c64ef94059`)。
- 生产库备份位于 `/opt/sub2api-sub/backups/20260721-1334-team-governance/`，SHA-256 校验通过。
- migration `187_team_governance.sql` 已登记；容器、PostgreSQL、Redis 和公网 `/health` 均正常，25 个现有团队全部标记待复审。
