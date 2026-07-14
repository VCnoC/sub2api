<template>
  <BaseDialog :show="show" :title="rule ? t('lotteryAdmin.editRule') : t('lotteryAdmin.addRule')" width="wide" @close="emit('cancel')">
    <div class="lottery_rule_dialog">
      <label>
        <span>{{ t('lotteryAdmin.name') }}</span>
        <input v-model.trim="form.name" class="input" maxlength="120" />
      </label>
      <div class="lottery_rule_dialog__grid">
        <label>
          <span>{{ t('lotteryAdmin.event') }}</span>
          <select v-model="form.event_type" class="input">
            <option value="signup">{{ t('lotteryAdmin.signup') }}</option>
            <option value="redeem">{{ t('lotteryAdmin.redeem') }}</option>
            <option value="recharge">{{ t('lotteryAdmin.recharge') }}</option>
          </select>
        </label>
        <label>
          <span>{{ t('lotteryAdmin.beneficiary') }}</span>
          <select v-model="form.beneficiary" class="input" :disabled="form.event_type !== 'signup'">
            <option value="inviter">{{ t('lotteryAdmin.inviter') }}</option>
            <option value="invitee">{{ t('lotteryAdmin.invitee') }}</option>
          </select>
        </label>
        <label>
          <span>{{ t('lotteryAdmin.normalChances') }}</span>
          <input v-model.number="form.normal_chances" class="input" type="number" min="0" max="100000" />
        </label>
        <label>
          <span>{{ t('lotteryAdmin.luxuryChances') }}</span>
          <input v-model.number="form.luxury_chances" class="input" type="number" min="0" max="100000" />
        </label>
        <template v-if="form.event_type === 'recharge'">
          <label>
            <span>{{ t('lotteryAdmin.rechargeMode') }}</span>
            <select v-model="form.recharge_mode" class="input">
              <option value="single">{{ t('lotteryAdmin.single') }}</option>
              <option value="cumulative">{{ t('lotteryAdmin.cumulative') }}</option>
            </select>
          </label>
          <label>
            <span>{{ t('lotteryAdmin.threshold') }}</span>
            <input v-model.number="form.recharge_threshold" class="input" type="number" min="0.00000001" step="0.01" />
          </label>
        </template>
      </div>
      <div class="lottery_rule_dialog__toggles">
        <label>
          <input v-model="form.enabled" type="checkbox" />
          <span>{{ t('lotteryAdmin.enabled') }}</span>
        </label>
        <label v-if="form.event_type === 'recharge'">
          <input v-model="form.repeatable" type="checkbox" />
          <span>{{ t('lotteryAdmin.repeatable') }}</span>
        </label>
      </div>
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

<style scoped>
.lottery_rule_dialog {
  display: grid;
  gap: 16px;
}

.lottery_rule_dialog > label,
.lottery_rule_dialog__grid label {
  display: grid;
  gap: 6px;
}

.lottery_rule_dialog label > span {
  color: rgb(75 85 99);
  font-size: 12px;
  font-weight: 600;
}

.lottery_rule_dialog__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.lottery_rule_dialog__toggles,
.lottery_rule_dialog__toggles label {
  display: flex;
  align-items: center;
}

.lottery_rule_dialog__toggles {
  flex-wrap: wrap;
  gap: 18px;
}

.lottery_rule_dialog__toggles label {
  gap: 8px;
  font-size: 13px;
}

:global(.dark) .lottery_rule_dialog label > span {
  color: rgb(161 161 170);
}

@media (max-width: 640px) {
  .lottery_rule_dialog__grid {
    grid-template-columns: 1fr;
  }
}
</style>
