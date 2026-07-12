/**
 * 站内工单的前端数据契约。
 */

import type { PaginatedResponse } from '@/types'

export type TicketStatus = 'pending_admin' | 'pending_user' | 'closed'
export type TicketCategory = 'account' | 'billing' | 'api' | 'model' | 'other'
export type TicketPriority = 'normal' | 'high' | 'urgent'
export type TicketMessageKind = 'public' | 'internal' | 'system'
export type TicketVisibility = 'user' | 'admin'

export const TICKET_STATUSES: TicketStatus[] = ['pending_admin', 'pending_user', 'closed']
export const TICKET_CATEGORIES: TicketCategory[] = ['account', 'billing', 'api', 'model', 'other']
export const TICKET_PRIORITIES: TicketPriority[] = ['normal', 'high', 'urgent']

export interface TicketUserSummary {
  id: number
  email: string
  username: string
}

export interface TicketAttachment {
  id: number
  message_id: number
  uploader_id: number
  original_name: string
  media_type: string
  size_bytes: number
  delete_after?: string
  deleted_at?: string
  deleted_by?: number
  delete_reason?: string
  created_at: string
}

export interface TicketMessage {
  id: number
  ticket_id: number
  author_id?: number
  author?: TicketUserSummary
  author_role?: 'admin' | 'user'
  kind: TicketMessageKind
  visibility: TicketVisibility
  body: string
  metadata?: Record<string, unknown>
  attachments: TicketAttachment[]
  created_at: string
}

export interface Ticket {
  id: number
  user_id: number
  user: TicketUserSummary
  subject: string
  category: TicketCategory
  status: TicketStatus
  priority: TicketPriority
  assignee_id?: number
  assignee?: TicketUserSummary
  closed_by?: number
  closed_at?: string
  last_message_at: string
  created_at: string
  updated_at: string
  unread: boolean
  messages?: TicketMessage[]
}

export interface TicketListParams {
  page?: number
  page_size?: number
  status?: TicketStatus | ''
  category?: TicketCategory | ''
}

export interface AdminTicketListParams extends TicketListParams {
  priority?: TicketPriority | ''
  assignee?: 'mine' | 'unassigned' | `${number}` | ''
  search?: string
}

export interface CreateTicketInput {
  subject: string
  category: TicketCategory
  body: string
  files?: File[]
}

export interface ReplyTicketInput {
  body: string
  files?: File[]
}

export interface UpdateTicketInput {
  priority?: TicketPriority
  assignee_id?: number | null
  closed?: boolean
}

export type TicketPage = PaginatedResponse<Ticket>
