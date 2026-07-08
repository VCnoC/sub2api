<template>
  <div class="spotlight-card rounded-3xl bg-white/60 dark:bg-dark-900/60 backdrop-blur-2xl border border-white/40 dark:border-white/10 shadow-[0_8px_32px_rgba(0,0,0,0.04)] dark:shadow-[0_8px_32px_rgba(0,0,0,0.16)] flex flex-col h-full">
    <div class="flex items-center justify-between border-b border-white/20 px-6 py-5 dark:border-white/5">
      <div class="flex items-center gap-3">
        <div class="flex h-8 w-8 items-center justify-center rounded-xl bg-primary-500/10 text-primary-600 dark:text-primary-400">
          <Icon name="clock" size="sm" />
        </div>
        <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('dashboard.recentUsage') }}</h2>
      </div>
      <span class="rounded-full bg-gray-100/80 px-3 py-1 text-xs font-medium text-gray-600 dark:bg-dark-800/80 dark:text-gray-400">{{ t('dashboard.last7Days') }}</span>
    </div>
    <div class="p-6 flex-1 flex flex-col">
      <div v-if="loading" class="flex flex-1 items-center justify-center py-12">
        <LoadingSpinner size="lg" />
      </div>
      <div v-else-if="data.length === 0" class="flex-1 py-8 flex items-center justify-center">
        <EmptyState :title="t('dashboard.noUsageRecords')" :description="t('dashboard.startUsingApi')" />
      </div>
      <div v-else class="space-y-3 flex-1">
        <div v-for="log in data" :key="log.id" class="group flex items-center justify-between rounded-2xl bg-white/40 p-4 transition-all hover:bg-white/80 hover:shadow-sm dark:bg-dark-800/40 dark:hover:bg-dark-800/80 border border-transparent hover:border-white/60 dark:hover:border-white/10">
          <div class="flex items-center gap-4">
            <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-gray-100 to-gray-200 dark:from-dark-700 dark:to-dark-800 shadow-inner group-hover:scale-110 transition-transform duration-300">
              <Icon name="beaker" size="md" class="text-gray-600 dark:text-gray-400" />
            </div>
            <div>
              <p class="text-sm font-bold text-gray-900 dark:text-white">{{ log.model }}</p>
              <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mt-0.5">{{ formatDateTime(log.created_at) }}</p>
            </div>
          </div>
          <div class="text-right">
            <p class="text-sm font-bold tabular-nums">
              <span class="text-green-600 dark:text-green-400" :title="t('dashboard.actual')">${{ formatCost(log.actual_cost) }}</span>
              <span class="font-normal text-gray-400 dark:text-gray-500" :title="t('dashboard.standard')"> / ${{ formatCost(log.total_cost) }}</span>
            </p>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mt-0.5">{{ (log.input_tokens + log.output_tokens).toLocaleString() }} tokens</p>
          </div>
        </div>

        <router-link to="/usage" class="mt-4 flex items-center justify-center gap-2 rounded-xl bg-gray-50/50 py-3 text-sm font-bold text-gray-600 transition-all hover:bg-gray-100 hover:text-gray-900 dark:bg-dark-800/50 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-white">
          {{ t('dashboard.viewAllUsage') }}
          <Icon name="arrowRight" size="sm" class="transition-transform group-hover:translate-x-1" />
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import type { UsageLog } from '@/types'

defineProps<{
  data: UsageLog[]
  loading: boolean
}>()
const { t } = useI18n()
const formatCost = (c: number) => c.toFixed(4)
</script>
