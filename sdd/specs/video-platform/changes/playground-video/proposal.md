# Proposal: 对话广场视频生成 (Proposal)

## 1. Context & Problem Statement
- **Current State**: 独立视频平台 API 已可用，但对话广场只调用 Chat Completions。
- **Pain Points**: 站点用户无法在网页内创建、查询和播放 Grok 视频任务。

## 2. Value Proposition
- 复用已有账号池、余额计费和失败退款，让站点用户不需要管理 API Key。
- 将异步创建、查询和结果播放整合进现有会话。

## 3. Alternatives Considered
- **直接从浏览器调用 `/v1/videos`**: 需要创建并暴露 API Key，拒绝。
- **新增独立视频任务服务**: 会复制已有调度、计费和退款逻辑，拒绝。
- **Playground 路由适配到现有处理器**: 采用，改动最小且行为一致。

## 4. Success Metrics
- [x] 文生视频可从对话广场完成
- [x] 1.5 模型可使用单张参考图生成
- [x] 成功视频可播放，失败可展示并沿用自动退款
