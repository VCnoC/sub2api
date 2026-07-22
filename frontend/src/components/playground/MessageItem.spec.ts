import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import MessageItem from './MessageItem.vue'
import type { Message } from '@/types/playground'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard: vi.fn() }),
}))

function videoMessage(status: 'creating' | 'completed', progress: number, url?: string): Message {
  return {
    key: 'assistant-video',
    from: 'assistant',
    status: status === 'creating' ? 'loading' : 'complete',
    versions: [
      {
        id: 'v0',
        content: '',
        video: { status, progress, url },
      },
    ],
  }
}

describe('MessageItem video progress', () => {
  it('shows 0% while creating and a player at 100% completion', async () => {
    const wrapper = mount(MessageItem, {
      props: { message: videoMessage('creating', 0) },
      global: { plugins: [createPinia()] },
    })

    expect(wrapper.get('[role="progressbar"]').attributes('aria-valuenow')).toBe('0')
    expect(wrapper.get('.message-video-percent').text()).toBe('0%')
    expect(wrapper.find('.message-loader').exists()).toBe(false)
    expect(wrapper.find('video').exists()).toBe(false)

    await wrapper.setProps({
      message: videoMessage(
        'completed',
        100,
        'https://cdn.example.com/video.mp4'
      ),
    })

    expect(wrapper.get('[role="progressbar"]').attributes('aria-valuenow')).toBe('100')
    expect(wrapper.get('.message-video-percent').text()).toBe('100%')
    expect(wrapper.get('video').attributes('src')).toBe('https://cdn.example.com/video.mp4')
  })

  it('renders generated images returned as a data URL', () => {
    const wrapper = mount(MessageItem, {
      props: {
        message: {
          key: 'assistant-image',
          from: 'assistant',
          status: 'complete',
          versions: [
            {
              id: 'v0',
              content: '![image-1](data:image/png;base64,aW1hZ2U=)',
            },
          ],
        },
      },
      global: { plugins: [createPinia()] },
    })

    expect(wrapper.get('.markdown-body img').attributes('src')).toBe(
      'data:image/png;base64,aW1hZ2U='
    )
  })
})
