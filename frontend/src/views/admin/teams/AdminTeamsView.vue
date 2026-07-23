<!-- 管理员团队总览、申请审核、详情与治理配置。 -->
<template>
  <AppLayout>
    <main class="admin_teams space-y-6">
      <!-- Top Banner -->
      <div class="relative overflow-hidden rounded-3xl bg-gradient-to-r from-slate-900 via-indigo-950 to-slate-900 p-6 text-white shadow-xl dark:border dark:border-white/10 dark:shadow-none">
        <div class="absolute -right-10 -top-10 h-64 w-64 rounded-full bg-indigo-500/10 blur-3xl"></div>
        <div class="relative z-10 flex flex-wrap items-center justify-between gap-4">
          <div>
            <div class="inline-flex items-center gap-2 rounded-full bg-white/10 px-3 py-1 text-xs font-medium text-indigo-200 backdrop-blur-md">
              <Icon name="shield" size="xs" />
              <span>团队系统治理</span>
            </div>
            <h1 class="mt-2 text-2xl font-extrabold text-white sm:text-3xl">{{ t('teamAdmin.title') }}</h1>
            <p class="mt-1 text-sm text-gray-300">{{ t('teamAdmin.description') }}</p>
          </div>
          <button class="btn btn-secondary border-white/20 bg-white/10 text-white hover:bg-white/20" :disabled="loading" @click="loadAll">
            <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
            <span>{{ t('common.refresh') }}</span>
          </button>
        </div>
      </div>

      <!-- Stats Grid -->
      <section class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div class="stat-card">
          <div class="stat-icon bg-indigo-100 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400">
            <Icon name="users" size="md" />
          </div>
          <div>
            <div class="stat-label">{{ t('teamAdmin.stats.total') }}</div>
            <div class="stat-value text-gray-900 dark:text-white">{{ stats.total_teams }}</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400">
            <Icon name="clock" size="md" />
          </div>
          <div>
            <div class="stat-label">{{ t('teamAdmin.stats.pending') }}</div>
            <div class="stat-value text-amber-600 dark:text-amber-400">{{ stats.pending_applications }}</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon bg-emerald-100 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400">
            <Icon name="checkCircle" size="md" />
          </div>
          <div>
            <div class="stat-label">{{ t('teamAdmin.stats.active') }}</div>
            <div class="stat-value text-emerald-600 dark:text-emerald-400">{{ activeCount }}</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400">
            <Icon name="exclamationCircle" size="md" />
          </div>
          <div>
            <div class="stat-label">{{ t('teamAdmin.stats.reviewRequired') }}</div>
            <div class="stat-value text-red-600 dark:text-red-400">{{ reviewRequiredCount }}</div>
          </div>
        </div>
      </section>

      <!-- Tabs Nav -->
      <nav class="flex gap-2 rounded-2xl bg-gray-100/80 p-1.5 dark:bg-dark-800/80" :aria-label="t('teamAdmin.title')">
        <button
          v-for="item in tabs"
          :key="item.key"
          class="rounded-xl px-5 py-2 text-sm font-semibold transition-all duration-200"
          :class="tab === item.key
            ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
            : 'text-gray-500 hover:text-gray-900 dark:text-dark-400 dark:hover:text-white'"
          @click="tab = item.key"
        >
          {{ item.label }}
        </button>
      </nav>

      <!-- Tab 1: Teams List -->
      <section v-if="tab === 'teams'" class="space-y-4">
        <div class="flex flex-wrap items-center gap-3">
          <div class="relative flex-1 min-w-[240px]">
            <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-gray-400">
              <Icon name="search" size="sm" />
            </div>
            <input
              v-model="filters.search"
              class="input w-full pl-9"
              :placeholder="t('teamAdmin.search')"
              @keyup.enter="loadTeams"
            />
          </div>
          <select v-model="filters.status" class="input w-40" @change="loadTeams">
            <option value="">{{ t('teamAdmin.allStatuses') }}</option>
            <option value="active">{{ t('teamAdmin.active') }}</option>
            <option value="disabled">{{ t('teamAdmin.frozen') }}</option>
          </select>
          <button class="btn btn-primary" @click="loadTeams">
            <Icon name="search" size="sm" />
            <span>{{ t('common.search') }}</span>
          </button>
        </div>

        <div class="card overflow-x-auto p-0">
          <table class="w-full min-w-[980px] text-left text-sm">
            <thead class="border-b border-gray-100 bg-gray-50/80 text-xs font-semibold text-gray-500 dark:border-dark-800 dark:bg-dark-800/50 dark:text-dark-400">
              <tr>
                <th class="p-3.5">{{ t('teamAdmin.team') }}</th>
                <th class="p-3.5">{{ t('teamAdmin.owner') }}</th>
                <th class="p-3.5">{{ t('teamAdmin.level') }}</th>
                <th class="p-3.5">{{ t('teamAdmin.members') }}</th>
                <th class="p-3.5">{{ t('teamAdmin.balance') }}</th>
                <th class="p-3.5">{{ t('teamAdmin.recharge') }}</th>
                <th class="p-3.5">{{ t('teamAdmin.spend7d') }}</th>
                <th class="p-3.5">{{ t('common.status') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
              <tr
                v-for="item in teams"
                :key="item.id"
                class="cursor-pointer transition-colors hover:bg-gray-50/80 dark:hover:bg-dark-800/40"
                @click="openTeam(item.id)"
              >
                <td class="p-3.5">
                  <p class="font-bold text-gray-900 dark:text-white">{{ item.name }}</p>
                  <span v-if="item.review_required" class="mt-0.5 inline-block text-xs font-semibold text-amber-600 dark:text-amber-400">
                    {{ t('teamAdmin.reviewRequired') }}
                  </span>
                </td>
                <td class="p-3.5 text-gray-600 dark:text-gray-300">{{ item.owner_email }}</td>
                <td class="p-3.5"><span class="badge badge-purple">Lv.{{ item.level }}</span></td>
                <td class="p-3.5 font-mono text-xs">{{ item.member_count }} / {{ item.member_limit }}</td>
                <td class="p-3.5 font-mono font-semibold tabular-nums text-emerald-600 dark:text-emerald-400">{{ formatCurrency(item.balance) }}</td>
                <td class="p-3.5 font-mono tabular-nums text-gray-600 dark:text-gray-300">{{ formatCurrency(item.effective_recharge) }}</td>
                <td class="p-3.5 font-mono tabular-nums text-amber-600 dark:text-amber-400">{{ formatCurrency(item.spend_7d) }}</td>
                <td class="p-3.5">
                  <span class="badge" :class="item.status === 'active' ? 'badge-success' : 'badge-danger'">
                    {{ item.status === 'active' ? t('teamAdmin.active') : t('teamAdmin.frozen') }}
                  </span>
                </td>
              </tr>
              <tr v-if="!teams.length">
                <td colspan="8" class="p-8 text-center text-gray-500">{{ t('teamAdmin.empty') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <Pagination v-if="pagination.total" :page="pagination.page" :page-size="pagination.page_size" :total="pagination.total" @update:page="changePage" />
      </section>

      <!-- Tab 2: Applications -->
      <section v-else-if="tab === 'applications'" class="card overflow-x-auto p-0">
        <table class="w-full min-w-[900px] text-left text-sm">
          <thead class="border-b border-gray-100 bg-gray-50/80 text-xs font-semibold text-gray-500 dark:border-dark-800 dark:bg-dark-800/50 dark:text-dark-400">
            <tr>
              <th class="p-3.5">ID</th>
              <th class="p-3.5">{{ t('teamAdmin.type') }}</th>
              <th class="p-3.5">{{ t('teamAdmin.applicant') }}</th>
              <th class="p-3.5">{{ t('teamAdmin.eligibility') }}</th>
              <th class="p-3.5">{{ t('teamAdmin.team') }}</th>
              <th class="p-3.5">{{ t('teamAdmin.targetLimit') }}</th>
              <th class="p-3.5">{{ t('teamAdmin.reason') }}</th>
              <th class="p-3.5">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
            <tr v-for="item in applications" :key="item.id" class="transition-colors hover:bg-gray-50/80 dark:hover:bg-dark-800/40">
              <td class="p-3.5 font-mono text-gray-400">#{{ item.id }}</td>
              <td class="p-3.5 font-medium">
                <span class="badge" :class="item.application_type === 'create' ? 'badge-primary' : 'badge-purple'">
                  {{ t(`teamAdmin.applicationType.${item.application_type}`) }}
                </span>
              </td>
              <td class="p-3.5 font-semibold text-gray-900 dark:text-white">{{ item.applicant_email }}</td>
              <td class="p-3.5 text-xs">
                <p class="font-medium text-gray-900 dark:text-white">{{ item.registration_days }} {{ t('teamAdmin.days') }}</p>
                <p class="text-emerald-600 dark:text-emerald-400 font-mono">{{ formatCurrency(item.effective_recharge) }}</p>
              </td>
              <td class="p-3.5 font-medium">{{ item.team_name || `#${item.team_id}` }}</td>
              <td class="p-3.5 font-mono text-xs">{{ item.target_limit || '-' }}</td>
              <td class="max-w-72 p-3.5"><p class="truncate text-gray-500 dark:text-dark-400 text-xs">{{ item.reason || '-' }}</p></td>
              <td class="p-3.5">
                <button class="btn btn-primary btn-sm" @click="openReview(item)">
                  <span>{{ t('teamAdmin.review') }}</span>
                </button>
              </td>
            </tr>
            <tr v-if="!applications.length">
              <td colspan="8" class="p-8 text-center text-gray-500">{{ t('teamAdmin.noApplications') }}</td>
            </tr>
          </tbody>
        </table>
      </section>

      <!-- Tab 3: Settings -->
      <section v-else class="space-y-4">
        <div class="card p-6 shadow-sm">
          <div class="grid gap-4 md:grid-cols-2">
            <div>
              <label class="input-label">{{ t('teamAdmin.settings.registrationDays') }}</label>
              <input v-model.number="settings.min_registration_days" type="number" min="0" class="input w-full" />
            </div>
            <div>
              <label class="input-label">{{ t('teamAdmin.settings.minRecharge') }}</label>
              <input v-model.number="settings.min_total_recharge" type="number" min="0" step="0.01" class="input w-full" />
            </div>
          </div>
        </div>

        <div class="card overflow-x-auto p-0 shadow-sm">
          <table class="w-full min-w-[720px] text-left text-sm">
            <thead class="border-b border-gray-100 bg-gray-50/80 text-xs font-semibold text-gray-500 dark:border-dark-800 dark:bg-dark-800/50 dark:text-dark-400">
              <tr>
                <th class="p-3.5">{{ t('teamAdmin.level') }}</th>
                <th class="p-3.5">{{ t('teamAdmin.recharge') }}</th>
                <th class="p-3.5">{{ t('teamAdmin.spend7d') }}</th>
                <th class="p-3.5">{{ t('teamAdmin.settings.mode') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
              <tr v-for="level in settings.levels" :key="level.limit" class="transition-colors hover:bg-gray-50/50 dark:hover:bg-dark-800/30">
                <td class="p-3.5 font-bold">上限 {{ level.limit }} 人</td>
                <td class="p-3.5"><input v-model.number="level.recharge" type="number" min="0" step="0.01" class="input w-full" /></td>
                <td class="p-3.5"><input v-model.number="level.spend_7d" type="number" min="0" step="0.01" class="input w-full" /></td>
                <td class="p-3.5">
                  <select v-model="level.mode" class="input w-full">
                    <option value="and">AND (且)</option>
                    <option value="or">OR (或)</option>
                  </select>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="flex justify-end">
          <button class="btn btn-primary" :disabled="savingSettings" @click="saveSettings">
            <Icon v-if="savingSettings" name="refresh" size="sm" class="animate-spin" />
            <span>{{ t('common.save') }}</span>
          </button>
        </div>
      </section>

      <!-- Team Detail Side Drawer -->
      <Teleport to="body">
        <div v-if="detail" class="fixed inset-0 z-50 flex justify-end bg-black/40 backdrop-blur-sm" @click.self="detail = null">
          <aside class="h-full w-full max-w-3xl overflow-y-auto bg-white p-6 shadow-2xl dark:bg-dark-900">
            <div class="flex items-start justify-between border-b border-gray-100 pb-4 dark:border-dark-800">
              <div>
                <h2 class="text-xl font-bold text-gray-900 dark:text-white">{{ detail.team.name }}</h2>
                <p class="text-xs text-gray-500 dark:text-dark-400">发起人: {{ detail.team.owner_email }}</p>
              </div>
              <button class="btn btn-ghost btn-icon" @click="detail = null">
                <Icon name="x" size="md" />
              </button>
            </div>

            <div class="mt-5 flex flex-wrap items-center gap-2">
              <span class="text-xs font-medium text-gray-500">修改上限:</span>
              <input v-model.number="detailLimit" type="number" min="1" class="input w-28 py-1 text-sm" />
              <button class="btn btn-secondary btn-sm" @click="saveLimit">{{ t('teamAdmin.saveLimit') }}</button>
              <button v-if="detail.team.review_required" class="btn btn-warning btn-sm" @click="completeReview">{{ t('teamAdmin.completeReview') }}</button>
              <button class="btn btn-sm ml-auto" :class="detail.team.status === 'active' ? 'btn-danger-outline' : 'btn-primary'" @click="toggleStatus">
                {{ detail.team.status === 'active' ? t('teamAdmin.freeze') : t('teamAdmin.restore') }}
              </button>
            </div>

            <h3 class="mt-8 font-bold text-gray-900 dark:text-white">{{ t('teamAdmin.members') }}</h3>
            <div class="mt-3 divide-y divide-gray-100 dark:divide-dark-800 rounded-xl border border-gray-100 p-2 dark:border-dark-800">
              <div v-for="member in detail.members" :key="member.id" class="grid items-center gap-2 p-3 sm:grid-cols-[1fr_auto_auto]">
                <div>
                  <p class="font-semibold text-gray-900 dark:text-white">{{ member.email }}</p>
                  <span class="badge badge-gray mt-0.5">{{ member.role }}</span>
                </div>
                <div class="text-right text-sm">
                  <p class="font-mono font-semibold text-emerald-600 dark:text-emerald-400">{{ formatCurrency(member.balance) }}</p>
                  <p class="text-xs text-gray-400">{{ t('teamAdmin.transferable') }} {{ formatCurrency(member.transferable_balance) }}</p>
                </div>
                <button v-if="member.role !== 'owner'" class="btn btn-danger-outline btn-sm ml-2" @click="removeDetailMember(member.id)">
                  {{ t('common.remove') }}
                </button>
              </div>
            </div>

            <h3 class="mt-8 font-bold text-gray-900 dark:text-white">{{ t('teamAdmin.fundLedger') }}</h3>
            <div class="mt-3 overflow-x-auto rounded-xl border border-gray-100 dark:border-dark-800">
              <table class="w-full min-w-[560px] text-left text-xs">
                <thead class="bg-gray-50 dark:bg-dark-800/50 text-gray-500">
                  <tr>
                    <th class="p-3">{{ t('teamAdmin.type') }}</th>
                    <th class="p-3">{{ t('teamAdmin.applicant') }}</th>
                    <th class="p-3">{{ t('teamAdmin.balance') }}</th>
                    <th class="p-3">{{ t('common.timeLabel') }}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
                  <tr v-for="entry in detail.fund_ledger" :key="entry.id">
                    <td class="p-3 font-medium">{{ entry.action }}</td>
                    <td class="p-3">{{ entry.user_id || '-' }}</td>
                    <td class="p-3 font-mono font-bold text-emerald-600 dark:text-emerald-400">{{ formatCurrency(entry.amount) }}</td>
                    <td class="p-3 text-gray-400 font-mono">{{ formatDateTime(entry.created_at) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </aside>
        </div>
      </Teleport>

      <!-- Review Application Modal -->
      <Teleport to="body">
        <div v-if="reviewing" class="modal-overlay" @click.self="reviewing = null">
          <div class="modal-content max-w-lg">
            <div class="modal-header">
              <h3 class="modal-title">{{ t('teamAdmin.review') }} #{{ reviewing.id }}</h3>
              <button class="btn btn-ghost btn-icon" @click="reviewing = null">
                <Icon name="x" size="sm" />
              </button>
            </div>
            <div class="modal-body space-y-4">
              <div>
                <label class="input-label">{{ t('teamAdmin.reviewReason') }}</label>
                <textarea v-model="reviewForm.review_reason" class="input min-h-24 w-full" :placeholder="t('teamAdmin.reviewReason')"></textarea>
              </div>

              <label v-if="reviewing.application_type === 'create'" class="flex items-center gap-2 text-sm">
                <input v-model="reviewForm.waive" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
                <span>{{ t('teamAdmin.waive') }}</span>
              </label>

              <div v-if="reviewing.application_type === 'expand'">
                <label class="input-label">{{ t('teamAdmin.targetLimit') }}</label>
                <input v-model.number="reviewForm.target_limit" type="number" min="41" class="input w-full" />
              </div>
            </div>
            <div class="modal-footer">
              <button class="btn btn-secondary" @click="reviewApplication(false)">{{ t('common.reject') }}</button>
              <button class="btn btn-primary" @click="reviewApplication(true)">{{ t('common.approve') }}</button>
            </div>
          </div>
        </div>
      </Teleport>
    </main>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'
import { adminTeamsAPI } from '@/api/admin/teams'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatCurrency, formatDateTime } from '@/utils/format'
import type { AdminTeamDetail, AdminTeamSummary, TeamAdminStats, TeamApplication, TeamGovernanceSettings } from '@/types'

type Tab = 'teams' | 'applications' | 'settings'
const { t } = useI18n()
const appStore = useAppStore()
const tab = ref<Tab>('teams')
const loading = ref(false)
const savingSettings = ref(false)
const stats = reactive<TeamAdminStats>({ total_teams: 0, pending_applications: 0 })
const teams = ref<AdminTeamSummary[]>([])
const applications = ref<TeamApplication[]>([])
const filters = reactive({ search: '', status: '' })
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
const detail = ref<AdminTeamDetail | null>(null)
const detailLimit = ref(5)
const reviewing = ref<TeamApplication | null>(null)
const reviewForm = reactive({ review_reason: '', waive: false, target_limit: 41 })
const settings = reactive<TeamGovernanceSettings>({ configured: false, min_registration_days: 0, min_total_recharge: 0, levels: [], updated_at: '' })

const tabs = computed(() => [{ key: 'teams' as const, label: t('teamAdmin.tabs.teams') }, { key: 'applications' as const, label: t('teamAdmin.tabs.applications') }, { key: 'settings' as const, label: t('teamAdmin.tabs.settings') }])
const activeCount = computed(() => teams.value.filter((item) => item.status === 'active').length)
const reviewRequiredCount = computed(() => teams.value.filter((item) => item.review_required).length)

onMounted(loadAll)

async function loadAll() {
  loading.value = true
  try {
    const [statsData, settingsData] = await Promise.all([adminTeamsAPI.getStats(), adminTeamsAPI.getSettings()])
    Object.assign(stats, statsData)
    Object.assign(settings, settingsData)
    await Promise.all([loadTeams(), loadApplications()])
  } catch (error) { appStore.showError(extractApiErrorMessage(error, t('teamAdmin.operationFailed'))) }
  finally { loading.value = false }
}

async function loadTeams() {
  const result = await adminTeamsAPI.listTeams({ ...filters, page: pagination.page, page_size: pagination.page_size })
  teams.value = result.items ?? []
  pagination.total = result.total
}

async function loadApplications() {
  const result = await adminTeamsAPI.listApplications({ status: 'pending', page: 1, page_size: 100 })
  applications.value = result.items ?? []
}

function changePage(page: number) { pagination.page = page; void loadTeams() }
async function openTeam(id: number) { detail.value = await adminTeamsAPI.getTeam(id); detailLimit.value = detail.value.team.member_limit }
function openReview(item: TeamApplication) { reviewing.value = item; reviewForm.review_reason = ''; reviewForm.waive = false; reviewForm.target_limit = item.target_limit ?? 41 }

async function reviewApplication(approve: boolean) {
  if (!reviewing.value) return
  try {
    await adminTeamsAPI.reviewApplication(reviewing.value.id, { approve, review_reason: reviewForm.review_reason, waive: reviewForm.waive, target_limit: reviewing.value.application_type === 'expand' ? reviewForm.target_limit : undefined })
    reviewing.value = null
    await loadAll()
  } catch (error) { appStore.showError(extractApiErrorMessage(error, t('teamAdmin.operationFailed'))) }
}

async function saveSettings() {
  savingSettings.value = true
  try { Object.assign(settings, await adminTeamsAPI.updateSettings(settings)); appStore.showSuccess(t('teamAdmin.saved')) }
  catch (error) { appStore.showError(extractApiErrorMessage(error, t('teamAdmin.operationFailed'))) }
  finally { savingSettings.value = false }
}

async function saveLimit() { if (!detail.value) return; await adminTeamsAPI.setMemberLimit(detail.value.team.id, detailLimit.value); await openTeam(detail.value.team.id); await loadTeams() }
async function completeReview() { if (!detail.value) return; await adminTeamsAPI.markReviewed(detail.value.team.id); await openTeam(detail.value.team.id); await loadTeams() }
async function toggleStatus() { if (!detail.value) return; await adminTeamsAPI.setStatus(detail.value.team.id, detail.value.team.status === 'active' ? 'frozen' : 'active'); await openTeam(detail.value.team.id); await loadTeams() }
async function removeDetailMember(memberId: number) { if (!detail.value || !confirm(t('teamAdmin.removeConfirm'))) return; await adminTeamsAPI.removeMember(detail.value.team.id, memberId); await openTeam(detail.value.team.id); await loadTeams() }
</script>
