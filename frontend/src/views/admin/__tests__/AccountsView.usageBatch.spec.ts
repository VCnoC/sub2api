import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import AccountsView from '../AccountsView.vue'
import type { Account, AccountUsageBatchResponse, AccountUsageInfo } from '@/types'

const {
  listAccounts,
  listWithEtag,
  getUsageBatch,
  getAllProxies,
  getAllGroups
} = vi.hoisted(() => ({
  listAccounts: vi.fn(),
  listWithEtag: vi.fn(),
  getUsageBatch: vi.fn(),
  getAllProxies: vi.fn(),
  getAllGroups: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      list: listAccounts,
      listWithEtag,
      getUsageBatch,
      delete: vi.fn(),
      batchClearError: vi.fn(),
      batchRefresh: vi.fn(),
      toggleSchedulable: vi.fn()
    },
    proxies: { getAll: getAllProxies },
    groups: { getAll: getAllGroups }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn(), showInfo: vi.fn() })
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ token: 'test-token' })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

const PaginationStub = {
  name: 'Pagination',
  props: ['page', 'total', 'pageSize'],
  emits: ['update:page', 'update:pageSize'],
  template: '<button data-test="next-page" @click="$emit(\'update:page\', 2)">next</button>'
}

const DataTableStub = {
  props: ['data'],
  template: `
    <div>
      <div v-for="row in data" :key="row.id" :data-test="'usage-row-' + row.id">
        <slot name="cell-usage" :row="row" />
      </div>
    </div>
  `
}

const AccountUsageCellStub = {
  props: ['account', 'usageInfo', 'usageLoading', 'usageError', 'todayStats', 'todayStatsLoading'],
  template: '<div class="usage-cell">{{ account.id }}|{{ usageInfo?.five_hour?.utilization ?? "none" }}|{{ usageError ?? "ok" }}</div>'
}

function makeAccount(id: number): Account {
  return {
    id,
    name: `account-${id}`,
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
    created_at: '2026-07-16T00:00:00Z',
    updated_at: '2026-07-16T00:00:00Z',
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null
  }
}

function makeUsage(utilization: number): AccountUsageInfo {
  return {
    source: 'passive',
    updated_at: '2026-07-16T08:00:00Z',
    five_hour: {
      utilization,
      resets_at: '2026-07-16T10:00:00Z',
      remaining_seconds: 3600
    },
    seven_day: null,
    seven_day_sonnet: null
  }
}

function makeBatch(ids: number[], utilization = 10): AccountUsageBatchResponse {
  return {
    usage: Object.fromEntries(ids.map(id => [String(id), makeUsage(utilization)])),
    today_stats: {},
    errors: {}
  }
}

function mountView() {
  return mount(AccountsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
        DataTable: DataTableStub,
        Pagination: PaginationStub,
        ConfirmDialog: true,
        AccountTableActions: {
          emits: ['refresh'],
          template: '<button data-test="manual-refresh" @click="$emit(\'refresh\')">refresh</button>'
        },
        AccountTableFilters: true,
        AccountBulkActionsBar: true,
        AccountActionMenu: true,
        ImportDataModal: true,
        ReAuthAccountModal: true,
        AccountTestModal: true,
        AccountStatsModal: true,
        ScheduledTestsPanel: true,
        SyncFromCrsModal: true,
        TempUnschedStatusModal: true,
        ErrorPassthroughRulesModal: true,
        TLSFingerprintProfilesModal: true,
        CreateAccountModal: true,
        EditAccountModal: true,
        BulkEditAccountModal: true,
        PlatformTypeBadge: true,
        AccountCapacityCell: true,
        AccountStatusIndicator: true,
        AccountTodayStatsCell: true,
        AccountGroupsCell: true,
        AccountUsageCell: AccountUsageCellStub,
        HelpTooltip: true,
        Icon: true
      }
    }
  })
}

describe('AccountsView usage batch snapshots', () => {
  beforeEach(() => {
    localStorage.clear()
    listAccounts.mockReset()
    listWithEtag.mockReset()
    getUsageBatch.mockReset()
    getAllProxies.mockReset()
    getAllGroups.mockReset()
    listWithEtag.mockResolvedValue({ notModified: true, etag: null, data: null })
    getAllProxies.mockResolvedValue([])
    getAllGroups.mockResolvedValue([])
  })

  it.each([50, 100])('loads %i visible accounts with one batch request', async (count) => {
    const rows = Array.from({ length: count }, (_, index) => makeAccount(index + 1))
    listAccounts.mockResolvedValue({
      items: rows,
      total: count,
      page: 1,
      page_size: count,
      pages: 1
    })
    getUsageBatch.mockResolvedValue(makeBatch(rows.map(row => row.id)))

    mountView()
    await vi.waitFor(() => expect(getUsageBatch).toHaveBeenCalledTimes(1))

    expect(getUsageBatch.mock.calls[0]?.[0]).toEqual(rows.map(row => row.id))
    expect(getUsageBatch.mock.calls[0]?.[1]?.signal).toBeInstanceOf(AbortSignal)
  })

  it('aborts the previous page request and ignores its late response', async () => {
    let resolveFirst!: (value: AccountUsageBatchResponse) => void
    const firstBatch = new Promise<AccountUsageBatchResponse>((resolve) => {
      resolveFirst = resolve
    })
    listAccounts.mockImplementation(async (page: number) => ({
      items: [makeAccount(page)],
      total: 2,
      page,
      page_size: 1,
      pages: 2
    }))
    getUsageBatch
      .mockReturnValueOnce(firstBatch)
      .mockResolvedValueOnce(makeBatch([2], 22))

    const wrapper = mountView()
    await vi.waitFor(() => expect(getUsageBatch).toHaveBeenCalledTimes(1))

    await wrapper.find('[data-test="next-page"]').trigger('click')
    await vi.waitFor(() => expect(getUsageBatch).toHaveBeenCalledTimes(2))

    const firstSignal = getUsageBatch.mock.calls[0]?.[1]?.signal as AbortSignal
    expect(firstSignal.aborted).toBe(true)
    expect(wrapper.find('[data-test="usage-row-2"]').text()).toContain('2|22|ok')

    resolveFirst(makeBatch([1], 99))
    await flushPromises()

    expect(wrapper.find('[data-test="usage-row-2"]').text()).toContain('2|22|ok')
    expect(wrapper.text()).not.toContain('99')
  })

  it('renders successful snapshots alongside per-account batch errors', async () => {
    listAccounts.mockResolvedValue({
      items: [makeAccount(1), makeAccount(2)],
      total: 2,
      page: 1,
      page_size: 2,
      pages: 1
    })
    getUsageBatch.mockResolvedValue({
      usage: { '1': makeUsage(35), '2': null },
      today_stats: {},
      errors: {
        '2': { code: 'snapshot_unavailable', message: 'snapshot unavailable' }
      }
    })

    const wrapper = mountView()
    await vi.waitFor(() => expect(getUsageBatch).toHaveBeenCalledTimes(1))
    await flushPromises()

    expect(wrapper.find('[data-test="usage-row-1"]').text()).toContain('1|35|ok')
    expect(wrapper.find('[data-test="usage-row-2"]').text()).toContain('2|none|snapshot unavailable')
  })

  it('manual refresh sends exactly one refreshed batch request', async () => {
    const rows = [makeAccount(1), makeAccount(2)]
    listAccounts.mockResolvedValue({
      items: rows,
      total: 2,
      page: 1,
      page_size: 2,
      pages: 1
    })
    getUsageBatch.mockResolvedValue(makeBatch([1, 2]))

    const wrapper = mountView()
    await vi.waitFor(() => expect(getUsageBatch).toHaveBeenCalledTimes(1))

    await wrapper.find('[data-test="manual-refresh"]').trigger('click')
    await vi.waitFor(() => expect(getUsageBatch).toHaveBeenCalledTimes(2))

    expect(listAccounts).toHaveBeenCalledTimes(2)
    expect(getUsageBatch.mock.calls[1]?.[0]).toEqual([1, 2])
    expect(getUsageBatch.mock.calls[1]?.[1]).toEqual(expect.objectContaining({ refresh: true }))
  })
})
