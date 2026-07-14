<template>
  <div class="lottery_reel" :class="{ 'lottery_reel--running': running }">
    <div v-if="running" class="lottery_reel__veil" aria-hidden="true"></div>
    <div class="lottery_reel__stage" aria-live="polite">
      <div class="lottery_reel__marker" aria-hidden="true"></div>
      <div ref="viewport" class="lottery_reel__viewport">
        <div ref="track" class="lottery_reel__track" :style="trackStyle">
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
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import type { LotteryDraw, LotteryPrize } from '@/types/lottery'
import LotteryPrizeCard from './LotteryPrizeCard.vue'

interface ReelItem {
  key: string
  prize: LotteryPrize | null
  none: boolean
}

const props = defineProps<{ prizes: LotteryPrize[] }>()
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
  const cardCenter = card ? card.offsetLeft + card.offsetWidth / 2 : winnerIndex.value * 164 + 74
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
.lottery_reel {
  position: relative;
}

.lottery_reel__veil {
  position: fixed;
  z-index: 39;
  inset: 0;
  background: rgb(17 24 39 / 18%);
  backdrop-filter: blur(5px);
}

.lottery_reel__stage {
  position: relative;
  z-index: 40;
  padding: 26px 0;
  overflow: hidden;
  border-block: 1px solid rgb(229 231 235);
  background: rgb(246 247 249);
}

.lottery_reel__viewport {
  width: 100%;
  overflow: hidden;
}

.lottery_reel__track {
  display: flex;
  width: max-content;
  gap: 16px;
  padding-inline: 24px;
  transition-property: transform;
  transition-timing-function: cubic-bezier(.12, .72, .08, 1);
  will-change: transform;
}

.lottery_reel__marker {
  position: absolute;
  z-index: 2;
  top: 10px;
  left: 50%;
  width: 0;
  height: 0;
  border-top: 12px solid rgb(232 93 74);
  border-right: 8px solid transparent;
  border-left: 8px solid transparent;
  transform: translateX(-50%);
}

.lottery_reel__marker::after {
  position: absolute;
  top: 182px;
  left: -1px;
  width: 2px;
  height: 10px;
  background: rgb(232 93 74);
  content: '';
}

:global(.dark) .lottery_reel__stage {
  border-color: rgb(63 63 70);
  background: rgb(15 15 18);
}

@media (max-width: 640px) {
  .lottery_reel__stage {
    padding-block: 22px;
  }

  .lottery_reel__track {
    gap: 12px;
    padding-inline: 16px;
  }

  .lottery_reel__marker::after {
    top: 166px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .lottery_reel__track {
    transition-duration: 0ms !important;
  }
}
</style>
