# 技术设计: 对话广场视频时长与比例选择

## 技术方案

### 核心技术
- Vue 3 Composition API、TypeScript、现有原生 `select` 和 Vitest。

### 实现要点
- 在 `PlaygroundConfig` 中增加 `videoSeconds`、`videoAspectRatio`，默认值为 `4`、`9:16`。
- 在共享常量中定义选项和 `isGrokImagineVideoModel` 判断，输入组件与请求编排共用。
- 输入组件只在视频分组且模型匹配时显示两个选择框。
- 视频创建 payload 只在模型匹配时追加 `seconds`、`aspect_ratio`。

## API设计

### POST /api/v1/playground/videos
- **请求:** `{ model, group, prompt, seconds?, aspect_ratio?, input_reference? }`
- **响应:** 不变。

## 安全与性能
- **安全:** 选择值来自固定列表；后端继续沿用现有鉴权、分组权限和上游校验。
- **性能:** 仅增加两个本地字段和条件渲染，无额外请求或依赖。

## 测试与部署
- **测试:** 输入组件可见性与事件、创建 payload、TypeScript、ESLint、生产构建。
- **部署:** 本次按用户要求先完成并验证本地代码，不部署。
