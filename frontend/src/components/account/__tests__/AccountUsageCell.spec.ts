import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import AccountUsageCell from '../AccountUsageCell.vue'
import type { Account, AccountUsageInfo } from '@/types'

const { getUsage } = vi.hoisted(() => ({
  getUsage: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: { getUsage }
  }
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

function makeAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: 'account',
    platform: 'anthropic',
    type: 'oauth',
    proxy_id: null,
    concurrency: 1,
    priority: 1,
    status: 'active',
    error_message: null,
    last_used_at: null,
    expires_at: null,
    auto_pause_on_expired: true,
    created_at: '2026-03-15T00:00:00Z',
    updated_at: '2026-03-15T00:00:00Z',
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null,
    ...overrides
  }
}

function makeUsage(overrides: Partial<AccountUsageInfo> = {}): AccountUsageInfo {
  return {
    source: 'passive',
    updated_at: '2026-07-16T08:00:00Z',
    five_hour: null,
    seven_day: null,
    seven_day_sonnet: null,
    ...overrides
  }
}

function mountCell(props: Record<string, unknown>) {
  return mount(AccountUsageCell, {
    props: {
      account: makeAccount(),
      ...props
    },
    global: {
      stubs: {
        UsageProgressBar: {
          props: ['label', 'utilization', 'resetsAt', 'windowStats', 'remainingCapacity'],
          template: '<div class="usage-bar">{{ label }}|{{ utilization }}|{{ resetsAt }}|{{ windowStats?.tokens }}|{{ remainingCapacity }}</div>'
        },
        AccountQuotaInfo: true,
        OpenAIQuotaResetCell: { template: '<div><slot name="pre-actions" /></div>' },
        GrokQuotaProbeCell: true
      }
    }
  })
}

describe('AccountUsageCell', () => {
  beforeEach(() => {
    getUsage.mockReset()
  })

  it('renders the parent snapshot without requesting per-row usage on mount', async () => {
    const wrapper = mountCell({
      usageInfo: makeUsage({
        five_hour: {
          utilization: 41,
          resets_at: '2026-07-16T10:00:00Z',
          remaining_seconds: 7200
        }
      })
    })

    await flushPromises()

    expect(getUsage).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('5h|41|2026-07-16T10:00:00Z')
  })

  it('does not request usage when the parent replaces the snapshot', async () => {
    const wrapper = mountCell({ usageInfo: makeUsage() })

    await wrapper.setProps({
      usageInfo: makeUsage({
        seven_day: {
          utilization: 72,
          resets_at: '2026-07-20T10:00:00Z',
          remaining_seconds: 3600
        }
      })
    })
    await flushPromises()

    expect(getUsage).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('7d|72|2026-07-20T10:00:00Z')
  })

  it('keeps the explicit active refresh as a single forced account request', async () => {
    getUsage.mockResolvedValue(makeUsage({
      source: 'active',
      five_hour: {
        utilization: 88,
        resets_at: '2026-07-16T11:00:00Z',
        remaining_seconds: 3600
      }
    }))
    const wrapper = mountCell({ usageInfo: makeUsage() })

    const activeButton = wrapper.findAll('button').find(button =>
      button.text().includes('admin.accounts.usageWindow.activeQuery')
    )
    expect(activeButton).toBeDefined()
    await activeButton!.trigger('click')
    await flushPromises()

    expect(getUsage).toHaveBeenCalledTimes(1)
    expect(getUsage).toHaveBeenCalledWith(1, 'active', true)
    expect(wrapper.text()).toContain('5h|88|2026-07-16T11:00:00Z')
  })

  it('renders batch loading and per-account errors from parent props', async () => {
    const loadingWrapper = mountCell({ usageLoading: true })
    expect(loadingWrapper.findAll('.animate-pulse').length).toBeGreaterThan(0)

    const errorWrapper = mountCell({ usageError: 'snapshot unavailable' })
    expect(errorWrapper.text()).toContain('snapshot unavailable')
    expect(getUsage).not.toHaveBeenCalled()
  })

  it('renders Antigravity image quota and credits from the passive snapshot', () => {
    const wrapper = mountCell({
      account: makeAccount({ platform: 'antigravity', type: 'oauth' }),
      usageInfo: makeUsage({
        antigravity_quota: {
          'gemini-3-pro-image': { utilization: 70, reset_time: '2026-07-16T09:00:00Z' }
        },
        ai_credits: [{ credit_type: 'GOOGLE_ONE_AI', amount: 25, minimum_balance: 5 }]
      })
    })

    expect(wrapper.text()).toContain('admin.accounts.usageWindow.gemini3Image|70')
    expect(wrapper.text()).toContain('admin.accounts.aiCreditsBalance')
    expect(wrapper.text()).toContain('25')
  })

  it('renders Grok local usage and remaining quota from the passive snapshot', () => {
    const wrapper = mountCell({
      account: makeAccount({ platform: 'grok', type: 'oauth' }),
      usageInfo: makeUsage({
        grok_local_usage: {
          requests: 4,
          tokens: 1200,
          cost: 0.12,
          standard_cost: 0.12,
          user_cost: 0.34
        },
        grok_request_quota: {
          limit: 10,
          remaining: 0,
          reset_at: '2026-07-16T16:00:00Z'
        },
        grok_quota_snapshot_state: 'observed'
      })
    })

    expect(wrapper.text()).toContain('4 req')
    expect(wrapper.text()).toContain('U $0.34')
    expect(wrapper.text()).toContain('admin.accounts.usageWindow.grokRequests|0')
  })

  it('renders Grok quota bars as 100% full and 25% remaining', () => {
    const wrapper = mountCell({
      account: makeAccount({ platform: 'grok', type: 'oauth' }),
      usageInfo: makeUsage({
        grok_request_quota: {
          limit: 100,
          remaining: 100,
          reset_at: '2026-07-16T16:00:00Z'
        },
        grok_token_quota: {
          limit: 1000,
          remaining: 250,
          reset_at: '2026-07-16T16:00:00Z'
        },
        grok_quota_snapshot_state: 'observed'
      })
    })

    expect(wrapper.text()).toContain('admin.accounts.usageWindow.grokRequests|100')
    expect(wrapper.text()).toContain('admin.accounts.usageWindow.grokTokens|25')
  })

  it('renders Grok weekly billing and the dynamic Free 24h token limit', () => {
    const weekly = mountCell({
      account: makeAccount({ platform: 'grok', type: 'oauth' }),
      usageInfo: makeUsage({
        grok_billing: {
          period_type: 'weekly',
          usage_percent: 42,
          period_end: '2026-07-23T00:00:00Z'
        }
      })
    })
    expect(weekly.text()).toContain('7d|42|2026-07-23T00:00:00Z')

    const free = mountCell({
      account: makeAccount({ platform: 'grok', type: 'oauth' }),
      usageInfo: makeUsage({
        subscription_tier: 'FREE',
        grok_free_token_limit: 1_000_000,
        grok_local_usage_24h: {
          requests: 2,
          tokens: 250_000,
          cost: 0,
          standard_cost: 0,
          user_cost: 0
        }
      })
    })
    expect(free.text()).toContain('24h|25')
  })

  it('renders today stats for key accounts without a usage snapshot', () => {
    const wrapper = mountCell({
      account: makeAccount({ type: 'apikey' }),
      todayStats: {
        requests: 1_000_000,
        tokens: 1_000_000_000,
        cost: 12.345,
        standard_cost: 12.345,
        user_cost: 6.789
      }
    })

    expect(wrapper.text()).toContain('1.0M req')
    expect(wrapper.text()).toContain('1.0B')
    expect(wrapper.text()).toContain('A $12.35')
    expect(wrapper.text()).toContain('U $6.79')
  })

  it('renders a today stats skeleton for key accounts while the batch is loading', () => {
    const wrapper = mountCell({
      account: makeAccount({ type: 'apikey' }),
      todayStatsLoading: true
    })

    expect(wrapper.findAll('.animate-pulse').length).toBeGreaterThan(0)
  })

  it('renders a dash for key accounts without stats or quota limits', () => {
    const wrapper = mountCell({
      account: makeAccount({
        type: 'apikey',
        quota_limit: 0,
        quota_daily_limit: 0,
        quota_weekly_limit: 0
      })
    })

    expect(wrapper.text().trim()).toBe('-')
  })

  it('renders Vertex today stats in the Gemini usage cell', () => {
    const wrapper = mountCell({
      account: makeAccount({
        platform: 'gemini',
        type: 'service_account',
        credentials: {
          tier_id: 'vertex',
          project_id: 'vertex-project',
          client_email: 'svc@vertex-project.iam.gserviceaccount.com',
          location: 'global'
        },
        extra: {}
      }),
      todayStats: {
        requests: 0,
        tokens: 0,
        cost: 0,
        standard_cost: 0,
        user_cost: 0
      }
    })

    expect(wrapper.text()).toContain('0 req')
    expect(wrapper.text()).toContain('A $0.00')
    expect(wrapper.text()).toContain('U $0.00')
  })

  it('renders Anthropic Sonnet and Fable windows from the parent snapshot', () => {
    const wrapper = mountCell({
      usageInfo: makeUsage({
        seven_day_sonnet: {
          utilization: 30,
          resets_at: '2026-07-20T10:00:00Z',
          remaining_seconds: 3600
        },
        seven_day_fable: {
          utilization: 100,
          resets_at: '2026-07-20T10:00:00Z',
          remaining_seconds: 3600
        }
      })
    })

    expect(wrapper.text()).toContain('7d S|30')
    expect(wrapper.text()).toContain('7d F|100')
  })

  it('does not render Anthropic Fable when the parent snapshot omits it', () => {
    const wrapper = mountCell({
      usageInfo: makeUsage({
        seven_day: {
          utilization: 56,
          resets_at: '2026-07-20T10:00:00Z',
          remaining_seconds: 3600
        }
      })
    })

    expect(wrapper.text()).toContain('7d|56')
    expect(wrapper.text()).not.toContain('7d F')
  })

  it('renders an OpenAI snapshot error without issuing a per-row request', () => {
    const wrapper = mountCell({
      account: makeAccount({ platform: 'openai', type: 'oauth' }),
      usageError: 'snapshot unavailable'
    })

    expect(wrapper.text()).toContain('snapshot unavailable')
    expect(getUsage).not.toHaveBeenCalled()
  })
})
