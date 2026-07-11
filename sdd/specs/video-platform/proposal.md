# Proposal: 独立视频平台与失败退款 (Proposal)

## 1. Context & Problem Statement
- **Current State**: 系统已有 Grok OAuth 视频生成与按秒计费，但没有可填写 Base URL、API Key 的独立视频平台，也未自动处理异步失败退款。
- **Pain Points**: `cpa-fan` 视频接口无法作为独立账号池接入；失败任务会保留创建时产生的余额扣费。

## 2. Value Proposition
- 复用账号池、分组、API Key、调度和统计能力接入视频上游。
- 自动跟踪终态并保证失败只退款一次。

## 3. Alternatives Considered
- **Option A**: 创建时冻结、成功结算。未采用：用户选择立即扣费后失败补偿。
- **Option B**: 创建时扣费、失败幂等退款。采用：只需处理余额，符合现有使用方式。

## 4. Success Metrics
- [ ] 管理员可创建视频平台账号并同步上游模型
- [ ] 六类视频接口可通过分组调度转发
- [ ] 按秒和按次价格计算准确
- [ ] 失败任务自动且仅退款一次
