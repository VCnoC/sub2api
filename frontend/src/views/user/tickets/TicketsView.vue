<!-- 用户工单列表及新建工单入口。 -->
<template>
  <AppLayout>
    <TablePageLayout>
      <template #actions>
        <div class="flex items-center justify-end">
          <button type="button" class="btn btn-primary" @click="openCreateDialog">
            <Icon name="plus" size="sm" />
            <span>{{ t('tickets.create') }}</span>
          </button>
        </div>
      </template>

      <template #filters>
        <div class="card p-4">
          <div class="grid gap-3 sm:grid-cols-2 lg:max-w-2xl">
            <Select v-model="filters.status" :options="statusOptions" @change="applyFilters" />
            <Select v-model="filters.category" :options="categoryOptions" @change="applyFilters" />
          </div>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="tickets"
          :loading="loading"
          :clickable-rows="true"
          row-key="id"
          @row-click="openTicket"
        >
          <template #cell-id="{ row }">
            <div class="flex items-center gap-2 font-mono text-xs">
              <span v-if="row.unread" class="h-2 w-2 flex-shrink-0 rounded-full bg-primary-500" :title="t('tickets.unread')"></span>
              <span>#{{ row.id }}</span>
            </div>
          </template>
          <template #cell-subject="{ row }">
            <span class="block max-w-md truncate font-medium text-gray-900 dark:text-white">{{ row.subject }}</span>
          </template>
          <template #cell-category="{ row }">
            <span class="badge badge-gray">{{ categoryLabel(row.category) }}</span>
          </template>
          <template #cell-status="{ row }"><TicketStatusBadge :status="row.status" /></template>
          <template #cell-updated_at="{ row }">{{ formatDate(row.last_message_at) }}</template>
          <template #cell-actions="{ row }">
            <button type="button" class="btn btn-ghost btn-sm" @click.stop="openTicket(row)">
              <Icon name="eye" size="sm" />
              <span>{{ t('common.view') }}</span>
            </button>
          </template>
          <template #empty>
            <EmptyState
              :title="t('tickets.empty')"
              :description="t('tickets.emptyDescription')"
              :action-text="t('tickets.create')"
              @action="openCreateDialog"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :page-size="pagination.page_size"
          :total="pagination.total"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <BaseDialog :show="showCreateDialog" :title="t('tickets.create')" width="wide" @close="closeCreateDialog">
      <form id="ticket-create-form" class="space-y-4" @submit.prevent="submitTicket">
        <Input
          v-model="createForm.subject"
          id="ticket-subject"
          :label="t('tickets.subject')"
          :placeholder="t('tickets.subjectPlaceholder')"
          :required="true"
        />
        <div>
          <label class="input-label mb-1.5 block">{{ t('tickets.categoryLabel') }}</label>
          <Select v-model="createForm.category" :options="createCategoryOptions" />
        </div>
        <TextArea
          v-model="createForm.body"
          id="ticket-body"
          :label="t('tickets.message')"
          :placeholder="t('tickets.messagePlaceholder')"
          :required="true"
          :rows="7"
        />
        <div>
          <label class="input-label mb-1.5 block">{{ t('tickets.attachments.label') }}</label>
          <TicketAttachmentPicker v-model="createForm.files" :disabled="creating" />
        </div>
      </form>
      <template #footer>
        <button type="button" class="btn btn-secondary" :disabled="creating" @click="closeCreateDialog">
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          form="ticket-create-form"
          class="btn btn-primary"
          :disabled="creating || !canCreate"
        >
          {{ creating ? t('common.submitting') : t('tickets.submit') }}
        </button>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ticketsAPI } from '@/api/tickets'
import { useAppStore } from '@/stores/app'
import { useTicketStore } from '@/stores/tickets'
import { useTicketFormat } from '@/composables/useTicketFormat'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatDate } from '@/utils/format'
import {
  TICKET_CATEGORIES,
  TICKET_STATUSES,
  type Ticket,
  type TicketCategory,
  type TicketStatus,
} from '@/types/ticket'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import Input from '@/components/common/Input.vue'
import TextArea from '@/components/common/TextArea.vue'
import Icon from '@/components/icons/Icon.vue'
import TicketStatusBadge from '@/components/tickets/TicketStatusBadge.vue'
import TicketAttachmentPicker from '@/components/tickets/TicketAttachmentPicker.vue'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const ticketStore = useTicketStore()
const { categoryLabel } = useTicketFormat()

const tickets = ref<Ticket[]>([])
const loading = ref(false)
const creating = ref(false)
const showCreateDialog = ref(false)
const filters = reactive<{ status: TicketStatus | ''; category: TicketCategory | '' }>({ status: '', category: '' })
const pagination = reactive({ page: 1, page_size: getPersistedPageSize(), total: 0 })
const createForm = reactive<{ subject: string; category: TicketCategory; body: string; files: File[] }>({
  subject: '',
  category: 'other',
  body: '',
  files: [],
})

const columns = computed<Column[]>(() => [
  { key: 'id', label: t('tickets.number') },
  { key: 'subject', label: t('tickets.subject') },
  { key: 'category', label: t('tickets.categoryLabel') },
  { key: 'status', label: t('common.status') },
  { key: 'updated_at', label: t('tickets.lastUpdated') },
  { key: 'actions', label: t('common.actions') },
])
const statusOptions = computed(() => [
  { value: '', label: t('tickets.filters.allStatuses') },
  ...TICKET_STATUSES.map((value) => ({ value, label: t(`tickets.status.${value}`) })),
])
const categoryOptions = computed(() => [
  { value: '', label: t('tickets.filters.allCategories') },
  ...TICKET_CATEGORIES.map((value) => ({ value, label: categoryLabel(value) })),
])
const createCategoryOptions = computed(() => TICKET_CATEGORIES.map((value) => ({ value, label: categoryLabel(value) })))
const canCreate = computed(() => createForm.subject.trim() !== '' && createForm.body.trim() !== '')

async function loadTickets(): Promise<void> {
  loading.value = true
  try {
    const result = await ticketsAPI.list({ ...filters, page: pagination.page, page_size: pagination.page_size })
    tickets.value = result.items ?? []
    pagination.total = result.total
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('tickets.errors.loadFailed')))
  } finally {
    loading.value = false
  }
}

function applyFilters(): void {
  pagination.page = 1
  void loadTickets()
}

function handlePageChange(page: number): void {
  pagination.page = page
  void loadTickets()
}

function handlePageSizeChange(pageSize: number): void {
  pagination.page = 1
  pagination.page_size = pageSize
  void loadTickets()
}

function openTicket(ticket: Ticket): void {
  void router.push(`/tickets/${ticket.id}`)
}

function openCreateDialog(): void {
  showCreateDialog.value = true
}

function closeCreateDialog(): void {
  if (!creating.value) showCreateDialog.value = false
}

function resetCreateForm(): void {
  Object.assign(createForm, { subject: '', category: 'other', body: '', files: [] })
}

async function submitTicket(): Promise<void> {
  if (!canCreate.value || creating.value) return
  creating.value = true
  try {
    const ticket = await ticketsAPI.create({
      subject: createForm.subject.trim(),
      category: createForm.category,
      body: createForm.body.trim(),
      files: createForm.files,
    })
    resetCreateForm()
    showCreateDialog.value = false
    await ticketStore.refreshUnreadCount(false)
    appStore.showSuccess(t('tickets.created'))
    await router.push(`/tickets/${ticket.id}`)
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('tickets.errors.createFailed')))
  } finally {
    creating.value = false
  }
}

onMounted(loadTickets)
</script>
