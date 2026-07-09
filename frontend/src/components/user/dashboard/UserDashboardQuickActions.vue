<template>
  <div class="spotlight-card rounded-3xl bg-white/60 dark:bg-dark-900/60 backdrop-blur-2xl border border-white/40 dark:border-white/10 shadow-[0_8px_32px_rgba(0,0,0,0.04)] dark:shadow-[0_8px_32px_rgba(0,0,0,0.16)] flex flex-col h-full">
    <div class="flex items-center gap-3 border-b border-white/20 px-6 py-5 dark:border-white/5">
      <div class="flex h-8 w-8 items-center justify-center rounded-xl bg-primary-500/10 text-primary-600 dark:text-primary-400">
        <Icon name="play" size="sm" />
      </div>
      <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('dashboard.quickActions') }}</h2>
    </div>
    <div class="space-y-3 p-6 flex-1 flex flex-col justify-center">
      <button @click="router.push('/keys')" class="group flex w-full items-center gap-4 rounded-2xl bg-white/40 p-4 text-left transition-all duration-300 hover:bg-white/80 hover:shadow-sm hover:-translate-y-0.5 dark:bg-dark-800/40 dark:hover:bg-dark-800/80 border border-transparent hover:border-white/60 dark:hover:border-white/10">
        <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-primary-400 to-cyan-500 text-white shadow-md shadow-primary-500/20 transition-transform group-hover:scale-110 group-hover:rotate-3">
          <Icon name="key" size="lg" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-bold text-gray-900 dark:text-white">{{ t('dashboard.createApiKey') }}</p>
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mt-0.5">{{ t('dashboard.generateNewKey') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-gray-400 transition-colors group-hover:text-primary-500 dark:text-dark-500"
        />
      </button>

      <button @click="router.push('/usage')" class="group flex w-full items-center gap-4 rounded-2xl bg-white/40 p-4 text-left transition-all duration-300 hover:bg-white/80 hover:shadow-sm hover:-translate-y-0.5 dark:bg-dark-800/40 dark:hover:bg-dark-800/80 border border-transparent hover:border-white/60 dark:hover:border-white/10">
        <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-emerald-400 to-teal-500 text-white shadow-md shadow-emerald-500/20 transition-transform group-hover:scale-110 group-hover:-rotate-3">
          <Icon name="chart" size="lg" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-bold text-gray-900 dark:text-white">{{ t('dashboard.viewUsage') }}</p>
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mt-0.5">{{ t('dashboard.checkDetailedLogs') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-gray-400 transition-colors group-hover:text-emerald-500 dark:text-dark-500"
        />
      </button>

      <button v-if="canUseBatchImage" @click="router.push('/batch-image')" class="group flex w-full items-center gap-4 rounded-2xl bg-white/40 p-4 text-left transition-all duration-300 hover:bg-white/80 hover:shadow-sm hover:-translate-y-0.5 dark:bg-dark-800/40 dark:hover:bg-dark-800/80 border border-transparent hover:border-white/60 dark:hover:border-white/10">
        <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-sky-400 to-blue-500 text-white shadow-md shadow-sky-500/20 transition-transform group-hover:scale-110 group-hover:-rotate-3">
          <Icon name="sparkles" size="lg" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-bold text-gray-900 dark:text-white">{{ t('dashboard.batchImageAgent') }}</p>
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mt-0.5">{{ t('dashboard.batchImageAgentDesc') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-gray-400 transition-colors group-hover:text-sky-500 dark:text-dark-500"
        />
      </button>

      <button @click="router.push('/redeem')" class="group flex w-full items-center gap-4 rounded-2xl bg-white/40 p-4 text-left transition-all duration-300 hover:bg-white/80 hover:shadow-sm hover:-translate-y-0.5 dark:bg-dark-800/40 dark:hover:bg-dark-800/80 border border-transparent hover:border-white/60 dark:hover:border-white/10">
        <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-amber-400 to-orange-500 text-white shadow-md shadow-amber-500/20 transition-transform group-hover:scale-110 group-hover:rotate-3">
          <Icon name="gift" size="lg" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-bold text-gray-900 dark:text-white">{{ t('dashboard.redeemCode') }}</p>
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mt-0.5">{{ t('dashboard.addBalanceWithCode') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-gray-400 transition-colors group-hover:text-amber-500 dark:text-dark-500"
        />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useBatchImageAccess } from '@/composables/useBatchImageAccess'
const router = useRouter()
const { t } = useI18n()
const { canUseBatchImage, refreshBatchImageAccess } = useBatchImageAccess()

onMounted(() => {
  void refreshBatchImageAccess()
})
</script>
