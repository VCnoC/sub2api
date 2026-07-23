<template>
  <BaseDialog :show="show" :title="prize ? t('lotteryAdmin.editPrize') : t('lotteryAdmin.addPrize')" width="wide" @close="emit('cancel')">
    <div class="space-y-5 py-2">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.name') }}
          </label>
          <input v-model.trim="form.name" class="input w-full" maxlength="120" />
        </div>

        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.prizeType') }}
          </label>
          <Select
            v-model="form.prize_type"
            :options="[
              { value: 'balance', label: t('lottery.balancePrize') },
              { value: 'subscription', label: t('lotteryAdmin.subscriptionPlan') }
            ]"
            class="w-full"
          />
        </div>

        <div v-if="form.prize_type === 'balance'">
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.balanceAmount') }}
          </label>
          <input v-model.number="form.balance_amount" class="input w-full" type="number" min="0.00000001" step="0.01" />
        </div>

        <div v-else>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.subscriptionPlan') }}
          </label>
          <Select
            v-model="form.plan_id"
            :options="planOptions"
            :placeholder="t('lotteryAdmin.selectSubscriptionPlan')"
            class="w-full"
          />
        </div>

        <div v-if="form.prize_type === 'subscription'">
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.subscriptionPlanInfo') }}
          </label>
          <input class="input w-full bg-gray-50 dark:bg-dark-900 text-gray-500" :value="selectedPlanInfo" disabled />
        </div>

        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.probability') }}
          </label>
          <input v-model.number="form.probability_percent" class="input w-full" type="number" min="0" max="100" step="0.0001" />
        </div>

        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.stock') }}
          </label>
          <Select
            v-model="form.stock_mode"
            :options="[
              { value: 'unlimited', label: t('lotteryAdmin.unlimited') },
              { value: 'fixed', label: t('lotteryAdmin.fixed') }
            ]"
            class="w-full"
          />
        </div>

        <div v-if="form.stock_mode === 'fixed'">
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.stock') }}
          </label>
          <input v-model.number="form.stock_total" class="input w-full" type="number" :min="prize?.stock_used ?? 0" step="1" />
        </div>

        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.descriptionLabel') }}
          </label>
          <input v-model.trim="form.description" class="input w-full" maxlength="500" />
        </div>

        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.sortOrder') }}
          </label>
          <input v-model.number="form.sort_order" class="input w-full" type="number" step="1" />
        </div>
      </div>

      <div>
        <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">
          {{ t('lotteryAdmin.image') }}
        </label>
        <ImageUpload
          v-model="form.image_data"
          mode="image"
          accept="image/png,image/jpeg,image/webp"
          :max-size="300 * 1024"
        />
      </div>

      <div class="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-dark-700">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('lotteryAdmin.enabled') }}</span>
        <Toggle v-model="form.enabled" />
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('cancel')">{{ t('common.cancel') }}</button>
        <button type="button" class="btn btn-primary" :disabled="saving || !canSave" @click="save">
          <Icon :name="saving ? 'refresh' : 'check'" size="sm" :class="{ 'animate-spin': saving }" class="mr-1.5" />
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
import Select from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
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

const planOptions = computed(() =>
  props.plans.map((plan) => ({
    value: plan.id,
    label: `${plan.name} · ${plan.validity_days} ${t('lotteryAdmin.days')}`
  }))
)

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
