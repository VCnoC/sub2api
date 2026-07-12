/**
 * 登录用户的站内工单 API。
 */

import { apiClient } from './client'
import type {
  CreateTicketInput,
  ReplyTicketInput,
  Ticket,
  TicketListParams,
  TicketPage,
} from '@/types/ticket'

function ticketFormData(input: CreateTicketInput | ReplyTicketInput): FormData {
  const form = new FormData()
  if ('subject' in input) {
    form.append('subject', input.subject)
    form.append('category', input.category)
  }
  form.append('body', input.body)
  for (const file of input.files ?? []) form.append('files', file)
  return form
}

export async function listTickets(params: TicketListParams = {}): Promise<TicketPage> {
  const { data } = await apiClient.get<TicketPage>('/tickets', { params })
  return data
}

export async function createTicket(input: CreateTicketInput): Promise<Ticket> {
  const { data } = await apiClient.post<Ticket>('/tickets', ticketFormData(input))
  return data
}

export async function getTicket(id: number): Promise<Ticket> {
  const { data } = await apiClient.get<Ticket>(`/tickets/${id}`)
  return data
}

export async function replyTicket(id: number, input: ReplyTicketInput): Promise<Ticket> {
  const { data } = await apiClient.post<Ticket>(`/tickets/${id}/replies`, ticketFormData(input))
  return data
}

export async function getTicketUnreadCount(): Promise<number> {
  const { data } = await apiClient.get<{ count: number }>('/tickets/unread-count')
  return data.count
}

export async function getTicketAttachment(id: number): Promise<Blob> {
  const { data } = await apiClient.get<Blob>(`/ticket-attachments/${id}`, { responseType: 'blob' })
  return data
}

export const ticketsAPI = {
  list: listTickets,
  create: createTicket,
  get: getTicket,
  reply: replyTicket,
  unreadCount: getTicketUnreadCount,
  attachment: getTicketAttachment,
}

export default ticketsAPI
