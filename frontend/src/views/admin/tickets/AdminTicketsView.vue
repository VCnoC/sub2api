<!-- 管理员工单队列、筛选和分页。 -->
<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="card p-4">
          <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-5">
            <SearchInput
              v-model="filters.search"
              :placeholder="t('tickets.filters.searchPlaceholder')"
              @search="applyFilters"
            />
            <Select v-model="filters.status" :options="statusOptions" @change="applyFilters" />
            <Select v-model="filters.category" :options="categoryOptions" @change="applyFilters" />
            <Select v-model="filters.priority" :options="priorityOptions" @change="applyFilters" />
            <Select v-model="filters.assignee" :options="assigneeOptions" @change="applyFilters" />
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
          <template #cell-user="{ row }">
            <div class="max-w-52">
              <p class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ row.user.username || row.user.email }}</p>
              <p v-if="row.user.username" class="truncate text-xs text-gray-500 dark:text-dark-400">{{ row.user.email }}</p>
            </div>
          </template>
          <template #cell-subject="{ row }">
            <span class="block max-w-sm truncate font-medium text-gray-900 dark:text-white">{{ row.subject }}</span>
          </template>
          <template #cell-category="{ row }"><span class="badge badge-gray">{{ categoryLabel(row.category) }}</span></template>
          <template #cell-status="{ row }"><TicketStatusBadge :status="row.status" /></template>
          <template #cell-priority="{ row }">
            <span class="badge" :class="priorityClass(row.priority)">{{ priorityLabel(row.priority) }}</span>
          </template>
          <template #cell-assignee="{ row }">{{ row.assignee?.username || row.assignee?.email || t('tickets.unassigned') }}</template>
          <template #cell-last_message_at="{ row }">{{ formatDate(row.last_message_at) }}</template>
          <template #cell-actions="{ row }">
            <button type="button" class="btn btn-ghost btn-sm" @click.stop="openTicket(row)">
              <Icon name="eye" size="sm" />
              <span>{{ t('common.view') }}</span>
            </button>
          </template>
          <template #empty>
            <EmptyState :title="t('tickets.emptyAdmin')" :description="t('tickets.emptyAdminDescription')" />
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
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { adminTicketsAPI } from '@/api/admin/tickets'
import { useAppStore } from '@/stores/app'
import { useTicketFormat } from '@/composables/useTicketFormat'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatDate } from '@/utils/format'
import {
  TICKET_CATEGORIES,
  TICKET_PRIORITIES,
  TICKET_STATUSES,
  type AdminTicketListParams,
  type Ticket,
} from '@/types/ticket'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import TicketStatusBadge from '@/components/tickets/TicketStatusBadge.vue'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const { categoryLabel, priorityLabel, priorityClass } = useTicketFormat()
const tickets = ref<Ticket[]>([])
const loading = ref(false)
const filters = reactive<Required<Pick<AdminTicketListParams, 'status' | 'category' | 'priority' | 'assignee' | 'search'>>>({
  status: '',
  category: '',
  priority: '',
  assignee: '',
  search: '',
})
const pagination = reactive({ page: 1, page_size: getPersistedPageSize(), total: 0 })

const columns = computed<Column[]>(() => [
  { key: 'id', label: t('tickets.number') },
  { key: 'user', label: t('tickets.requester') },
  { key: 'subject', label: t('tickets.subject') },
  { key: 'category', label: t('tickets.categoryLabel') },
  { key: 'status', label: t('common.status') },
  { key: 'priority', label: t('tickets.priorityLabel') },
  { key: 'assignee', label: t('tickets.assignee') },
  { key: 'last_message_at', label: t('tickets.lastUpdated') },
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
const priorityOptions = computed(() => [
  { value: '', label: t('tickets.filters.allPriorities') },
  ...TICKET_PRIORITIES.map((value) => ({ value, label: priorityLabel(value) })),
])
const assigneeOptions = computed(() => [
  { value: '', label: t('tickets.filters.allAssignees') },
  { value: 'mine', label: t('tickets.filters.mine') },
  { value: 'unassigned', label: t('tickets.unassigned') },
])

async function loadTickets(): Promise<void> {
  loading.value = true
  try {
    const result = await adminTicketsAPI.list({ ...filters, page: pagination.page, page_size: pagination.page_size })
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
  void router.push(`/admin/tickets/${ticket.id}`)
}

onMounted(loadTickets)
</script>
