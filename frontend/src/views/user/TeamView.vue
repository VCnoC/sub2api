<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('team.title') }}</h1>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <!-- Not in any team -->
      <template v-else-if="!team">
        <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
          <div class="card p-6">
            <div class="flex items-center gap-3">
              <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-blue-100 text-2xl dark:bg-blue-900/30">🏗️</div>
              <div>
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('team.create.title') }}</h2>
                <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('team.create.description') }}</p>
              </div>
            </div>
            <div class="mt-4 space-y-3">
              <input
                v-model="createName"
                type="text"
                class="input w-full"
                :placeholder="t('team.create.namePlaceholder')"
                maxlength="100"
              />
              <button class="btn btn-primary w-full" :disabled="creating" @click="handleCreate">
                <Icon v-if="creating" name="refresh" size="sm" class="animate-spin" />
                <Icon v-else name="plus" size="sm" />
                <span>{{ creating ? t('team.create.creating') : t('team.create.button') }}</span>
              </button>
            </div>
          </div>

          <div class="card p-6">
            <div class="flex items-center gap-3">
              <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-emerald-100 text-2xl dark:bg-emerald-900/30">🚪</div>
              <div>
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('team.join.title') }}</h2>
                <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('team.join.description') }}</p>
              </div>
            </div>
            <div class="mt-4 space-y-3">
              <input
                v-model="joinCode"
                type="text"
                class="input w-full"
                :placeholder="t('team.join.codePlaceholder')"
              />
              <button class="btn btn-primary w-full" :disabled="joining" @click="handleJoin">
                <Icon v-if="joining" name="refresh" size="sm" class="animate-spin" />
                <Icon v-else name="arrowRight" size="sm" />
                <span>{{ joining ? t('team.join.joining') : t('team.join.button') }}</span>
              </button>
            </div>
          </div>
        </div>
      </template>

      <!-- In a team -->
      <template v-else>
        <!-- Team info card -->
        <div class="card p-6">
          <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div>
              <div class="flex items-center gap-2">
                <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ team.name }}</h2>
                <span
                  class="badge"
                  :class="team.role === 'owner' ? 'badge-primary' : 'badge-gray'"
                >
                  {{ team.role === 'owner' ? t('team.role.owner') : t('team.role.member') }}
                </span>
              </div>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                {{ t('team.members.count', { count: filteredMembers.length }) }}
              </p>
            </div>

            <div class="flex items-center gap-3">
              <template v-if="isOwner">
                <div class="flex items-center gap-2 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                  <code class="text-sm font-mono font-semibold text-gray-900 dark:text-white">{{ team.invite_code }}</code>
                  <button class="btn btn-secondary btn-sm" @click="copyCode">
                    <Icon name="copy" size="sm" />
                  </button>
                </div>
                <button class="btn btn-secondary" :disabled="refreshing" @click="handleRefreshCode">
                  <Icon v-if="refreshing" name="refresh" size="sm" class="animate-spin" />
                  <Icon v-else name="refresh" size="sm" />
                  <span>{{ t('team.inviteCode.refresh') }}</span>
                </button>
              </template>
              <button
                v-else
                class="btn btn-danger-outline"
                :disabled="leaving"
                @click="handleLeave"
              >
                <Icon v-if="leaving" name="refresh" size="sm" class="animate-spin" />
                <span>{{ leaving ? t('team.leave.leaving') : t('team.leave.button') }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Member list -->
        <div class="card p-6">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex items-center gap-2">
              <span class="h-4 w-1 rounded-full bg-gradient-to-b from-primary-400 to-primary-600"></span>
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('team.members.title') }}</h3>
            </div>
            <div class="relative max-w-xs">
              <Icon name="search" size="sm" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
              <input
                v-model="memberSearch"
                type="text"
                class="input w-full pl-9"
                :placeholder="t('team.members.searchPlaceholder')"
              />
            </div>
          </div>

          <div v-if="membersLoading" class="flex justify-center py-8">
            <div class="h-6 w-6 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
          </div>

          <div v-else-if="filteredMembers.length === 0" class="mt-4 rounded-xl border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
            {{ t('team.members.empty') }}
          </div>

          <div v-else class="mt-4 overflow-x-auto">
            <table class="w-full min-w-[640px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ t('team.members.columns.user') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ t('team.members.columns.balance') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ t('team.members.columns.usage') }}</th>
                  <th v-if="isOwner" class="px-3 py-2 font-medium text-right">{{ t('team.members.columns.actions') }}</th>
                </tr>
              </thead>
              <tbody>
                <template v-for="member in filteredMembers" :key="member.id">
                  <tr
                    class="cursor-pointer border-b border-gray-100 transition-colors hover:bg-gray-50 dark:border-dark-800 dark:hover:bg-dark-900/50"
                    :class="expandedMemberId === member.id ? 'bg-gray-50 dark:bg-dark-900/50' : ''"
                    @click="toggleMember(member)"
                  >
                    <td class="px-3 py-3">
                      <div class="flex items-center gap-2">
                        <Icon
                          name="chevronRight"
                          size="sm"
                          class="text-gray-400 transition-transform"
                          :class="expandedMemberId === member.id ? 'rotate-90' : ''"
                        />
                        <div>
                          <div class="text-base font-semibold text-gray-900 dark:text-white">{{ member.email }}</div>
                          <span class="badge badge-gray mt-1">{{ member.role === 'owner' ? t('team.role.owner') : t('team.role.member') }}</span>
                        </div>
                      </div>
                    </td>
                    <td class="px-3 py-3 text-right font-mono tabular-nums text-gray-900 dark:text-white">{{ formatCurrency(member.balance) }}</td>
                    <td class="px-3 py-3 text-right font-mono tabular-nums text-gray-700 dark:text-gray-300">{{ formatCurrency(member.total_usage) }}</td>
                    <td v-if="isOwner" class="px-3 py-3 text-right">
                      <div v-if="member.id !== currentUserId" class="flex justify-end gap-2" @click.stop>
                        <button class="btn btn-primary btn-sm" @click="openTransfer(member)">
                          <Icon name="creditCard" size="sm" />
                          <span>{{ t('team.transfer.button') }}</span>
                        </button>
                        <button class="btn btn-danger-outline btn-sm" :disabled="removingId === member.id" @click="handleRemove(member)">
                          <Icon v-if="removingId === member.id" name="refresh" size="sm" class="animate-spin" />
                          <Icon v-else name="trash" size="sm" />
                          <span>{{ t('team.members.remove') }}</span>
                        </button>
                      </div>
                    </td>
                  </tr>

                  <!-- Expanded usage panel -->
                  <tr v-if="expandedMemberId === member.id" class="border-b border-gray-100 dark:border-dark-800">
                    <td :colspan="isOwner ? 4 : 3" class="bg-gray-50/50 px-3 py-4 dark:bg-dark-900/30">
                      <div v-if="!canViewUsage(member)" class="rounded-xl border border-dashed border-gray-300 p-4 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
                        {{ t('team.members.usage.noPermission') }}
                      </div>
                      <div v-else class="space-y-4">
                        <div class="flex flex-col gap-3 lg:flex-row lg:items-end">
                          <div class="flex flex-wrap gap-2">
                            <button
                              v-for="preset in datePresets"
                              :key="preset.value"
                              class="btn btn-sm"
                              :class="activePreset === preset.value ? 'btn-primary' : 'btn-secondary'"
                              @click="applyDatePreset(preset.value)"
                            >
                              {{ t(preset.labelKey) }}
                            </button>
                          </div>
                          <div class="flex items-end gap-2">
                            <div>
                              <label class="input-label">{{ t('team.members.usage.startDate') }}</label>
                              <input v-model="usageStartDate" type="date" class="input" />
                            </div>
                            <div>
                              <label class="input-label">{{ t('team.members.usage.endDate') }}</label>
                              <input v-model="usageEndDate" type="date" class="input" />
                            </div>
                            <button
                              class="btn btn-secondary"
                              :disabled="usageLoading[member.id]"
                              @click="loadMemberUsage(member)"
                            >
                              <Icon v-if="usageLoading[member.id]" name="refresh" size="sm" class="animate-spin" />
                              <Icon v-else name="search" size="sm" />
                              <span>{{ t('team.members.usage.query') }}</span>
                            </button>
                          </div>
                        </div>

                        <div v-if="usageLoading[member.id]" class="flex justify-center py-6">
                          <div class="h-6 w-6 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
                        </div>

                        <div v-else-if="memberUsage[member.id]?.length === 0" class="rounded-xl border border-dashed border-gray-300 p-4 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
                          {{ t('team.members.usage.empty') }}
                        </div>

                        <div v-else-if="memberUsage[member.id]?.length > 0" class="overflow-x-auto rounded-xl border border-gray-200 dark:border-dark-700">
                          <table class="w-full min-w-[900px] text-left text-xs">
                            <thead>
                              <tr class="border-b border-gray-200 bg-gray-100 dark:border-dark-700 dark:bg-dark-800">
                                <th class="px-3 py-2 font-medium text-gray-700 dark:text-gray-300">{{ t('team.members.usage.time') }}</th>
                                <th class="px-3 py-2 font-medium text-gray-700 dark:text-gray-300">{{ t('team.members.usage.model') }}</th>
                                <th class="px-3 py-2 font-medium text-gray-700 dark:text-gray-300">{{ t('team.members.usage.type') }}</th>
                                <th class="px-3 py-2 font-medium text-gray-700 dark:text-gray-300">{{ t('team.members.usage.tokens') }}</th>
                                <th class="px-3 py-2 font-medium text-gray-700 dark:text-gray-300">{{ t('team.members.usage.cost') }}</th>
                                <th class="px-3 py-2 font-medium text-gray-700 dark:text-gray-300">{{ t('usage.firstToken') }}</th>
                                <th class="px-3 py-2 font-medium text-gray-700 dark:text-gray-300">{{ t('team.members.usage.duration') }}</th>
                              </tr>
                            </thead>
                            <tbody>
                              <tr
                                v-for="log in memberUsage[member.id]"
                                :key="log.id"
                                class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
                              >
                                <td class="px-3 py-2 text-gray-600 dark:text-gray-400">{{ formatDateTime(log.created_at) }}</td>
                                <td class="px-3 py-2 font-medium text-gray-900 dark:text-white">{{ log.model }}</td>
                                <td class="px-3 py-2">
                                  <span
                                    class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium"
                                    :class="getRequestTypeBadgeClass(log)"
                                  >
                                    {{ getRequestTypeLabel(log) }}
                                  </span>
                                </td>
                                <td class="px-3 py-2">
                                  <div class="space-y-1">
                                    <div class="flex items-center gap-2">
                                      <span class="inline-flex items-center gap-1 text-emerald-600 dark:text-emerald-400">
                                        <Icon name="arrowDown" size="xs" />
                                        <span class="tabular-nums">{{ log.input_tokens.toLocaleString() }}</span>
                                      </span>
                                      <span class="inline-flex items-center gap-1 text-violet-600 dark:text-violet-400">
                                        <Icon name="arrowUp" size="xs" />
                                        <span class="tabular-nums">{{ log.output_tokens.toLocaleString() }}</span>
                                      </span>
                                    </div>
                                    <div v-if="log.cache_read_tokens > 0 || log.cache_creation_tokens > 0" class="flex items-center gap-2">
                                      <span v-if="log.cache_read_tokens > 0" class="inline-flex items-center gap-1 text-sky-600 dark:text-sky-400">
                                        <Icon name="inbox" size="xs" />
                                        <span class="tabular-nums">{{ formatCacheTokens(log.cache_read_tokens) }}</span>
                                      </span>
                                      <span v-if="log.cache_creation_tokens > 0" class="inline-flex items-center gap-1 text-amber-600 dark:text-amber-400">
                                        <Icon name="edit" size="xs" />
                                        <span class="tabular-nums">{{ formatCacheTokens(log.cache_creation_tokens) }}</span>
                                      </span>
                                    </div>
                                    <div v-if="log.image_count > 0" class="flex items-center gap-1 text-pink-600 dark:text-pink-400">
                                      <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
                                      <span class="tabular-nums">{{ log.image_count }}{{ t('usage.imageUnit') }}</span>
                                    </div>
                                  </div>
                                </td>
                                <td class="px-3 py-2 tabular-nums font-medium text-emerald-600 dark:text-emerald-400">
                                  ${{ (log.actual_cost ?? 0).toFixed(6) }}
                                </td>
                                <td class="px-3 py-2 text-gray-600 dark:text-gray-400">{{ log.first_token_ms != null ? formatDuration(log.first_token_ms) : '-' }}</td>
                                <td class="px-3 py-2 text-gray-600 dark:text-gray-400">{{ formatDuration(log.duration_ms) }}</td>
                              </tr>
                            </tbody>
                          </table>
                        </div>

                        <Pagination
                          v-if="usagePagination[member.id]?.total > 0"
                          :page="usagePagination[member.id].page"
                          :total="usagePagination[member.id].total"
                          :page-size="usagePagination[member.id].page_size"
                          @update:page="(page: number) => loadMemberUsage(member, page)"
                        />
                      </div>
                    </td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </div>

    <!-- Transfer modal -->
    <Teleport to="body">
      <div
        v-if="transferTarget"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
        @click.self="closeTransfer"
      >
        <div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl dark:bg-dark-800">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('team.transfer.title') }}
          </h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
            {{ t('team.transfer.to', { email: transferTarget.email }) }}
          </p>

          <div class="mt-4 space-y-4">
            <div>
              <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('team.transfer.amount') }}</label>
              <input
                v-model.number="transferAmount"
                type="number"
                step="0.01"
                min="0.01"
                class="input w-full"
                :placeholder="t('team.transfer.amountPlaceholder')"
              />
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('team.transfer.password') }}</label>
              <input
                v-model="transferPassword"
                type="password"
                class="input w-full"
                :placeholder="t('team.transfer.passwordPlaceholder')"
              />
            </div>
          </div>

          <div class="mt-6 flex justify-end gap-2">
            <button class="btn btn-secondary" @click="closeTransfer">{{ t('common.cancel') }}</button>
            <button class="btn btn-primary" :disabled="transferring" @click="handleTransfer">
              <Icon v-if="transferring" name="refresh" size="sm" class="animate-spin" />
              <span>{{ transferring ? t('team.transfer.transferring') : t('team.transfer.confirm') }}</span>
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useClipboard } from '@/composables/useClipboard'
import { formatCurrency, formatDateTime } from '@/utils/format'
import { formatCacheTokens } from '@/utils/formatters'
import { extractApiErrorMessage } from '@/utils/apiError'
import { resolveUsageRequestType } from '@/utils/usageRequestType'
import {
  createTeam,
  getMyTeam,
  joinTeam,
  leaveTeam,
  listMembers,
  refreshInviteCode,
  removeMember,
  transferBalance,
  getMemberUsage
} from '@/api/team'
import type { Team, TeamMember, UsageLog } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const team = ref<Team | null>(null)
const members = ref<TeamMember[]>([])
const membersLoading = ref(false)
const memberSearch = ref('')

const createName = ref('')
const joinCode = ref('')
const creating = ref(false)
const joining = ref(false)
const leaving = ref(false)
const refreshing = ref(false)
const removingId = ref<number | null>(null)

const transferTarget = ref<TeamMember | null>(null)
const transferAmount = ref<number>(0)
const transferPassword = ref('')
const transferring = ref(false)

const expandedMemberId = ref<number | null>(null)
const memberUsage = ref<Record<number, UsageLog[]>>({})
const usageLoading = ref<Record<number, boolean>>({})
const usagePagination = ref<Record<number, { page: number; page_size: number; total: number }>>({})

const now = new Date()
const weekAgo = new Date(now)
weekAgo.setDate(weekAgo.getDate() - 6)
const formatLocalDate = (date: Date): string => {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}
const usageStartDate = ref(formatLocalDate(weekAgo))
const usageEndDate = ref(formatLocalDate(now))
const activePreset = ref('last7Days')

const datePresets = [
  { labelKey: 'dates.today', value: 'today' },
  { labelKey: 'dates.yesterday', value: 'yesterday' },
  { labelKey: 'dates.last7Days', value: 'last7Days' },
  { labelKey: 'dates.last30Days', value: 'last30Days' }
]

function applyDatePreset(value: string) {
  activePreset.value = value
  const today = new Date()
  const start = new Date(today)
  switch (value) {
    case 'today':
      break
    case 'yesterday':
      start.setDate(start.getDate() - 1)
      today.setDate(today.getDate() - 1)
      break
    case 'last7Days':
      start.setDate(start.getDate() - 6)
      break
    case 'last30Days':
      start.setDate(start.getDate() - 29)
      break
  }
  usageStartDate.value = formatLocalDate(start)
  usageEndDate.value = formatLocalDate(today)
}

const currentUserId = computed(() => authStore.user?.id ?? 0)
const isOwner = computed(() => team.value?.role === 'owner')

const filteredMembers = computed(() => {
  const query = memberSearch.value.trim().toLowerCase()
  if (!query) return members.value
  return members.value.filter(
    (m) =>
      m.email.toLowerCase().includes(query) ||
      (m.username && m.username.toLowerCase().includes(query))
  )
})

onMounted(() => {
  loadTeam()
})

function formatDuration(ms: number | null | undefined): string {
  if (ms == null) return '-'
  if (ms < 1000) return `${ms.toFixed(0)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

async function loadTeam() {
  loading.value = true
  try {
    team.value = await getMyTeam()
    if (team.value) {
      await loadMembers()
    }
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.loadError')))
  } finally {
    loading.value = false
  }
}

async function loadMembers() {
  membersLoading.value = true
  try {
    const result = await listMembers({ page: 1, page_size: 100 })
    members.value = result.items
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.members.loadError')))
  } finally {
    membersLoading.value = false
  }
}

// Owner can view any member's usage; regular members can only view their own.
function canViewUsage(member: TeamMember): boolean {
  return isOwner.value || member.id === currentUserId.value
}

function toggleMember(member: TeamMember) {
  if (expandedMemberId.value === member.id) {
    expandedMemberId.value = null
  } else {
    expandedMemberId.value = member.id
    if (canViewUsage(member) && !memberUsage.value[member.id]) {
      loadMemberUsage(member)
    }
  }
}

async function loadMemberUsage(member: TeamMember, page = 1) {
  usageLoading.value[member.id] = true
  try {
    const result = await getMemberUsage(member.id, {
      page,
      page_size: 10,
      start_date: usageStartDate.value,
      end_date: usageEndDate.value
    })
    memberUsage.value[member.id] = result.items
    usagePagination.value[member.id] = {
      page: result.page,
      page_size: result.page_size,
      total: result.total
    }
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.members.usage.loadError')))
  } finally {
    usageLoading.value[member.id] = false
  }
}

function getRequestTypeLabel(log: UsageLog): string {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'cyber') return t('usage.cyber')
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
}

function getRequestTypeBadgeClass(log: UsageLog): string {
  const requestType = resolveUsageRequestType(log)
  if (requestType === 'cyber') return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
  if (requestType === 'ws_v2') return 'bg-violet-100 text-violet-800 dark:bg-violet-900 dark:text-violet-200'
  if (requestType === 'stream') return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
  if (requestType === 'sync') return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
  return 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200'
}

async function handleCreate() {
  const name = createName.value.trim()
  if (!name) {
    appStore.showError(t('team.create.nameRequired'))
    return
  }
  creating.value = true
  try {
    team.value = await createTeam(name)
    createName.value = ''
    appStore.showSuccess(t('team.create.success'))
    await loadMembers()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.create.error')))
  } finally {
    creating.value = false
  }
}

async function handleJoin() {
  const code = joinCode.value.trim()
  if (!code) {
    appStore.showError(t('team.join.codeRequired'))
    return
  }
  joining.value = true
  try {
    await joinTeam(code)
    joinCode.value = ''
    appStore.showSuccess(t('team.join.success'))
    await loadTeam()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.join.error')))
  } finally {
    joining.value = false
  }
}

async function handleRefreshCode() {
  refreshing.value = true
  try {
    const result = await refreshInviteCode()
    if (team.value) {
      team.value.invite_code = result.invite_code
    }
    appStore.showSuccess(t('team.inviteCode.refreshSuccess'))
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.inviteCode.refreshError')))
  } finally {
    refreshing.value = false
  }
}

async function handleLeave() {
  if (!confirm(t('team.leave.confirm'))) return
  leaving.value = true
  try {
    await leaveTeam()
    team.value = null
    members.value = []
    appStore.showSuccess(t('team.leave.success'))
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.leave.error')))
  } finally {
    leaving.value = false
  }
}

async function handleRemove(member: TeamMember) {
  if (!confirm(t('team.members.removeConfirm', { email: member.email }))) return
  removingId.value = member.id
  try {
    await removeMember(member.id)
    appStore.showSuccess(t('team.members.removeSuccess'))
    await loadMembers()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.members.removeError')))
  } finally {
    removingId.value = null
  }
}

function openTransfer(member: TeamMember) {
  transferTarget.value = member
  transferAmount.value = 0
  transferPassword.value = ''
}

function closeTransfer() {
  transferTarget.value = null
  transferAmount.value = 0
  transferPassword.value = ''
}

async function handleTransfer() {
  if (!transferTarget.value) return
  if (transferAmount.value <= 0) {
    appStore.showError(t('team.transfer.amountRequired'))
    return
  }
  if (!transferPassword.value) {
    appStore.showError(t('team.transfer.passwordRequired'))
    return
  }
  transferring.value = true
  try {
    await transferBalance(transferTarget.value.id, transferAmount.value, transferPassword.value)
    appStore.showSuccess(t('team.transfer.success'))
    closeTransfer()
    await loadMembers()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('team.transfer.error')))
  } finally {
    transferring.value = false
  }
}

async function copyCode() {
  if (!team.value?.invite_code) return
  await copyToClipboard(team.value.invite_code, t('team.inviteCode.copied'))
}
</script>
