import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post, put } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: { get, post, put },
}))

import { createConversation, getConversation, updateConversation } from '@/api/playground'

describe('playground conversation API', () => {
  beforeEach(() => {
    get.mockReset().mockResolvedValue({ data: { id: 337, messages: [] } })
    post.mockReset().mockResolvedValue({ data: { id: 338, messages: [] } })
    put.mockReset().mockResolvedValue({ data: undefined })
  })

  it('uses a 60 second timeout for large conversation payloads', async () => {
    await getConversation(337)
    await createConversation({ title: 'test', messages: [] })
    await updateConversation(337, { model: null, group_name: null, messages: [] })

    expect(get).toHaveBeenCalledWith('/playground/conversations/337', { timeout: 60_000 })
    expect(post).toHaveBeenCalledWith(
      '/playground/conversations',
      { title: 'test', messages: [] },
      { timeout: 60_000 },
    )
    expect(put).toHaveBeenCalledWith(
      '/playground/conversations/337',
      { model: null, group_name: null, messages: [] },
      { timeout: 60_000 },
    )
  })
})
