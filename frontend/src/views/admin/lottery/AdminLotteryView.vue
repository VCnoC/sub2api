<template>
  <AppLayout>
    <main class="admin_lottery">
      <header class="admin_lottery__header">
        <div>
          <h1>{{ t('lotteryAdmin.title') }}</h1>
          <p>{{ t('lotteryAdmin.description') }}</p>
        </div>
      </header>

      <nav class="admin_lottery__tabs">
        <button v-for="item in tabs" :key="item.key" type="button" :class="{ active: tab === item.key }" @click="tab = item.key">
          {{ item.label }}
        </button>
      </nav>

      <div v-if="loading" class="admin_lottery__loading"><Icon name="refresh" class="animate-spin" /></div>

      <LotteryPoolPanel
        v-else-if="tab === 'pools'"
        :pools="pools"
        :saving-key="savingPool"
        @save="savePool"
      />

      <section v-else-if="tab === 'prizes'" class="admin_lottery__section">
        <div class="admin_lottery__toolbar">
          <div class="admin_lottery__segments">
            <button v-for="pool in pools" :key="pool.id" type="button" :class="{ active: selectedPoolId === pool.id }" @click="selectPool(pool.id)">
              {{ pool.key === 'normal' ? t('lottery.normal') : t('lottery.luxury') }}
            </button>
          </div>
          <div class="admin_lottery__toolbar_actions">
            <span>{{ probabilityTotal }}%</span>
            <button type="button" class="btn btn-primary" :disabled="selectedPoolId <= 0" @click="openPrize()">
              <Icon name="plus" size="sm" />
              <span>{{ t('lotteryAdmin.addPrize') }}</span>
            </button>
          </div>
        </div>

        <div class="admin_lottery__table_wrap">
          <table>
            <thead><tr><th>{{ t('lotteryAdmin.name') }}</th><th>{{ t('lotteryAdmin.prizeType') }}</th><th>{{ t('lotteryAdmin.probability') }}</th><th>{{ t('lotteryAdmin.stock') }}</th><th>{{ t('lotteryAdmin.enabled') }}</th><th></th></tr></thead>
            <tbody>
              <tr v-for="prize in prizes" :key="prize.id">
                <td>
                  <div class="admin_lottery__prize_name">
                    <img v-if="prize.image_data" :src="prize.image_data" :alt="prize.name" />
                    <Icon v-else :name="prize.prize_type === 'subscription' ? 'creditCard' : 'dollar'" size="sm" />
                    <span>{{ prize.name }}</span>
                  </div>
                </td>
                <td>{{ prize.prize_type === 'balance' ? t('lottery.balancePrize') : subscriptionPrizeText(prize) }}</td>
                <td class="tabular-nums">{{ probability(prize.probability_ppm) }}</td>
                <td>{{ prize.stock_total == null ? t('lotteryAdmin.unlimited') : `${prize.stock_used}/${prize.stock_total}` }}</td>
                <td><span class="admin_lottery__status" :class="{ active: prize.enabled }">{{ prize.enabled ? t('common.enabled') : t('common.disabled') }}</span></td>
                <td>
                  <div class="admin_lottery__row_actions">
                    <button type="button" :title="t('lotteryAdmin.editPrize')" @click="openPrize(prize)"><Icon name="edit" size="sm" /></button>
                    <button type="button" :title="t('lotteryAdmin.delete')" @click="confirmDelete('prize', prize.id, prize.name)"><Icon name="trash" size="sm" /></button>
                  </div>
                </td>
              </tr>
              <tr v-if="prizes.length === 0"><td colspan="6" class="admin_lottery__empty">{{ t('lotteryAdmin.empty') }}</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      <section v-else-if="tab === 'rules'" class="admin_lottery__section">
        <div class="admin_lottery__toolbar">
          <span></span>
          <button type="button" class="btn btn-primary" @click="openRule()">
            <Icon name="plus" size="sm" />
            <span>{{ t('lotteryAdmin.addRule') }}</span>
          </button>
        </div>
        <div class="admin_lottery__table_wrap">
          <table>
            <thead><tr><th>{{ t('lotteryAdmin.name') }}</th><th>{{ t('lotteryAdmin.event') }}</th><th>{{ t('lotteryAdmin.beneficiary') }}</th><th>{{ t('lottery.normal') }}</th><th>{{ t('lottery.luxury') }}</th><th>{{ t('lotteryAdmin.enabled') }}</th><th></th></tr></thead>
            <tbody>
              <tr v-for="rule in rules" :key="rule.id">
                <td>{{ rule.name }}</td>
                <td>{{ t(`lotteryAdmin.${rule.event_type}`) }}</td>
                <td>{{ t(`lotteryAdmin.${rule.beneficiary}`) }}</td>
                <td class="tabular-nums">+{{ rule.normal_chances }}</td>
                <td class="tabular-nums">+{{ rule.luxury_chances }}</td>
                <td><span class="admin_lottery__status" :class="{ active: rule.enabled }">{{ rule.enabled ? t('common.enabled') : t('common.disabled') }}</span></td>
                <td>
                  <div class="admin_lottery__row_actions">
                    <button type="button" :title="t('lotteryAdmin.editRule')" @click="openRule(rule)"><Icon name="edit" size="sm" /></button>
                    <button type="button" :title="t('lotteryAdmin.delete')" @click="confirmDelete('rule', rule.id, rule.name)"><Icon name="trash" size="sm" /></button>
                  </div>
                </td>
              </tr>
              <tr v-if="rules.length === 0"><td colspan="7" class="admin_lottery__empty">{{ t('lotteryAdmin.empty') }}</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      <section v-else class="admin_lottery__section">
        <div class="admin_lottery__toolbar">
          <div class="admin_lottery__segments">
            <button type="button" :class="{ active: recordMode === 'draws' }" @click="switchRecords('draws')">{{ t('lotteryAdmin.records') }}</button>
            <button type="button" :class="{ active: recordMode === 'ledger' }" @click="switchRecords('ledger')">{{ t('lotteryAdmin.ledger') }}</button>
          </div>
          <div class="admin_lottery__filters">
            <input v-model="recordUserId" class="input" type="number" min="1" :placeholder="t('lotteryAdmin.userId')" @change="loadRecords" />
            <select v-model="recordPool" class="input" @change="loadRecords">
              <option value="">-</option><option value="normal">{{ t('lottery.normal') }}</option><option value="luxury">{{ t('lottery.luxury') }}</option>
            </select>
          </div>
        </div>
        <div class="admin_lottery__table_wrap">
          <table v-if="recordMode === 'draws'">
            <thead><tr><th>ID</th><th>{{ t('lotteryAdmin.userId') }}</th><th>{{ t('lotteryAdmin.outcome') }}</th><th>{{ t('lotteryAdmin.name') }}</th><th>{{ t('lotteryAdmin.time') }}</th></tr></thead>
            <tbody>
              <tr v-for="item in draws" :key="item.id"><td>{{ item.id }}</td><td>{{ item.user_id }}</td><td>{{ item.outcome === 'win' ? item.prize?.name : t('lottery.noPrize') }}</td><td>{{ item.pool_key }}</td><td>{{ formatDateTime(item.created_at) }}</td></tr>
              <tr v-if="draws.length === 0"><td colspan="5" class="admin_lottery__empty">{{ t('lotteryAdmin.empty') }}</td></tr>
            </tbody>
          </table>
          <table v-else>
            <thead><tr><th>ID</th><th>{{ t('lotteryAdmin.userId') }}</th><th>{{ t('lotteryAdmin.source') }}</th><th>{{ t('lotteryAdmin.delta') }}</th><th>{{ t('lotteryAdmin.time') }}</th></tr></thead>
            <tbody>
              <tr v-for="item in ledger" :key="item.id"><td>{{ item.id }}</td><td>{{ item.user_id }}</td><td>{{ item.source_type }} / {{ item.source_id }}</td><td class="tabular-nums">{{ item.base_delta || item.extra_delta }}</td><td>{{ formatDateTime(item.created_at) }}</td></tr>
              <tr v-if="ledger.length === 0"><td colspan="5" class="admin_lottery__empty">{{ t('lotteryAdmin.empty') }}</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      <LotteryPrizeDialog :show="prizeDialog.show" :prize="prizeDialog.item" :pool-id="selectedPoolId" :plans="subscriptionPlans" :saving="prizeDialog.saving" @cancel="closePrize" @save="savePrize" />
      <LotteryRuleDialog :show="ruleDialog.show" :rule="ruleDialog.item" :saving="ruleDialog.saving" @cancel="closeRule" @save="saveRule" />
      <ConfirmDialog :show="deleteDialog.show" :title="t('lotteryAdmin.delete')" :message="deleteDialog.name" danger @cancel="deleteDialog.show = false" @confirm="performDelete" />
    </main>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import adminLotteryAPI from '@/api/admin/lottery'
import { adminPaymentAPI } from '@/api/admin/payment'
import type { SubscriptionPlan } from '@/types/payment'
import type { LotteryChanceLedgerEntry, LotteryDraw, LotteryPool, LotteryPoolInput, LotteryPoolKey, LotteryPrize, LotteryPrizeInput, LotteryRule, LotteryRuleInput } from '@/types/lottery'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'
import LotteryPoolPanel from './LotteryPoolPanel.vue'
import LotteryPrizeDialog from './LotteryPrizeDialog.vue'
import LotteryRuleDialog from './LotteryRuleDialog.vue'

type TabKey = 'pools' | 'prizes' | 'rules' | 'records'
const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(true)
const tab = ref<TabKey>('pools')
const pools = ref<LotteryPool[]>([])
const prizes = ref<LotteryPrize[]>([])
const rules = ref<LotteryRule[]>([])
const subscriptionPlans = ref<SubscriptionPlan[]>([])
const selectedPoolId = ref(0)
const savingPool = ref<LotteryPoolKey | ''>('')
const recordMode = ref<'draws' | 'ledger'>('draws')
const recordUserId = ref<number | null>(null)
const recordPool = ref('')
const draws = ref<LotteryDraw[]>([])
const ledger = ref<LotteryChanceLedgerEntry[]>([])
const prizeDialog = reactive<{ show: boolean; item: LotteryPrize | null; saving: boolean }>({ show: false, item: null, saving: false })
const ruleDialog = reactive<{ show: boolean; item: LotteryRule | null; saving: boolean }>({ show: false, item: null, saving: false })
const deleteDialog = reactive<{ show: boolean; kind: 'prize' | 'rule'; id: number; name: string }>({ show: false, kind: 'prize', id: 0, name: '' })

const tabs = computed(() => [
  { key: 'pools' as const, label: t('lotteryAdmin.pools') },
  { key: 'prizes' as const, label: t('lotteryAdmin.prizes') },
  { key: 'rules' as const, label: t('lotteryAdmin.rules') },
  { key: 'records' as const, label: t('lotteryAdmin.records') },
])
const probabilityTotal = computed(() => Number((prizes.value.filter((item) => item.enabled).reduce((sum, item) => sum + item.probability_ppm, 0) / 10_000).toFixed(4)))

async function load(): Promise<void> {
  loading.value = true
  try {
    const [poolItems, ruleItems, planRes] = await Promise.all([adminLotteryAPI.listPools(), adminLotteryAPI.listRules(), adminPaymentAPI.getPlans()])
    pools.value = poolItems
    rules.value = ruleItems
    subscriptionPlans.value = planRes.data || []
    selectedPoolId.value = poolItems[0]?.id ?? 0
    await loadPrizes()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('lotteryAdmin.operationFailed')))
  } finally {
    loading.value = false
  }
}

async function loadPrizes(): Promise<void> {
  prizes.value = selectedPoolId.value > 0 ? await adminLotteryAPI.listPrizes(selectedPoolId.value) : []
}

async function selectPool(id: number): Promise<void> { selectedPoolId.value = id; await loadPrizes() }
function probability(ppm: number): string { return `${Number((ppm / 10_000).toFixed(4))}%` }
function subscriptionPrizeText(prize: LotteryPrize): string {
  const plan = subscriptionPlans.value.find((item) => item.group_id === prize.group_id && item.validity_days === prize.validity_days)
  return plan?.name ?? t('lottery.subscriptionPrize', { days: prize.validity_days ?? 0 })
}

async function savePool(key: LotteryPoolKey, input: LotteryPoolInput): Promise<void> {
  savingPool.value = key
  try {
    const saved = await adminLotteryAPI.updatePool(key, input)
    pools.value = pools.value.map((pool) => pool.key === key ? saved : pool)
    appStore.showSuccess(t('lotteryAdmin.saveSuccess'))
  } catch (error) { appStore.showError(extractApiErrorMessage(error, t('lotteryAdmin.operationFailed'))) }
  finally { savingPool.value = '' }
}

function openPrize(item: LotteryPrize | null = null): void { prizeDialog.item = item; prizeDialog.show = true }
function closePrize(): void { if (!prizeDialog.saving) prizeDialog.show = false }
async function savePrize(input: LotteryPrizeInput): Promise<void> {
  const otherProbability = prizes.value
    .filter((item) => item.enabled && item.id !== prizeDialog.item?.id)
    .reduce((total, item) => total + item.probability_ppm, 0)
  if (input.enabled && otherProbability + input.probability_ppm > 1_000_000) {
    appStore.showError(t('lotteryAdmin.probabilityOverflow'))
    return
  }
  prizeDialog.saving = true
  try {
    if (prizeDialog.item) await adminLotteryAPI.updatePrize(prizeDialog.item.id, input)
    else await adminLotteryAPI.createPrize(input)
    await loadPrizes(); prizeDialog.show = false; appStore.showSuccess(t('lotteryAdmin.saveSuccess'))
  } catch (error) { appStore.showError(extractApiErrorMessage(error, t('lotteryAdmin.operationFailed'))) }
  finally { prizeDialog.saving = false }
}

function openRule(item: LotteryRule | null = null): void { ruleDialog.item = item; ruleDialog.show = true }
function closeRule(): void { if (!ruleDialog.saving) ruleDialog.show = false }
async function saveRule(input: LotteryRuleInput): Promise<void> {
  ruleDialog.saving = true
  try {
    if (ruleDialog.item) await adminLotteryAPI.updateRule(ruleDialog.item.id, input)
    else await adminLotteryAPI.createRule(input)
    rules.value = await adminLotteryAPI.listRules(); ruleDialog.show = false; appStore.showSuccess(t('lotteryAdmin.saveSuccess'))
  } catch (error) { appStore.showError(extractApiErrorMessage(error, t('lotteryAdmin.operationFailed'))) }
  finally { ruleDialog.saving = false }
}

function confirmDelete(kind: 'prize' | 'rule', id: number, name: string): void { Object.assign(deleteDialog, { show: true, kind, id, name }) }
async function performDelete(): Promise<void> {
  try {
    if (deleteDialog.kind === 'prize') { await adminLotteryAPI.deletePrize(deleteDialog.id); await loadPrizes() }
    else { await adminLotteryAPI.deleteRule(deleteDialog.id); rules.value = await adminLotteryAPI.listRules() }
    deleteDialog.show = false; appStore.showSuccess(t('lotteryAdmin.deleteSuccess'))
  } catch (error) { appStore.showError(extractApiErrorMessage(error, t('lotteryAdmin.operationFailed'))) }
}

async function loadRecords(): Promise<void> {
  const params = { page: 1, page_size: 50, user_id: recordUserId.value || undefined, pool: recordPool.value || undefined }
  try {
    if (recordMode.value === 'draws') draws.value = (await adminLotteryAPI.listDraws(params)).items ?? []
    else ledger.value = (await adminLotteryAPI.listLedger(params)).items ?? []
  } catch (error) { appStore.showError(extractApiErrorMessage(error, t('lotteryAdmin.operationFailed'))) }
}

async function switchRecords(mode: 'draws' | 'ledger'): Promise<void> { recordMode.value = mode; await loadRecords() }
watch(tab, (value) => { if (value === 'records') void loadRecords() })
onMounted(load)
</script>

<style scoped>
.admin_lottery { display: flex; max-width: 1240px; margin: 0 auto; flex-direction: column; gap: 20px; padding-bottom: 48px; }
.admin_lottery__header h1 { color: rgb(17 24 39); font-size: 24px; font-weight: 720; }
.admin_lottery__header p { margin-top: 3px; color: rgb(107 114 128); font-size: 13px; }
.admin_lottery__tabs { display: flex; gap: 24px; overflow-x: auto; border-bottom: 1px solid rgb(229 231 235); }
.admin_lottery__tabs button { min-height: 42px; flex: none; border-bottom: 2px solid transparent; color: rgb(107 114 128); font-size: 13px; font-weight: 620; }
.admin_lottery__tabs button.active { border-color: rgb(232 93 74); color: rgb(17 24 39); }
.admin_lottery__loading { display: flex; min-height: 360px; align-items: center; justify-content: center; color: rgb(15 138 120); }
.admin_lottery__section { display: flex; min-width: 0; flex-direction: column; gap: 14px; }
.admin_lottery__toolbar, .admin_lottery__toolbar_actions, .admin_lottery__filters, .admin_lottery__segments, .admin_lottery__prize_name, .admin_lottery__row_actions { display: flex; align-items: center; }
.admin_lottery__toolbar { min-height: 42px; justify-content: space-between; gap: 12px; }
.admin_lottery__toolbar_actions, .admin_lottery__filters { gap: 10px; }
.admin_lottery__toolbar_actions > span { color: rgb(15 138 120); font-size: 13px; font-weight: 700; font-variant-numeric: tabular-nums; }
.admin_lottery__segments { padding: 3px; border: 1px solid rgb(229 231 235); border-radius: 8px; background: rgb(243 244 246); }
.admin_lottery__segments button { min-height: 32px; padding: 0 14px; border-radius: 6px; color: rgb(107 114 128); font-size: 12px; font-weight: 600; }
.admin_lottery__segments button.active { background: white; color: rgb(17 24 39); box-shadow: 0 1px 4px rgb(15 23 42 / 10%); }
.admin_lottery__filters .input { width: 148px; }
.admin_lottery__table_wrap { overflow-x: auto; border-block: 1px solid rgb(229 231 235); }
.admin_lottery table { width: 100%; min-width: 720px; border-collapse: collapse; font-size: 13px; }
.admin_lottery th { padding: 10px 12px; color: rgb(107 114 128); font-size: 11px; font-weight: 650; text-align: left; text-transform: uppercase; }
.admin_lottery td { padding: 11px 12px; border-top: 1px solid rgb(243 244 246); color: rgb(55 65 81); }
.admin_lottery__prize_name { gap: 9px; min-width: 180px; color: rgb(17 24 39); font-weight: 620; }
.admin_lottery__prize_name img { width: 32px; height: 32px; border-radius: 6px; object-fit: cover; }
.admin_lottery__status { display: inline-flex; min-width: 48px; justify-content: center; padding: 2px 7px; border-radius: 999px; background: rgb(243 244 246); color: rgb(107 114 128); font-size: 11px; }
.admin_lottery__status.active { background: rgb(220 252 231); color: rgb(21 128 61); }
.admin_lottery__row_actions { justify-content: flex-end; gap: 5px; }
.admin_lottery__row_actions button { display: inline-flex; width: 30px; height: 30px; align-items: center; justify-content: center; border-radius: 6px; color: rgb(107 114 128); }
.admin_lottery__row_actions button:hover { background: rgb(243 244 246); color: rgb(17 24 39); }
.admin_lottery__empty { height: 96px; color: rgb(107 114 128) !important; text-align: center; }
:global(.dark) .admin_lottery__header h1, :global(.dark) .admin_lottery__tabs button.active, :global(.dark) .admin_lottery__prize_name { color: rgb(244 244 245); }
:global(.dark) .admin_lottery__tabs, :global(.dark) .admin_lottery__table_wrap { border-color: rgb(63 63 70); }
:global(.dark) .admin_lottery__segments { border-color: rgb(63 63 70); background: rgb(24 24 27); }
:global(.dark) .admin_lottery__segments button.active { background: rgb(63 63 70); color: white; }
:global(.dark) .admin_lottery td { border-color: rgb(39 39 42); color: rgb(212 212 216); }
@media (max-width: 640px) { .admin_lottery__toolbar { align-items: stretch; flex-direction: column; } .admin_lottery__toolbar_actions, .admin_lottery__filters { justify-content: space-between; } .admin_lottery__filters .input { width: 48%; } }
</style>
