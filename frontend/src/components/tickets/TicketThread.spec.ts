import { createI18n } from 'vue-i18n'
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { TicketMessage } from '@/types/ticket'
import TicketThread from './TicketThread.vue'

describe('TicketThread', () => {
  it('renders message bodies as text and never as HTML', () => {
    const i18n = createI18n({
      legacy: false,
      locale: 'zh',
      messages: { zh: { tickets: { unknownAuthor: () => '未知用户' } } },
    })
    const messages: TicketMessage[] = [{
      id: 1,
      ticket_id: 2,
      author_role: 'user',
      kind: 'public',
      visibility: 'user',
      body: '<img src=x onerror=alert(1)>',
      attachments: [],
      created_at: '2026-07-12T00:00:00Z',
    }]

    const wrapper = mount(TicketThread, { props: { messages }, global: { plugins: [i18n] } })

    expect(wrapper.text()).toContain('<img src=x onerror=alert(1)>')
    expect(wrapper.find('img').exists()).toBe(false)
  })
})
