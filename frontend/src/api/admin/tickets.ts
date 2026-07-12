/**
 * 管理员站内工单 API。
 */

import { apiClient } from '../client'
import type {
  AdminTicketListParams,
  ReplyTicketInput,
  Ticket,
  TicketPage,
  UpdateTicketInput,
} from '@/types/ticket'

function replyFormData(input: ReplyTicketInput, internal: boolean): FormData {
  const form = new FormData()
  form.append('body', input.body)
  form.append('internal', String(internal))
  for (const file of input.files ?? []) form.append('files', file)
  return form
}

export async function list(params: AdminTicketListParams = {}): Promise<TicketPage> {
  const { data } = await apiClient.get<TicketPage>('/admin/tickets', { params })
  return data
}

export async function get(id: number): Promise<Ticket> {
  const { data } = await apiClient.get<Ticket>(`/admin/tickets/${id}`)
  return data
}

export async function reply(
  id: number,
  input: ReplyTicketInput,
  internal = false,
): Promise<Ticket> {
  const { data } = await apiClient.post<Ticket>(
    `/admin/tickets/${id}/replies`,
    replyFormData(input, internal),
  )
  return data
}

export async function update(id: number, input: UpdateTicketInput): Promise<Ticket> {
  const { data } = await apiClient.patch<Ticket>(`/admin/tickets/${id}`, input)
  return data
}

export async function unreadCount(): Promise<number> {
  const { data } = await apiClient.get<{ count: number }>('/admin/tickets/unread-count')
  return data.count
}

export async function deleteAttachment(id: number, reason: string): Promise<void> {
  await apiClient.delete(`/admin/ticket-attachments/${id}`, { data: { reason } })
}

export const adminTicketsAPI = { list, get, reply, update, unreadCount, deleteAttachment }

export default adminTicketsAPI
