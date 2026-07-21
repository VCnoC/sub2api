# Proposal: 纯次数订阅套餐 (Proposal)

## 1. Context & Problem Statement
- **Current State**: 订阅分组仅按 USD 日、周、月额度计费；API Key 的 5 小时和 1 天限制同样按 USD 且可被多个 Key 分散使用。
- **Pain Points**: 管理员无法通过现有“分组 + 套餐”流程销售固定成功调用次数，用户也无法看到次数剩余和重置时间。

## 2. Value Proposition
- 保持现有分组、套餐、支付、兑换和多订阅流程不变，只扩展订阅权益的计量方式。
- 固定次数比 Token/USD 更容易向用户解释，并通过数据库占位保证并发下不超发。

## 3. Alternatives Considered
- **Option A**: PostgreSQL 占位账本与订阅计数（采用）。一致性明确，可与现有计费事务组合。
- **Option B**: 成功后直接累加。实现较少，但高并发可能突破上限。
- **Option C**: Redis 占位、数据库异步落账。吞吐较高，但双写和故障恢复复杂，当前没有必要。

## 4. Success Metrics
- [x] 管理员可创建按次数计费的订阅分组并照常绑定套餐
- [x] 5 小时和 24 小时窗口从首次成功调用起算
- [x] 失败请求释放占位，成功文本请求最终保留 1 次
- [x] 并发、重试和跨组 failover 不会重复扣次或超发
- [x] 图片、视频、count_tokens 和只读请求不计次
- [x] 现有 USD 订阅行为保持兼容
