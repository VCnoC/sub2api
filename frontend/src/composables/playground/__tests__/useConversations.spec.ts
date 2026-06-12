/**
 * useConversations 单元测试
 *
 * 覆盖：列表加载与旧数据迁移 / 切换会话 / 草稿态创建 / 防抖保存语义 /
 * 删除会话 / PUT 必带 model+group_name（后端清空语义防护）
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { defineComponent } from 'vue'
import { mount, type VueWrapper } from '@vue/test-utils'
import { createConversations } from '../useConversations'
import { playgroundAPI } from '@/api/playground'
import { CONVERSATION_SAVE_DEBOUNCE_MS, STORAGE_KEYS } from '@/constants/playground'
import type { Message } from '@/types/playground'

vi.mock('@/api/playground', () => ({
  playgroundAPI: {
    listConversations: vi.fn(),
    getConversation: vi.fn(),
    createConversation: vi.fn(),
    updateConversation: vi.fn(),
    deleteConversation: vi.fn(),
  },
}))

const mockAPI = vi.mocked(playgroundAPI)

function makeMessage(from: 'user' | 'assistant', content: string): Message {
  return { key: `${from}-${content}`, from, versions: [{ id: 'v0', content }] }
}

function makeDetail(id: number, overrides: Record<string, unknown> = {}) {
  return {
    id,
    title: `会话${id}`,
    model: 'gpt-5.5',
    group_name: 'default',
    messages: [makeMessage('user', 'hello')],
    last_activity_at: '2026-06-11T00:00:00Z',
    created_at: '2026-06-11T00:00:00Z',
    updated_at: '2026-06-11T00:00:00Z',
    ...overrides,
  }
}

describe('useConversations', () => {
  let messages: Message[] = []
  let wrapper: VueWrapper | null = null

  function setup() {
    let api!: ReturnType<typeof createConversations>
    const Harness = defineComponent({
      setup() {
        api = createConversations({
          getMessages: () => messages,
          setMessages: (next) => {
            messages = next
          },
          getModel: () => 'gpt-5.5',
          getGroupName: () => 'default',
          defaultTitle: () => '新对话',
        })
        return () => null
      },
    })
    wrapper = mount(Harness)
    return api
  }

  beforeEach(() => {
    vi.useFakeTimers()
    messages = []
    localStorage.clear()
    mockAPI.listConversations.mockResolvedValue([])
    mockAPI.updateConversation.mockResolvedValue(undefined)
    mockAPI.deleteConversation.mockResolvedValue(undefined)
  })

  afterEach(() => {
    wrapper?.unmount()
    wrapper = null
    vi.useRealTimers()
    vi.clearAllMocks()
  })

  it('加载列表：无旧数据时直接拉取', async () => {
    mockAPI.listConversations.mockResolvedValue([
      makeDetail(1),
      makeDetail(2),
    ])
    const api = setup()
    await api.loadConversations()

    expect(api.conversations.value).toHaveLength(2)
    expect(mockAPI.createConversation).not.toHaveBeenCalled()
  })

  it('旧 localStorage 消息应迁移为新会话并清除存储', async () => {
    const legacy = [makeMessage('user', '旧消息内容')]
    localStorage.setItem(STORAGE_KEYS.MESSAGES, JSON.stringify(legacy))
    mockAPI.createConversation.mockResolvedValue(makeDetail(9))

    const api = setup()
    await api.loadConversations()

    expect(mockAPI.createConversation).toHaveBeenCalledWith(
      expect.objectContaining({
        title: '旧消息内容',
        model: 'gpt-5.5',
        group_name: 'default',
      })
    )
    expect(localStorage.getItem(STORAGE_KEYS.MESSAGES)).toBeNull()
  })

  it('切换会话：懒加载详情并覆盖消息列表', async () => {
    mockAPI.getConversation.mockResolvedValue(makeDetail(3))
    const api = setup()

    await api.selectConversation(3)

    expect(api.activeConversationId.value).toBe(3)
    expect(messages).toHaveLength(1)
    expect(messages[0].versions[0].content).toBe('hello')
  })

  it('切换会话：messages 为 null 时落空数组', async () => {
    mockAPI.getConversation.mockResolvedValue(makeDetail(4, { messages: null }))
    const api = setup()

    await api.selectConversation(4)
    expect(messages).toEqual([])
  })

  it('草稿态防抖保存：有消息时创建会话（标题取首条用户消息前 20 字）', async () => {
    const created = makeDetail(7)
    mockAPI.createConversation.mockResolvedValue(created)
    const api = setup()

    messages = [makeMessage('user', '这是一条非常非常非常长的用户消息内容用来测试标题截断')]
    api.scheduleSave()

    // 防抖窗口内不应触发
    expect(mockAPI.createConversation).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(CONVERSATION_SAVE_DEBOUNCE_MS + 10)

    expect(mockAPI.createConversation).toHaveBeenCalledTimes(1)
    const payload = mockAPI.createConversation.mock.calls[0][0]
    expect(Array.from(payload.title).length).toBeLessThanOrEqual(20)
    expect(api.activeConversationId.value).toBe(7)
    expect(api.conversations.value[0].id).toBe(7)
  })

  it('草稿态空消息不落库', async () => {
    const api = setup()
    api.scheduleSave()
    await vi.advanceTimersByTimeAsync(CONVERSATION_SAVE_DEBOUNCE_MS + 10)

    expect(mockAPI.createConversation).not.toHaveBeenCalled()
    expect(mockAPI.updateConversation).not.toHaveBeenCalled()
  })

  it('已有会话保存：PUT 必须带 model 与 group_name（后端清空语义防护）', async () => {
    mockAPI.getConversation.mockResolvedValue(makeDetail(5))
    const api = setup()
    await api.selectConversation(5)

    messages = [...messages, makeMessage('assistant', '回复')]
    api.scheduleSave()
    await vi.advanceTimersByTimeAsync(CONVERSATION_SAVE_DEBOUNCE_MS + 10)

    expect(mockAPI.updateConversation).toHaveBeenCalledTimes(1)
    const [id, payload] = mockAPI.updateConversation.mock.calls[0]
    expect(id).toBe(5)
    expect(payload.model).toBe('gpt-5.5')
    expect(payload.group_name).toBe('default')
    expect(payload.messages).toHaveLength(2)
  })

  it('连续 scheduleSave 在防抖窗口内只保存一次', async () => {
    mockAPI.getConversation.mockResolvedValue(makeDetail(6))
    const api = setup()
    await api.selectConversation(6)

    api.scheduleSave()
    await vi.advanceTimersByTimeAsync(CONVERSATION_SAVE_DEBOUNCE_MS / 2)
    api.scheduleSave()
    await vi.advanceTimersByTimeAsync(CONVERSATION_SAVE_DEBOUNCE_MS + 10)

    expect(mockAPI.updateConversation).toHaveBeenCalledTimes(1)
  })

  it('删除当前会话后回到草稿态并清空消息', async () => {
    mockAPI.listConversations.mockResolvedValue([makeDetail(8)])
    mockAPI.getConversation.mockResolvedValue(makeDetail(8))
    const api = setup()
    await api.loadConversations()
    await api.selectConversation(8)

    await api.removeConversation(8)

    expect(mockAPI.deleteConversation).toHaveBeenCalledWith(8)
    expect(api.conversations.value).toHaveLength(0)
    expect(api.activeConversationId.value).toBeNull()
    expect(messages).toEqual([])
  })

  it('切换会话前 flush 未保存的内容', async () => {
    mockAPI.getConversation.mockImplementation(async (id: number) => makeDetail(id))
    const api = setup()
    await api.selectConversation(10)

    // 修改消息后立即切换：保存应在切换前 flush 执行
    messages = [...messages, makeMessage('user', '未保存的输入')]
    api.scheduleSave()
    await api.selectConversation(11)

    expect(mockAPI.updateConversation).toHaveBeenCalledTimes(1)
    expect(mockAPI.updateConversation.mock.calls[0][0]).toBe(10)
    expect(api.activeConversationId.value).toBe(11)
  })
})
