# 任务清单: 对话广场视频生成

目录: `helloagents/plan/202607111841_playground_video/`

---

## 1. Playground 视频网关
- [√] 1.1 扩展 `backend/internal/server/middleware/playground_context.go` 支持 GET group query，验证 why.md#需求-对话广场视频生成-场景-失败与中断
- [√] 1.2 在 `backend/internal/server/routes/playground.go` 注册视频创建与查询路由并复用现有处理器，验证 why.md#需求-对话广场视频生成-场景-文生视频
- [√] 1.3 增加后端最小路由/中间件测试，依赖任务 1.1-1.2

## 2. 前端视频生成
- [√] 2.1 在 `frontend/src/types/playground.ts` 与 `frontend/src/api/playground.ts` 增加视频协议类型和 API，验证 why.md#需求-对话广场视频生成-场景-文生视频
- [√] 2.2 在 `frontend/src/composables/playground/useChatHandler.ts` 复用消息状态实现创建、轮询、中断与错误，验证文生/图生/失败场景，依赖任务 2.1
- [√] 2.3 在 `frontend/src/views/user/PlaygroundView.vue` 与 `frontend/src/components/playground/MessageItem.vue` 按视频分组发送并播放结果，依赖任务 2.2

## 3. 安全检查
- [√] 3.1 检查 JWT、分组权限、余额校验、任务 ID 转义、附件边界和上游 Key 不泄露

## 4. 文档更新
- [√] 4.1 更新 SDD、知识库 API/架构/模块文档与 CHANGELOG

## 5. 测试
- [√] 5.1 运行后端定向测试、前端类型检查与生产构建
