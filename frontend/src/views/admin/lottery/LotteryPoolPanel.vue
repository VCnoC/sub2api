<template>
  <div class="space-y-6">
    <div
      v-for="pool in forms"
      :key="pool.key"
      class="card p-6 bg-white dark:bg-dark-800 rounded-2xl border border-gray-200 dark:border-dark-700 shadow-sm"
    >
      <!-- Card Header -->
      <div class="flex flex-wrap items-center justify-between gap-4 pb-5 border-b border-gray-100 dark:border-dark-700/80">
        <div class="flex items-center gap-3">
          <div
            :class="[
              'flex h-10 w-10 items-center justify-center rounded-xl shadow-sm',
              pool.key === 'normal'
                ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400'
                : 'bg-purple-50 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400'
            ]"
          >
            <Icon :name="pool.key === 'normal' ? 'gift' : 'sparkles'" size="md" />
          </div>
          <div>
            <div class="flex items-center gap-2">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                {{ pool.key === 'normal' ? t('lottery.normal') : t('lottery.luxury') }}
              </h2>
              <span
                class="rounded-md bg-gray-100 dark:bg-dark-700 px-2 py-0.5 text-xs font-mono font-medium text-gray-500 dark:text-gray-400"
              >
                {{ pool.key }}
              </span>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <span
            :class="[
              'inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium',
              pool.enabled
                ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400'
                : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'
            ]"
          >
            <span
              :class="[
                'h-1.5 w-1.5 rounded-full',
                pool.enabled ? 'bg-emerald-500' : 'bg-gray-400'
              ]"
            ></span>
            {{ pool.enabled ? t('common.enabled') : t('common.disabled') }}
          </span>
          <Toggle v-model="pool.enabled" />
        </div>
      </div>

      <!-- Card Form Grid -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5 pt-5">
        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.name') }}
          </label>
          <input
            v-model.trim="pool.name"
            class="input w-full"
            maxlength="80"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.cycle') }}
          </label>
          <Select
            v-model="pool.cycle_type"
            :options="[
              { value: 'daily', label: t('lottery.daily') },
              { value: 'weekly', label: t('lottery.weekly') }
            ]"
            class="w-full"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.cycleChances') }}
          </label>
          <input
            v-model.number="pool.cycle_chances"
            class="input w-full"
            type="number"
            min="0"
            max="100"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.startsAt') }}
          </label>
          <input
            v-model="pool.starts_at"
            class="input w-full text-xs"
            type="datetime-local"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.endsAt') }}
          </label>
          <input
            v-model="pool.ends_at"
            class="input w-full text-xs"
            type="datetime-local"
          />
        </div>
      </div>

      <!-- Card Footer -->
      <div class="mt-6 flex justify-end border-t border-gray-100 dark:border-dark-700/80 pt-4">
        <button
          type="button"
          class="btn btn-primary"
          :disabled="savingKey === pool.key"
          @click="save(pool)"
        >
          <Icon
            :name="savingKey === pool.key ? 'refresh' : 'check'"
            size="sm"
            :class="{ 'animate-spin': savingKey === pool.key }"
            class="mr-1.5"
          />
          <span>{{ t('lotteryAdmin.save') }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import Toggle from '@/components/common/Toggle.vue'
import Select from '@/components/common/Select.vue'
import type { LotteryPool, LotteryPoolInput, LotteryPoolKey } from '@/types/lottery'

interface PoolForm {
  key: LotteryPoolKey
  name: string
  enabled: boolean
  cycle_type: 'daily' | 'weekly'
  cycle_chances: number
  starts_at: string
  ends_at: string
}

const props = defineProps<{ pools: LotteryPool[]; savingKey: LotteryPoolKey | '' }>()
const emit = defineEmits<{ save: [key: LotteryPoolKey, input: LotteryPoolInput] }>()
const { t } = useI18n()
const forms = ref<PoolForm[]>([])

function localDateTime(value?: string | null): string {
  if (!value) return ''
  const date = new Date(value)
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000)
  return local.toISOString().slice(0, 16)
}

watch(
  () => props.pools,
  (pools) => {
    forms.value = pools.map((pool) => ({
      key: pool.key,
      name: pool.name,
      enabled: pool.enabled,
      cycle_type: pool.cycle_type,
      cycle_chances: pool.cycle_chances,
      starts_at: localDateTime(pool.starts_at),
      ends_at: localDateTime(pool.ends_at),
    }))
  },
  { deep: true, immediate: true },
)

function save(form: PoolForm): void {
  emit('save', form.key, {
    name: form.name,
    enabled: form.enabled,
    cycle_type: form.cycle_type,
    cycle_chances: Number(form.cycle_chances),
    starts_at: form.starts_at ? new Date(form.starts_at).toISOString() : null,
    ends_at: form.ends_at ? new Date(form.ends_at).toISOString() : null,
  })
}
</script>
