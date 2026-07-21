# 任务清单: 团队治理与资金防滥用

目录: `helloagents/plan/202607210307_team_governance/`

---

## 1. 数据与核心服务
- [√] 1.1 增加团队治理 migration 和回归测试，验证 why.md#需求-受控创建与加入-场景-并发批准加入
- [√] 1.2 实现 `TeamGovernanceRepository` 的配置、申请、指标、管理和事务额度操作，依赖任务1.1
- [√] 1.3 扩展 `TeamService` 实现创建审核、加入审批、等级升级、扩容和管理员治理，依赖任务1.2

## 2. API 与前端
- [√] 2.1 扩展用户 Handler/路由并新增管理员 Team Handler/路由，依赖任务1.3
- [√] 2.2 接入 Repository、Service、Handler 的 Wire 依赖注入，依赖任务2.1
- [√] 2.3 扩展用户 TeamView 与团队 API，完成申请、审批、升级和可转赠余额交互，依赖任务2.1
- [√] 2.4 新增管理端团队页面、路由、侧边栏和中英文文案，依赖任务2.1

## 3. 安全检查
- [√] 3.1 执行安全检查：主体鉴权、IDOR、SQL 参数化、并发超员、余额重复扣减、冻结状态和审计完整性

## 4. 文档更新
- [√] 4.1 更新 `helloagents/wiki/api.md`、`helloagents/wiki/data.md`、团队模块文档和 `helloagents/CHANGELOG.md`
- [√] 4.2 更新 SDD 验收项、任务状态和 `sdd/project.md`

## 5. 测试
- [√] 5.1 运行 migration、repository 和 TeamService 相关 Go 测试
- [√] 5.2 运行前端 TypeScript、ESLint 和生产构建

## 6. 生产部署
- [√] 6.1 备份海外 D 散户站 PostgreSQL 并校验 SHA-256
- [√] 6.2 以 `144a75ba7` 为干净基线构建团队治理镜像，未混入未提交的次数订阅与抽奖日期改动
- [√] 6.3 仅切换 `sub2api-sub` 应用容器，验证 migration 187、新路由鉴权、公网健康和既有请求

> 部署记录: 镜像 `sub2api-local:team-governance-20260721`，回滚标签 `sub2api-local:rollback-before-team-governance-20260721`。Sun 容灾复制槽在验收时为 inactive，且 Sun SSH 超时，需独立恢复容灾链路。
