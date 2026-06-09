/**
 * 对话广场（Playground）类型定义
 * 移植自 new-api/web/default/src/features/playground/types.ts
 * 适配 sub2api Vue 3 技术栈
 */

// ==================== 消息相关 ====================

export type MessageRole = 'user' | 'assistant' | 'system'

export type MessageStatus = 'loading' | 'streaming' | 'complete' | 'error'

export interface MessageVersion {
  /** 版本唯一 ID（用于多版本切换的 key） */
  id: string
  /** 该版本对应的文本内容 */
  content: string
}

export interface MessageReasoning {
  /** 推理内容（思考链） */
  content: string
  /** 推理耗时（毫秒） */
  duration: number
}

export interface MessageSource {
  href: string
  title: string
}

export type PlaygroundAttachmentKind = 'image' | 'document'

export interface PlaygroundAttachment {
  id: string
  kind: PlaygroundAttachmentKind
  name: string
  type: string
  size: number
  /** 图片使用 data URL 直传给 OpenAI 兼容 image_url。 */
  dataUrl?: string
  /** 文档在浏览器本地读取为纯文本后拼入用户消息。 */
  text?: string
}

export interface Message {
  /** 消息唯一 key */
  key: string
  /** 消息来源（user/assistant/system） */
  from: MessageRole
  /** 多版本数组：versions[0] 为当前显示版本；regenerate 时追加新版本 */
  versions: MessageVersion[]
  /** 引用来源列表 */
  sources?: MessageSource[]
  /** 推理内容（reasoning_content / <think> 标签解析） */
  reasoning?: MessageReasoning
  /** 推理是否在流式输出中 */
  isReasoningStreaming?: boolean
  /** 推理是否已完成 */
  isReasoningComplete?: boolean
  /** 内容是否已完成 */
  isContentComplete?: boolean
  /** 当前状态 */
  status?: MessageStatus
  /** 错误码（用于本地化错误提示） */
  errorCode?: string | null
  /** 用户上传的图片/文本类文档附件 */
  attachments?: PlaygroundAttachment[]
}

// ==================== API 请求/响应 ====================

export interface ContentPart {
  type: 'text' | 'image_url'
  text?: string
  image_url?: { url: string }
}

export interface ChatCompletionMessage {
  role: MessageRole
  content: string | ContentPart[]
}

export interface ChatCompletionRequest {
  model: string
  group: string
  messages: ChatCompletionMessage[]
  stream: boolean
  temperature?: number
  top_p?: number
  max_tokens?: number
  frequency_penalty?: number
  presence_penalty?: number
  seed?: number | null
}

export interface ChatCompletionChunkDelta {
  role?: MessageRole
  content?: string
  reasoning_content?: string
}

export interface ChatCompletionChunk {
  id: string
  object: string
  created: number
  model: string
  choices: Array<{
    index: number
    delta: ChatCompletionChunkDelta
    finish_reason: string | null
  }>
}

export interface ChatCompletionResponse {
  id: string
  object: string
  created: number
  model: string
  choices: Array<{
    index: number
    message: {
      role: MessageRole
      content: string
      reasoning_content?: string
    }
    finish_reason: string
  }>
  usage?: {
    prompt_tokens: number
    completion_tokens: number
    total_tokens: number
  }
}

// ==================== 配置 ====================

export interface PlaygroundConfig {
  model: string
  group: string
  temperature: number
  top_p: number
  max_tokens: number
  frequency_penalty: number
  presence_penalty: number
  seed: number | null
  stream: boolean
  systemPrompt: string
}

export interface ParameterEnabled {
  temperature: boolean
  top_p: boolean
  max_tokens: boolean
  frequency_penalty: boolean
  presence_penalty: boolean
  seed: boolean
}

// ==================== 模型/分组 ====================

export interface ModelOption {
  /** 显示文本 */
  label: string
  /** 实际 model ID */
  value: string
  /** 所属平台（用于按平台禁用某些参数） */
  platform?: string
}

export interface GroupOption {
  /** 显示文本 */
  label: string
  /** 实际 group name */
  value: string
  /** 分组费率倍数 */
  ratio: number
  /** 平台（openai / anthropic / gemini / antigravity） */
  platform?: string
  /** 简介 */
  desc?: string
}

// ==================== 后端响应 ====================

export interface PlaygroundAvailableModel {
  id: string
  platform: string
  group_id: number
  group_name: string
}

export interface PlaygroundAvailableModelsResponse {
  models: PlaygroundAvailableModel[]
}

// ==================== 错误类型 ====================

export interface PlaygroundError {
  type: string
  message: string
  code?: string
}
