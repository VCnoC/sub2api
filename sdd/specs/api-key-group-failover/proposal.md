# Proposal: API Key 有序多分组容灾 (Proposal)

## 1. Context & Problem Statement
- **Current State**: 一个 API Key 仅绑定一个分组，现有 failover 只在该分组的账号之间切换。
- **Pain Points**: 分组账号全部故障、持续 429/5xx、余额不足或订阅耗尽时，无人值守任务会中断。

## 2. Value Proposition
- 用户可用一个 Key 组合余额和订阅分组，并自行安排消费与容灾顺序。
- 降低单分组故障导致长时间任务失败的概率，同时保持实际分组准确计费。

## 3. Alternatives Considered
- **Option A**: `api_key_groups` 有序关联表（采用）。优点是外键、查询、顺序和缓存失效可控。
- **Option B**: 在 API Key 保存 JSONB/逗号字符串。缺点是引用完整性、删除和统计困难。
- **Option C**: 复用全局分组 fallback。缺点是无法按 Key 自定义顺序，且会影响共享分组的其他用户。

## 4. Success Metrics
- [x] 一个 API Key 可配置 1-5 个同平台分组并拖拽排序
- [x] 每个新请求严格从最高优先级开始
- [x] 余额/订阅不可用或安全可重试故障时自动推进下一组
- [x] 响应提交或异步任务可能创建后不会重复请求
- [x] 实际成功分组和订阅准确计费、记录
- [x] 旧 `group_id` 客户端和现有 Key 保持兼容
