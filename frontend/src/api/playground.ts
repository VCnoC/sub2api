/**
 * 对话广场 API 客户端
 *
 * 端点：
 *   GET  /api/v1/playground/models?group={name}   返回该分组下用户可用模型
 *   POST /api/v1/playground/chat/completions      OpenAI 兼容流式聊天
 *
 * 流式请求不走 axios（无法处理 SSE）→ 见 composables/playground/useStreamChat.ts
 */

import { apiClient } from './client'
import { API_ENDPOINTS } from '@/constants/playground'
import type {
  ChatCompletionRequest,
  ChatCompletionResponse,
  PlaygroundAvailableModel,
  PlaygroundAvailableModelsResponse,
} from '@/types/playground'

/**
 * 查询指定分组下用户可用的模型列表
 */
export async function getAvailableModels(
  groupName: string
): Promise<PlaygroundAvailableModel[]> {
  const { data } = await apiClient.get<PlaygroundAvailableModelsResponse>(
    API_ENDPOINTS.AVAILABLE_MODELS,
    { params: { group: groupName } }
  )
  return data?.models ?? []
}

/**
 * 非流式 Chat Completions（stream=false 时使用）
 */
export async function sendChatCompletionNonStream(
  payload: ChatCompletionRequest
): Promise<ChatCompletionResponse> {
  const { data } = await apiClient.post<ChatCompletionResponse>(
    API_ENDPOINTS.CHAT_COMPLETIONS,
    payload
  )
  return data
}

export const playgroundAPI = {
  getAvailableModels,
  sendChatCompletionNonStream,
}

export default playgroundAPI
