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

export const GPT_IMAGE_2_RESOLUTION_OPTIONS = ['1K', '2K', '4K'] as const
export const GPT_IMAGE_2_ASPECT_RATIO_OPTIONS = [
  '1:1',
  '4:3',
  '3:4',
  '3:2',
  '2:3',
  '16:9',
  '9:16',
  '21:9',
] as const

export type GptImage2Resolution = typeof GPT_IMAGE_2_RESOLUTION_OPTIONS[number]
export type GptImage2AspectRatio = typeof GPT_IMAGE_2_ASPECT_RATIO_OPTIONS[number]

const GPT_IMAGE_2_SIZE_MAP: Record<
  GptImage2Resolution,
  Record<GptImage2AspectRatio, string>
> = {
  '1K': {
    '1:1': '1024x1024',
    '4:3': '1024x768',
    '3:4': '768x1024',
    '3:2': '1008x672',
    '2:3': '672x1008',
    '16:9': '1280x720',
    '9:16': '720x1280',
    '21:9': '1344x576',
  },
  '2K': {
    '1:1': '2048x2048',
    '4:3': '2304x1728',
    '3:4': '1728x2304',
    '3:2': '2496x1664',
    '2:3': '1664x2496',
    '16:9': '2560x1440',
    '9:16': '1440x2560',
    '21:9': '3024x1296',
  },
  '4K': {
    '1:1': '2880x2880',
    '4:3': '3264x2448',
    '3:4': '2448x3264',
    '3:2': '3504x2336',
    '2:3': '2336x3504',
    '16:9': '3840x2160',
    '9:16': '2160x3840',
    '21:9': '3808x1632',
  },
}

export const GPT_IMAGE_2_QUALITY_OPTIONS = ['auto', 'low', 'medium', 'high'] as const
export const GPT_IMAGE_2_RESPONSE_FORMAT_OPTIONS = ['url', 'b64_json'] as const
export const GPT_IMAGE_2_BACKGROUND_OPTIONS = ['auto', 'transparent', 'opaque'] as const

export function isGrokImagineVideoModel(model: string): boolean {
  return model.trim().toLowerCase().startsWith('grok-imagine-video')
}

export function isGptImageModel(model: string): boolean {
  const normalized = model.trim().toLowerCase()
  return normalized.startsWith('gpt-image-') || /^image-2(?:-|$)/.test(normalized)
}

export function getGptImage2Resolution(model: string): GptImage2Resolution | null {
  const suffix = model.trim().match(/^(?:gpt-)?image-2-(1k|2k|4k)$/i)?.[1]
  return suffix ? (suffix.toUpperCase() as GptImage2Resolution) : null
}

export function getGptImage2ResolutionForSize(size: string): GptImage2Resolution | null {
  for (const resolution of GPT_IMAGE_2_RESOLUTION_OPTIONS) {
    if (Object.values(GPT_IMAGE_2_SIZE_MAP[resolution]).includes(size)) return resolution
  }
  return null
}

export function getGptImage2AspectRatio(size: string): GptImage2AspectRatio | null {
  for (const resolution of GPT_IMAGE_2_RESOLUTION_OPTIONS) {
    for (const ratio of GPT_IMAGE_2_ASPECT_RATIO_OPTIONS) {
      if (GPT_IMAGE_2_SIZE_MAP[resolution][ratio] === size) return ratio
    }
  }
  return null
}

export function getGptImage2Size(
  resolution: GptImage2Resolution,
  ratio: GptImage2AspectRatio
): string {
  return GPT_IMAGE_2_SIZE_MAP[resolution][ratio]
}

export function withGptImage2Resolution(
  model: string,
  resolution: GptImage2Resolution
): string {
  const match = model.trim().match(/^((?:gpt-)?image-2)(?:-(?:1k|2k|4k))?$/i)
  return match ? `${match[1]}-${resolution.toLowerCase()}` : model.trim()
}

// 默认配置
export const DEFAULT_CONFIG: PlaygroundConfig = {
  model: '',
  group: DEFAULT_GROUP,
  videoSeconds: '4',
  videoAspectRatio: '9:16',
  imageSize: '1024x1024',
  imageQuality: 'high',
  imageResponseFormat: 'b64_json',
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
  imageResponseFormat: true,
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
