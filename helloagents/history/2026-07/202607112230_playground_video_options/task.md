# 任务清单: 对话广场视频时长与比例选择

归档: `helloagents/history/2026-07/202607112230_playground_video_options/`

## 1. 配置与输入控件
- [√] 1.1 扩展 `frontend/src/types/playground.ts` 与 `frontend/src/constants/playground.ts` 的视频配置和选项。
- [√] 1.2 在 `PlaygroundInput.vue` 与 `PlaygroundView.vue` 中显示并绑定时长、比例选择，依赖任务 1.1。

## 2. 请求与文案
- [√] 2.1 在 `useChatHandler.ts` 中按模型追加 `seconds` 与 `aspect_ratio`，依赖任务 1.1。
- [√] 2.2 更新中英文对话广场文案。

## 3. 测试
- [√] 3.1 更新输入组件与视频请求测试，覆盖显示、隐藏和 payload。
- [√] 3.2 运行相关 Vitest、TypeScript、ESLint 和生产构建。

## 4. 安全检查
- [√] 4.1 检查固定选项、模型隔离、移动端布局和敏感信息。

## 5. 文档更新
- [√] 5.1 同步 `helloagents/wiki/modules/gateway-video.md`、`helloagents/wiki/api.md` 与 `helloagents/CHANGELOG.md`。

## 执行结果
- 定向 Vitest：2 个文件、6 个测试通过。
- `vue-tsc --noEmit`、目标文件 ESLint 和生产构建通过。
- 构建仅保留项目既有的 Browserslist、动态导入和 chunk 体积警告。
