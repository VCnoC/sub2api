<!-- 管理员工单详情、协作控制和回复。 -->
<template>
  <AppLayout>
    <div class="mx-auto flex w-full max-w-6xl flex-col gap-4">
      <button type="button" class="btn btn-ghost btn-sm self-start" @click="router.push('/admin/tickets')">
        <Icon name="arrowLeft" size="sm" />
        <span>{{ t('tickets.backToQueue') }}</span>
      </button>

      <div v-if="loading" class="flex min-h-64 items-center justify-center"><LoadingSpinner /></div>
      <EmptyState v-else-if="!ticket" :title="t('tickets.notFound')" :description="t('tickets.notFoundDescription')" />

      <template v-else>
        <header class="border-b border-gray-200 pb-4 dark:border-dark-700">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="mb-1 font-mono text-xs text-gray-500 dark:text-dark-400">#{{ ticket.id }}</p>
              <h2 class="break-words text-xl font-semibold text-gray-900 dark:text-white">{{ ticket.subject }}</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                {{ ticket.user.username || ticket.user.email }}<span v-if="ticket.user.username"> · {{ ticket.user.email }}</span>
              </p>
            </div>
            <TicketStatusBadge :status="ticket.status" />
          </div>
        </header>

        <section class="grid gap-3 rounded-md border border-gray-200 bg-white/60 p-4 dark:border-dark-700 dark:bg-dark-900/60 md:grid-cols-3" aria-label="Ticket controls">
          <div>
            <label class="input-label mb-1.5 block">{{ t('tickets.priorityLabel') }}</label>
            <Select
              v-model="prioritySelection"
              :options="priorityOptions"
              :disabled="updating"
              @change="changePriority"
            />
          </div>
          <div>
            <label class="input-label mb-1.5 block">{{ t('tickets.assignee') }}</label>
            <Select
              v-model="assigneeSelection"
              :options="assigneeOptions"
              :disabled="updating"
              searchable
              @change="changeAssignee"
            />
          </div>
          <div class="flex flex-wrap items-end gap-2">
            <button
              v-if="ticket.assignee_id !== authStore.user?.id"
              type="button"
              class="btn btn-secondary"
              :disabled="updating"
              @click="claimTicket"
            >
              <Icon name="user" size="sm" />
              <span>{{ t('tickets.claim') }}</span>
            </button>
            <button
              type="button"
              class="btn"
              :class="ticket.status === 'closed' ? 'btn-secondary' : 'btn-danger'"
              :disabled="updating"
              @click="toggleClosed"
            >
              <Icon :name="ticket.status === 'closed' ? 'refresh' : 'xCircle'" size="sm" />
              <span>{{ ticket.status === 'closed' ? t('tickets.reopen') : t('tickets.closeTicket') }}</span>
            </button>
          </div>
        </section>

        <div class="flex flex-wrap gap-2 text-xs text-gray-500 dark:text-dark-400">
          <span class="badge badge-gray">{{ categoryLabel(ticket.category) }}</span>
          <span class="badge" :class="priorityClass(ticket.priority)">{{ priorityLabel(ticket.priority) }}</span>
          <span>{{ t('tickets.createdAt', { date: formatDate(ticket.created_at) }) }}</span>
        </div>

        <section aria-labelledby="admin-ticket-thread-title">
          <h3 id="admin-ticket-thread-title" class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{{ t('tickets.conversation') }}</h3>
          <TicketThread :messages="ticket.messages ?? []" @open-attachment="handleOpenAttachment">
            <template #attachment-actions="{ attachment }">
              <button
                type="button"
                class="btn-icon h-7 w-7 flex-shrink-0 text-gray-400 hover:text-red-600"
                :title="t('tickets.attachments.delete')"
                @click="openDeleteDialog(attachment)"
              >
                <Icon name="trash" size="sm" />
              </button>
            </template>
          </TicketThread>
        </section>

        <section class="border-t border-gray-200 pt-4 dark:border-dark-700" aria-labelledby="admin-ticket-reply-title">
          <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
            <h3 id="admin-ticket-reply-title" class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('tickets.reply') }}</h3>
            <div class="inline-flex rounded-md border border-gray-200 p-1 dark:border-dark-700" role="group" :aria-label="t('tickets.replyMode')">
              <button
                type="button"
                class="rounded px-3 py-1.5 text-xs font-medium"
                :class="replyMode === 'public' ? 'bg-primary-500 text-white' : 'text-gray-600 dark:text-gray-300'"
                @click="replyMode = 'public'"
              >
                {{ t('tickets.publicReply') }}
              </button>
              <button
                type="button"
                class="rounded px-3 py-1.5 text-xs font-medium"
                :class="replyMode === 'internal' ? 'bg-amber-500 text-white' : 'text-gray-600 dark:text-gray-300'"
                @click="replyMode = 'internal'"
              >
                {{ t('tickets.internalNote') }}
              </button>
            </div>
          </div>

          <form class="space-y-3" @submit.prevent="submitReply">
            <p v-if="replyMode === 'internal'" class="text-xs text-amber-700 dark:text-amber-400">{{ t('tickets.internalHint') }}</p>
            <TextArea
              v-model="replyForm.body"
              id="admin-ticket-reply"
              :placeholder="replyMode === 'internal' ? t('tickets.internalPlaceholder') : t('tickets.replyPlaceholder')"
              :rows="5"
              :disabled="sending"
            />
            <TicketAttachmentPicker v-if="replyMode === 'public'" v-model="replyForm.files" :disabled="sending" />
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

    <BaseDialog :show="!!deletingAttachment" :title="t('tickets.attachments.delete')" width="narrow" @close="closeDeleteDialog">
      <div class="space-y-3">
        <p class="text-sm text-gray-600 dark:text-gray-300">{{ t('tickets.attachments.deleteDescription') }}</p>
        <TextArea
          v-model="deleteReason"
          id="ticket-attachment-delete-reason"
          :label="t('tickets.attachments.deleteReason')"
          :placeholder="t('tickets.attachments.deleteReasonPlaceholder')"
          :rows="3"
          :disabled="deleting"
        />
      </div>
      <template #footer>
        <button type="button" class="btn btn-secondary" :disabled="deleting" @click="closeDeleteDialog">{{ t('common.cancel') }}</button>
        <button type="button" class="btn btn-danger" :disabled="deleting || !deleteReason.trim()" @click="deleteAttachment">
          {{ deleting ? t('common.processing') : t('common.delete') }}
        </button>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { adminTicketsAPI } from '@/api/admin/tickets'
import { usersAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useTicketStore } from '@/stores/tickets'
import { useTicketFormat } from '@/composables/useTicketFormat'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatDate } from '@/utils/format'
import { openTicketAttachment } from '@/utils/ticketAttachment'
import {
  TICKET_PRIORITIES,
  type Ticket,
  type TicketAttachment,
  type TicketPriority,
  type TicketUserSummary,
  type UpdateTicketInput,
} from '@/types/ticket'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Select from '@/components/common/Select.vue'
import TextArea from '@/components/common/TextArea.vue'
import Icon from '@/components/icons/Icon.vue'
import TicketStatusBadge from '@/components/tickets/TicketStatusBadge.vue'
import TicketThread from '@/components/tickets/TicketThread.vue'
import TicketAttachmentPicker from '@/components/tickets/TicketAttachmentPicker.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()
const ticketStore = useTicketStore()
const { categoryLabel, priorityLabel, priorityClass } = useTicketFormat()
const ticketId = Number(route.params.id)
const ticket = ref<Ticket | null>(null)
const admins = ref<TicketUserSummary[]>([])
const loading = ref(false)
const updating = ref(false)
const sending = ref(false)
const deleting = ref(false)
const prioritySelection = ref<TicketPriority>('normal')
const assigneeSelection = ref<number | null>(null)
const replyMode = ref<'public' | 'internal'>('public')
const replyForm = reactive<{ body: string; files: File[] }>({ body: '', files: [] })
const deletingAttachment = ref<TicketAttachment | null>(null)
const deleteReason = ref('')

const canReply = computed(() => replyForm.body.trim() !== '' || (replyMode.value === 'public' && replyForm.files.length > 0))
const priorityOptions = computed(() => TICKET_PRIORITIES.map((value) => ({ value, label: priorityLabel(value) })))
const assigneeOptions = computed(() => [
  { value: null, label: t('tickets.unassigned') },
  ...admins.value.map((admin) => ({ value: admin.id, label: admin.username || admin.email })),
])

function syncSelections(): void {
  if (!ticket.value) return
  prioritySelection.value = ticket.value.priority
  assigneeSelection.value = ticket.value.assignee_id ?? null
}

async function loadTicket(): Promise<void> {
  if (!Number.isInteger(ticketId) || ticketId <= 0) {
    await router.replace('/admin/tickets')
    return
  }
  loading.value = true
  try {
    ticket.value = await adminTicketsAPI.get(ticketId)
    syncSelections()
    await ticketStore.refreshUnreadCount(true)
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('tickets.errors.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function loadAdmins(): Promise<void> {
  try {
    const items: TicketUserSummary[] = []
    let page = 1
    while (true) {
      const result = await usersAPI.list(page, 100, { role: 'admin', status: 'active' })
      items.push(...result.items.map(({ id, email, username }) => ({ id, email, username })))
      if (page >= result.pages) break
      page++
    }
    admins.value = items
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('tickets.errors.adminsFailed')))
  }
}

async function applyUpdate(input: UpdateTicketInput, successMessage: string): Promise<void> {
  if (updating.value) return
  updating.value = true
  try {
    ticket.value = await adminTicketsAPI.update(ticketId, input)
    syncSelections()
    appStore.showSuccess(successMessage)
  } catch (error) {
    syncSelections()
    appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('tickets.errors.updateFailed')))
  } finally {
    updating.value = false
  }
}

function changePriority(value: string | number | boolean | null): void {
  if (typeof value !== 'string' || !TICKET_PRIORITIES.includes(value as TicketPriority)) return
  void applyUpdate({ priority: value as TicketPriority }, t('tickets.priorityUpdated'))
}

function changeAssignee(value: string | number | boolean | null): void {
  if (value !== null && typeof value !== 'number') return
  void applyUpdate({ assignee_id: value }, t('tickets.assigneeUpdated'))
}

function claimTicket(): void {
  const adminId = authStore.user?.id
  if (adminId) void applyUpdate({ assignee_id: adminId }, t('tickets.claimed'))
}

function toggleClosed(): void {
  if (!ticket.value) return
  const closing = ticket.value.status !== 'closed'
  void applyUpdate({ closed: closing }, closing ? t('tickets.closed') : t('tickets.reopened'))
}

async function submitReply(): Promise<void> {
  if (!canReply.value || sending.value) return
  sending.value = true
  try {
    ticket.value = await adminTicketsAPI.reply(
      ticketId,
      { body: replyForm.body.trim(), files: replyMode.value === 'public' ? replyForm.files : [] },
      replyMode.value === 'internal',
    )
    Object.assign(replyForm, { body: '', files: [] })
    await ticketStore.refreshUnreadCount(true)
    appStore.showSuccess(replyMode.value === 'internal' ? t('tickets.noteAdded') : t('tickets.replySent'))
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

function openDeleteDialog(attachment: TicketAttachment): void {
  deletingAttachment.value = attachment
  deleteReason.value = ''
}

function closeDeleteDialog(): void {
  if (!deleting.value) deletingAttachment.value = null
}

async function deleteAttachment(): Promise<void> {
  if (!deletingAttachment.value || !deleteReason.value.trim() || deleting.value) return
  deleting.value = true
  try {
    await adminTicketsAPI.deleteAttachment(deletingAttachment.value.id, deleteReason.value.trim())
    deletingAttachment.value = null
    deleteReason.value = ''
    await loadTicket()
    appStore.showSuccess(t('tickets.attachments.deletedSuccess'))
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('tickets.errors.deleteAttachmentFailed')))
  } finally {
    deleting.value = false
  }
}

onMounted(() => {
  void Promise.all([loadTicket(), loadAdmins()])
})
</script>
