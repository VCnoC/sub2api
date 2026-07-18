# 任务清单: API Key 有序多分组容灾

目录: `helloagents/plan/202607181905_api_key_group_failover/`

---

## 1. 数据模型与迁移
- [√] 1.1 在 `backend/ent/schema/api_key_group.go` 和 `backend/ent/schema/api_key.go` 定义有序分组关联，验证 why.md#需求-用户配置有序候选分组-场景-旧-key-升级
- [√] 1.2 新增 `backend/migrations/185_api_key_ordered_groups.sql` 并回填旧 `group_id`，增加 migration 回归测试，依赖任务1.1
- [√] 1.3 重新生成 Ent 并执行生成结果一致性检查，依赖任务1.1至1.2

## 2. Repository 与 API Key 服务
- [√] 2.1 扩展 `backend/internal/repository/api_key_repo.go` 读写有序分组并事务同步首组镜像，验证 why.md#需求-用户配置有序候选分组
- [√] 2.2 扩展 `backend/internal/service/api_key.go` 与 `backend/internal/service/api_key_service.go`，实现 1-5 个、去重、同平台和逐组权限校验，依赖任务2.1
- [√] 2.3 更新分组删除/迁移、列表过滤和统计查询以覆盖关联表，依赖任务2.1

## 3. 鉴权缓存与候选激活
- [√] 3.1 扩展 `backend/internal/service/api_key_auth_cache.go` 和 `api_key_auth_cache_impl.go` 保存有序候选快照
- [√] 3.2 调整 `backend/internal/server/middleware/api_key_auth.go`，认证 Key 后加载候选链并激活第一个有计费资格的分组，验证 why.md#需求-请求按优先级跨组容灾-场景-分组计费资格不可用，依赖任务3.1
- [√] 3.3 增加分组切换上下文更新、订阅选择和缓存失效测试，依赖任务3.1至3.2

## 4. 网关跨组 Failover
- [√] 4.1 扩展 `backend/internal/handler/failover_loop.go`，在组内耗尽后推进候选分组并重置账号状态
- [√] 4.2 接入通用网关、Chat Completions 与 Responses 循环，验证连接错误、超时、429、5xx 和无账号切换，依赖任务4.1
- [√] 4.3 接入 Gemini 与其他共享调度入口，保持参数/权限错误不切换，依赖任务4.1
- [√] 4.4 增加 Writer 已提交、客户端取消和异步任务可能创建后的禁止切换测试，依赖任务4.2至4.3

## 5. 计费与用量一致性
- [√] 5.1 验证 `backend/internal/service/billing_cache_service.go` 按候选组重新检查余额或选择最早到期订阅
- [√] 5.2 验证 usage billing 与日志只记录实际成功的 `group_id`、`subscription_id` 和账号，增加跨组回归测试，依赖任务5.1

## 6. 用户端交互
- [√] 6.1 更新 `frontend/src/api/keys.ts`、`frontend/src/types/index.ts` 支持兼容 `group_id` 与有序 `group_ids`
- [√] 6.2 更新 `frontend/src/views/user/KeysView.vue`，实现同平台多选、最多 5 个和拖拽排序，依赖任务6.1
- [√] 6.3 更新中英文文案并增加创建、编辑、排序和校验测试，依赖任务6.2

## 7. 安全、文档与验证
- [√] 7.1 执行安全检查：分组权限、跨平台拒绝、循环/重复数据、非幂等重放和计费幂等
- [√] 7.2 更新 SDD、知识库、API/数据模型文档和 CHANGELOG
- [√] 7.3 运行 Ent 生成、后端完整测试与构建、前端测试/类型检查/生产构建及 `git diff --check`
- [√] 7.4 更新任务状态并迁移方案包至 `helloagents/history/2026-07/`
