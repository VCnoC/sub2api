<template>
  <AppLayout>
    <div class="lottery-page" :class="themeClass">
      <div class="lottery-page__glow" aria-hidden="true"></div>
      <div class="lottery-page__glow lottery-page__glow--secondary" aria-hidden="true"></div>

      <main class="lottery-page__main">
        <!-- Header -->
        <header class="lottery-hero">
          <div class="lottery-hero__copy">
            <p class="lottery-hero__eyebrow">
              <Icon name="sparkles" size="xs" />
              {{ activeKey === 'luxury' ? t('lottery.luxury') : t('lottery.normal') }}
              <span aria-hidden="true">·</span>
              Lucky Draw
            </p>
            <h1 class="lottery-hero__title">{{ t('lottery.title') }}</h1>
            <p class="lottery-hero__desc">{{ t('lottery.pageIntro') }}</p>
          </div>
          <RouterLink to="/affiliate" class="lottery-hero__invite">
            <Icon name="users" size="sm" />
            <span>{{ t('lottery.invite') }}</span>
          </RouterLink>
        </header>

        <!-- Loading -->
        <div v-if="loading" class="lottery-loading">
          <div class="lottery-loading__ring"></div>
          <p>{{ t('common.loading') }}</p>
        </div>

        <template v-else-if="activePool">
          <!-- Controls -->
          <section class="lottery-panel">
            <div class="lottery-panel__top">
              <nav class="lottery-tabs" aria-label="pool">
                <button
                  v-for="pool in summary?.pools ?? []"
                  :key="pool.pool.key"
                  type="button"
                  class="lottery-tabs__item"
                  :class="{ 'is-active': activeKey === pool.pool.key }"
                  @click="activeKey = pool.pool.key"
                >
                  <Icon :name="pool.pool.key === 'luxury' ? 'sparkles' : 'gift'" size="xs" />
                  {{ pool.pool.key === 'normal' ? t('lottery.normal') : t('lottery.luxury') }}
                </button>
              </nav>

              <div class="lottery-chances">
                <div class="lottery-chance">
                  <span class="lottery-chance__label">{{ t('lottery.baseChances') }}</span>
                  <strong class="lottery-chance__value">{{ activePool.base_remaining }}</strong>
                </div>
                <div class="lottery-chance lottery-chance--bonus">
                  <span class="lottery-chance__label">{{ t('lottery.extraChances') }}</span>
                  <strong class="lottery-chance__value">{{ activePool.extra_remaining }}</strong>
                </div>
              </div>
            </div>

            <div class="lottery-panel__meta">
              <div class="lottery-panel__pool">
                <span class="lottery-panel__dot"></span>
                <strong>{{ activePool.pool.name }}</strong>
                <span class="lottery-panel__chip">
                  {{ activePool.pool.cycle_type === 'daily' ? t('lottery.daily') : t('lottery.weekly') }}
                </span>
              </div>
              <dl v-if="activePool.pool.starts_at || activePool.pool.ends_at" class="lottery-panel__schedule">
                <div v-if="activePool.pool.starts_at">
                  <dt>{{ t('lottery.startsAt') }}</dt>
                  <dd>{{ formatDateTime(activePool.pool.starts_at) }}</dd>
                </div>
                <div v-if="activePool.pool.ends_at">
                  <dt>{{ t('lottery.endsAt') }}</dt>
                  <dd>{{ formatDateTime(activePool.pool.ends_at) }}</dd>
                </div>
              </dl>
            </div>
          </section>

          <!-- Reel -->
          <LotteryReel ref="reel" :prizes="activePool.prizes" :theme="activeKey" />

          <!-- Draw CTA -->
          <div class="lottery-cta">
            <button
              type="button"
              class="lottery-cta__btn"
              :disabled="!canDraw"
              @click="draw"
            >
              <Icon :name="drawing ? 'refresh' : 'sparkles'" size="md" :class="{ 'animate-spin': drawing }" />
              <span>
                {{ drawing ? t('lottery.drawing') : activePool.active ? t('lottery.start') : t('lottery.inactive') }}
              </span>
            </button>
            <p class="lottery-cta__hint">
              {{ t('lottery.remainingHint', { n: activePool.base_remaining + activePool.extra_remaining }) }}
            </p>
          </div>

          <!-- Result -->
          <transition name="lottery-fade">
            <div
              v-if="lastDraw"
              class="lottery-result"
              :class="lastDraw.outcome === 'win' ? 'is-win' : 'is-miss'"
              role="status"
            >
              <Icon :name="lastDraw.outcome === 'win' ? 'gift' : 'xCircle'" size="md" />
              <span>
                {{
                  lastDraw.outcome === 'win'
                    ? t('lottery.won', { name: lastDraw.prize?.name || '' })
                    : t('lottery.noPrize')
                }}
              </span>
            </div>
          </transition>

          <!-- Prize catalog -->
          <section class="lottery-section">
            <div class="lottery-section__head">
              <h2>{{ t('lottery.prizes') }}</h2>
              <span class="lottery-section__count">{{ activePool.prizes.length }}</span>
            </div>
            <div class="lottery-catalog">
              <LotteryPrizeCard
                v-for="prize in activePool.prizes"
                :key="prize.id"
                :prize="prize"
                show-probability
              />
            </div>
          </section>

          <!-- History -->
          <section class="lottery-section lottery-history">
            <div class="lottery-section__head">
              <h2>{{ t('lottery.history') }}</h2>
            </div>

            <div v-if="history.length === 0" class="lottery-history__empty">
              <Icon name="clock" size="lg" />
              <p>{{ t('lottery.noHistory') }}</p>
            </div>

            <ul v-else class="lottery-history__list">
              <li v-for="item in history" :key="item.id" class="lottery-history__item">
                <div
                  class="lottery-history__icon"
                  :class="item.outcome === 'win' ? 'is-win' : 'is-miss'"
                >
                  <Icon :name="item.outcome === 'win' ? 'gift' : 'xCircle'" size="sm" />
                </div>
                <div class="lottery-history__body">
                  <span class="lottery-history__name">
                    {{ item.outcome === 'win' ? item.prize?.name : t('lottery.noPrize') }}
                  </span>
                  <time>{{ formatDateTime(item.created_at) }}</time>
                </div>
              </li>
            </ul>
          </section>
        </template>
      </main>
    </div>
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
const themeClass = computed(() => (activeKey.value === 'luxury' ? 'is-gold' : 'is-teal'))
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
.lottery-page {
  /* Light defaults — tinted paper, not flat white */
  --lp-accent: #0d9488;
  --lp-accent-deep: #0f766e;
  --lp-accent-soft: rgba(13, 148, 136, 0.14);
  --lp-glow: rgba(13, 148, 136, 0.22);
  --lp-ink: #0f172a;
  --lp-muted: #64748b;
  --lp-surface: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
  --lp-surface-2: #f0fdfa;
  --lp-line: color-mix(in srgb, #0d9488 14%, #e2e8f0);
  --lp-hero-bg:
    radial-gradient(circle at 0% 0%, rgba(45, 212, 191, 0.2), transparent 48%),
    linear-gradient(145deg, #ecfeff 0%, #f0fdfa 45%, #ccfbf1 100%);
  --lp-hero-ink: #0f172a;
  --lp-hero-muted: #475569;
  --lp-tabs-bg: #e2e8f0;
  --lp-tabs-ink: #475569;
  --lp-invite-bg: #fffbeb;
  --lp-invite-ink: #b45309;
  --lp-invite-line: rgba(245, 158, 11, 0.35);

  position: relative;
  min-height: 100%;
  color: var(--lp-ink);
}

.lottery-page.is-gold {
  --lp-accent: #d97706;
  --lp-accent-deep: #b45309;
  --lp-accent-soft: rgba(217, 119, 6, 0.14);
  --lp-glow: rgba(217, 119, 6, 0.22);
  --lp-surface-2: #fffbeb;
  --lp-line: color-mix(in srgb, #d97706 16%, #e2e8f0);
  --lp-hero-bg:
    radial-gradient(circle at 0% 0%, rgba(251, 191, 36, 0.22), transparent 48%),
    linear-gradient(145deg, #fffbeb 0%, #fef3c7 45%, #fde68a 100%);
}

.dark .lottery-page {
  --lp-accent: #2dd4bf;
  --lp-accent-deep: #14b8a6;
  --lp-accent-soft: rgba(45, 212, 191, 0.16);
  --lp-glow: rgba(45, 212, 191, 0.28);
  --lp-ink: #f8fafc;
  --lp-muted: #94a3b8;
  --lp-surface: #111827;
  --lp-surface-2: #0f172a;
  --lp-line: #1e293b;
  --lp-hero-bg:
    radial-gradient(circle at 0% 0%, var(--lp-accent-soft), transparent 48%),
    linear-gradient(145deg, #1e293b 0%, #0f172a 55%, #020617 100%);
  --lp-hero-ink: #f8fafc;
  --lp-hero-muted: #cbd5e1;
  --lp-tabs-bg: #020617;
  --lp-tabs-ink: #94a3b8;
  --lp-invite-bg: rgba(245, 158, 11, 0.14);
  --lp-invite-ink: #fde68a;
  --lp-invite-line: rgba(251, 191, 36, 0.35);
}

.dark .lottery-page.is-gold {
  --lp-accent: #fbbf24;
  --lp-accent-deep: #f59e0b;
  --lp-accent-soft: rgba(251, 191, 36, 0.16);
  --lp-glow: rgba(245, 158, 11, 0.28);
  --lp-hero-bg:
    radial-gradient(circle at 0% 0%, var(--lp-accent-soft), transparent 48%),
    linear-gradient(145deg, #292524 0%, #1c1917 55%, #0c0a09 100%);
}

.lottery-page__glow {
  position: fixed;
  top: -10%;
  left: 15%;
  z-index: 0;
  width: 420px;
  height: 420px;
  border-radius: 50%;
  background: var(--lp-glow);
  filter: blur(120px);
  pointer-events: none;
  opacity: 0.45;
}

.lottery-page__glow--secondary {
  top: auto;
  right: 8%;
  bottom: 10%;
  left: auto;
  width: 360px;
  height: 360px;
  opacity: 0.25;
  background: rgba(99, 102, 241, 0.16);
}

.dark .lottery-page__glow {
  opacity: 0.55;
}

.dark .lottery-page__glow--secondary {
  opacity: 0.35;
  background: rgba(99, 102, 241, 0.22);
}

.lottery-page__main {
  position: relative;
  z-index: 1;
  display: flex;
  max-width: 960px;
  margin: 0 auto;
  flex-direction: column;
  gap: 22px;
  padding-bottom: 56px;
}

.lottery-hero {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  padding: 22px 24px;
  border: 1px solid var(--lp-line);
  border-radius: 24px;
  background: var(--lp-hero-bg);
  box-shadow:
    0 1px 0 rgba(255, 255, 255, 0.8) inset,
    0 14px 36px rgba(15, 23, 42, 0.08);
  color: var(--lp-hero-ink);
}

.dark .lottery-hero {
  box-shadow: 0 18px 40px rgba(2, 6, 23, 0.35);
}

.lottery-hero__eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin: 0 0 8px;
  color: var(--lp-accent-deep);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.dark .lottery-hero__eyebrow {
  color: var(--lp-accent);
}

.lottery-hero__title {
  margin: 0;
  font-size: clamp(1.6rem, 2.4vw, 2.1rem);
  font-weight: 800;
  letter-spacing: -0.03em;
  line-height: 1.15;
}

.lottery-hero__desc {
  max-width: 36rem;
  margin: 8px 0 0;
  color: var(--lp-hero-muted);
  font-size: 14px;
  line-height: 1.6;
}

.lottery-hero__invite {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  border: 1px solid var(--lp-invite-line);
  border-radius: 999px;
  background: var(--lp-invite-bg);
  padding: 10px 16px;
  color: var(--lp-invite-ink);
  font-size: 13px;
  font-weight: 700;
  transition: background 200ms ease, transform 200ms ease;
}

.lottery-hero__invite:hover {
  transform: translateY(-1px);
  filter: brightness(0.98);
}

.dark .lottery-hero__invite:hover {
  filter: brightness(1.12);
}

.lottery-loading {
  display: flex;
  min-height: 280px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  color: var(--lp-muted);
  font-size: 13px;
  font-weight: 600;
}

.lottery-loading__ring {
  width: 42px;
  height: 42px;
  border: 3px solid color-mix(in srgb, var(--lp-accent) 25%, transparent);
  border-top-color: var(--lp-accent);
  border-radius: 50%;
  animation: lottery-spin 0.9s linear infinite;
}

.lottery-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 18px 20px;
  border: 1px solid var(--lp-line);
  border-radius: 22px;
  background: var(--lp-surface);
  box-shadow:
    0 1px 0 rgba(255, 255, 255, 0.85) inset,
    0 12px 30px rgba(15, 23, 42, 0.07);
}

.lottery-panel__top {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
}

.lottery-tabs {
  display: inline-flex;
  gap: 4px;
  padding: 4px;
  border-radius: 14px;
  background: var(--lp-tabs-bg);
}

.lottery-tabs__item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border-radius: 10px;
  padding: 9px 14px;
  color: var(--lp-tabs-ink);
  font-size: 13px;
  font-weight: 700;
  transition: all 180ms ease;
}

.lottery-tabs__item.is-active {
  background: linear-gradient(135deg, var(--lp-accent-deep), var(--lp-accent));
  color: #fff;
  box-shadow: 0 8px 16px var(--lp-glow);
}

.lottery-chances {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.lottery-chance {
  display: flex;
  min-width: 118px;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border: 1px solid var(--lp-line);
  border-radius: 14px;
  background: var(--lp-surface-2);
  padding: 10px 14px;
}

.lottery-chance--bonus {
  border-color: rgba(245, 158, 11, 0.35);
  background: linear-gradient(135deg, #fffbeb, #fef3c7);
}

.dark .lottery-chance--bonus {
  background: linear-gradient(135deg, rgba(120, 53, 15, 0.55), rgba(69, 26, 3, 0.75));
  border-color: rgba(251, 191, 36, 0.28);
}

.lottery-chance__label {
  color: var(--lp-muted);
  font-size: 12px;
  font-weight: 650;
}

.lottery-chance__value {
  color: var(--lp-ink);
  font-size: 22px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.lottery-chance--bonus .lottery-chance__value {
  color: #b45309;
}

.dark .lottery-chance--bonus .lottery-chance__value {
  color: #fbbf24;
}

.lottery-panel__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-top: 1px solid var(--lp-line);
  padding-top: 14px;
}

.lottery-panel__pool {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: var(--lp-ink);
  font-size: 14px;
}

.lottery-panel__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--lp-accent);
  box-shadow: 0 0 0 4px var(--lp-accent-soft);
}

.lottery-panel__chip {
  border-radius: 999px;
  background: var(--lp-surface-2);
  padding: 3px 9px;
  color: var(--lp-muted);
  font-size: 11px;
  font-weight: 700;
  border: 1px solid var(--lp-line);
}

.lottery-panel__schedule {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  margin: 0;
}

.lottery-panel__schedule div {
  display: grid;
  gap: 2px;
}

.lottery-panel__schedule dt {
  color: var(--lp-muted);
  font-size: 11px;
}

.lottery-panel__schedule dd {
  margin: 0;
  color: var(--lp-ink);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  font-weight: 650;
}

.lottery-cta {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.lottery-cta__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-width: 220px;
  border: 0;
  border-radius: 999px;
  background: linear-gradient(135deg, var(--lp-accent-deep), var(--lp-accent));
  padding: 15px 34px;
  color: #fff;
  font-size: 16px;
  font-weight: 800;
  letter-spacing: 0.02em;
  box-shadow: 0 14px 30px var(--lp-glow);
  transition: transform 160ms ease, filter 160ms ease, opacity 160ms ease;
}

.lottery-cta__btn:hover:not(:disabled) {
  transform: translateY(-2px);
  filter: brightness(1.05);
}

.lottery-cta__btn:active:not(:disabled) {
  transform: translateY(0) scale(0.98);
}

.lottery-cta__btn:disabled {
  cursor: not-allowed;
  opacity: 0.45;
  box-shadow: none;
}

.lottery-cta__hint {
  margin: 0;
  color: var(--lp-muted);
  font-size: 12px;
  font-weight: 600;
}

.lottery-result {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin: 0 auto;
  max-width: 420px;
  border-radius: 16px;
  padding: 14px 18px;
  font-size: 15px;
  font-weight: 750;
}

.lottery-result.is-win {
  border: 1px solid rgba(245, 158, 11, 0.4);
  background: linear-gradient(135deg, #fffbeb, #fef3c7);
  color: #92400e;
  box-shadow: 0 12px 28px rgba(245, 158, 11, 0.18);
}

.lottery-result.is-miss {
  border: 1px solid var(--lp-line);
  background: var(--lp-surface-2);
  color: var(--lp-muted);
}

.lottery-section {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.lottery-section__head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.lottery-section__head h2 {
  margin: 0;
  color: var(--lp-ink);
  font-size: 17px;
  font-weight: 800;
  letter-spacing: -0.02em;
}

.lottery-section__count {
  border-radius: 999px;
  background: var(--lp-accent-soft);
  padding: 2px 9px;
  color: var(--lp-accent-deep);
  font-size: 12px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

.lottery-catalog {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  justify-items: center;
  gap: 14px;
}

.lottery-history {
  border: 1px solid var(--lp-line);
  border-radius: 22px;
  background: var(--lp-surface);
  padding: 18px 20px 10px;
  box-shadow:
    0 1px 0 rgba(255, 255, 255, 0.85) inset,
    0 12px 30px rgba(15, 23, 42, 0.07);
}

.lottery-history__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 36px 12px;
  color: var(--lp-muted);
  font-size: 13px;
  font-weight: 600;
}

.lottery-history__list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.lottery-history__item {
  display: flex;
  align-items: center;
  gap: 12px;
  border-top: 1px solid var(--lp-line);
  padding: 12px 2px;
}

.lottery-history__icon {
  display: flex;
  width: 36px;
  height: 36px;
  flex: none;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
}

.lottery-history__icon.is-win {
  background: #fef3c7;
  color: #d97706;
}

.dark .lottery-history__icon.is-win {
  background: rgba(120, 53, 15, 0.45);
  color: #fbbf24;
}

.lottery-history__icon.is-miss {
  background: var(--lp-surface-2);
  color: var(--lp-muted);
  border: 1px solid var(--lp-line);
}

.lottery-history__body {
  display: flex;
  min-width: 0;
  flex: 1;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.lottery-history__name {
  overflow: hidden;
  color: var(--lp-ink);
  font-size: 14px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.lottery-history__body time {
  flex: none;
  color: var(--lp-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 11px;
}

.lottery-fade-enter-active,
.lottery-fade-leave-active {
  transition: all 280ms ease;
}

.lottery-fade-enter-from,
.lottery-fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

@keyframes lottery-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 640px) {
  .lottery-hero,
  .lottery-panel,
  .lottery-history {
    padding-left: 16px;
    padding-right: 16px;
  }

  .lottery-cta__btn {
    width: 100%;
  }
}
</style>
