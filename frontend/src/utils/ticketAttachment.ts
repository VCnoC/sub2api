/**
 * 通过鉴权 API 打开或下载工单附件。
 */

import { getTicketAttachment } from '@/api/tickets'
import type { TicketAttachment } from '@/types/ticket'

export async function openTicketAttachment(attachment: TicketAttachment): Promise<void> {
  const previewWindow = attachment.media_type.startsWith('image/') ? window.open('about:blank', '_blank') : null
  if (previewWindow) previewWindow.opener = null
  try {
    const blob = await getTicketAttachment(attachment.id)
    const url = URL.createObjectURL(blob)
    if (previewWindow) {
      previewWindow.location.replace(url)
    } else {
      const link = document.createElement('a')
      link.href = url
      link.download = attachment.original_name
      link.click()
    }
    setTimeout(() => URL.revokeObjectURL(url), 60_000)
  } catch (error) {
    previewWindow?.close()
    throw error
  }
}
