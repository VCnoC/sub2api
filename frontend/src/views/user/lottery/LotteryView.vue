<template>
  <AppLayout>
    <main class="lottery_page">
      <header class="lottery_page__header">
        <div>
          <h1>{{ t('lottery.title') }}</h1>
          <p>{{ t('lottery.description') }}</p>
        </div>
        <RouterLink to="/affiliate" class="btn btn-secondary">
          <Icon name="users" size="sm" />
          <span>{{ t('lottery.invite') }}</span>
        </RouterLink>
      </header>

      <div v-if="loading" class="lottery_page__loading">
        <Icon name="refresh" class="animate-spin" />
      </div>

      <template v-else-if="activePool">
        <nav class="lottery_page__segments" :aria-label="t('lottery.title')">
          <button
            v-for="pool in summary?.pools ?? []"
            :key="pool.pool.key"
            type="button"
            :class="{ 'lottery_page__segment--active': activeKey === pool.pool.key }"
            @click="activeKey = pool.pool.key"
          >
            {{ pool.pool.key === 'normal' ? t('lottery.normal') : t('lottery.luxury') }}
          </button>
        </nav>

        <section class="lottery_page__status" aria-live="polite">
          <div class="lottery_page__pool_name">
            <span class="lottery_page__pool_marker" :class="`lottery_page__pool_marker--${activeKey}`"></span>
            <strong>{{ activePool.pool.name }}</strong>
            <span>{{ activePool.pool.cycle_type === 'daily' ? t('lottery.daily') : t('lottery.weekly') }}</span>
          </div>
          <div class="lottery_page__chance_group">
            <div>
              <span>{{ t('lottery.baseChances') }}</span>
              <strong>{{ activePool.base_remaining }}</strong>
            </div>
            <div>
              <span>{{ t('lottery.extraChances') }}</span>
              <strong>{{ activePool.extra_remaining }}</strong>
            </div>
          </div>
        </section>

        <dl
          v-if="activePool.pool.starts_at || activePool.pool.ends_at"
          class="lottery_page__schedule"
        >
          <div v-if="activePool.pool.starts_at">
            <dt>{{ t('lottery.startsAt') }}</dt>
            <dd><time :datetime="activePool.pool.starts_at">{{ formatDateTime(activePool.pool.starts_at) }}</time></dd>
          </div>
          <div v-if="activePool.pool.ends_at">
            <dt>{{ t('lottery.endsAt') }}</dt>
            <dd><time :datetime="activePool.pool.ends_at">{{ formatDateTime(activePool.pool.ends_at) }}</time></dd>
          </div>
        </dl>

        <LotteryReel ref="reel" :prizes="activePool.prizes" />

        <div class="lottery_page__action">
          <button
            type="button"
            class="lottery_page__draw_button"
            :disabled="!canDraw"
            @click="draw"
          >
            <Icon :name="drawing ? 'refresh' : 'sparkles'" :class="{ 'animate-spin': drawing }" />
            <span>{{ drawing ? t('lottery.drawing') : activePool.active ? t('lottery.start') : t('lottery.inactive') }}</span>
          </button>
        </div>

        <div v-if="lastDraw" class="lottery_page__result" role="status">
          <Icon :name="lastDraw.outcome === 'win' ? 'gift' : 'refresh'" />
          <strong>
            {{ lastDraw.outcome === 'win'
              ? t('lottery.won', { name: lastDraw.prize?.name || '' })
              : t('lottery.noPrize') }}
          </strong>
        </div>

        <section class="lottery_page__prizes">
          <div class="lottery_page__section_heading">
            <h2>{{ t('lottery.prizes') }}</h2>
            <span>{{ activePool.prizes.length }}</span>
          </div>
          <div class="lottery_page__prize_grid">
            <LotteryPrizeCard
              v-for="prize in activePool.prizes"
              :key="prize.id"
              :prize="prize"
              show-probability
            />
          </div>
        </section>

        <section class="lottery_page__history">
          <div class="lottery_page__section_heading">
            <h2>{{ t('lottery.history') }}</h2>
          </div>
          <div v-if="history.length === 0" class="lottery_page__empty">{{ t('lottery.noHistory') }}</div>
          <div v-else class="lottery_page__history_list">
            <div v-for="item in history" :key="item.id" class="lottery_page__history_row">
              <Icon :name="item.outcome === 'win' ? 'gift' : 'refresh'" size="sm" />
              <span>{{ item.outcome === 'win' ? item.prize?.name : t('lottery.noPrize') }}</span>
              <time>{{ formatDateTime(item.created_at) }}</time>
            </div>
          </div>
        </section>
      </template>
    </main>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import lotteryAPI from '@/api/lottery'
import type { LotteryDraw, LotteryPoolKey, LotterySummary } from '@/types/lottery'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'
import LotteryPrizeCard from './components/LotteryPrizeCard.vue'
import LotteryReel from './components/LotteryReel.vue'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(true)
const drawing = ref(false)
const summary = ref<LotterySummary | null>(null)
const activeKey = ref<LotteryPoolKey>('normal')
const history = ref<LotteryDraw[]>([])
const lastDraw = ref<LotteryDraw | null>(null)
const reel = ref<InstanceType<typeof LotteryReel> | null>(null)

const activePool = computed(() => summary.value?.pools.find((item) => item.pool.key === activeKey.value) ?? null)
const canDraw = computed(() => {
  const pool = activePool.value
  return !!pool && pool.active && !drawing.value && pool.base_remaining + pool.extra_remaining > 0
})

async function load(): Promise<void> {
  loading.value = true
  try {
    const [summaryResult, historyResult] = await Promise.all([
      lotteryAPI.summary(),
      lotteryAPI.history(1, 20),
    ])
    summary.value = summaryResult
    history.value = historyResult.items ?? []
    if (!summaryResult.pools.some((item) => item.pool.key === activeKey.value)) {
      activeKey.value = summaryResult.pools[0]?.pool.key ?? 'normal'
    }
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('lottery.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function draw(): Promise<void> {
  if (!canDraw.value || !activePool.value) return
  drawing.value = true
  lastDraw.value = null
  try {
    const result = await lotteryAPI.draw(activeKey.value)
    await reel.value?.play(result)
    lastDraw.value = result
    activePool.value.base_remaining = result.base_remaining ?? activePool.value.base_remaining
    activePool.value.extra_remaining = result.extra_remaining ?? activePool.value.extra_remaining
    history.value = [result, ...history.value].slice(0, 20)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('lottery.drawFailed')))
  } finally {
    drawing.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.lottery_page {
  --lottery-accent: rgb(232 93 74);
  --lottery-premium: rgb(73 92 125);
  --lottery-success: rgb(15 138 120);
  display: flex;
  max-width: 1120px;
  margin: 0 auto;
  flex-direction: column;
  gap: 24px;
  padding-bottom: 48px;
}

.lottery_page__header,
.lottery_page__status,
.lottery_page__chance_group,
.lottery_page__pool_name,
.lottery_page__section_heading,
.lottery_page__history_row {
  display: flex;
  align-items: center;
}

.lottery_page__header {
  justify-content: space-between;
  gap: 16px;
}

.lottery_page__header h1 {
  color: rgb(17 24 39);
  font-size: 24px;
  font-weight: 720;
}

.lottery_page__header p {
  margin-top: 2px;
  color: rgb(107 114 128);
  font-size: 13px;
}

.lottery_page__loading {
  display: flex;
  min-height: 360px;
  align-items: center;
  justify-content: center;
  color: var(--lottery-success);
}

.lottery_page__segments {
  display: grid;
  width: min(360px, 100%);
  grid-template-columns: repeat(2, minmax(0, 1fr));
  padding: 3px;
  border: 1px solid rgb(229 231 235);
  border-radius: 8px;
  background: rgb(243 244 246);
}

.lottery_page__segments button {
  min-height: 36px;
  border-radius: 6px;
  color: rgb(107 114 128);
  font-size: 14px;
  font-weight: 600;
}

.lottery_page__segments .lottery_page__segment--active {
  background: rgb(255 255 255);
  color: rgb(17 24 39);
  box-shadow: 0 1px 4px rgb(15 23 42 / 10%);
}

.lottery_page__status {
  justify-content: space-between;
  gap: 20px;
  padding-block: 2px;
}

.lottery_page__pool_name {
  gap: 9px;
  min-width: 0;
}

.lottery_page__pool_name strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.lottery_page__pool_name > span:last-child {
  color: rgb(107 114 128);
  font-size: 12px;
}

.lottery_page__pool_marker {
  width: 4px;
  height: 28px;
  flex: none;
  border-radius: 2px;
  background: var(--lottery-accent);
}

.lottery_page__pool_marker--luxury {
  background: var(--lottery-premium);
}

.lottery_page__chance_group {
  gap: 24px;
}

.lottery_page__chance_group > div {
  display: grid;
  grid-template-columns: auto auto;
  align-items: baseline;
  gap: 8px;
}

.lottery_page__chance_group span {
  color: rgb(107 114 128);
  font-size: 12px;
}

.lottery_page__chance_group strong {
  color: var(--lottery-success);
  font-size: 24px;
  font-variant-numeric: tabular-nums;
}

.lottery_page__schedule {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 24px;
  color: rgb(107 114 128);
  font-size: 12px;
}

.lottery_page__schedule > div {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.lottery_page__schedule dd {
  color: rgb(55 65 81);
  font-variant-numeric: tabular-nums;
}

.lottery_page__action {
  display: flex;
  justify-content: center;
}

.lottery_page__draw_button {
  display: inline-flex;
  min-width: 188px;
  min-height: 46px;
  align-items: center;
  justify-content: center;
  gap: 9px;
  border-radius: 8px;
  background: var(--lottery-accent);
  color: white;
  font-size: 15px;
  font-weight: 700;
  box-shadow: 0 10px 22px rgb(232 93 74 / 22%);
}

.lottery_page__draw_button:disabled {
  cursor: not-allowed;
  background: rgb(156 163 175);
  box-shadow: none;
}

.lottery_page__result {
  display: flex;
  min-height: 48px;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-block: 1px solid rgb(232 93 74 / 35%);
  color: rgb(185 54 41);
}

.lottery_page__prizes,
.lottery_page__history {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.lottery_page__section_heading {
  gap: 8px;
}

.lottery_page__section_heading h2 {
  color: rgb(17 24 39);
  font-size: 16px;
  font-weight: 680;
}

.lottery_page__section_heading span {
  color: rgb(107 114 128);
  font-size: 12px;
}

.lottery_page__prize_grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(148px, 1fr));
  gap: 12px;
}

.lottery_page__prize_grid :deep(.lottery_prize_card) {
  width: 100%;
}

.lottery_page__history_list {
  border-block: 1px solid rgb(229 231 235);
}

.lottery_page__history_row {
  min-height: 48px;
  gap: 10px;
  border-bottom: 1px solid rgb(243 244 246);
  color: rgb(55 65 81);
  font-size: 13px;
}

.lottery_page__history_row:last-child {
  border-bottom: 0;
}

.lottery_page__history_row span {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.lottery_page__history_row time,
.lottery_page__empty {
  color: rgb(107 114 128);
  font-size: 12px;
}

.lottery_page__empty {
  padding: 28px 0;
  border-block: 1px dashed rgb(209 213 219);
  text-align: center;
}

:global(.dark) .lottery_page__header h1,
:global(.dark) .lottery_page__pool_name strong,
:global(.dark) .lottery_page__section_heading h2 {
  color: rgb(244 244 245);
}

:global(.dark) .lottery_page__segments {
  border-color: rgb(63 63 70);
  background: rgb(24 24 27);
}

:global(.dark) .lottery_page__segments .lottery_page__segment--active {
  background: rgb(63 63 70);
  color: white;
}

:global(.dark) .lottery_page__history_list,
:global(.dark) .lottery_page__history_row {
  border-color: rgb(63 63 70);
}

:global(.dark) .lottery_page__history_row {
  color: rgb(212 212 216);
}

:global(.dark) .lottery_page__schedule dd {
  color: rgb(212 212 216);
}

@media (max-width: 640px) {
  .lottery_page {
    gap: 18px;
  }

  .lottery_page__header,
  .lottery_page__status {
    align-items: flex-start;
  }

  .lottery_page__status {
    flex-direction: column;
  }

  .lottery_page__chance_group {
    width: 100%;
    justify-content: space-between;
  }

  .lottery_page__prize_grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .lottery_page__history_row time {
    max-width: 112px;
    text-align: right;
  }
}
</style>
