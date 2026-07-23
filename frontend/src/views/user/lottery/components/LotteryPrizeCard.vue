<template>
  <article
    class="foil-card lottery_prize_card"
    :class="[
      `foil-card--${variant}`,
      {
        'foil-card--winner lottery_prize_card--winner': winner,
        'lottery_prize_card--empty': none,
      },
    ]"
  >
    <div class="foil-card__shell">
      <div class="foil-card__edge" aria-hidden="true"></div>

      <!-- Top foil band -->
      <div class="foil-card__band">
        <span class="foil-card__rank">{{ rankText }}</span>
        <span v-if="winner" class="foil-card__badge foil-card__badge--win">
          {{ t('lottery.wonLabel') }}
        </span>
        <span v-else class="foil-card__badge">{{ kindText }}</span>
      </div>

      <!-- Medallion -->
      <div class="foil-card__stage">
        <div class="foil-card__halo" aria-hidden="true"></div>
        <div class="foil-card__medallion">
          <img
            v-if="imageData && !none"
            class="foil-card__photo"
            :src="imageData"
            :alt="displayName"
          />
          <Icon
            v-else
            class="foil-card__icon"
            :name="none ? 'xCircle' : prizeType === 'subscription' ? 'creditCard' : 'dollar'"
            size="lg"
            :stroke-width="1.75"
          />
        </div>
      </div>

      <!-- Caption plate -->
      <div class="foil-card__plate">
        <p class="foil-card__title">{{ displayName }}</p>
        <div class="foil-card__row">
          <span class="foil-card__amount">{{ amountText }}</span>
          <span v-if="showProbability && probabilityPPM != null" class="foil-card__rate">
            {{ rateText }}
          </span>
        </div>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { LotteryPrize, LotteryPrizeSnapshot } from '@/types/lottery'

const props = withDefaults(
  defineProps<{
    prize?: LotteryPrize | LotteryPrizeSnapshot | null
    none?: boolean
    winner?: boolean
    showProbability?: boolean
  }>(),
  {
    prize: null,
    none: false,
    winner: false,
    showProbability: false,
  },
)

const { t } = useI18n()

const prizeType = computed(() => props.prize?.prize_type || 'balance')
const imageData = computed(() =>
  'image_data' in (props.prize ?? {}) ? String((props.prize as LotteryPrize).image_data || '') : '',
)
const probabilityPPM = computed(() => props.prize?.probability_ppm)

const isHigh = computed(() => {
  if (prizeType.value === 'subscription') return true
  return Number(props.prize?.balance_amount ?? 0) >= 10
})

const variant = computed<'void' | 'teal' | 'gold'>(() => {
  if (props.none) return 'void'
  return isHigh.value ? 'gold' : 'teal'
})

const rankText = computed(() => {
  if (variant.value === 'void') return 'VOID'
  return variant.value === 'gold' ? 'GOLD' : 'TEAL'
})

const kindText = computed(() => {
  if (props.none) return '—'
  if (prizeType.value === 'subscription') return t('lottery.cardTypeSubscription')
  return t('lottery.balancePrize')
})

const displayName = computed(() => (props.none ? t('lottery.noPrize') : props.prize?.name || t('lottery.noPrize')))

const amountText = computed(() => {
  if (props.none) return '· · ·'
  if (prizeType.value === 'subscription') {
    return t('lottery.subscriptionPrize', { days: props.prize?.validity_days ?? 0 })
  }
  return `$${Number(props.prize?.balance_amount ?? 0).toFixed(2)}`
})

const rateText = computed(() => `${Number(((probabilityPPM.value ?? 0) / 10_000).toFixed(2))}%`)
</script>

<style scoped>
.foil-card {
  /* Light: tinted ticket with foil edge — not flat white paper */
  --fc-w: 150px;
  --fc-h: 218px;
  --fc-radius: 18px;
  --fc-ink: #0f172a;
  --fc-soft: #64748b;
  --fc-accent: #0f766e;
  --fc-band: linear-gradient(105deg, #115e59 0%, #0d9488 42%, #2dd4bf 72%, #0f766e 100%);
  --fc-shell:
    radial-gradient(ellipse at 50% 0%, rgba(45, 212, 191, 0.28), transparent 55%),
    linear-gradient(165deg, #ccfbf1 0%, #99f6e4 38%, #5eead4 100%);
  --fc-edge: #0d9488;
  --fc-glow: 0 12px 24px rgba(13, 148, 136, 0.22), 0 2px 0 rgba(255, 255, 255, 0.85);
  --fc-halo: rgba(20, 184, 166, 0.45);
  --fc-medal: linear-gradient(145deg, #134e4a, #14b8a6 48%, #0f766e);
  --fc-icon: #ecfeff;
  --fc-plate: rgba(255, 255, 255, 0.92);
  --fc-plate-line: rgba(13, 148, 136, 0.35);
  --fc-badge-bg: rgba(0, 0, 0, 0.22);

  position: relative;
  width: var(--fc-w);
  min-width: var(--fc-w);
  height: var(--fc-h);
  flex: none;
  border-radius: var(--fc-radius);
  color: var(--fc-ink);
  transition: transform 240ms ease, filter 240ms ease;
  filter: drop-shadow(var(--fc-glow));
  user-select: none;
}

.foil-card:hover {
  transform: translateY(-4px);
}

.foil-card--gold {
  --fc-accent: #b45309;
  --fc-band: linear-gradient(105deg, #92400e 0%, #d97706 40%, #fbbf24 70%, #b45309 100%);
  --fc-shell:
    radial-gradient(ellipse at 50% 0%, rgba(251, 191, 36, 0.32), transparent 55%),
    linear-gradient(165deg, #fef3c7 0%, #fde68a 40%, #fcd34d 100%);
  --fc-edge: #d97706;
  --fc-glow: 0 12px 24px rgba(217, 119, 6, 0.26), 0 2px 0 rgba(255, 255, 255, 0.85);
  --fc-halo: rgba(245, 158, 11, 0.5);
  --fc-medal: linear-gradient(145deg, #92400e, #f59e0b 48%, #b45309);
  --fc-icon: #fffbeb;
  --fc-plate-line: rgba(217, 119, 6, 0.4);
}

.foil-card--void {
  --fc-accent: #475569;
  --fc-band: linear-gradient(105deg, #334155 0%, #64748b 48%, #94a3b8 78%, #475569 100%);
  --fc-shell:
    radial-gradient(ellipse at 50% 0%, rgba(148, 163, 184, 0.24), transparent 55%),
    linear-gradient(165deg, #e2e8f0 0%, #cbd5e1 42%, #94a3b8 100%);
  --fc-edge: #64748b;
  --fc-glow: 0 12px 24px rgba(71, 85, 105, 0.18), 0 2px 0 rgba(255, 255, 255, 0.8);
  --fc-halo: rgba(148, 163, 184, 0.38);
  --fc-medal: linear-gradient(145deg, #334155, #94a3b8 52%, #475569);
  --fc-icon: #f8fafc;
  --fc-soft: #475569;
  --fc-plate-line: rgba(100, 116, 139, 0.32);
}

.foil-card--winner {
  --fc-band: linear-gradient(90deg, #b45309, #fbbf24 45%, #f59e0b 70%, #ea580c);
  --fc-edge: #f59e0b;
  --fc-glow: 0 0 0 1px rgba(245, 158, 11, 0.35), 0 16px 34px rgba(245, 158, 11, 0.28);
  z-index: 20;
  transform: scale(1.06);
  animation: foil-pulse 1.7s ease-in-out infinite;
}

/*
  Dark overrides live in the UNSCOPED block below.
  Scoped `.dark .x` fails because cards are rendered inside LotteryReel,
  so their data-v- hash differs from this component's hash.
*/

.foil-card__shell {
  position: relative;
  display: flex;
  height: 100%;
  flex-direction: column;
  overflow: hidden;
  border-radius: var(--fc-radius);
  background: var(--fc-shell);
  isolation: isolate;
}

.foil-card__shell::before {
  content: '';
  position: absolute;
  inset: 0;
  z-index: 0;
  background: linear-gradient(
    125deg,
    transparent 35%,
    rgba(255, 255, 255, 0.45) 48%,
    transparent 62%
  );
  opacity: 0.55;
  pointer-events: none;
}

.foil-card__edge {
  position: absolute;
  inset: 0;
  border-radius: inherit;
  border: 1.5px solid var(--fc-edge);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.65),
    inset 0 -2px 0 rgba(15, 23, 42, 0.08),
    0 0 0 1px color-mix(in srgb, var(--fc-edge) 28%, transparent);
  pointer-events: none;
  z-index: 5;
}

.foil-card__band {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-height: 36px;
  padding: 0 10px;
  background: var(--fc-band);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.35),
    inset 0 -1px 0 rgba(0, 0, 0, 0.18),
    0 8px 16px rgba(15, 23, 42, 0.14);
}

.foil-card__rank {
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.14em;
  line-height: 1;
  color: #fff;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.35);
}

.foil-card__badge {
  max-width: 64px;
  overflow: hidden;
  border-radius: 999px;
  background: var(--fc-badge-bg);
  padding: 3px 8px;
  color: rgba(255, 255, 255, 0.95);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.02em;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.foil-card__badge--win {
  background: rgba(255, 255, 255, 0.28);
  color: #fffbeb;
  font-weight: 800;
}

.foil-card__stage {
  position: relative;
  z-index: 1;
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: center;
  min-height: 0;
  padding: 14px 12px 8px;
}

.foil-card__halo {
  position: absolute;
  width: 92px;
  height: 92px;
  border-radius: 50%;
  background: radial-gradient(circle, var(--fc-halo) 0%, transparent 70%);
  filter: blur(2px);
  pointer-events: none;
}

.foil-card__medallion {
  position: relative;
  z-index: 1;
  display: flex;
  width: 74px;
  height: 74px;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 50%;
  background: var(--fc-medal);
  box-shadow:
    0 0 0 3px rgba(255, 255, 255, 0.7),
    0 0 0 7px color-mix(in srgb, var(--fc-accent) 28%, transparent),
    0 12px 20px rgba(15, 23, 42, 0.2),
    inset 0 2px 0 rgba(255, 255, 255, 0.35),
    inset 0 -3px 8px rgba(0, 0, 0, 0.2);
}

.foil-card__photo {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.foil-card__icon {
  color: var(--fc-icon);
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.25));
}

.foil-card__plate {
  position: relative;
  z-index: 2;
  margin: 0 8px 8px;
  border-radius: 12px;
  background: var(--fc-plate);
  padding: 9px 10px;
  border: 1px solid var(--fc-plate-line);
  backdrop-filter: blur(8px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.75),
    0 4px 10px rgba(15, 23, 42, 0.06);
}

.foil-card__title {
  display: -webkit-box;
  overflow: hidden;
  margin: 0;
  color: var(--fc-ink);
  font-size: 12.5px;
  font-weight: 750;
  letter-spacing: -0.01em;
  line-height: 1.3;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.foil-card__row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
  margin-top: 5px;
  min-width: 0;
}

.foil-card__amount {
  overflow: hidden;
  color: var(--fc-accent);
  font-size: 12px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.01em;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.foil-card__rate {
  flex: none;
  color: var(--fc-soft);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 10px;
  font-variant-numeric: tabular-nums;
}

.foil-card--winner::after {
  content: '';
  position: absolute;
  inset: 0;
  z-index: 6;
  overflow: hidden;
  border-radius: inherit;
  pointer-events: none;
  background: linear-gradient(
    110deg,
    transparent 20%,
    rgba(255, 255, 255, 0.28) 45%,
    transparent 70%
  );
  background-size: 220% 100%;
  animation: foil-sheen 2s ease-in-out infinite;
}

@keyframes foil-pulse {
  0%,
  100% {
    filter: drop-shadow(0 14px 28px rgba(245, 158, 11, 0.28));
  }
  50% {
    filter: drop-shadow(0 18px 36px rgba(245, 158, 11, 0.42));
  }
}

@keyframes foil-sheen {
  0% {
    background-position: 120% 0;
    opacity: 0.35;
  }
  40% {
    opacity: 0.9;
  }
  100% {
    background-position: -40% 0;
    opacity: 0.35;
  }
}

@media (prefers-reduced-motion: reduce) {
  .foil-card,
  .foil-card--winner,
  .foil-card--winner::after {
    animation: none !important;
    transition: none !important;
  }
}
</style>

<style>
/*
  Dark-mode overrides: UNSCOPED block.
  Scoped `.dark .x` fails because cards are rendered inside LotteryReel,
  so their data-v- hash differs from this component's hash.
  See SettingsView.vue for the same pattern.
*/
.dark .foil-card {
  --fc-ink: #f8fafc;
  --fc-soft: rgba(248, 250, 252, 0.68);
  --fc-accent: #2dd4bf;
  --fc-band: linear-gradient(90deg, #115e59, #14b8a6 55%, #0d9488);
  --fc-shell: linear-gradient(165deg, #1e293b 0%, #0f172a 48%, #020617 100%);
  --fc-edge: rgba(45, 212, 191, 0.55);
  --fc-glow: 0 14px 28px rgba(15, 118, 110, 0.28);
  --fc-halo: rgba(45, 212, 191, 0.35);
  --fc-medal: linear-gradient(145deg, #134e4a, #0f766e 55%, #115e59);
  --fc-icon: #99f6e4;
  --fc-plate: rgba(2, 6, 23, 0.55);
  --fc-plate-line: rgba(255, 255, 255, 0.08);
  --fc-badge-bg: rgba(0, 0, 0, 0.22);
}

.dark .foil-card--gold {
  --fc-accent: #fbbf24;
  --fc-band: linear-gradient(90deg, #92400e, #f59e0b 50%, #d97706);
  --fc-shell: linear-gradient(165deg, #292524 0%, #1c1917 46%, #0c0a09 100%);
  --fc-edge: rgba(251, 191, 36, 0.62);
  --fc-glow: 0 14px 30px rgba(180, 83, 9, 0.32);
  --fc-halo: rgba(251, 191, 36, 0.38);
  --fc-medal: linear-gradient(145deg, #92400e, #d97706 52%, #78350f);
  --fc-icon: #fde68a;
  --fc-plate-line: rgba(251, 191, 36, 0.16);
}

.dark .foil-card--void {
  --fc-accent: #94a3b8;
  --fc-band: linear-gradient(90deg, #334155, #64748b 55%, #475569);
  --fc-shell: linear-gradient(165deg, #1e293b 0%, #0f172a 50%, #020617 100%);
  --fc-edge: rgba(148, 163, 184, 0.35);
  --fc-glow: 0 10px 22px rgba(15, 23, 42, 0.28);
  --fc-halo: rgba(148, 163, 184, 0.18);
  --fc-medal: linear-gradient(145deg, #334155, #475569 55%, #1e293b);
  --fc-icon: #cbd5e1;
  --fc-soft: rgba(203, 213, 225, 0.55);
  --fc-plate-line: rgba(148, 163, 184, 0.16);
}

.dark .foil-card--winner {
  --fc-band: linear-gradient(90deg, #b45309, #fbbf24 45%, #f59e0b 70%, #ea580c);
  --fc-edge: rgba(251, 191, 36, 0.95);
  --fc-glow: 0 0 0 1px rgba(251, 191, 36, 0.55), 0 16px 36px rgba(245, 158, 11, 0.5);
}

.dark .foil-card__shell::before {
  background: linear-gradient(
    125deg,
    transparent 35%,
    rgba(255, 255, 255, 0.12) 48%,
    transparent 62%
  );
  opacity: 0.7;
}

.dark .foil-card__edge {
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.18),
    inset 0 -1px 0 rgba(0, 0, 0, 0.35);
}

.dark .foil-card__medallion {
  box-shadow:
    0 0 0 3px rgba(255, 255, 255, 0.12),
    0 0 0 6px color-mix(in srgb, var(--fc-accent) 28%, transparent),
    0 10px 18px rgba(0, 0, 0, 0.35),
    inset 0 2px 0 rgba(255, 255, 255, 0.22);
}

.dark .foil-card__plate {
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.06);
}
</style>
