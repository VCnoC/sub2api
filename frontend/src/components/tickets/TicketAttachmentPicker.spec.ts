import { createI18n } from 'vue-i18n'
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import TicketAttachmentPicker from './TicketAttachmentPicker.vue'

function mountPicker(files: File[] = []) {
  const i18n = createI18n({
    legacy: false,
    locale: 'zh',
    messages: {
      zh: {
        tickets: {
          attachments: {
            choose: () => '选择文件',
            rules: () => '上传规则',
            remove: () => '移除附件',
          },
          errors: {
            tooManyFiles: () => '每条消息最多上传 5 个文件',
            fileTooLarge: () => '文件过大',
            filesTooLarge: () => '附件总大小过大',
            invalidFileType: () => '仅支持指定文件类型',
          },
        },
      },
    },
  })
  return mount(TicketAttachmentPicker, { props: { modelValue: files }, global: { plugins: [i18n] } })
}

async function chooseFiles(wrapper: ReturnType<typeof mountPicker>, files: File[]): Promise<void> {
  const input = wrapper.get('input[type="file"]')
  Object.defineProperty(input.element, 'files', { configurable: true, value: files })
  await input.trigger('change')
}

describe('TicketAttachmentPicker', () => {
  it('accepts valid image and text files', async () => {
    const wrapper = mountPicker()
    const files = [
      new File(['image'], 'screen.png', { type: 'image/png' }),
      new File(['details'], 'request.log', { type: 'text/plain' }),
    ]

    await chooseFiles(wrapper, files)

    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([files])
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
  })

  it('rejects more than five files without changing the model', async () => {
    const wrapper = mountPicker()
    const files = Array.from({ length: 6 }, (_, index) => new File(['x'], `${index}.txt`))

    await chooseFiles(wrapper, files)

    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    expect(wrapper.get('[role="alert"]').text()).toContain('最多上传 5 个')
  })

  it('rejects unsupported extensions', async () => {
    const wrapper = mountPicker()

    await chooseFiles(wrapper, [new File(['x'], 'payload.svg', { type: 'image/svg+xml' })])

    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    expect(wrapper.get('[role="alert"]').text()).toContain('仅支持')
  })
})
