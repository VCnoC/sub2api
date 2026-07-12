<!-- 用户工单详情和公开回复。 -->
<template>
  <AppLayout>
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-4">
      <button type="button" class="btn btn-ghost btn-sm self-start" @click="router.push('/tickets')">
        <Icon name="arrowLeft" size="sm" />
        <span>{{ t('tickets.backToList') }}</span>
      </button>

      <div v-if="loading" class="flex min-h-64 items-center justify-center">
        <LoadingSpinner />
      </div>

      <EmptyState v-else-if="!ticket" :title="t('tickets.notFound')" :description="t('tickets.notFoundDescription')" />

      <template v-else>
        <header class="border-b border-gray-200 pb-4 dark:border-dark-700">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="mb-1 font-mono text-xs text-gray-500 dark:text-dark-400">#{{ ticket.id }}</p>
              <h2 class="break-words text-xl font-semibold text-gray-900 dark:text-white">{{ ticket.subject }}</h2>
            </div>
            <TicketStatusBadge :status="ticket.status" />
          </div>
          <div class="mt-3 flex flex-wrap gap-2 text-xs text-gray-500 dark:text-dark-400">
            <span class="badge badge-gray">{{ categoryLabel(ticket.category) }}</span>
            <span>{{ t('tickets.createdAt', { date: formatDate(ticket.created_at) }) }}</span>
            <span v-if="ticket.assignee">{{ t('tickets.assigneeValue', { name: ticket.assignee.username || ticket.assignee.email }) }}</span>
          </div>
        </header>

        <section aria-labelledby="ticket-thread-title">
          <h3 id="ticket-thread-title" class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{{ t('tickets.conversation') }}</h3>
          <TicketThread
            :messages="ticket.messages ?? []"
            @open-attachment="handleOpenAttachment"
          />
        </section>

        <section class="border-t border-gray-200 pt-4 dark:border-dark-700" aria-labelledby="ticket-reply-title">
          <div class="mb-3">
            <h3 id="ticket-reply-title" class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('tickets.reply') }}</h3>
            <p v-if="ticket.status === 'closed'" class="mt-1 text-xs text-amber-700 dark:text-amber-400">
              {{ t('tickets.closedReplyHint') }}
            </p>
          </div>
          <form class="space-y-3" @submit.prevent="submitReply">
            <TextArea
              v-model="replyForm.body"
              id="ticket-reply"
              :placeholder="t('tickets.replyPlaceholder')"
              :rows="5"
              :disabled="sending"
            />
            <TicketAttachmentPicker v-model="replyForm.files" :disabled="sending" />
            <div class="flex justify-end">
              <button type="submit" class="btn btn-primary" :disabled="sending || !canReply">
                <Icon name="chat" size="sm" />
                <span>{{ sending ? t('common.submitting') : t('tickets.sendReply') }}</span>
              </button>
            </div>
          </form>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ticketsAPI } from '@/api/tickets'
import { useAppStore } from '@/stores/app'
import { useTicketStore } from '@/stores/tickets'
import { useTicketFormat } from '@/composables/useTicketFormat'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatDate } from '@/utils/format'
import { openTicketAttachment } from '@/utils/ticketAttachment'
import type { Ticket, TicketAttachment } from '@/types/ticket'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import TextArea from '@/components/common/TextArea.vue'
import Icon from '@/components/icons/Icon.vue'
import TicketStatusBadge from '@/components/tickets/TicketStatusBadge.vue'
import TicketThread from '@/components/tickets/TicketThread.vue'
import TicketAttachmentPicker from '@/components/tickets/TicketAttachmentPicker.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const ticketStore = useTicketStore()
const { categoryLabel } = useTicketFormat()
const ticket = ref<Ticket | null>(null)
const loading = ref(false)
const sending = ref(false)
const replyForm = reactive<{ body: string; files: File[] }>({ body: '', files: [] })
const ticketId = Number(route.params.id)
const canReply = computed(() => replyForm.body.trim() !== '' || replyForm.files.length > 0)

async function loadTicket(): Promise<void> {
  if (!Number.isInteger(ticketId) || ticketId <= 0) {
    await router.replace('/tickets')
    return
  }
  loading.value = true
  try {
    ticket.value = await ticketsAPI.get(ticketId)
    await ticketStore.refreshUnreadCount(false)
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('tickets.errors.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function submitReply(): Promise<void> {
  if (!canReply.value || sending.value) return
  sending.value = true
  try {
    ticket.value = await ticketsAPI.reply(ticketId, {
      body: replyForm.body.trim(),
      files: replyForm.files,
    })
    Object.assign(replyForm, { body: '', files: [] })
    await ticketStore.refreshUnreadCount(false)
    appStore.showSuccess(t('tickets.replySent'))
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('tickets.errors.replyFailed')))
  } finally {
    sending.value = false
  }
}

async function handleOpenAttachment(attachment: TicketAttachment): Promise<void> {
  try {
    await openTicketAttachment(attachment)
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('tickets.errors.attachmentFailed')))
  }
}

onMounted(loadTicket)
</script>
