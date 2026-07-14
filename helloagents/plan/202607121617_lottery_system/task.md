# 任务清单: 双奖池邀请抽奖系统

目录: `helloagents/plan/202607121617_lottery_system/`

---

## 1. 数据模型与领域契约
- [ ] 1.1 在 `backend/migrations/176_lottery_system.sql` 创建并约束六张抽奖领域表、索引及普通/豪华种子数据，验证 why.md#需求-双奖池和周期次数-场景-周期刷新
- [ ] 1.2 在 `backend/migrations/lottery_migration_test.go` 增加迁移结构、约束和重复执行测试，依赖任务 1.1
- [ ] 1.3 在 `backend/internal/domain/lottery.go` 定义奖池、奖品、规则、次数、流水、抽奖记录及枚举，验证 why.md#核心规则
- [ ] 1.4 在 `backend/internal/service/lottery.go` 定义输入输出、repository 接口、错误码和图片/概率/规则校验，依赖任务 1.3

## 2. Repository 与周期次数
- [ ] 2.1 在 `backend/internal/repository/lottery_repo.go` 实现奖池、奖品、规则和用户摘要查询，复用 `clientFromContext`，依赖任务 1.1、1.4
- [ ] 2.2 在 `backend/internal/repository/lottery_repo.go` 实现次数账户行锁、惰性周期刷新、流水幂等和抽奖记录事务写入，依赖任务 2.1
- [ ] 2.3 在 `backend/internal/repository/lottery_repo_integration_test.go` 验证日/周刷新、额外次数长期保留、重复来源幂等和并发不超扣，依赖任务 2.2
- [ ] 2.4 在 `backend/internal/service/lottery_period.go` 与 `backend/internal/service/lottery_period_test.go` 复用站点时区计算每日/每周周期键及边界测试，依赖任务 1.4
- [ ] 2.5 在 `backend/internal/repository/wire.go` 注册 `LotteryRepository` provider，依赖任务 2.1

## 3. 邀请奖励规则
- [ ] 3.1 在 `backend/internal/service/lottery_chance_service.go` 实现注册、首次兑换、单笔充值和累计充值规则计算，依赖任务 2.2
- [ ] 3.2 在 `backend/internal/service/lottery_chance_service.go` 实现退款后单笔/累计档位冲正，仅扣可用额外次数并记录不足，依赖任务 3.1
- [ ] 3.3 在 `backend/internal/service/lottery_chance_service_test.go` 覆盖多规则叠加、双奖池目标、首次有效、门槛倍数、重复事件和退款不足，依赖任务 3.2
- [ ] 3.4 在 `backend/internal/service/affiliate_service.go` 接入邀请关系首次绑定后的注册规则事件，并保持注册主流程现有容错语义，依赖任务 3.1
- [ ] 3.5 在 `backend/internal/service/affiliate_service_test.go` 覆盖邀请人/被邀请人奖励、自邀拒绝和重复绑定不重复发放，依赖任务 3.4

## 4. 统一兑换码事务入口
- [ ] 4.1 在 `backend/internal/repository/redeem_code_repo.go` 让创建、读取、使用兑换码统一选择 `clientFromContext`，依赖任务 1.4
- [ ] 4.2 在 `backend/internal/repository/redeem_code_repo_integration_test.go` 验证外层事务提交及回滚均正确覆盖兑换码，依赖任务 4.1
- [ ] 4.3 在 `backend/internal/service/redeem_service.go` 提取可复用外层事务的兑换核心，并增加系统订阅码创建/兑换入口及提交后缓存失效结果，依赖任务 4.1
- [ ] 4.4 在 `backend/internal/service/redeem_service.go` 将用户首次成功兑换事件接入 `LotteryChanceService`，同时跳过抽奖内部订阅码，依赖任务 3.1、4.3
- [ ] 4.5 在 `backend/internal/service/redeem_service_redeem_test.go` 覆盖公开兑换回归、外层事务回滚、订阅续期、首次事件和内部码跳过，依赖任务 4.4

## 5. 充值与退款接入
- [ ] 5.1 在 `backend/internal/service/payment_fulfillment.go` 于余额/订阅订单完成路径接入充值规则，并使用订单站内金额和订单 id，依赖任务 3.1
- [ ] 5.2 在 `backend/internal/service/payment_fulfillment_test.go` 覆盖履约重试、单笔/累计规则及相同订单幂等，依赖任务 5.1
- [ ] 5.3 在 `backend/internal/service/payment_refund.go` 于退款最终成功路径执行机会冲正并写支付审计，依赖任务 3.2
- [ ] 5.4 在 `backend/internal/service/payment_refund_test.go` 覆盖全额/部分退款、重复完成通知、已消费差额和累计档位下降，依赖任务 5.3

## 6. 抽奖和自动发奖
- [ ] 6.1 在 `backend/internal/service/lottery_service.go` 实现用户摘要、历史及安全随机区间选择，确保缺省概率和无库存区间为未中奖，依赖任务 2.2
- [ ] 6.2 在 `backend/internal/service/lottery_service.go` 实现原子抽奖：锁次数、优先扣基础次数、占库存、余额到账或订阅码兑换、记录快照及统一回滚，依赖任务 4.3、6.1
- [ ] 6.3 在 `backend/internal/service/lottery_service_test.go` 覆盖概率边界、活动时间、无次数、余额奖品、订阅奖品和任一步骤失败回滚，依赖任务 6.2
- [ ] 6.4 在 `backend/internal/repository/lottery_repo_integration_test.go` 增加同用户并发抽奖及固定库存最后一份的竞争测试，依赖任务 6.2
- [ ] 6.5 在 `backend/internal/service/lottery_admin_service.go` 实现奖池更新、奖品/规则维护、软停用和审计分页，依赖任务 2.1
- [ ] 6.6 在 `backend/internal/service/lottery_admin_service_test.go` 覆盖概率总和、类型互斥字段、固定奖池、图片安全校验及历史引用保护，依赖任务 6.5

## 7. 后端 API 与依赖注入
- [ ] 7.1 在 `backend/internal/handler/lottery_handler.go` 实现用户摘要、幂等抽奖和本人历史接口及 DTO 校验，依赖任务 6.2
- [ ] 7.2 在 `backend/internal/handler/admin/lottery_handler.go` 实现奖池、奖品、规则 CRUD 和记录分页接口，依赖任务 6.5
- [ ] 7.3 在 `backend/internal/handler/lottery_handler_test.go` 与 `backend/internal/handler/admin/lottery_handler_test.go` 覆盖输入、权限、分页和错误响应，依赖任务 7.1、7.2
- [ ] 7.4 在 `backend/internal/server/routes/user.go` 与 `backend/internal/server/routes/admin.go` 注册用户及管理员抽奖路由，依赖任务 7.1、7.2
- [ ] 7.5 在 `backend/internal/service/wire.go` 注册 chance、lottery 和 admin service providers，解除兑换码依赖环，依赖任务 6.2、6.5
- [ ] 7.6 在 `backend/internal/handler/wire.go` 注册用户及管理员 lottery handlers，依赖任务 7.1、7.2、7.5
- [ ] 7.7 运行 Wire 生成流程并更新实际生成文件，仅保留生成器产生的必要差异，依赖任务 2.5、7.6
- [ ] 7.8 在 `backend/internal/server/api_contract_test.go` 增加抽奖接口契约、用户隔离和管理员鉴权回归，依赖任务 7.7

## 8. 前端 API 与导航
- [ ] 8.1 在 `frontend/src/api/lottery.ts` 定义用户/管理 DTO、摘要、抽奖、历史及 CRUD 请求，依赖任务 7.4
- [ ] 8.2 在 `frontend/src/router/index.ts` 增加 `/lottery` 和 `/admin/lottery` 路由，依赖任务 8.1
- [ ] 8.3 在 `frontend/src/components/layout/AppSidebar.vue` 增加用户抽奖和管理员抽奖管理入口，并按可用状态展示，依赖任务 8.2
- [ ] 8.4 在 `frontend/src/i18n/locales/zh/misc.ts` 与 `frontend/src/i18n/locales/en/misc.ts` 增加用户抽奖、状态和结果文案，依赖任务 8.2
- [ ] 8.5 在 `frontend/src/i18n/locales/zh/admin/resources.ts` 与 `frontend/src/i18n/locales/en/admin/resources.ts` 增加管理配置、验证和记录文案，依赖任务 8.2

## 9. 用户抽奖体验
- [ ] 9.1 在 `frontend/src/components/lottery/LotteryPrizeCard.vue` 实现固定尺寸奖品卡、默认余额/订阅图标和安全图片展示，验证 why.md#需求-用户与管理界面-场景-用户抽奖
- [ ] 9.2 在 `frontend/src/components/lottery/LotteryReel.vue` 实现横向循环轮带、背景虚化、服务端结果落点和 `prefers-reduced-motion`，依赖任务 9.1
- [ ] 9.3 在 `frontend/src/views/user/LotteryView.vue` 实现普通/豪华页签、次数、奖品概率、抽奖按钮、结果和历史，依赖任务 8.1、9.2
- [ ] 9.4 在 `frontend/src/components/lottery/__tests__/LotteryReel.spec.ts` 验证固定布局、动画终点、重复点击保护和减少动态效果，依赖任务 9.2
- [ ] 9.5 在 `frontend/src/views/user/__tests__/LotteryView.spec.ts` 覆盖奖池状态、准确概率、未中奖、剩余次数和历史加载，依赖任务 9.3

## 10. 管理端配置
- [ ] 10.1 在 `frontend/src/views/admin/lottery/AdminLotteryView.vue` 实现奖池、奖品、规则、记录四个页签和数据装载，依赖任务 8.1
- [ ] 10.2 在 `frontend/src/views/admin/lottery/LotteryPoolPanel.vue` 实现周期、次数、启停和起止时间表单，依赖任务 10.1
- [ ] 10.3 在 `frontend/src/views/admin/lottery/LotteryPrizeDialog.vue` 复用 `ImageUpload`，实现余额/订阅字段、概率和库存校验，依赖任务 10.1
- [ ] 10.4 在 `frontend/src/views/admin/lottery/LotteryRuleDialog.vue` 实现固定事件、受益人、双奖池次数、充值口径和重复模式，依赖任务 10.1
- [ ] 10.5 在 `frontend/src/views/admin/lottery/LotteryRecordsTable.vue` 实现抽奖与次数流水筛选分页，依赖任务 10.1
- [ ] 10.6 在 `frontend/src/views/admin/lottery/__tests__/AdminLotteryView.spec.ts` 覆盖奖池保存、概率总和、订阅奖品、规则字段联动和记录筛选，依赖任务 10.2 至 10.5

## 11. 安全检查
- [ ] 11.1 检查用户/管理员权限、数值上下界、Data URL MIME/体积、服务端随机、SQL 参数化和敏感审计字段，验证 G9 要求
- [ ] 11.2 检查抽奖、规则事件、充值履约及退款路径的事务与幂等键，确认无负次数、超库存、重复奖励或部分发奖

## 12. 文档与知识库
- [ ] 12.1 新建 `helloagents/wiki/modules/lottery.md`，同步用户规则、管理能力、API、数据模型和依赖
- [ ] 12.2 更新 `helloagents/wiki/overview.md` 与 `helloagents/wiki/arch.md`，登记抽奖模块和 ADR，依赖任务 12.1
- [ ] 12.3 更新 `helloagents/wiki/api.md` 与 `helloagents/wiki/data.md`，登记接口和六张数据表，依赖任务 12.1
- [ ] 12.4 更新 `helloagents/CHANGELOG.md`，记录新增双奖池邀请抽奖功能，依赖全部实现任务

## 13. 验证与交付
- [ ] 13.1 运行后端 lottery/redeem/affiliate/payment 定向单元与集成测试，修复失败，依赖任务 11.2
- [ ] 13.2 运行 `go test ./...` 和迁移回归测试，记录环境性跳过项，依赖任务 13.1
- [ ] 13.3 运行前端类型检查、Vitest、lint 和生产构建，依赖任务 9.5、10.6
- [ ] 13.4 使用 Playwright 检查桌面和移动端轮带非空、落点一致、无文字/控件重叠及减少动态效果，依赖任务 13.3
- [ ] 13.5 更新本清单为完成状态，迁移方案包至 `helloagents/history/2026-07/` 并更新 `helloagents/history/index.md`，依赖任务 12.4、13.4
