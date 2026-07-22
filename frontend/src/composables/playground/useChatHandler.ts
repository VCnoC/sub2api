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

import { computed, type Ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  MESSAGE_ROLES,
  MESSAGE_STATUS,
  ERROR_CODE_I18N_MAP,
  isGrokImagineVideoModel,
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
  PlaygroundVideoResponse,
  PlaygroundVideoState,
  PlaygroundImageResponse,
} from '@/types/playground'
import {
  playgroundVideoGenerating,
  useStreamChat,
  type StreamUpdateType,
} from './useStreamChat'
import {
  createPlaygroundImage,
  createPlaygroundVideo,
  getPlaygroundVideo,
} from '@/api/playground'

const VIDEO_POLL_INTERVAL_MS = 2000
const VIDEO_POLL_TIMEOUT_MS = 10 * 60 * 1000

interface UseChatHandlerOptions {
  config: Ref<PlaygroundConfig>
  parameterEnabled: Ref<ParameterEnabled>
  messages: Ref<Message[]>
  updateMessages: (updater: (prev: Message[]) => Message[]) => void
  /**
   * 每轮流式请求收尾后回调（正常完成 / 出错 / 中断均触发）。
   * 用于触发会话保存——以前由视图层 watch(isGenerating) 实现，
   * 但组件级 watch 在路由切走后被销毁，后台完成的生成无法落库；
   * 改为闭包回调后与组件生命周期解耦。
   */
  onSettled?: () => void
}

/**
 * 模块级单例缓存：首次（组件 setup 中）调用时创建并缓存。
 * 流式状态机闭包脱离组件生命周期，路由切换后生成继续在后台跑。
 * 后续调用忽略传入的 opts（所有 opts 均为同一组全局单例的引用）。
 */
let _singleton: ReturnType<typeof createChatHandler> | null = null

export function useChatHandler(opts: UseChatHandlerOptions) {
  if (!_singleton) {
    _singleton = createChatHandler(opts)
  }
  return _singleton
}

function createChatHandler(opts: UseChatHandlerOptions) {
  const { t, te } = useI18n()
  const { send, stop, isStreaming } = useStreamChat()
  const isGenerating = computed(
    () => isStreaming.value || playgroundVideoGenerating.value
  )
  let videoAbortController: AbortController | null = null
  let imageAbortController: AbortController | null = null

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
    const payloadParams = payload as unknown as Record<string, unknown>
    for (const key of paramKeys) {
      if (enabled[key]) {
        const v = cfg[key]
        if (v !== undefined && v !== null) {
          payloadParams[key] = v
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

  function patchVideoMessage(
    video: PlaygroundVideoState,
    status: Message['status'],
    content?: string
  ) {
    patchLastAssistant((msg) => ({
      ...msg,
      status,
      versions: [
        {
          ...(msg.versions?.[0] || { id: 'v0', content: '' }),
          ...(content === undefined ? {} : { content }),
          video,
        },
        ...(msg.versions?.slice(1) || []),
      ],
    }))
  }

  // ==================== 业务 API ====================

  /** 发送一轮对话；调用前应已把 user 消息 + loading assistant 占位添加到 messages */
  async function sendChat(messagesIncludingPlaceholder: Message[]) {
    const payload = buildPayload(
      // 不包含最后一条 loading 占位
      messagesIncludingPlaceholder.slice(0, -1)
    )

    try {
      await doSend(payload)
    } finally {
      // 不论完成 / 出错 / 中断，都通知调用方收尾（防抖保存会话）
      opts.onSettled?.()
    }
  }

  async function doSend(payload: ChatCompletionRequest) {
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

  async function sendVideo(messagesIncludingPlaceholder: Message[]) {
    const { model, group, videoSeconds, videoAspectRatio } = opts.config.value
    const userMessage = messagesIncludingPlaceholder
      .slice(0, -1)
      .reverse()
      .find((message) => message.from === MESSAGE_ROLES.USER)
    const image = userMessage?.attachments?.find(
      (item) => item.kind === 'image' && item.dataUrl
    )
    const controller = new AbortController()
    videoAbortController = controller
    playgroundVideoGenerating.value = true
    patchVideoMessage({ status: 'creating', progress: 0 }, MESSAGE_STATUS.LOADING)

    try {
      const created = await createPlaygroundVideo(
        {
          model,
          group,
          prompt: userMessage?.versions?.[0]?.content?.trim() || '',
          ...(isGrokImagineVideoModel(model)
            ? { seconds: videoSeconds, aspect_ratio: videoAspectRatio }
            : {}),
          ...(image?.dataUrl
            ? { input_reference: { image_url: image.dataUrl } }
            : {}),
        },
        controller.signal
      )
      if (!created.id?.trim()) {
        throw new Error(t('playground.video.missingTaskId'))
      }

      let progress = 0
      const deadline = Date.now() + VIDEO_POLL_TIMEOUT_MS
      let response = created
      while (!controller.signal.aborted) {
        const video = normalizePlaygroundVideoResponse(response)
        progress = Math.max(progress, video.progress)
        video.progress = progress

        if (video.status === 'completed') {
          if (!video.url) throw new Error(t('playground.video.missingUrl'))
          patchVideoMessage(video, MESSAGE_STATUS.COMPLETE)
          return
        }
        if (video.status === 'failed') {
          patchVideoMessage(
            video,
            MESSAGE_STATUS.ERROR,
            videoResponseError(response) || t('playground.video.failed')
          )
          return
        }

        patchVideoMessage(video, MESSAGE_STATUS.STREAMING)
        if (Date.now() >= deadline) {
          patchVideoMessage(
            { ...video, status: 'stopped' },
            MESSAGE_STATUS.ERROR,
            t('playground.video.timeout')
          )
          return
        }
        await waitForVideoPoll(controller.signal)
        if (controller.signal.aborted) return
        response = await getPlaygroundVideo(
          created.id,
          group,
          controller.signal
        )
      }
    } catch (error) {
      if (controller.signal.aborted) return
      patchVideoMessage(
        { status: 'failed', progress: currentVideoProgress(opts.messages.value) },
        MESSAGE_STATUS.ERROR,
        requestErrorMessage(error)
      )
    } finally {
      if (videoAbortController === controller) {
        videoAbortController = null
        playgroundVideoGenerating.value = false
      }
      opts.onSettled?.()
    }
  }

  async function sendImage(messagesIncludingPlaceholder: Message[]) {
    const cfg = opts.config.value
    const enabled = opts.parameterEnabled.value
    const userMessage = messagesIncludingPlaceholder
      .slice(0, -1)
      .reverse()
      .find((message) => message.from === MESSAGE_ROLES.USER)
    const images = (userMessage?.attachments || [])
      .filter((item) => item.kind === 'image' && item.dataUrl)
      .map((item) => item.dataUrl!)
    // Prefer an inline image by default. Upstream URL results are commonly
    // short-lived or reject browser hotlinking, which leaves a broken image in
    // the conversation even though generation succeeded.
    const responseFormat =
      enabled.imageResponseFormat && cfg.imageResponseFormat
        ? cfg.imageResponseFormat
        : 'b64_json'
    const controller = new AbortController()
    imageAbortController = controller
    playgroundVideoGenerating.value = true

    try {
      const response = await createPlaygroundImage(
        {
          model: cfg.model,
          group: cfg.group,
          prompt: userMessage?.versions?.[0]?.content?.trim() || '',
          n: 1,
          size: cfg.imageSize,
          ...(images.length === 1 ? { image: images[0] } : {}),
          ...(images.length > 1 ? { image: images } : {}),
          ...(enabled.imageQuality && cfg.imageQuality ? { quality: cfg.imageQuality } : {}),
          response_format: responseFormat,
          ...(enabled.imageStyle && cfg.imageStyle ? { style: cfg.imageStyle } : {}),
          ...(enabled.imageBackground && cfg.imageBackground ? { background: cfg.imageBackground } : {}),
          ...(enabled.imageWatermark ? { watermark: cfg.imageWatermark } : {}),
        },
        controller.signal
      )
      patchLastAssistant((msg) => ({
        ...msg,
        status: MESSAGE_STATUS.COMPLETE,
        isContentComplete: true,
        versions: [
          {
            ...(msg.versions?.[0] || { id: 'v0', content: '' }),
            content: imageResponseMarkdown(response),
          },
          ...(msg.versions?.slice(1) || []),
        ],
      }))
    } catch (error) {
      if (controller.signal.aborted) return
      patchLastAssistant((m) => ({
        ...m,
        status: MESSAGE_STATUS.ERROR,
        versions: [
          {
            ...(m.versions?.[0] || { id: 'v0', content: '' }),
            content: requestErrorMessage(error),
          },
        ],
      }))
    } finally {
      if (imageAbortController === controller) {
        imageAbortController = null
        playgroundVideoGenerating.value = false
      }
      opts.onSettled?.()
    }
  }

  /** 中断当前生成（仅停流，最后一条消息标记为完成） */
  function stopGeneration() {
    stop()
    imageAbortController?.abort()
    imageAbortController = null
    const stoppedVideo = videoAbortController !== null
    videoAbortController?.abort()
    videoAbortController = null
    playgroundVideoGenerating.value = false
    patchLastAssistant((m) => {
      if (
        m.status === MESSAGE_STATUS.STREAMING ||
        m.status === MESSAGE_STATUS.LOADING
      ) {
        if (stoppedVideo && m.versions?.[0]?.video) {
          return {
            ...finalizeMessage(m),
            versions: [
              {
                ...m.versions[0],
                content: t('playground.video.stopped'),
                video: { ...m.versions[0].video, status: 'stopped' },
              },
              ...m.versions.slice(1),
            ],
          }
        }
        return finalizeMessage(m)
      }
      return m
    })
  }

  return {
    sendChat,
    sendImage,
    sendVideo,
    stopGeneration,
    isGenerating,
  }
}

export function imageResponseMarkdown(response: PlaygroundImageResponse): string {
  const items = response.data || []
  const lines = items.flatMap((item, index) => {
    // Some compatible upstreams include both fields. The base64 payload is the
    // stable display source; URLs can expire or reject browser-side hotlinking.
    const src = item.b64_json
      ? `data:image/png;base64,${item.b64_json}`
      : item.url || ''
    if (!src) return []
    const title = item.revised_prompt?.trim()
    return [
      `![image-${index + 1}](${src})`,
      ...(title ? [`> ${title}`] : []),
    ]
  })
  return lines.join('\n\n') || 'Image generated, but no image URL was returned.'
}

export function normalizePlaygroundVideoResponse(
  response: PlaygroundVideoResponse
): PlaygroundVideoState {
  const rawStatus = response.status?.trim().toLowerCase()
  const status = (() => {
    if (['completed', 'done', 'succeeded', 'success'].includes(rawStatus)) {
      return 'completed'
    }
    if (['failed', 'error', 'expired', 'cancelled', 'canceled'].includes(rawStatus)) {
      return 'failed'
    }
    if (['queued', 'pending'].includes(rawStatus)) return 'queued'
    return 'in_progress'
  })() as PlaygroundVideoState['status']
  const parsedProgress = Number(response.progress ?? 0)
  const progress =
    status === 'completed'
      ? 100
      : Math.min(100, Math.max(0, Number.isFinite(parsedProgress) ? parsedProgress : 0))

  return {
    id: response.id,
    status,
    progress,
    url: response.video_url || response.video?.url || undefined,
  }
}

function videoResponseError(response: PlaygroundVideoResponse): string {
  if (typeof response.error === 'string') return response.error
  return response.error?.message || response.error?.error || response.message || ''
}

function requestErrorMessage(error: unknown): string {
  const value = error as {
    message?: string
    error?: string | { message?: string; error?: string }
  }
  if (typeof value?.error === 'string') return value.error
  return value?.error?.message || value?.error?.error || value?.message || 'Video request failed'
}

function currentVideoProgress(messages: Message[]): number {
  for (let i = messages.length - 1; i >= 0; i--) {
    const progress = messages[i].versions?.[0]?.video?.progress
    if (progress !== undefined) return progress
  }
  return 0
}

function waitForVideoPoll(signal: AbortSignal): Promise<void> {
  return new Promise((resolve) => {
    if (signal.aborted) {
      resolve()
      return
    }
    const onAbort = () => {
      window.clearTimeout(timer)
      resolve()
    }
    const timer = window.setTimeout(() => {
      signal.removeEventListener('abort', onAbort)
      resolve()
    }, VIDEO_POLL_INTERVAL_MS)
    signal.addEventListener('abort', onAbort, { once: true })
  })
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
