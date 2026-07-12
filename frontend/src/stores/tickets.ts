/**
 * 工单未读状态及页面可见时的轻量轮询。
 */

import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getTicketUnreadCount } from '@/api/tickets'
import { unreadCount as getAdminTicketUnreadCount } from '@/api/admin/tickets'

const POLL_INTERVAL_MS = 60_000

export const useTicketStore = defineStore('tickets', () => {
  const unreadCount = ref(0)
  const loading = ref(false)
  let pollTimer: ReturnType<typeof setInterval> | null = null
  let adminViewer = false

  async function refreshUnreadCount(isAdmin = adminViewer): Promise<void> {
    if (loading.value) return
    adminViewer = isAdmin
    loading.value = true
    try {
      unreadCount.value = await (isAdmin ? getAdminTicketUnreadCount() : getTicketUnreadCount())
    } catch (error) {
      console.error('Failed to refresh ticket unread count:', error)
    } finally {
      loading.value = false
    }
  }

  function refreshWhenVisible(): void {
    if (!document.hidden) void refreshUnreadCount()
  }

  function startUnreadPolling(isAdmin: boolean): void {
    stopUnreadPolling()
    adminViewer = isAdmin
    void refreshUnreadCount(isAdmin)
    pollTimer = setInterval(refreshWhenVisible, POLL_INTERVAL_MS)
    window.addEventListener('focus', refreshWhenVisible)
    document.addEventListener('visibilitychange', refreshWhenVisible)
  }

  function stopUnreadPolling(): void {
    if (pollTimer) clearInterval(pollTimer)
    pollTimer = null
    window.removeEventListener('focus', refreshWhenVisible)
    document.removeEventListener('visibilitychange', refreshWhenVisible)
  }

  function reset(): void {
    stopUnreadPolling()
    unreadCount.value = 0
    loading.value = false
  }

  return { unreadCount, loading, refreshUnreadCount, startUnreadPolling, stopUnreadPolling, reset }
})
