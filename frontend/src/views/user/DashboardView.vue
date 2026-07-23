<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- 系统醒目提醒：由「系统设置 → 通用设置」配置，空内容不展示 -->
      <div
        v-if="dashboardNotice"
        class="relative overflow-hidden rounded-2xl border border-amber-200/80 bg-gradient-to-r from-amber-50 via-orange-50 to-amber-50 p-4 shadow-sm dark:border-amber-800/50 dark:from-amber-950/40 dark:via-orange-950/30 dark:to-amber-950/40 sm:p-5"
      >
        <div class="pointer-events-none absolute -right-6 -top-8 h-28 w-28 rounded-full bg-amber-300/20 blur-2xl dark:bg-amber-500/10"></div>
        <div class="relative z-10 flex items-start gap-3">
          <div class="mt-0.5 flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl bg-amber-100 text-amber-600 dark:bg-amber-900/50 dark:text-amber-400">
            <Icon name="exclamationTriangle" size="md" />
          </div>
          <div class="min-w-0 flex-1">
            <p class="text-xs font-semibold uppercase tracking-wide text-amber-700/80 dark:text-amber-400/90">
              {{ t('dashboard.noticeLabel') }}
            </p>
            <p class="mt-1 whitespace-pre-wrap text-sm font-medium leading-relaxed text-amber-950 dark:text-amber-100">
              {{ dashboardNotice }}
            </p>
          </div>
        </div>
      </div>

      <!-- Welcome hero banner -->
      <div class="relative overflow-hidden rounded-3xl bg-gradient-to-r from-primary-600 via-teal-500 to-cyan-500 p-8 text-white shadow-[0_8px_32px_rgba(20,184,166,0.25)] animate-fade-in dark:from-primary-700 dark:via-teal-600 dark:to-cyan-700">
        <!-- 装饰光斑 -->
        <div class="pointer-events-none absolute -right-10 -top-16 h-64 w-64 rounded-full bg-white/20 blur-3xl mix-blend-overlay"></div>
        <div class="pointer-events-none absolute -bottom-20 left-1/3 h-48 w-48 rounded-full bg-cyan-300/30 blur-3xl mix-blend-overlay"></div>
        <div class="pointer-events-none absolute -left-8 -bottom-10 h-40 w-40 rounded-full bg-white/15 blur-2xl mix-blend-overlay"></div>
        <div class="relative flex flex-wrap items-end justify-between gap-4 z-10">
          <div>
            <h1 class="text-3xl font-extrabold tracking-tight drop-shadow-sm">
              {{ greeting }}<template v-if="displayName">，{{ displayName }}</template> 👋
            </h1>
            <p class="mt-2 text-sm font-medium text-white/90">{{ todayLabel }} · {{ t('dashboard.welcomeMessage') }}</p>
          </div>
          <button @click="refreshAll" :disabled="loading || loadingCharts" class="btn btn-sm rounded-full border border-white/30 bg-white/20 px-4 py-2 text-sm font-medium text-white backdrop-blur-md transition-all hover:bg-white/30 hover:scale-105 active:scale-95" :title="t('common.refresh')">
            <Icon name="refresh" size="sm" :class="loadingCharts ? 'animate-spin' : ''" />
            {{ t('common.refresh') }}
          </button>
        </div>
      </div>

      <div v-if="loading" class="flex items-center justify-center py-12"><LoadingSpinner /></div>
      <template v-else-if="stats">
        <UserDashboardStats :stats="stats" :balance="user?.balance || 0" :is-simple="authStore.isSimpleMode" :platform-quotas="platformQuotas" />
        <UserDashboardCharts v-model:startDate="startDate" v-model:endDate="endDate" v-model:granularity="granularity" :loading="loadingCharts" :trend="trendData" :models="modelStats" @dateRangeChange="loadCharts" @granularityChange="loadCharts" @refresh="refreshAll" />
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div class="lg:col-span-2"><UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" /></div>
          <div class="lg:col-span-1"><UserDashboardQuickActions /></div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'
import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'
import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import type { UsageLog, TrendDataPoint, ModelStat, PlatformQuotaItem } from '@/types'
import { getMyPlatformQuotas } from '@/api/user'
import { formatDateLocalInput } from '@/utils/format'

const { t, locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const user = computed(() => authStore.user)
const displayName = computed(() => user.value?.username || user.value?.email?.split('@')[0] || '')
const greeting = computed(() => {
  const h = new Date().getHours()
  if (h >= 5 && h < 12) return t('dashboard.greetingMorning')
  if (h >= 12 && h < 18) return t('dashboard.greetingAfternoon')
  return t('dashboard.greetingEvening')
})
const todayLabel = computed(() =>
  new Date().toLocaleDateString(locale.value.startsWith('zh') ? 'zh-CN' : 'en-US', {
    month: 'long',
    day: 'numeric',
    weekday: 'long'
  })
)

const dashboardNotice = computed(() => (appStore.cachedPublicSettings?.dashboard_notice || '').trim())

const stats = ref<UserStatsType | null>(null)
const loading = ref(false)
const loadingUsage = ref(false)
const loadingCharts = ref(false)
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const recentUsage = ref<UsageLog[]>([])
const platformQuotas = ref<PlatformQuotaItem[] | null>(null)

const startDate = ref(formatDateLocalInput(new Date(Date.now() - 6 * 86400000)))
const endDate = ref(formatDateLocalInput(new Date()))
const granularity = ref('day')

const loadStats = async () => {
  loading.value = true
  try {
    await authStore.refreshUser()
    stats.value = await usageAPI.getDashboardStats()
  } catch (error) {
    console.error('Failed to load dashboard stats:', error)
  } finally {
    loading.value = false
  }
}
const loadCharts = async () => {
  loadingCharts.value = true
  try {
    const res = await Promise.all([
      usageAPI.getDashboardTrend({
        start_date: startDate.value,
        end_date: endDate.value,
        granularity: granularity.value as any
      }),
      usageAPI.getDashboardModels({ start_date: startDate.value, end_date: endDate.value })
    ])
    trendData.value = res[0].trend || []
    modelStats.value = res[1].models || []
  } catch (error) {
    console.error('Failed to load charts:', error)
  } finally {
    loadingCharts.value = false
  }
}
const loadRecent = async () => {
  loadingUsage.value = true
  try {
    const res = await usageAPI.getByDateRange(startDate.value, endDate.value)
    recentUsage.value = res.items.slice(0, 5)
  } catch (error) {
    console.error('Failed to load recent usage:', error)
  } finally {
    loadingUsage.value = false
  }
}
const loadPlatformQuotas = async () => {
  try {
    const data = await getMyPlatformQuotas()
    platformQuotas.value = data.platform_quotas ?? []
  } catch (error) {
    console.warn('Failed to load platform quotas:', error)
    platformQuotas.value = []
  }
}
const refreshAll = () => {
  loadStats()
  loadCharts()
  loadRecent()
  loadPlatformQuotas()
}

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    void appStore.fetchPublicSettings()
  }
  refreshAll()
})
</script>
