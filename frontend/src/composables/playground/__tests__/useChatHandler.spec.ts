import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import {
  createLoadingAssistantMessage,
  createUserMessage,
  normalizePlaygroundVideoResponse,
  useChatHandler,
} from '../useChatHandler'
import type {
  Message,
  ParameterEnabled,
  PlaygroundConfig,
  PlaygroundVideoState,
} from '@/types/playground'

const apiMocks = vi.hoisted(() => ({
  create: vi.fn(),
  get: vi.fn(),
}))

vi.mock('@/api/playground', () => ({
  createPlaygroundVideo: apiMocks.create,
  getPlaygroundVideo: apiMocks.get,
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key, te: () => true }),
}))

vi.mock('../useStreamChat', () => ({
  playgroundVideoGenerating: { value: false },
  useStreamChat: () => ({
    send: vi.fn(),
    stop: vi.fn(),
    isStreaming: { value: false },
  }),
}))

beforeEach(() => {
  vi.useFakeTimers()
  vi.clearAllMocks()
})

afterEach(() => vi.useRealTimers())

describe('normalizePlaygroundVideoResponse', () => {
  it.each([
    [{ id: 'queued', status: 'queued' }, 'queued', 0, undefined],
    [{ id: 'running', status: 'processing', progress: '37' }, 'in_progress', 37, undefined],
    [
      {
        id: 'done',
        status: 'completed',
        progress: 82,
        video_url: 'https://cdn.example.com/video.mp4',
      },
      'completed',
      100,
      'https://cdn.example.com/video.mp4',
    ],
    [{ id: 'failed', status: 'failed', progress: 61 }, 'failed', 61, undefined],
  ] as const)('normalizes %s', (response, status, progress, url) => {
    expect(normalizePlaygroundVideoResponse(response)).toEqual({
      id: response.id,
      status,
      progress,
      url,
    })
  })
})

describe('useChatHandler video flow', () => {
  it('sends a 1.5 reference image and reaches completed at 100%', async () => {
    apiMocks.create.mockResolvedValue({ id: 'task-1', status: 'queued', progress: 0 })
    apiMocks.get
      .mockResolvedValueOnce({ id: 'task-1', status: 'processing', progress: 43 })
      .mockResolvedValueOnce({
        id: 'task-1',
        status: 'completed',
        progress: 97,
        video_url: 'https://cdn.example.com/task-1.mp4',
      })

    const config = ref<PlaygroundConfig>({
      model: 'grok-imagine-video-1.5-preview',
      group: 'video',
      temperature: 1,
      top_p: 1,
      max_tokens: 4096,
      frequency_penalty: 0,
      presence_penalty: 0,
      seed: null,
      stream: true,
      systemPrompt: '',
    })
    const parameterEnabled = ref<ParameterEnabled>({
      temperature: false,
      top_p: false,
      max_tokens: false,
      frequency_penalty: false,
      presence_penalty: false,
      seed: false,
    })
    const user = createUserMessage('animate this image', [
      {
        id: 'image-1',
        kind: 'image',
        name: 'reference.png',
        type: 'image/png',
        size: 5,
        dataUrl: 'data:image/png;base64,aW1hZ2U=',
      },
    ])
    const messages = ref<Message[]>([user, createLoadingAssistantMessage()])
    const snapshots: PlaygroundVideoState[] = []
    const onSettled = vi.fn()
    const handler = useChatHandler({
      config,
      parameterEnabled,
      messages,
      updateMessages: (updater) => {
        messages.value = updater(messages.value)
        const video = messages.value.at(-1)?.versions?.[0]?.video
        if (video) snapshots.push({ ...video })
      },
      onSettled,
    })

    const pending = handler.sendVideo(messages.value)
    await vi.advanceTimersByTimeAsync(2000)
    await vi.advanceTimersByTimeAsync(2000)
    await pending

    expect(apiMocks.create).toHaveBeenCalledWith(
      {
        model: 'grok-imagine-video-1.5-preview',
        group: 'video',
        prompt: 'animate this image',
        input_reference: { image_url: 'data:image/png;base64,aW1hZ2U=' },
      },
      expect.any(AbortSignal)
    )
    expect(apiMocks.get).toHaveBeenNthCalledWith(1, 'task-1', 'video', expect.any(AbortSignal))
    expect(snapshots.map(({ status, progress }) => [status, progress])).toEqual([
      ['creating', 0],
      ['queued', 0],
      ['in_progress', 43],
      ['completed', 100],
    ])
    expect(messages.value.at(-1)?.versions?.[0]?.video?.url).toBe(
      'https://cdn.example.com/task-1.mp4'
    )
    expect(onSettled).toHaveBeenCalledOnce()
  })
})
