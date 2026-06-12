/**
 * 对话广场常量定义
 * 移植自 new-api/web/default/src/features/playground/constants.ts
 */

import type { PlaygroundConfig, ParameterEnabled } from '@/types/playground'

// 消息角色
export const MESSAGE_ROLES = {
  USER: 'user',
  ASSISTANT: 'assistant',
  SYSTEM: 'system',
} as const

// 消息状态
export const MESSAGE_STATUS = {
  LOADING: 'loading',
  STREAMING: 'streaming',
  COMPLETE: 'complete',
  ERROR: 'error',
} as const

// API 端点
export const API_ENDPOINTS = {
  CHAT_COMPLETIONS: '/playground/chat/completions',
  AVAILABLE_MODELS: '/playground/models',
  AVAILABLE_GROUPS: '/groups/available',
  CONVERSATIONS: '/playground/conversations',
} as const

// 会话标题自动取首条用户消息的前 N 个字符
export const CONVERSATION_TITLE_MAX_CHARS = 20

// 消息保存防抖窗口（流式回复结束后延迟保存，毫秒）
export const CONVERSATION_SAVE_DEBOUNCE_MS = 1000

// 默认分组（兜底值）
export const DEFAULT_GROUP = 'default' as const

// 默认配置
export const DEFAULT_CONFIG: PlaygroundConfig = {
  model: '',
  group: DEFAULT_GROUP,
  temperature: 0.7,
  top_p: 1,
  max_tokens: 4096,
  frequency_penalty: 0,
  presence_penalty: 0,
  seed: null,
  stream: true,
  systemPrompt: '',
}

// 默认参数启用开关
export const DEFAULT_PARAMETER_ENABLED: ParameterEnabled = {
  temperature: true,
  top_p: true,
  max_tokens: false,
  frequency_penalty: true,
  presence_penalty: true,
  seed: false,
}

// localStorage 键（分治三套独立持久化）
export const STORAGE_KEYS = {
  CONFIG: 'sub2api_playground_config',
  MESSAGES: 'sub2api_playground_messages',
  PARAMETER_ENABLED: 'sub2api_playground_parameter_enabled',
} as const

// 历史消息上限（防止 localStorage 溢出 5MB 配额）
export const MAX_HISTORY_MESSAGES = 200

// 错误码 i18n key 映射
export const ERROR_CODE_I18N_MAP: Record<string, string> = {
  authentication_error: 'playground.error.authError',
  permission_error: 'playground.error.permissionError',
  invalid_request_error: 'playground.error.invalidRequest',
  insufficient_quota: 'playground.error.insufficientQuota',
  api_error: 'playground.error.apiError',
  upstream_error: 'playground.error.upstreamError',
  server_error: 'playground.error.serverError',
  rate_limit_error: 'playground.error.rateLimitError',
}
