<template>
  <article
    class="lottery_prize_card"
    :class="{ 'lottery_prize_card--empty': none, 'lottery_prize_card--winner': winner }"
  >
    <div class="lottery_prize_card__visual">
      <img
        v-if="imageData && !none"
        :src="imageData"
        :alt="name"
        class="lottery_prize_card__image"
      />
      <Icon
        v-else
        :name="none ? 'refresh' : prizeType === 'subscription' ? 'creditCard' : 'dollar'"
        size="lg"
        :stroke-width="1.7"
      />
    </div>
    <div class="lottery_prize_card__copy">
      <p class="lottery_prize_card__name">{{ none ? t('lottery.noPrize') : name }}</p>
      <p v-if="!none" class="lottery_prize_card__value">{{ valueLabel }}</p>
      <p v-if="showProbability && probabilityPPM != null" class="lottery_prize_card__probability">
        {{ t('lottery.probability') }} {{ probabilityLabel }}
      </p>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { LotteryPrize, LotteryPrizeSnapshot } from '@/types/lottery'

const props = withDefaults(defineProps<{
  prize?: LotteryPrize | LotteryPrizeSnapshot | null
  none?: boolean
  winner?: boolean
  showProbability?: boolean
}>(), {
  prize: null,
  none: false,
  winner: false,
  showProbability: false,
})

const { t } = useI18n()
const name = computed(() => props.prize?.name || t('lottery.noPrize'))
const imageData = computed(() => 'image_data' in (props.prize ?? {}) ? String((props.prize as LotteryPrize).image_data || '') : '')
const prizeType = computed(() => props.prize?.prize_type || 'balance')
const probabilityPPM = computed(() => props.prize?.probability_ppm)
const probabilityLabel = computed(() => {
  const value = (probabilityPPM.value ?? 0) / 10_000
  return `${Number(value.toFixed(4))}%`
})
const valueLabel = computed(() => {
  if (prizeType.value === 'subscription') {
    return t('lottery.subscriptionPrize', { days: props.prize?.validity_days ?? 0 })
  }
  return `${t('lottery.balancePrize')} $${Number(props.prize?.balance_amount ?? 0).toFixed(2)}`
})
</script>

<style scoped>
.lottery_prize_card {
  display: flex;
  width: 148px;
  min-width: 148px;
  height: 188px;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgb(229 231 235);
  border-radius: 8px;
  background: rgb(255 255 255);
  color: rgb(17 24 39);
  box-shadow: 0 10px 24px rgb(15 23 42 / 8%);
}

.lottery_prize_card--winner {
  border-color: rgb(232 93 74);
  box-shadow: 0 0 0 2px rgb(232 93 74 / 20%), 0 18px 36px rgb(232 93 74 / 18%);
}

.lottery_prize_card--empty {
  border-style: dashed;
  color: rgb(107 114 128);
}

.lottery_prize_card__visual {
  display: flex;
  height: 116px;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: rgb(246 247 249);
  color: rgb(15 138 120);
}

.lottery_prize_card--empty .lottery_prize_card__visual {
  color: rgb(156 163 175);
}

.lottery_prize_card__image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.lottery_prize_card__copy {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  justify-content: center;
  padding: 10px 12px;
}

.lottery_prize_card__name,
.lottery_prize_card__value,
.lottery_prize_card__probability {
  overflow: hidden;
  text-overflow: ellipsis;
}

.lottery_prize_card__name {
  display: -webkit-box;
  overflow-wrap: anywhere;
  font-size: 14px;
  font-weight: 650;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.lottery_prize_card__value {
  margin-top: 2px;
  color: rgb(15 138 120);
  font-size: 12px;
  white-space: nowrap;
}

.lottery_prize_card__probability {
  margin-top: 2px;
  color: rgb(107 114 128);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

:global(.dark) .lottery_prize_card {
  border-color: rgb(63 63 70);
  background: rgb(24 24 27);
  color: rgb(244 244 245);
}

:global(.dark) .lottery_prize_card__visual {
  background: rgb(39 39 42);
}

@media (max-width: 640px) {
  .lottery_prize_card {
    width: 132px;
    min-width: 132px;
    height: 172px;
  }

  .lottery_prize_card__visual {
    height: 102px;
  }
}
</style>
