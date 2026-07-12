import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useTicketStore } from '@/stores/tickets'

const mocks = vi.hoisted(() => ({ userUnread: vi.fn(), adminUnread: vi.fn() }))

vi.mock('@/api/tickets', () => ({ getTicketUnreadCount: mocks.userUnread }))
vi.mock('@/api/admin/tickets', () => ({ unreadCount: mocks.adminUnread }))

describe('useTicketStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
    vi.clearAllMocks()
    Object.defineProperty(document, 'hidden', { configurable: true, value: false })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('loads the correct unread endpoint for users and administrators', async () => {
    mocks.userUnread.mockResolvedValue(3)
    mocks.adminUnread.mockResolvedValue(7)
    const store = useTicketStore()

    await store.refreshUnreadCount(false)
    expect(store.unreadCount).toBe(3)
    expect(mocks.userUnread).toHaveBeenCalledOnce()

    await store.refreshUnreadCount(true)
    expect(store.unreadCount).toBe(7)
    expect(mocks.adminUnread).toHaveBeenCalledOnce()
  })

  it('refreshes every 60 seconds only while the page is visible', async () => {
    mocks.userUnread.mockResolvedValue(2)
    const store = useTicketStore()
    store.startUnreadPolling(false)
    await vi.advanceTimersByTimeAsync(0)

    expect(mocks.userUnread).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(60_000)
    expect(mocks.userUnread).toHaveBeenCalledTimes(2)

    Object.defineProperty(document, 'hidden', { configurable: true, value: true })
    await vi.advanceTimersByTimeAsync(60_000)
    expect(mocks.userUnread).toHaveBeenCalledTimes(2)

    store.stopUnreadPolling()
  })
})
