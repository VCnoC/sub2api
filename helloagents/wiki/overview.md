# Sub2API

## 1. 项目概述

Sub2API 是多上游 AI 账号池、分组路由、API Key 分发与用量计费网关。

## 2. 模块索引

| 模块 | 职责 | 状态 | 文档 |
|------|------|------|------|
| 账号管理 | 上游凭证、模型限制、代理与分组绑定 | ✅稳定 | [账号管理](modules/accounts.md) |
| 分组与计费 | 平台隔离、倍率、媒体价格和配额 | ✅稳定 | [分组与计费](modules/groups-billing.md) |
| 网关与任务 | 请求调度、上游转发、异步媒体任务和用量记录 | ✅稳定 | [网关与任务](modules/gateway-video.md) |
| 站内工单 | 用户双向支持、管理员协作、私有附件和未读通知 | 🚧已部署待验收 | [站内工单](modules/support-tickets.md) |
| 邀请抽奖 | 普通/豪华双奖池、邀请机会、余额与订阅自动发奖 | 🚧待部署验收 | [邀请抽奖](modules/lottery.md) |

## 3. 快速链接
- [技术约定](../project.md)
- [架构设计](arch.md)
- [API 手册](api.md)
- [数据模型](data.md)
- [变更历史](../history/index.md)
