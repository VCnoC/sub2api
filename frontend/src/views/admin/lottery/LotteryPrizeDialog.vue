<template>
  <BaseDialog :show="show" :title="prize ? t('lotteryAdmin.editPrize') : t('lotteryAdmin.addPrize')" width="wide" @close="emit('cancel')">
    <div class="lottery_prize_dialog">
      <div class="lottery_prize_dialog__grid">
        <label>
          <span>{{ t('lotteryAdmin.name') }}</span>
          <input v-model.trim="form.name" class="input" maxlength="120" />
        </label>
        <label>
          <span>{{ t('lotteryAdmin.prizeType') }}</span>
          <select v-model="form.prize_type" class="input">
            <option value="balance">{{ t('lottery.balancePrize') }}</option>
            <option value="subscription">{{ t('lotteryAdmin.subscriptionPlan') }}</option>
          </select>
        </label>
        <label v-if="form.prize_type === 'balance'">
          <span>{{ t('lotteryAdmin.balanceAmount') }}</span>
          <input v-model.number="form.balance_amount" class="input" type="number" min="0.00000001" step="0.01" />
        </label>
        <label v-else>
          <span>{{ t('lotteryAdmin.subscriptionPlan') }}</span>
          <select v-model.number="form.plan_id" class="input">
            <option :value="null">{{ t('lotteryAdmin.selectSubscriptionPlan') }}</option>
            <option v-if="plans.length === 0" :value="null" disabled>{{ t('lotteryAdmin.noSubscriptionPlans') }}</option>
            <option v-for="plan in plans" :key="plan.id" :value="plan.id">
              {{ plan.name }} · {{ plan.validity_days }} {{ t('lotteryAdmin.days') }}
            </option>
          </select>
        </label>
        <label v-if="form.prize_type === 'subscription'">
          <span>{{ t('lotteryAdmin.subscriptionPlanInfo') }}</span>
          <input class="input" :value="selectedPlanInfo" disabled />
        </label>
        <label>
          <span>{{ t('lotteryAdmin.probability') }}</span>
          <input v-model.number="form.probability_percent" class="input" type="number" min="0" max="100" step="0.0001" />
        </label>
        <label>
          <span>{{ t('lotteryAdmin.stock') }}</span>
          <select v-model="form.stock_mode" class="input">
            <option value="unlimited">{{ t('lotteryAdmin.unlimited') }}</option>
            <option value="fixed">{{ t('lotteryAdmin.fixed') }}</option>
          </select>
        </label>
        <label v-if="form.stock_mode === 'fixed'">
          <span>{{ t('lotteryAdmin.stock') }}</span>
          <input v-model.number="form.stock_total" class="input" type="number" :min="prize?.stock_used ?? 0" step="1" />
        </label>
        <label>
          <span>{{ t('lotteryAdmin.descriptionLabel') }}</span>
          <input v-model.trim="form.description" class="input" maxlength="500" />
        </label>
        <label>
          <span>{{ t('lotteryAdmin.sortOrder') }}</span>
          <input v-model.number="form.sort_order" class="input" type="number" step="1" />
        </label>
      </div>

      <div>
        <p class="lottery_prize_dialog__label">{{ t('lotteryAdmin.image') }}</p>
        <ImageUpload
          v-model="form.image_data"
          mode="image"
          accept="image/png,image/jpeg,image/webp"
          :max-size="300 * 1024"
        />
      </div>

      <label class="lottery_prize_dialog__enabled">
        <input v-model="form.enabled" type="checkbox" />
        <span>{{ t('lotteryAdmin.enabled') }}</span>
      </label>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('cancel')">{{ t('common.cancel') }}</button>
        <button type="button" class="btn btn-primary" :disabled="saving || !canSave" @click="save">
          <Icon :name="saving ? 'refresh' : 'check'" size="sm" :class="{ 'animate-spin': saving }" />
          <span>{{ t('lotteryAdmin.save') }}</span>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ImageUpload from '@/components/common/ImageUpload.vue'
import Icon from '@/components/icons/Icon.vue'
import type { SubscriptionPlan } from '@/types/payment'
import type { LotteryPrize, LotteryPrizeInput, LotteryPrizeType } from '@/types/lottery'

interface PrizeForm {
  name: string
  description: string
  image_data: string
  prize_type: LotteryPrizeType
  balance_amount: number | null
  plan_id: number | null
  probability_percent: number
  stock_mode: 'unlimited' | 'fixed'
  stock_total: number | null
  enabled: boolean
  sort_order: number
}

const props = defineProps<{
  show: boolean
  prize: LotteryPrize | null
  poolId: number
  plans: SubscriptionPlan[]
  saving: boolean
}>()
const emit = defineEmits<{ cancel: []; save: [input: LotteryPrizeInput] }>()
const { t } = useI18n()

const form = reactive<PrizeForm>({
  name: '', description: '', image_data: '', prize_type: 'balance', balance_amount: 1,
  plan_id: null, probability_percent: 0, stock_mode: 'unlimited',
  stock_total: null, enabled: true, sort_order: 0,
})

const selectedPlan = computed(() => props.plans.find((plan) => plan.id === Number(form.plan_id)) ?? null)
const selectedPlanInfo = computed(() => {
  const plan = selectedPlan.value
  if (!plan) return '-'
  return `${plan.group_name ?? `#${plan.group_id}`} · ${plan.validity_days} ${t('lotteryAdmin.days')}`
})

watch(
  () => [props.show, props.prize] as const,
  () => {
    const prize = props.prize
    const plan = prize?.prize_type === 'subscription'
      ? props.plans.find((item) => item.group_id === prize.group_id && item.validity_days === prize.validity_days)
      : null
    Object.assign(form, {
      name: prize?.name ?? '',
      description: prize?.description ?? '',
      image_data: prize?.image_data ?? '',
      prize_type: prize?.prize_type ?? 'balance',
      balance_amount: prize?.balance_amount ?? 1,
      plan_id: plan?.id ?? null,
      probability_percent: (prize?.probability_ppm ?? 0) / 10_000,
      stock_mode: prize?.stock_total == null ? 'unlimited' : 'fixed',
      stock_total: prize?.stock_total ?? null,
      enabled: prize?.enabled ?? true,
      sort_order: prize?.sort_order ?? 0,
    })
  },
  { immediate: true },
)

const canSave = computed(() => {
  if (!form.name || props.poolId <= 0 || form.probability_percent < 0 || form.probability_percent > 100) return false
  if (form.stock_mode === 'fixed' && (
    form.stock_total == null
    || !Number.isSafeInteger(Number(form.stock_total))
    || form.stock_total < (props.prize?.stock_used ?? 0)
  )) return false
  if (form.prize_type === 'balance') {
    const amount = Number(form.balance_amount)
    return Number.isFinite(amount) && amount > 0 && amount <= 1_000_000_000_000
  }
  return selectedPlan.value != null
})

function save(): void {
  if (!canSave.value) return
  emit('save', {
    pool_id: props.poolId,
    name: form.name,
    description: form.description,
    image_data: form.image_data,
    prize_type: form.prize_type,
    balance_amount: form.prize_type === 'balance' ? Number(form.balance_amount) : null,
    group_id: form.prize_type === 'subscription' ? selectedPlan.value?.group_id ?? null : null,
    validity_days: form.prize_type === 'subscription' ? selectedPlan.value?.validity_days ?? null : null,
    probability_ppm: Math.round(Number(form.probability_percent) * 10_000),
    stock_total: form.stock_mode === 'fixed' ? Number(form.stock_total) : null,
    enabled: form.enabled,
    sort_order: Number(form.sort_order) || 0,
  })
}
</script>

<style scoped>
.lottery_prize_dialog {
  display: grid;
  gap: 20px;
}

.lottery_prize_dialog__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.lottery_prize_dialog__grid label {
  display: grid;
  gap: 6px;
}

.lottery_prize_dialog__grid label > span,
.lottery_prize_dialog__label {
  color: rgb(75 85 99);
  font-size: 12px;
  font-weight: 600;
}

.lottery_prize_dialog__label {
  margin-bottom: 8px;
}

.lottery_prize_dialog__enabled {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

:global(.dark) .lottery_prize_dialog__grid label > span,
:global(.dark) .lottery_prize_dialog__label {
  color: rgb(161 161 170);
}

@media (max-width: 640px) {
  .lottery_prize_dialog__grid {
    grid-template-columns: 1fr;
  }
}
</style>
