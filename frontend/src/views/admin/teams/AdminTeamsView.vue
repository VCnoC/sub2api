<!-- 管理员团队总览、申请审核、详情与治理配置。 -->
<template>
  <AppLayout>
    <main class="admin_teams space-y-6">
      <header class="flex flex-wrap items-end justify-between gap-4">
        <div><h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('teamAdmin.title') }}</h1><p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('teamAdmin.description') }}</p></div>
        <button class="btn btn-secondary" :disabled="loading" @click="loadAll"><Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />{{ t('common.refresh') }}</button>
      </header>

      <section class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div class="card p-4"><p class="text-xs text-gray-500">{{ t('teamAdmin.stats.total') }}</p><strong class="mt-1 block text-2xl">{{ stats.total_teams }}</strong></div>
        <div class="card p-4"><p class="text-xs text-gray-500">{{ t('teamAdmin.stats.pending') }}</p><strong class="mt-1 block text-2xl text-amber-600">{{ stats.pending_applications }}</strong></div>
        <div class="card p-4"><p class="text-xs text-gray-500">{{ t('teamAdmin.stats.active') }}</p><strong class="mt-1 block text-2xl">{{ activeCount }}</strong></div>
        <div class="card p-4"><p class="text-xs text-gray-500">{{ t('teamAdmin.stats.reviewRequired') }}</p><strong class="mt-1 block text-2xl">{{ reviewRequiredCount }}</strong></div>
      </section>

      <nav class="flex gap-1 border-b border-gray-200 dark:border-dark-700" :aria-label="t('teamAdmin.title')">
        <button v-for="item in tabs" :key="item.key" class="border-b-2 px-4 py-3 text-sm font-medium" :class="tab === item.key ? 'border-primary-500 text-primary-600' : 'border-transparent text-gray-500'" @click="tab = item.key">{{ item.label }}</button>
      </nav>

      <section v-if="tab === 'teams'" class="space-y-4">
        <div class="flex flex-wrap gap-3">
          <input v-model="filters.search" class="input min-w-64 flex-1" :placeholder="t('teamAdmin.search')" @keyup.enter="loadTeams" />
          <select v-model="filters.status" class="input w-40" @change="loadTeams"><option value="">{{ t('teamAdmin.allStatuses') }}</option><option value="active">{{ t('teamAdmin.active') }}</option><option value="disabled">{{ t('teamAdmin.frozen') }}</option></select>
          <button class="btn btn-primary" @click="loadTeams"><Icon name="search" size="sm" />{{ t('common.search') }}</button>
        </div>
        <div class="card overflow-x-auto p-0">
          <table class="w-full min-w-[980px] text-left text-sm">
            <thead class="border-b border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-800"><tr><th class="p-3">{{ t('teamAdmin.team') }}</th><th class="p-3">{{ t('teamAdmin.owner') }}</th><th class="p-3">{{ t('teamAdmin.level') }}</th><th class="p-3">{{ t('teamAdmin.members') }}</th><th class="p-3">{{ t('teamAdmin.balance') }}</th><th class="p-3">{{ t('teamAdmin.recharge') }}</th><th class="p-3">{{ t('teamAdmin.spend7d') }}</th><th class="p-3">{{ t('common.status') }}</th></tr></thead>
            <tbody><tr v-for="item in teams" :key="item.id" class="cursor-pointer border-b border-gray-100 hover:bg-gray-50 dark:border-dark-800 dark:hover:bg-dark-900" @click="openTeam(item.id)"><td class="p-3"><p class="font-medium">{{ item.name }}</p><span v-if="item.review_required" class="text-xs text-amber-600">{{ t('teamAdmin.reviewRequired') }}</span></td><td class="p-3">{{ item.owner_email }}</td><td class="p-3">{{ item.level }}</td><td class="p-3">{{ item.member_count }} / {{ item.member_limit }}</td><td class="p-3 tabular-nums">{{ formatCurrency(item.balance) }}</td><td class="p-3 tabular-nums">{{ formatCurrency(item.effective_recharge) }}</td><td class="p-3 tabular-nums">{{ formatCurrency(item.spend_7d) }}</td><td class="p-3"><span class="badge" :class="item.status === 'active' ? 'badge-success' : 'badge-danger'">{{ item.status === 'active' ? t('teamAdmin.active') : t('teamAdmin.frozen') }}</span></td></tr><tr v-if="!teams.length"><td colspan="8" class="p-8 text-center text-gray-500">{{ t('teamAdmin.empty') }}</td></tr></tbody>
          </table>
        </div>
        <Pagination v-if="pagination.total" :page="pagination.page" :page-size="pagination.page_size" :total="pagination.total" @update:page="changePage" />
      </section>

      <section v-else-if="tab === 'applications'" class="card overflow-x-auto p-0">
        <table class="w-full min-w-[900px] text-left text-sm">
          <thead class="border-b border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-800"><tr><th class="p-3">ID</th><th class="p-3">{{ t('teamAdmin.type') }}</th><th class="p-3">{{ t('teamAdmin.applicant') }}</th><th class="p-3">{{ t('teamAdmin.eligibility') }}</th><th class="p-3">{{ t('teamAdmin.team') }}</th><th class="p-3">{{ t('teamAdmin.targetLimit') }}</th><th class="p-3">{{ t('teamAdmin.reason') }}</th><th class="p-3">{{ t('common.actions') }}</th></tr></thead>
          <tbody><tr v-for="item in applications" :key="item.id" class="border-b border-gray-100 dark:border-dark-800"><td class="p-3">#{{ item.id }}</td><td class="p-3">{{ t(`teamAdmin.applicationType.${item.application_type}`) }}</td><td class="p-3">{{ item.applicant_email }}</td><td class="p-3"><p>{{ item.registration_days }} {{ t('teamAdmin.days') }}</p><p class="text-xs text-gray-500">{{ formatCurrency(item.effective_recharge) }}</p></td><td class="p-3">{{ item.team_name || `#${item.team_id}` }}</td><td class="p-3">{{ item.target_limit || '-' }}</td><td class="max-w-72 p-3"><p class="truncate">{{ item.reason || '-' }}</p></td><td class="p-3"><button class="btn btn-primary btn-sm" @click="openReview(item)">{{ t('teamAdmin.review') }}</button></td></tr><tr v-if="!applications.length"><td colspan="8" class="p-8 text-center text-gray-500">{{ t('teamAdmin.noApplications') }}</td></tr></tbody>
        </table>
      </section>

      <section v-else class="space-y-4">
        <div class="card p-6">
          <div class="grid gap-4 md:grid-cols-2"><label><span class="input-label">{{ t('teamAdmin.settings.registrationDays') }}</span><input v-model.number="settings.min_registration_days" type="number" min="0" class="input w-full" /></label><label><span class="input-label">{{ t('teamAdmin.settings.minRecharge') }}</span><input v-model.number="settings.min_total_recharge" type="number" min="0" step="0.01" class="input w-full" /></label></div>
        </div>
        <div class="card overflow-x-auto p-0"><table class="w-full min-w-[720px] text-left text-sm"><thead class="border-b border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-800"><tr><th class="p-3">{{ t('teamAdmin.level') }}</th><th class="p-3">{{ t('teamAdmin.recharge') }}</th><th class="p-3">{{ t('teamAdmin.spend7d') }}</th><th class="p-3">{{ t('teamAdmin.settings.mode') }}</th></tr></thead><tbody><tr v-for="level in settings.levels" :key="level.limit" class="border-b border-gray-100 dark:border-dark-800"><td class="p-3">{{ level.limit }}</td><td class="p-3"><input v-model.number="level.recharge" type="number" min="0" step="0.01" class="input w-full" /></td><td class="p-3"><input v-model.number="level.spend_7d" type="number" min="0" step="0.01" class="input w-full" /></td><td class="p-3"><select v-model="level.mode" class="input w-full"><option value="and">AND</option><option value="or">OR</option></select></td></tr></tbody></table></div>
        <div class="flex justify-end"><button class="btn btn-primary" :disabled="savingSettings" @click="saveSettings"><Icon v-if="savingSettings" name="refresh" size="sm" class="animate-spin" />{{ t('common.save') }}</button></div>
      </section>

      <Teleport to="body">
        <div v-if="detail" class="fixed inset-0 z-50 flex justify-end bg-black/40" @click.self="detail = null">
          <aside class="h-full w-full max-w-3xl overflow-y-auto bg-white p-6 shadow-xl dark:bg-dark-900">
            <div class="flex items-start justify-between"><div><h2 class="text-xl font-semibold">{{ detail.team.name }}</h2><p class="text-sm text-gray-500">{{ detail.team.owner_email }}</p></div><button class="btn btn-ghost" @click="detail = null"><Icon name="x" /></button></div>
            <div class="mt-5 flex flex-wrap gap-2"><input v-model.number="detailLimit" type="number" min="1" class="input w-28" /><button class="btn btn-secondary" @click="saveLimit">{{ t('teamAdmin.saveLimit') }}</button><button v-if="detail.team.review_required" class="btn btn-secondary" @click="completeReview">{{ t('teamAdmin.completeReview') }}</button><button class="btn" :class="detail.team.status === 'active' ? 'btn-danger-outline' : 'btn-primary'" @click="toggleStatus">{{ detail.team.status === 'active' ? t('teamAdmin.freeze') : t('teamAdmin.restore') }}</button></div>
            <h3 class="mt-8 font-semibold">{{ t('teamAdmin.members') }}</h3><div class="mt-3 divide-y divide-gray-200 dark:divide-dark-700"><div v-for="member in detail.members" :key="member.id" class="grid gap-2 py-3 sm:grid-cols-[1fr_auto_auto]"><div><p class="font-medium">{{ member.email }}</p><p class="text-xs text-gray-500">{{ member.role }}</p></div><div class="text-right text-sm"><p>{{ formatCurrency(member.balance) }}</p><p class="text-xs text-gray-500">{{ t('teamAdmin.transferable') }} {{ formatCurrency(member.transferable_balance) }}</p></div><button v-if="member.role !== 'owner'" class="btn btn-danger-outline btn-sm" @click="removeDetailMember(member.id)">{{ t('common.remove') }}</button></div></div>
            <h3 class="mt-8 font-semibold">{{ t('teamAdmin.fundLedger') }}</h3><div class="mt-3 overflow-x-auto"><table class="w-full min-w-[560px] text-left text-sm"><thead><tr><th class="p-2">{{ t('teamAdmin.type') }}</th><th class="p-2">{{ t('teamAdmin.applicant') }}</th><th class="p-2">{{ t('teamAdmin.balance') }}</th><th class="p-2">{{ t('common.timeLabel') }}</th></tr></thead><tbody><tr v-for="entry in detail.fund_ledger" :key="entry.id" class="border-t border-gray-100 dark:border-dark-800"><td class="p-2">{{ entry.action }}</td><td class="p-2">{{ entry.user_id || '-' }}</td><td class="p-2">{{ formatCurrency(entry.amount) }}</td><td class="p-2">{{ formatDateTime(entry.created_at) }}</td></tr></tbody></table></div>
          </aside>
        </div>
      </Teleport>

      <Teleport to="body"><div v-if="reviewing" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" @click.self="reviewing = null"><div class="card w-full max-w-lg p-6"><h3 class="text-lg font-semibold">{{ t('teamAdmin.review') }} #{{ reviewing.id }}</h3><textarea v-model="reviewForm.review_reason" class="input mt-4 min-h-24 w-full" :placeholder="t('teamAdmin.reviewReason')"></textarea><label v-if="reviewing.application_type === 'create'" class="mt-3 flex items-center gap-2"><input v-model="reviewForm.waive" type="checkbox" />{{ t('teamAdmin.waive') }}</label><label v-if="reviewing.application_type === 'expand'" class="mt-3 block"><span class="input-label">{{ t('teamAdmin.targetLimit') }}</span><input v-model.number="reviewForm.target_limit" type="number" min="41" class="input w-full" /></label><div class="mt-5 flex justify-end gap-2"><button class="btn btn-secondary" @click="reviewApplication(false)">{{ t('common.reject') }}</button><button class="btn btn-primary" @click="reviewApplication(true)">{{ t('common.approve') }}</button></div></div></div></Teleport>
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
