import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import LotteryReel from './LotteryReel.vue'
import type { LotteryDraw, LotteryPrize } from '@/types/lottery'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

const prizes: LotteryPrize[] = [
  {
    id: 1,
    pool_id: 1,
    name: 'Balance prize',
    description: '',
    prize_type: 'balance',
    balance_amount: 5,
    probability_ppm: 200_000,
    stock_used: 0,
    enabled: true,
    sort_order: 0,
  },
  {
    id: 2,
    pool_id: 1,
    name: 'Subscription prize',
    description: '',
    prize_type: 'subscription',
    group_id: 2,
    validity_days: 30,
    probability_ppm: 100_000,
    stock_used: 0,
    enabled: true,
    sort_order: 1,
  },
]

function draw(overrides: Partial<LotteryDraw> = {}): LotteryDraw {
  return {
    id: 10,
    pool_id: 1,
    outcome: 'win',
    chance_source: 'base',
    prize_id: 2,
    prize: { id: 2, name: 'Subscription prize', prize_type: 'subscription', validity_days: 30 },
    created_at: '2026-07-12T00:00:00Z',
    ...overrides,
  }
}

describe('LotteryReel', () => {
  beforeEach(() => {
    window.matchMedia = vi.fn().mockReturnValue({ matches: true }) as unknown as typeof window.matchMedia
  })

  it('reduced motion 会直接停在服务端返回的奖品', async () => {
    const wrapper = mount(LotteryReel, { props: { prizes }, global: { stubs: { Icon: true } } })

    await (wrapper.vm as unknown as { play: (value: LotteryDraw) => Promise<void> }).play(draw())

    expect(wrapper.find('.lottery_reel__veil').exists()).toBe(false)
    expect(wrapper.find('.lottery_prize_card--winner').text()).toContain('Subscription prize')
  })

  it('未中奖结果会停在未中奖卡牌', async () => {
    const wrapper = mount(LotteryReel, { props: { prizes }, global: { stubs: { Icon: true } } })

    await (wrapper.vm as unknown as { play: (value: LotteryDraw) => Promise<void> }).play(draw({ outcome: 'none', prize_id: null, prize: undefined }))

    expect(wrapper.find('.lottery_prize_card--winner').classes()).toContain('lottery_prize_card--empty')
  })
})
