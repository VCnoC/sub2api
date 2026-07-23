<template>
  <BaseDialog :show="show" :title="rule ? t('lotteryAdmin.editRule') : t('lotteryAdmin.addRule')" width="wide" @close="emit('cancel')">
    <div class="space-y-5 py-2">
      <div>
        <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
          {{ t('lotteryAdmin.name') }}
        </label>
        <input v-model.trim="form.name" class="input w-full" maxlength="120" />
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.event') }}
          </label>
          <Select
            v-model="form.event_type"
            :options="[
              { value: 'signup', label: t('lotteryAdmin.signup') },
              { value: 'redeem', label: t('lotteryAdmin.redeem') },
              { value: 'recharge', label: t('lotteryAdmin.recharge') }
            ]"
            class="w-full"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.beneficiary') }}
          </label>
          <Select
            v-model="form.beneficiary"
            :options="[
              { value: 'inviter', label: t('lotteryAdmin.inviter') },
              { value: 'invitee', label: t('lotteryAdmin.invitee') }
            ]"
            :disabled="form.event_type !== 'signup'"
            class="w-full"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.normalChances') }}
          </label>
          <input v-model.number="form.normal_chances" class="input w-full" type="number" min="0" max="100000" />
        </div>

        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
            {{ t('lotteryAdmin.luxuryChances') }}
          </label>
          <input v-model.number="form.luxury_chances" class="input w-full" type="number" min="0" max="100000" />
        </div>

        <template v-if="form.event_type === 'recharge'">
          <div>
            <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
              {{ t('lotteryAdmin.rechargeMode') }}
            </label>
            <Select
              v-model="form.recharge_mode"
              :options="[
                { value: 'single', label: t('lotteryAdmin.single') },
                { value: 'cumulative', label: t('lotteryAdmin.cumulative') }
              ]"
              class="w-full"
            />
          </div>

          <div>
            <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">
              {{ t('lotteryAdmin.threshold') }}
            </label>
            <input v-model.number="form.recharge_threshold" class="input w-full" type="number" min="0.00000001" step="0.01" />
          </div>
        </template>
      </div>

      <div class="space-y-3 pt-2 border-t border-gray-100 dark:border-dark-700">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('lotteryAdmin.enabled') }}</span>
          <Toggle v-model="form.enabled" />
        </div>

        <div v-if="form.event_type === 'recharge'" class="flex items-center justify-between">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('lotteryAdmin.repeatable') }}</span>
          <Toggle v-model="form.repeatable" />
        </div>
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
import Select from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import type { LotteryEventType, LotteryRule, LotteryRuleInput } from '@/types/lottery'

interface RuleForm {
  name: string
  event_type: LotteryEventType
  beneficiary: 'inviter' | 'invitee'
  normal_chances: number
  luxury_chances: number
  recharge_mode: 'single' | 'cumulative'
  recharge_threshold: number
  repeatable: boolean
  enabled: boolean
}

const props = defineProps<{ show: boolean; rule: LotteryRule | null; saving: boolean }>()
const emit = defineEmits<{ cancel: []; save: [input: LotteryRuleInput] }>()
const { t } = useI18n()
const form = reactive<RuleForm>({
  name: '', event_type: 'signup', beneficiary: 'inviter', normal_chances: 1,
  luxury_chances: 0, recharge_mode: 'single', recharge_threshold: 1,
  repeatable: false, enabled: true,
})

watch(
  () => [props.show, props.rule] as const,
  () => {
    const rule = props.rule
    Object.assign(form, {
      name: rule?.name ?? '',
      event_type: rule?.event_type ?? 'signup',
      beneficiary: rule?.beneficiary ?? 'inviter',
      normal_chances: rule?.normal_chances ?? 1,
      luxury_chances: rule?.luxury_chances ?? 0,
      recharge_mode: rule?.recharge_mode ?? 'single',
      recharge_threshold: rule?.recharge_threshold ?? 1,
      repeatable: rule?.repeatable ?? false,
      enabled: rule?.enabled ?? true,
    })
  },
  { immediate: true },
)

watch(() => form.event_type, (event) => {
  if (event !== 'signup') form.beneficiary = 'inviter'
  if (event !== 'recharge') form.repeatable = false
})

const canSave = computed(() => {
  const normal = Number(form.normal_chances)
  const luxury = Number(form.luxury_chances)
  if (!form.name || !Number.isInteger(normal) || !Number.isInteger(luxury)) return false
  if (normal < 0 || normal > 100_000 || luxury < 0 || luxury > 100_000 || normal + luxury <= 0) return false
  if (form.event_type !== 'recharge') return true
  return Number.isFinite(form.recharge_threshold) && form.recharge_threshold > 0 && form.recharge_threshold <= 1_000_000_000_000
})

function save(): void {
  if (!canSave.value) return
  emit('save', {
    name: form.name,
    event_type: form.event_type,
    beneficiary: form.beneficiary,
    normal_chances: Number(form.normal_chances),
    luxury_chances: Number(form.luxury_chances),
    recharge_mode: form.event_type === 'recharge' ? form.recharge_mode : null,
    recharge_threshold: form.event_type === 'recharge' ? Number(form.recharge_threshold) : null,
    repeatable: form.event_type === 'recharge' && form.repeatable,
    enabled: form.enabled,
  })
}
</script>
