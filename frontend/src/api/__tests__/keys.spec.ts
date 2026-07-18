import { beforeEach, describe, expect, it, vi } from 'vitest'

const { post, put } = vi.hoisted(() => ({
  post: vi.fn(),
  put: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: { post, put },
}))

import { keysAPI } from '@/api/keys'

describe('keys api ordered groups', () => {
  beforeEach(() => {
    post.mockReset().mockResolvedValue({ data: {} })
    put.mockReset().mockResolvedValue({ data: {} })
  })

  it('preserves group priority when creating a key', async () => {
    await keysAPI.create('night-job', [3, 1, 2])

    expect(post).toHaveBeenCalledWith('/keys', {
      name: 'night-job',
      group_id: 3,
      group_ids: [3, 1, 2],
    })
  })

  it('preserves group priority when updating a key', async () => {
    await keysAPI.update(7, { group_id: 2, group_ids: [2, 4, 1] })

    expect(put).toHaveBeenCalledWith('/keys/7', {
      group_id: 2,
      group_ids: [2, 4, 1],
    })
  })
})
