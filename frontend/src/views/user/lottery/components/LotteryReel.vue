<template>
  <div class="reel" :class="[`reel--${theme}`, { 'is-running': running }]">
    <div v-if="running" class="reel__veil lottery_reel__veil" aria-hidden="true"></div>

    <div class="reel__cabinet">
      <div class="reel__marquee">
        <div class="reel__leds" aria-hidden="true">
          <span v-for="i in 6" :key="`l-${i}`" class="reel__led" :style="{ animationDelay: `${i * 80}ms` }"></span>
        </div>
        <p class="reel__title">
          <Icon name="sparkles" size="xs" />
          <span>{{ t('lottery.reelTitle') }}</span>
          <Icon name="sparkles" size="xs" />
        </p>
        <div class="reel__leds" aria-hidden="true">
          <span v-for="i in 6" :key="`r-${i}`" class="reel__led" :style="{ animationDelay: `${i * 80}ms` }"></span>
        </div>
      </div>

      <div class="reel__stage">
        <div class="reel__bezel reel__bezel--top" aria-hidden="true"></div>
        <div class="reel__bezel reel__bezel--bottom" aria-hidden="true"></div>
        <div class="reel__fade reel__fade--left" aria-hidden="true"></div>
        <div class="reel__fade reel__fade--right" aria-hidden="true"></div>

        <div class="reel__pointer" aria-hidden="true">
          <span class="reel__crown"></span>
          <span class="reel__beam"></span>
        </div>

        <div ref="viewport" class="reel__viewport">
          <div ref="track" class="reel__track" :style="trackStyle">
            <LotteryPrizeCard
              v-for="(item, index) in reelItems"
              :key="`${item.key}-${index}`"
              :prize="item.prize"
              :none="item.none"
              :winner="index === winnerIndex && !running"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { LotteryDraw, LotteryPoolKey, LotteryPrize } from '@/types/lottery'
import LotteryPrizeCard from './LotteryPrizeCard.vue'

interface ReelItem {
  key: string
  prize: LotteryPrize | null
  none: boolean
}

const props = withDefaults(
  defineProps<{
    prizes: LotteryPrize[]
    theme?: LotteryPoolKey
  }>(),
  { theme: 'normal' },
)

const { t } = useI18n()
const viewport = ref<HTMLElement | null>(null)
const track = ref<HTMLElement | null>(null)
const running = ref(false)
const translateX = ref(0)
const durationMs = ref(0)
const winnerIndex = ref(-1)
const reelItems = ref<ReelItem[]>([])

const trackStyle = computed(() => ({
  transform: `translate3d(${-translateX.value}px, 0, 0)`,
  transitionDuration: `${durationMs.value}ms`,
}))

function baseItems(): ReelItem[] {
  const prizes = props.prizes.map((prize) => ({ key: `prize-${prize.id}`, prize, none: false }))
  return [...prizes, { key: 'none', prize: null, none: true }]
}

function resultOffset(draw: LotteryDraw, items: ReelItem[]): number {
  if (draw.outcome === 'none') return items.findIndex((item) => item.none)
  const resultId = Number(draw.prize?.id ?? draw.prize_id ?? 0)
  const byId = items.findIndex((item) => item.prize?.id === resultId)
  if (byId >= 0) return byId
  const byName = items.findIndex((item) => item.prize?.name === draw.prize?.name)
  return byName >= 0 ? byName : items.findIndex((item) => item.none)
}

async function play(draw: LotteryDraw): Promise<void> {
  const base = baseItems()
  const cycles = Math.max(7, Math.ceil(28 / base.length))
  reelItems.value = Array.from({ length: cycles }, () => base).flat()
  const target = resultOffset(draw, base)
  winnerIndex.value = (cycles - 2) * base.length + Math.max(0, target)
  durationMs.value = 0
  translateX.value = 0
  running.value = true
  await nextTick()

  const card = track.value?.children.item(winnerIndex.value) as HTMLElement | null
  const cardCenter = card ? card.offsetLeft + card.offsetWidth / 2 : winnerIndex.value * 166 + 75
  const viewportCenter = (viewport.value?.clientWidth ?? 320) / 2
  const targetTranslate = Math.max(0, cardCenter - viewportCenter)
  const reduced = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false
  if (reduced) {
    translateX.value = targetTranslate
    running.value = false
    return
  }

  durationMs.value = 4200
  await nextTick()
  translateX.value = targetTranslate

  await new Promise<void>((resolve) => {
    let settled = false
    const finish = () => {
      if (settled) return
      settled = true
      track.value?.removeEventListener('transitionend', finish)
      resolve()
    }
    track.value?.addEventListener('transitionend', finish, { once: true })
    window.setTimeout(finish, 4600)
  })
  running.value = false
}

watch(
  () => props.prizes,
  () => {
    if (running.value) return
    reelItems.value = Array.from({ length: 3 }, () => baseItems()).flat()
    translateX.value = 0
    durationMs.value = 0
    winnerIndex.value = -1
  },
  { deep: true, immediate: true },
)

defineExpose({ play })
</script>

<style scoped>
.reel {
  --reel-accent: #0d9488;
  --reel-frame:
    linear-gradient(180deg, #f0fdfa 0%, #e2e8f0 42%, #cbd5e1 100%);
  --reel-frame-line: color-mix(in srgb, #0d9488 35%, #94a3b8);
  --reel-marquee: linear-gradient(180deg, #ffffff 0%, #f0fdfa 100%);
  --reel-marquee-line: rgba(13, 148, 136, 0.22);
  --reel-stage:
    radial-gradient(ellipse at 50% 0%, rgba(45, 212, 191, 0.18), transparent 48%),
    linear-gradient(180deg, #134e4a 0%, #0f172a 48%, #020617 100%);
  --reel-fade: #0f172a;
  --reel-title: #0f766e;
  --reel-shadow:
    0 1px 0 rgba(255, 255, 255, 0.75) inset,
    0 18px 44px rgba(15, 23, 42, 0.16),
    0 4px 12px rgba(13, 148, 136, 0.12);
  position: relative;
  width: 100%;
  user-select: none;
}

.reel--luxury {
  --reel-accent: #d97706;
  --reel-frame:
    linear-gradient(180deg, #fffbeb 0%, #fde68a 42%, #fbbf24 100%);
  --reel-frame-line: color-mix(in srgb, #d97706 40%, #a8a29e);
  --reel-marquee: linear-gradient(180deg, #ffffff 0%, #fffbeb 100%);
  --reel-marquee-line: rgba(217, 119, 6, 0.28);
  --reel-stage:
    radial-gradient(ellipse at 50% 0%, rgba(251, 191, 36, 0.2), transparent 48%),
    linear-gradient(180deg, #78350f 0%, #1c1917 48%, #0c0a09 100%);
  --reel-fade: #1c1917;
  --reel-title: #b45309;
  --reel-shadow:
    0 1px 0 rgba(255, 255, 255, 0.75) inset,
    0 18px 44px rgba(120, 53, 15, 0.18),
    0 4px 12px rgba(217, 119, 6, 0.14);
}

/*
  Dark overrides: `.dark .x` — do NOT use `:global(.dark) .x`
  (Vue scoped drops the descendant selector).
*/
.dark .reel {
  --reel-accent: #2dd4bf;
  --reel-frame: linear-gradient(180deg, #1e293b 0%, #0f172a 42%, #020617 100%);
  --reel-frame-line: color-mix(in srgb, #2dd4bf 35%, #1e293b);
  --reel-marquee: rgba(2, 6, 23, 0.55);
  --reel-marquee-line: rgba(255, 255, 255, 0.06);
  --reel-stage:
    radial-gradient(circle at 50% 0%, rgba(45, 212, 191, 0.16), transparent 42%),
    linear-gradient(180deg, #111827 0%, #020617 100%);
  --reel-fade: #020617;
  --reel-title: #2dd4bf;
  --reel-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.04) inset,
    0 22px 48px rgba(2, 6, 23, 0.35);
}

.dark .reel--luxury {
  --reel-accent: #fbbf24;
  --reel-frame-line: color-mix(in srgb, #fbbf24 35%, #1e293b);
  --reel-stage:
    radial-gradient(circle at 50% 0%, rgba(251, 191, 36, 0.16), transparent 42%),
    linear-gradient(180deg, #1c1917 0%, #0c0a09 100%);
  --reel-title: #fbbf24;
}

.reel.is-running {
  pointer-events: none;
}

.reel__veil {
  position: fixed;
  inset: 0;
  z-index: 30;
  background: rgba(15, 23, 42, 0.28);
  backdrop-filter: blur(2px);
}

.dark .reel__veil {
  background: rgba(2, 6, 23, 0.45);
}

.reel__cabinet {
  position: relative;
  z-index: 40;
  overflow: hidden;
  border: 1.5px solid var(--reel-frame-line);
  border-radius: 26px;
  background: var(--reel-frame);
  padding: 12px;
  box-shadow: var(--reel-shadow);
}

.reel__cabinet::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.45) 0%,
    transparent 35%,
    transparent 65%,
    rgba(15, 23, 42, 0.06) 100%
  );
  pointer-events: none;
  z-index: 0;
}

.dark .reel__cabinet::before {
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.08) 0%,
    transparent 40%,
    transparent 70%,
    rgba(0, 0, 0, 0.2) 100%
  );
}

.reel__marquee {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  border: 1px solid var(--reel-marquee-line);
  border-radius: 16px;
  background: var(--reel-marquee);
  padding: 10px 14px;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    0 4px 12px rgba(15, 23, 42, 0.06);
}

.dark .reel__marquee {
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.05);
}

.reel__leds {
  display: flex;
  gap: 6px;
}

.reel__led {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: color-mix(in srgb, var(--reel-accent) 55%, #94a3b8);
  box-shadow: 0 0 8px color-mix(in srgb, var(--reel-accent) 35%, transparent);
  opacity: 0.55;
}

.reel.is-running .reel__led {
  animation: reel-blink 0.55s ease-in-out infinite alternate;
  opacity: 1;
}

.reel__title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  color: var(--reel-title);
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.reel__stage {
  position: relative;
  z-index: 1;
  overflow: hidden;
  border-radius: 18px;
  background: var(--reel-stage);
  padding: 28px 0;
  box-shadow:
    inset 0 0 0 1px rgba(255, 255, 255, 0.08),
    inset 0 12px 28px rgba(0, 0, 0, 0.35),
    inset 0 -8px 18px rgba(0, 0, 0, 0.22);
}

.dark .reel__stage {
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.04);
}

.reel__bezel {
  position: absolute;
  inset-inline: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--reel-accent), transparent);
  box-shadow: 0 0 12px var(--reel-accent);
  pointer-events: none;
}

.reel__bezel--top {
  top: 0;
}

.reel__bezel--bottom {
  bottom: 0;
}

.reel__fade {
  position: absolute;
  inset-block: 0;
  z-index: 20;
  width: clamp(48px, 12vw, 120px);
  pointer-events: none;
}

.reel__fade--left {
  left: 0;
  background: linear-gradient(90deg, var(--reel-fade) 10%, transparent);
}

.reel__fade--right {
  right: 0;
  background: linear-gradient(270deg, var(--reel-fade) 10%, transparent);
}

.reel__pointer {
  position: absolute;
  top: 4px;
  left: 50%;
  z-index: 30;
  display: flex;
  flex-direction: column;
  align-items: center;
  transform: translateX(-50%);
  pointer-events: none;
}

.reel__crown {
  width: 0;
  height: 0;
  border-left: 12px solid transparent;
  border-right: 12px solid transparent;
  border-top: 16px solid var(--reel-accent);
  filter: drop-shadow(0 4px 10px color-mix(in srgb, var(--reel-accent) 70%, transparent));
}

.reel__beam {
  width: 2px;
  height: 218px;
  background: linear-gradient(180deg, var(--reel-accent), transparent);
  box-shadow: 0 0 12px var(--reel-accent);
  opacity: 0.85;
}

.reel__viewport {
  width: 100%;
  overflow: hidden;
}

.reel__track {
  display: flex;
  width: max-content;
  gap: 16px;
  padding: 0 28px;
  transition-property: transform;
  transition-timing-function: cubic-bezier(0.12, 0.72, 0.08, 1);
  will-change: transform;
}

@keyframes reel-blink {
  from {
    filter: brightness(0.7);
  }
  to {
    filter: brightness(1.35);
  }
}

@media (prefers-reduced-motion: reduce) {
  .reel.is-running .reel__led {
    animation: none;
  }
}
</style>
