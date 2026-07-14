<template>
  <div class="lottery_pool_panel">
    <section v-for="pool in forms" :key="pool.key" class="lottery_pool_panel__section">
      <header>
        <div>
          <h2>{{ pool.key === 'normal' ? t('lottery.normal') : t('lottery.luxury') }}</h2>
          <span>{{ pool.key }}</span>
        </div>
        <label class="lottery_pool_panel__toggle">
          <input v-model="pool.enabled" type="checkbox" />
          <span>{{ t('lotteryAdmin.enabled') }}</span>
        </label>
      </header>

      <div class="lottery_pool_panel__grid">
        <label>
          <span>{{ t('lotteryAdmin.name') }}</span>
          <input v-model.trim="pool.name" class="input" maxlength="80" />
        </label>
        <label>
          <span>{{ t('lotteryAdmin.cycle') }}</span>
          <select v-model="pool.cycle_type" class="input">
            <option value="daily">{{ t('lottery.daily') }}</option>
            <option value="weekly">{{ t('lottery.weekly') }}</option>
          </select>
        </label>
        <label>
          <span>{{ t('lotteryAdmin.cycleChances') }}</span>
          <input v-model.number="pool.cycle_chances" class="input" type="number" min="0" max="100" />
        </label>
        <label>
          <span>{{ t('lotteryAdmin.startsAt') }}</span>
          <input v-model="pool.starts_at" class="input" type="datetime-local" />
        </label>
        <label>
          <span>{{ t('lotteryAdmin.endsAt') }}</span>
          <input v-model="pool.ends_at" class="input" type="datetime-local" />
        </label>
      </div>

      <footer>
        <button type="button" class="btn btn-primary" :disabled="savingKey === pool.key" @click="save(pool)">
          <Icon :name="savingKey === pool.key ? 'refresh' : 'check'" size="sm" :class="{ 'animate-spin': savingKey === pool.key }" />
          <span>{{ t('lotteryAdmin.save') }}</span>
        </button>
      </footer>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
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

<style scoped>
.lottery_pool_panel {
  display: grid;
  gap: 20px;
}

.lottery_pool_panel__section {
  border-top: 3px solid rgb(232 93 74);
  border-bottom: 1px solid rgb(229 231 235);
  padding: 18px 0 20px;
}

.lottery_pool_panel__section:nth-child(2) {
  border-top-color: rgb(73 92 125);
}

.lottery_pool_panel__section header,
.lottery_pool_panel__section footer,
.lottery_pool_panel__toggle {
  display: flex;
  align-items: center;
}

.lottery_pool_panel__section header {
  justify-content: space-between;
  gap: 16px;
}

.lottery_pool_panel__section h2 {
  color: rgb(17 24 39);
  font-size: 17px;
  font-weight: 680;
}

.lottery_pool_panel__section header span {
  color: rgb(107 114 128);
  font-size: 11px;
}

.lottery_pool_panel__toggle {
  gap: 8px;
  font-size: 13px;
}

.lottery_pool_panel__grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
  margin-top: 18px;
}

.lottery_pool_panel__grid label {
  display: grid;
  gap: 6px;
}

.lottery_pool_panel__grid label > span {
  color: rgb(75 85 99);
  font-size: 12px;
  font-weight: 600;
}

.lottery_pool_panel__section footer {
  justify-content: flex-end;
  margin-top: 16px;
}

:global(.dark) .lottery_pool_panel__section {
  border-bottom-color: rgb(63 63 70);
}

:global(.dark) .lottery_pool_panel__section h2 {
  color: rgb(244 244 245);
}

:global(.dark) .lottery_pool_panel__grid label > span {
  color: rgb(161 161 170);
}

@media (max-width: 860px) {
  .lottery_pool_panel__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 560px) {
  .lottery_pool_panel__grid {
    grid-template-columns: 1fr;
  }
}
</style>
