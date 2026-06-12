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
  ConversationDetail,
  ConversationSummary,
  CreateConversationRequest,
  PlaygroundAvailableModel,
  PlaygroundAvailableModelsResponse,
  UpdateConversationRequest,
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

// ==================== 多会话持久化 CRUD ====================

/** 获取当前用户的会话摘要列表（按最后活动时间倒序，不含 messages） */
export async function listConversations(): Promise<ConversationSummary[]> {
  const { data } = await apiClient.get<ConversationSummary[]>(
    API_ENDPOINTS.CONVERSATIONS
  )
  return data ?? []
}

/** 获取单个会话详情（含完整消息列表） */
export async function getConversation(id: number): Promise<ConversationDetail> {
  const { data } = await apiClient.get<ConversationDetail>(
    `${API_ENDPOINTS.CONVERSATIONS}/${id}`
  )
  return data
}

/** 新建会话 */
export async function createConversation(
  payload: CreateConversationRequest
): Promise<ConversationDetail> {
  const { data } = await apiClient.post<ConversationDetail>(
    API_ENDPOINTS.CONVERSATIONS,
    payload
  )
  return data
}

/**
 * 更新会话（保存消息/改标题）。
 * ⚠️ model/group_name 缺省 = 清空，调用方每次都必须带上（见 UpdateConversationRequest 注释）。
 */
export async function updateConversation(
  id: number,
  payload: UpdateConversationRequest
): Promise<void> {
  await apiClient.put(`${API_ENDPOINTS.CONVERSATIONS}/${id}`, payload)
}

/** 删除会话 */
export async function deleteConversation(id: number): Promise<void> {
  await apiClient.delete(`${API_ENDPOINTS.CONVERSATIONS}/${id}`)
}

export const playgroundAPI = {
  getAvailableModels,
  sendChatCompletionNonStream,
  listConversations,
  getConversation,
  createConversation,
  updateConversation,
  deleteConversation,
}

export default playgroundAPI
