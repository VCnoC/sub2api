import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import { describe, expect, it } from 'vitest'
import SubscriptionPlanCard from '@/components/payment/SubscriptionPlanCard.vue'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'

const plan: SubscriptionPlan = {
  id: 1,
  group_id: 7,
  name: 'Test plan',
  description: '',
  price: 10,
  validity_days: 30,
  validity_unit: 'day',
  features: [],
  for_sale: true,
  sort_order: 0
}

function mountCard(
  activeSubscriptions: UserSubscription[] = [],
  planOverrides: Partial<SubscriptionPlan> = {}
) {
  const i18n = createI18n({
    legacy: false,
    locale: 'en',
    messages: {
      en: {
        payment: {
          buyAgain: () => 'Buy again',
          subscribeNow: () => 'Subscribe',
          days: () => ' days',
          weeks: () => ' weeks',
          months: () => ' months',
          perMonth: () => 'month',
          planCard: {
            rate: () => 'Rate',
            quota: () => 'Quota',
            unlimited: () => 'Unlimited',
            models: () => 'Models',
            requestLimit5h: () => '5-hour requests',
            requestLimit1d: () => '24-hour requests'
          }
        }
      }
    }
  })
  return mount(SubscriptionPlanCard, {
    props: { plan: { ...plan, ...planOverrides }, activeSubscriptions },
    global: { plugins: [createPinia(), i18n] }
  })
}

describe('SubscriptionPlanCard', () => {
  it('shows buy again for another active entitlement in the same group', () => {
    const active = [{ id: 9, group_id: plan.group_id, status: 'active' }] as UserSubscription[]
    expect(mountCard(active).get('button').text()).toBe('Buy again')
  })

  it('shows subscribe for a group without an active entitlement', () => {
    expect(mountCard().get('button').text()).toBe('Subscribe')
  })

  it('does not show Antigravity model scopes for OpenAI plans', () => {
    const text = mountCard([], {
      group_platform: 'openai',
      supported_model_scopes: ['claude', 'gemini_text', 'gemini_image']
    }).text()

    expect(text).not.toContain('Claude')
    expect(text).not.toContain('Gemini')
    expect(text).not.toContain('Imagen')
  })

  it('shows model scopes for Antigravity plans', () => {
    const text = mountCard([], {
      group_platform: 'antigravity',
      supported_model_scopes: ['claude', 'gemini_text', 'gemini_image']
    }).text()

    expect(text).toContain('Claude')
    expect(text).toContain('Gemini')
    expect(text).toContain('Imagen')
  })

  it('shows request-count limits instead of USD quota labels', () => {
    const text = mountCard([], {
      subscription_billing_mode: 'request_count',
      request_limit_5h: 20,
      request_limit_1d: 50,
      daily_limit_usd: null,
      weekly_limit_usd: null,
      monthly_limit_usd: null
    }).text()

    expect(text).toContain('5-hour requests')
    expect(text).toContain('20')
    expect(text).toContain('24-hour requests')
    expect(text).toContain('50')
    expect(text).not.toContain('Unlimited')
  })

  it('normalizes plural validity units saved by the admin form', () => {
    expect(mountCard([], { validity_days: 1, validity_unit: 'months' }).text()).toContain('/ month')
    expect(mountCard([], { validity_days: 3, validity_unit: 'months' }).text()).toContain('/ 3 months')
    expect(mountCard([], { validity_days: 2, validity_unit: 'weeks' }).text()).toContain('/ 2 weeks')
  })

  it('uses the configured currency symbol and preserves legacy USD defaults', () => {
    expect(mountCard([], { currency: 'CNY', original_price: 20 }).text()).toContain('¥10CNY')
    expect(mountCard([], { currency: 'USD' }).text()).toContain('$10USD')
    expect(mountCard([], { currency: '' }).text()).toContain('$10')
  })
})
