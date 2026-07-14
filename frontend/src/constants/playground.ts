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
  IMAGES: '/playground/images',
  VIDEOS: '/playground/videos',
} as const

// 会话标题自动取首条用户消息的前 N 个字符
export const CONVERSATION_TITLE_MAX_CHARS = 20

// 消息保存防抖窗口（流式回复结束后延迟保存，毫秒）
export const CONVERSATION_SAVE_DEBOUNCE_MS = 1000

// 默认分组（兜底值）
export const DEFAULT_GROUP = 'default' as const

export const VIDEO_SECONDS_OPTIONS = Array.from(
  { length: 15 },
  (_, index) => String(index + 1)
)

export const VIDEO_ASPECT_RATIO_OPTIONS = [
  '1:1',
  '16:9',
  '9:16',
  '4:3',
  '3:4',
  '3:2',
  '2:3',
  '2:1',
  '1:2',
  '19.5:9',
  '9:19.5',
  '20:9',
  '9:20',
] as const

export const GPT_IMAGE_2_VIP_MODEL = 'gpt-image-2-vip'

export const GPT_IMAGE_2_SIZE_OPTIONS = [
  { label: '1:1', value: '1024x1024' },
  { label: '4:3', value: '1024x768' },
  { label: '3:4', value: '768x1024' },
  { label: '3:2', value: '1008x672' },
  { label: '2:3', value: '672x1008' },
  { label: '16:9', value: '1280x720' },
  { label: '9:16', value: '720x1280' },
  { label: '21:9', value: '1344x576' },
  { label: '1:1(2K)', value: '2048x2048' },
  { label: '4:3(2K)', value: '2304x1728' },
  { label: '3:4(2K)', value: '1728x2304' },
  { label: '3:2(2K)', value: '2496x1664' },
  { label: '2:3(2K)', value: '1664x2496' },
  { label: '16:9(2K)', value: '2560x1440' },
  { label: '9:16(2K)', value: '1440x2560' },
  { label: '21:9(2K)', value: '3024x1296' },
  { label: '1:1(4K)', value: '2880x2880' },
  { label: '4:3(4K)', value: '3264x2448' },
  { label: '3:4(4K)', value: '2448x3264' },
  { label: '3:2(4K)', value: '3504x2336' },
  { label: '2:3(4K)', value: '2336x3504' },
  { label: '16:9(4K)', value: '3840x2160' },
  { label: '9:16(4K)', value: '2160x3840' },
  { label: '21:9(4K)', value: '3808x1632' },
] as const

export const GPT_IMAGE_2_QUALITY_OPTIONS = ['auto', 'low', 'medium', 'high'] as const
export const GPT_IMAGE_2_RESPONSE_FORMAT_OPTIONS = ['url', 'b64_json'] as const
export const GPT_IMAGE_2_BACKGROUND_OPTIONS = ['auto', 'transparent', 'opaque'] as const

export function isGrokImagineVideoModel(model: string): boolean {
  return model.trim().toLowerCase().startsWith('grok-imagine-video')
}

export function isGptImage2VipModel(model: string): boolean {
  return model.trim().toLowerCase() === GPT_IMAGE_2_VIP_MODEL
}

// 默认配置
export const DEFAULT_CONFIG: PlaygroundConfig = {
  model: '',
  group: DEFAULT_GROUP,
  videoSeconds: '4',
  videoAspectRatio: '9:16',
  imageSize: '1024x1024',
  imageQuality: 'high',
  imageResponseFormat: '',
  imageStyle: '',
  imageBackground: '',
  imageWatermark: false,
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
  imageQuality: true,
  imageResponseFormat: false,
  imageStyle: false,
  imageBackground: false,
  imageWatermark: false,
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
