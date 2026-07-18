# Task Breakdown & Execution Board: API Key 有序多分组容灾 (Tasks)

## Phase 1: 数据模型与兼容迁移
- [x] Task 1.1: 增加 `api_key_groups` Ent schema、约束和关联
- [x] Task 1.2: 增加 185 migration 并回填旧 `group_id`
- [x] Task 1.3: 生成 Ent 并验证迁移、顺序和唯一性

## Phase 2: API Key 领域与缓存
- [x] Task 2.1: Repository 事务读写有序分组并同步首组镜像
- [x] Task 2.2: Service 校验 1-5 个、同平台、去重和逐组权限
- [x] Task 2.3: Handler/API 兼容 `group_id` 并新增 `group_ids`
- [x] Task 2.4: 认证缓存保存候选链并正确失效

## Phase 3: 计费资格与跨组 Failover
- [x] Task 3.1: 鉴权阶段按序选择首个余额/订阅可用分组
- [x] Task 3.2: 扩展 `FailoverState` 在组内耗尽后安全推进下一组
- [x] Task 3.3: 接入通用网关、Chat Completions、Responses 和 Gemini
- [x] Task 3.4: 保证实际组、订阅和账号计费及日志一致
- [x] Task 3.5: 阻止响应提交、取消和异步任务后的不安全重放

## Phase 4: UI & Interaction
- [x] Task 4.1: API 类型支持有序 `group_ids`
- [x] Task 4.2: Key 创建/编辑实现多选和拖拽排序
- [x] Task 4.3: 增加数量、同平台、回填和载荷测试及中英文文案

## Phase 5: Integration & Refinement
- [x] Task 5.1: 完成 migration、repository、service、cache 和 failover 回归测试
- [x] Task 5.2: 完成前端测试、类型检查、后端完整测试和生产构建
- [x] Task 5.3: 更新知识库、CHANGELOG、方案状态并归档
