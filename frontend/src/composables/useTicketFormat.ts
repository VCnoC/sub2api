/**
 * 工单枚举的本地化标签和稳定视觉样式。
 */

import { useI18n } from 'vue-i18n'
import type { TicketCategory, TicketPriority, TicketStatus } from '@/types/ticket'

export function useTicketFormat() {
  const { t } = useI18n()

  const statusLabel = (value: TicketStatus) => t(`tickets.status.${value}`)
  const categoryLabel = (value: TicketCategory) => t(`tickets.category.${value}`)
  const priorityLabel = (value: TicketPriority) => t(`tickets.priority.${value}`)

  const statusClass = (value: TicketStatus) => ({
    pending_admin: 'badge-warning',
    pending_user: 'badge-primary',
    closed: 'badge-gray',
  })[value]

  const priorityClass = (value: TicketPriority) => ({
    normal: 'badge-gray',
    high: 'badge-warning',
    urgent: 'badge-danger',
  })[value]

  return { statusLabel, categoryLabel, priorityLabel, statusClass, priorityClass }
}
