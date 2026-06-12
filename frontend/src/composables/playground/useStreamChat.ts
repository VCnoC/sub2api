/**
 * 对话广场 SSE 流式核心 composable
 *
 * 用 fetch + ReadableStream 原生实现，零新增依赖。
 * 自动处理：
 *   - JWT Bearer header 注入
 *   - SSE 协议拆分（处理跨 chunk 截断）
 *   - reasoning_content / content 分流回调
 *   - AbortController 中断
 *   - 401 / 4xx / 5xx 错误本地化
 *
 * 移植自 new-api/web/default/src/features/playground/hooks/use-stream-request.ts
 * 适配 Vue 3 Composition API + 原生 fetch（替代 sse.js）
 */

import { ref } from 'vue'
import { API_ENDPOINTS } from '@/constants/playground'
import type { ChatCompletionRequest, ChatCompletionChunk } from '@/types/playground'

/** SSE 增量类型 */
export type StreamUpdateType = 'reasoning' | 'content'

/** 流式回调集合 */
export interface StreamCallbacks {
  /** 增量到达时回调（reasoning 或 content） */
  onUpdate: (type: StreamUpdateType, chunk: string) => void
  /** 流式正常结束（收到 [DONE]） */
  onComplete: () => void
  /** 错误回调 */
  onError: (error: string, errorCode?: string) => void
}

/** API base URL（与 axios 客户端一致） */
function getApiBaseUrl(): string {
  return import.meta.env.VITE_API_BASE_URL || '/api/v1'
}

/** 从 localStorage 读取当前 JWT token（与 axios 客户端共用） */
function getAuthToken(): string {
  return localStorage.getItem('auth_token') || ''
}

// ==================== 模块级单例状态 ====================
// 流式状态与中断控制器提升到模块作用域：SSE 消费循环不依赖组件生命周期，
// 路由切换（组件卸载）后生成继续在后台进行，重新进入页面时无缝衔接。
const isStreaming = ref(false)
let abortController: AbortController | null = null

/** 全局只读流式状态（供布局层显示「生成中」指示用） */
export const playgroundStreaming = isStreaming

export function useStreamChat() {

  /**
   * 发起 SSE 流式聊天请求
   *
   * 失败模式：
   *   - 401 → onError('需要重新登录', 'authentication_error')
   *   - 4xx/5xx + OpenAI 错误体 → onError(error.message, error.type)
   *   - 网络/解析异常 → onError(本地化兜底)
   *   - 用户主动中止 → 静默关闭，不触发 onError
   */
  async function send(
    payload: ChatCompletionRequest,
    callbacks: StreamCallbacks
  ): Promise<void> {
    if (isStreaming.value) {
      // eslint-disable-next-line no-console
      console.warn('[useStreamChat] Stream is already in progress')
      return
    }

    const url = `${getApiBaseUrl()}${API_ENDPOINTS.CHAT_COMPLETIONS}`
    abortController = new AbortController()
    isStreaming.value = true

    try {
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${getAuthToken()}`,
          Accept: 'text/event-stream',
        },
        body: JSON.stringify(payload),
        signal: abortController.signal,
      })

      if (!response.ok) {
        const errBody = await safeParseErrorBody(response)
        callbacks.onError(
          errBody?.message || `HTTP ${response.status}`,
          errBody?.type || mapStatusToErrorType(response.status)
        )
        return
      }

      if (!response.body) {
        callbacks.onError('Response body is empty', 'server_error')
        return
      }

      await consumeSSE(response.body, callbacks)
    } catch (err) {
      // AbortController 中断 → 静默不报错
      if ((err as Error)?.name === 'AbortError') {
        return
      }
      callbacks.onError(
        (err as Error)?.message || 'Network error',
        'server_error'
      )
    } finally {
      isStreaming.value = false
      abortController = null
    }
  }

  /** 主动中止当前流式响应 */
  function stop() {
    if (abortController) {
      abortController.abort()
      abortController = null
    }
    isStreaming.value = false
  }

  return {
    send,
    stop,
    isStreaming,
  }
}

// ==================== 工具函数 ====================

/**
 * 消费 SSE ReadableStream
 *
 * SSE 协议规范：
 *   - 每条事件以空行 `\n\n` 分隔
 *   - 数据行格式：`data: <payload>`
 *   - `data: [DONE]` 表示流结束
 *
 * 难点：跨 chunk 截断 —— `data:` 可能横跨两次 reader.read() 返回的字节块
 * 解决：用 buffer 累积，按 `\n` 拆分，最后一个不完整片段留在 buffer 等下一轮
 */
async function consumeSSE(
  body: ReadableStream<Uint8Array>,
  callbacks: StreamCallbacks
): Promise<void> {
  const reader = body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { value, done } = await reader.read()
    if (done) {
      // 流自然结束（上游主动关闭、未发 [DONE]）
      callbacks.onComplete()
      return
    }

    buffer += decoder.decode(value, { stream: true })

    // 按行拆分；最后一行可能不完整，留在 buffer
    const lines = buffer.split('\n')
    buffer = lines.pop() ?? ''

    for (const rawLine of lines) {
      const line = rawLine.trim()
      if (!line) continue
      if (!line.startsWith('data:')) continue

      const data = line.slice(5).trim()
      if (data === '[DONE]') {
        callbacks.onComplete()
        return
      }

      try {
        const chunk = JSON.parse(data) as ChatCompletionChunk
        const delta = chunk.choices?.[0]?.delta
        if (!delta) continue
        if (delta.reasoning_content) {
          callbacks.onUpdate('reasoning', delta.reasoning_content)
        }
        if (delta.content) {
          callbacks.onUpdate('content', delta.content)
        }
      } catch {
        // 跳过无法解析的 chunk（可能是心跳 ping）
      }
    }
  }
}

/** 安全解析错误响应体（OpenAI 兼容 { error: { type, message } }） */
async function safeParseErrorBody(
  response: Response
): Promise<{ type: string; message: string } | null> {
  try {
    const data = (await response.json()) as {
      error?: { type?: string; message?: string }
    }
    if (data?.error) {
      return {
        type: data.error.type || mapStatusToErrorType(response.status),
        message: data.error.message || `HTTP ${response.status}`,
      }
    }
  } catch {
    /* not JSON */
  }
  return null
}

/** HTTP 状态码 → OpenAI 错误类型映射 */
function mapStatusToErrorType(status: number): string {
  if (status === 401) return 'authentication_error'
  if (status === 403) return 'permission_error'
  if (status === 400) return 'invalid_request_error'
  if (status === 402) return 'insufficient_quota'
  if (status === 429) return 'rate_limit_error'
  if (status >= 500) return 'server_error'
  return 'api_error'
}
