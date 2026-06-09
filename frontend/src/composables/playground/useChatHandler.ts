/**
 * 对话广场业务编排 composable
 *
 * 职责：
 *   - 拼装 OpenAI 兼容 payload（含参数开关 + system prompt）
 *   - 调用 useStreamChat 发送流式请求
 *   - 处理流式增量更新消息状态机
 *   - 解析 <think> 标签为 reasoning 字段
 *   - 错误本地化映射
 *   - finalize 流结束 / 中断
 *
 * 移植自 new-api/web/default/src/features/playground/hooks/use-chat-handler.ts
 * 适配 Vue 3 Composition API
 */

import type { Ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  MESSAGE_ROLES,
  MESSAGE_STATUS,
  ERROR_CODE_I18N_MAP,
} from '@/constants/playground'
import type {
  Message,
  PlaygroundConfig,
  ParameterEnabled,
  ChatCompletionRequest,
  ChatCompletionMessage,
  MessageRole,
  ContentPart,
  PlaygroundAttachment,
} from '@/types/playground'
import { useStreamChat, type StreamUpdateType } from './useStreamChat'

interface UseChatHandlerOptions {
  config: Ref<PlaygroundConfig>
  parameterEnabled: Ref<ParameterEnabled>
  messages: Ref<Message[]>
  updateMessages: (updater: (prev: Message[]) => Message[]) => void
}

export function useChatHandler(opts: UseChatHandlerOptions) {
  const { t, te } = useI18n()
  const { send, stop, isStreaming } = useStreamChat()

  // ==================== Payload 构造 ====================

  function buildPayload(messagesForRequest: Message[]): ChatCompletionRequest {
    const cfg = opts.config.value
    const enabled = opts.parameterEnabled.value

    // 1. 把 Message 转 OpenAI ChatCompletionMessage
    const chatMessages: ChatCompletionMessage[] = []

    // System prompt（若有）
    if (cfg.systemPrompt && cfg.systemPrompt.trim()) {
      chatMessages.push({
        role: MESSAGE_ROLES.SYSTEM as MessageRole,
        content: cfg.systemPrompt.trim(),
      })
    }

    // 历史消息（过滤错误/加载中、取 versions[0].content）
    for (const msg of messagesForRequest) {
      if (msg.status === MESSAGE_STATUS.ERROR) continue
      if (msg.status === MESSAGE_STATUS.LOADING) continue
      const content = msg.versions?.[0]?.content
      const attachments = msg.attachments || []
      if (
        (!content || !content.trim()) &&
        !attachments.some(hasUsableAttachment)
      ) {
        continue
      }
      chatMessages.push({
        role: msg.from,
        content: buildMessageContent(content || '', attachments),
      })
    }

    // 2. 基础 payload
    const payload: ChatCompletionRequest = {
      model: cfg.model,
      group: cfg.group,
      messages: chatMessages,
      stream: cfg.stream,
    }

    // 3. 按启用开关追加可选参数
    const paramKeys = [
      'temperature',
      'top_p',
      'max_tokens',
      'frequency_penalty',
      'presence_penalty',
      'seed',
    ] as const
    for (const key of paramKeys) {
      if (enabled[key]) {
        const v = cfg[key]
        if (v !== undefined && v !== null) {
          ;(payload as unknown as Record<string, unknown>)[key] = v
        }
      }
    }

    return payload
  }

  // ==================== 流式状态机 ====================

  /** 处理 `<think>...</think>` 标签：把内容转入 reasoning 字段，仅 content 部分留作可见输出 */
  function processStreamChunk(
    msg: Message,
    type: StreamUpdateType,
    chunk: string
  ): Message {
    if (type === 'reasoning') {
      return {
        ...msg,
        reasoning: {
          content: (msg.reasoning?.content || '') + chunk,
          duration: msg.reasoning?.duration ?? 0,
        },
        isReasoningStreaming: true,
        status: MESSAGE_STATUS.STREAMING,
      }
    }

    // content 增量：识别 <think> 包裹
    const version = msg.versions?.[0]
    const oldContent = version?.content || ''
    const accumulated = oldContent + chunk

    // 简单状态机：检测是否处于 <think> 块中
    const thinkOpenIdx = accumulated.lastIndexOf('<think>')
    const thinkCloseIdx = accumulated.lastIndexOf('</think>')

    let visibleContent = accumulated
    let reasoningContent = msg.reasoning?.content || ''
    let isReasoningStreaming = msg.isReasoningStreaming || false

    if (thinkOpenIdx >= 0) {
      if (thinkCloseIdx > thinkOpenIdx) {
        // <think>...</think> 闭合完成
        const beforeThink = accumulated.slice(0, thinkOpenIdx)
        const insideThink = accumulated.slice(thinkOpenIdx + 7, thinkCloseIdx)
        const afterThink = accumulated.slice(thinkCloseIdx + 8)
        reasoningContent = insideThink
        visibleContent = beforeThink + afterThink
        isReasoningStreaming = false
      } else {
        // <think> 开了但没闭 → 还在思考中
        const beforeThink = accumulated.slice(0, thinkOpenIdx)
        const insideThink = accumulated.slice(thinkOpenIdx + 7)
        reasoningContent = insideThink
        visibleContent = beforeThink
        isReasoningStreaming = true
      }
    }

    return {
      ...msg,
      versions: [
        {
          ...(version || { id: 'v0', content: '' }),
          content: visibleContent,
        },
        ...(msg.versions?.slice(1) || []),
      ],
      reasoning: reasoningContent
        ? {
            content: reasoningContent,
            duration: msg.reasoning?.duration ?? 0,
          }
        : msg.reasoning,
      isReasoningStreaming,
      status: MESSAGE_STATUS.STREAMING,
    }
  }

  /** 把最后一条 assistant 消息标记为完成 */
  function finalizeMessage(msg: Message): Message {
    return {
      ...msg,
      isReasoningStreaming: false,
      isReasoningComplete: true,
      isContentComplete: true,
      status: MESSAGE_STATUS.COMPLETE,
    }
  }

  /** 更新最后一条 assistant 消息（找不到则不动） */
  function patchLastAssistant(updater: (m: Message) => Message) {
    opts.updateMessages((prev) => {
      const next = [...prev]
      for (let i = next.length - 1; i >= 0; i--) {
        if (next[i].from === MESSAGE_ROLES.ASSISTANT) {
          next[i] = updater(next[i])
          return next
        }
      }
      return next
    })
  }

  // ==================== 业务 API ====================

  /** 发送一轮对话；调用前应已把 user 消息 + loading assistant 占位添加到 messages */
  async function sendChat(messagesIncludingPlaceholder: Message[]) {
    const payload = buildPayload(
      // 不包含最后一条 loading 占位
      messagesIncludingPlaceholder.slice(0, -1)
    )

    await send(payload, {
      onUpdate: (type, chunk) => {
        patchLastAssistant((m) => {
          if (m.status === MESSAGE_STATUS.ERROR) return m
          return processStreamChunk(m, type, chunk)
        })
      },
      onComplete: () => {
        patchLastAssistant((m) => {
          if (
            m.status === MESSAGE_STATUS.COMPLETE ||
            m.status === MESSAGE_STATUS.ERROR
          ) {
            return m
          }
          return finalizeMessage(m)
        })
      },
      onError: (errMessage, errCode) => {
        const i18nKey =
          (errCode && ERROR_CODE_I18N_MAP[errCode]) ||
          'playground.error.serverError'
        const localized = te(i18nKey) ? t(i18nKey) : errMessage

        patchLastAssistant((m) => ({
          ...m,
          status: MESSAGE_STATUS.ERROR,
          errorCode: errCode ?? null,
          versions: [
            {
              ...(m.versions?.[0] || { id: 'v0', content: '' }),
              content: localized,
            },
          ],
        }))
      },
    })
  }

  /** 中断当前生成（仅停流，最后一条消息标记为完成） */
  function stopGeneration() {
    stop()
    patchLastAssistant((m) => {
      if (
        m.status === MESSAGE_STATUS.STREAMING ||
        m.status === MESSAGE_STATUS.LOADING
      ) {
        return finalizeMessage(m)
      }
      return m
    })
  }

  return {
    sendChat,
    stopGeneration,
    isGenerating: isStreaming,
  }
}

// ==================== 工厂函数（消息创建） ====================

function hasUsableAttachment(item: PlaygroundAttachment): boolean {
  return (
    (item.kind === 'image' && !!item.dataUrl) ||
    (item.kind === 'document' && !!item.text?.trim())
  )
}

function buildMessageContent(
  text: string,
  attachments: PlaygroundAttachment[]
): string | ContentPart[] {
  const usableAttachments = attachments.filter(hasUsableAttachment)
  if (usableAttachments.length === 0) return text

  const documentBlocks = usableAttachments
    .filter((item) => item.kind === 'document' && item.text)
    .map(
      (item) =>
        `--- ${item.name} (${item.type || 'text/plain'}, ${item.size} bytes) ---\n${item.text}`
    )

  const textParts = [text.trim()]
  if (documentBlocks.length > 0) {
    textParts.push(
      `Attached documents:\n\n${documentBlocks.join('\n\n')}`
    )
  }

  const parts: ContentPart[] = []
  const mergedText = textParts.filter(Boolean).join('\n\n')
  if (mergedText) {
    parts.push({ type: 'text', text: mergedText })
  }

  for (const item of usableAttachments) {
    if (item.kind !== 'image' || !item.dataUrl) continue
    parts.push({
      type: 'image_url',
      image_url: { url: item.dataUrl },
    })
  }

  return parts.length === 1 && parts[0].type === 'text'
    ? parts[0].text || ''
    : parts
}

export function createUserMessage(
  text: string,
  attachments: PlaygroundAttachment[] = []
): Message {
  return {
    key: `user-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    from: MESSAGE_ROLES.USER,
    versions: [{ id: 'v0', content: text }],
    status: MESSAGE_STATUS.COMPLETE,
    attachments,
  }
}

export function createLoadingAssistantMessage(): Message {
  return {
    key: `asst-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    from: MESSAGE_ROLES.ASSISTANT,
    versions: [{ id: 'v0', content: '' }],
    status: MESSAGE_STATUS.LOADING,
  }
}
