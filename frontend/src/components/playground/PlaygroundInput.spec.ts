import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import PlaygroundInput from './PlaygroundInput.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

describe('PlaygroundInput video mode', () => {
  afterEach(() => vi.unstubAllGlobals())

  it('shows video options, hides documents, and keeps the reference image removable', async () => {
    vi.stubGlobal('FileReader', class {
      result = ''
      onload: (() => void) | null = null

      readAsDataURL() {
        this.result = 'data:image/png;base64,aW1hZ2U='
        this.onload?.()
      }
    })

    const wrapper = mount(PlaygroundInput, {
      props: {
        modelValue: 'grok-imagine-video-1.5-preview',
        models: [{ label: 'Video 1.5', value: 'grok-imagine-video-1.5-preview' }],
        groupValue: 'video',
        groups: [{ label: 'Video', value: 'video', ratio: 1, platform: 'video' }],
        videoMode: true,
        videoSeconds: '4',
        videoAspectRatio: '9:16',
      },
      global: { plugins: [createPinia()] },
    })

    const videoSelects = wrapper.findAll('.playground-video-select')
    expect(videoSelects).toHaveLength(2)
    await videoSelects[0].setValue('8')
    await videoSelects[1].setValue('16:9')
    expect(wrapper.emitted('videoSecondsChange')?.at(-1)).toEqual(['8'])
    expect(wrapper.emitted('videoAspectRatioChange')?.at(-1)).toEqual(['16:9'])

    expect(wrapper.find('[title="playground.input.attachDocument"]').exists()).toBe(false)
    expect(wrapper.find('input[accept*=".txt"]').exists()).toBe(false)

    const input = wrapper.get('input[accept^="image/png"]')
    Object.defineProperty(input.element, 'files', {
      configurable: true,
      value: [new File(['image'], 'reference.png', { type: 'image/png' })],
    })
    await input.trigger('change')
    await flushPromises()

    expect(wrapper.find('.playground-attachment-remove').exists()).toBe(true)

    await wrapper.setProps({ modelValue: 'grok-video-3-pro' })
    expect(wrapper.find('.playground-video-options').exists()).toBe(false)
  })
})
