<template>
  <AppLayout>
    <!--
      固定壳层：顶部 Tab 与工具栏高度不变，仅下方内容区随 Tab 切换（v-show，不整页重挂）。
      标题已由 AppHeader 展示，此处不再重复。
    -->
    <div class="flex flex-col gap-4">
      <!-- ① 固定 Tab 栏 -->
      <div class="shrink-0 border-b border-gray-200 dark:border-dark-700">
        <nav class="-mb-px flex gap-6 overflow-x-auto" aria-label="Tabs">
          <button
            v-for="item in tabs"
            :key="item.key"
            type="button"
            :class="[
              'inline-flex shrink-0 items-center gap-2 border-b-2 px-1 py-3 text-sm font-medium whitespace-nowrap transition-colors',
              tab === item.key
                ? 'border-primary-600 text-primary-600 dark:border-primary-500 dark:text-primary-400'
                : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-dark-400 dark:hover:border-dark-600 dark:hover:text-dark-200'
            ]"
            @click="tab = item.key"
          >
            <Icon :name="tabIcons[item.key]" size="sm" />
            <span>{{ item.label }}</span>
          </button>
        </nav>
      </div>

      <!-- ② 固定高度工具栏：所有 Tab 共用同一行槽位，避免切换时上下跳动 -->
      <div class="card flex min-h-[56px] shrink-0 flex-wrap items-center justify-between gap-3 px-4 py-3">
        <!-- pools -->
        <template v-if="tab === 'pools'">
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('lotteryAdmin.description') }}
          </p>
          <button type="button" class="btn btn-secondary" :title="t('common.refresh')" @click="load">
            <Icon name="refresh" size="md" :class="{ 'animate-spin': loading }" />
          </button>
        </template>

        <!-- prizes -->
        <template v-else-if="tab === 'prizes'">
          <div class="inline-flex rounded-xl bg-gray-100 p-1 dark:bg-dark-800">
            <button
              v-for="pool in pools"
              :key="pool.id"
              type="button"
              :class="[
                'rounded-lg px-3.5 py-1.5 text-xs font-semibold transition-all',
                selectedPoolId === pool.id
                  ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
                  : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'
              ]"
              @click="selectPool(pool.id)"
            >
              {{ pool.key === 'normal' ? t('lottery.normal') : t('lottery.luxury') }}
            </button>
          </div>
          <div class="flex flex-wrap items-center gap-3">
            <div
              :class="[
                'inline-flex items-center gap-1.5 rounded-lg border px-3 py-1.5 text-xs font-medium',
                probabilityTotal > 100
                  ? 'border-red-200 bg-red-50 text-red-700 dark:border-red-800/40 dark:bg-red-950/30 dark:text-red-400'
                  : 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800/40 dark:bg-emerald-950/30 dark:text-emerald-400'
              ]"
            >
              <Icon name="chart" size="xs" />
              <span>{{ t('lotteryAdmin.probability') }}:</span>
              <span class="font-bold tabular-nums">{{ probabilityTotal }}%</span>
            </div>
            <button type="button" class="btn btn-secondary" :title="t('common.refresh')" @click="loadPrizes">
              <Icon name="refresh" size="md" />
            </button>
            <button type="button" class="btn btn-primary" :disabled="selectedPoolId <= 0" @click="openPrize()">
              <Icon name="plus" size="md" class="mr-1" />
              <span>{{ t('lotteryAdmin.addPrize') }}</span>
            </button>
          </div>
        </template>

        <!-- rules -->
        <template v-else-if="tab === 'rules'">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('lotteryAdmin.rules') }}</span>
          <div class="flex items-center gap-2">
            <button type="button" class="btn btn-secondary" :title="t('common.refresh')" @click="loadRules">
              <Icon name="refresh" size="md" />
            </button>
            <button type="button" class="btn btn-primary" @click="openRule()">
              <Icon name="plus" size="md" class="mr-1" />
              <span>{{ t('lotteryAdmin.addRule') }}</span>
            </button>
          </div>
        </template>

        <!-- records -->
        <template v-else>
          <div class="flex flex-wrap items-center gap-3">
            <div class="inline-flex rounded-xl bg-gray-100 p-1 dark:bg-dark-800">
              <button
                type="button"
                :class="[
                  'rounded-lg px-3.5 py-1.5 text-xs font-semibold transition-all',
                  recordMode === 'draws'
                    ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
                    : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'
                ]"
                @click="switchRecords('draws')"
              >
                {{ t('lotteryAdmin.records') }}
              </button>
              <button
                type="button"
                :class="[
                  'rounded-lg px-3.5 py-1.5 text-xs font-semibold transition-all',
                  recordMode === 'ledger'
                    ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
                    : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'
                ]"
                @click="switchRecords('ledger')"
              >
                {{ t('lotteryAdmin.ledger') }}
              </button>
            </div>
            <div class="w-36">
              <input
                v-model.number="recordUserId"
                class="input w-full"
                type="number"
                min="1"
                :placeholder="t('lotteryAdmin.userId')"
                @change="loadRecords"
              />
            </div>
            <div class="w-36">
              <Select
                v-model="recordPool"
                :options="[
                  { value: '', label: t('common.all') },
                  { value: 'normal', label: t('lottery.normal') },
                  { value: 'luxury', label: t('lottery.luxury') }
                ]"
                @change="loadRecords"
              />
            </div>
          </div>
          <button type="button" class="btn btn-secondary" :title="t('common.refresh')" @click="loadRecords">
            <Icon name="refresh" size="md" />
          </button>
        </template>
      </div>

      <!-- ③ 内容区：固定最小高度，仅内部用 v-show 切换，不卸载顶部 -->
      <div class="relative min-h-[520px]">
        <div
          v-if="loading"
          class="absolute inset-0 z-10 flex items-center justify-center rounded-2xl border border-gray-200 bg-white/80 dark:border-dark-700 dark:bg-dark-900/80"
        >
          <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-600 border-t-transparent"></div>
        </div>

        <!-- pools -->
        <div v-show="tab === 'pools'">
          <LotteryPoolPanel :pools="pools" :saving-key="savingPool" @save="savePool" />
        </div>

        <!-- prizes -->
        <div
          v-show="tab === 'prizes'"
          class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800"
        >
          <DataTable :columns="prizeColumns" :data="prizes">
            <template #cell-name="{ row }">
              <div class="flex items-center gap-3">
                <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center overflow-hidden rounded-lg border border-gray-200 bg-gray-100 dark:border-dark-600 dark:bg-dark-700">
                  <img v-if="row.image_data" :src="row.image_data" :alt="row.name" class="h-full w-full object-cover" />
                  <Icon
                    v-else
                    :name="row.prize_type === 'subscription' ? 'creditCard' : 'gift'"
                    size="sm"
                    class="text-gray-500 dark:text-gray-400"
                  />
                </div>
                <div>
                  <span class="font-medium text-gray-900 dark:text-white">{{ row.name }}</span>
                  <p v-if="row.description" class="line-clamp-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ row.description }}
                  </p>
                </div>
              </div>
            </template>

            <template #cell-prize_type="{ row }">
              <span
                :class="[
                  'inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-xs font-medium',
                  row.prize_type === 'balance'
                    ? 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400'
                    : 'bg-indigo-50 text-indigo-700 dark:bg-indigo-950/40 dark:text-indigo-400'
                ]"
              >
                <Icon :name="row.prize_type === 'balance' ? 'bolt' : 'creditCard'" size="xs" />
                {{ row.prize_type === 'balance' ? t('lottery.balancePrize') : subscriptionPrizeText(row) }}
              </span>
            </template>

            <template #cell-probability="{ row }">
              <span class="font-mono text-sm font-semibold tabular-nums text-gray-900 dark:text-gray-100">
                {{ probability(row.probability_ppm) }}
              </span>
            </template>

            <template #cell-stock="{ row }">
              <span class="text-sm text-gray-600 dark:text-gray-300">
                {{ row.stock_total == null ? t('lotteryAdmin.unlimited') : `${row.stock_used} / ${row.stock_total}` }}
              </span>
            </template>

            <template #cell-status="{ row }">
              <span
                :class="[
                  'inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium',
                  row.enabled
                    ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400'
                    : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'
                ]"
              >
                <span :class="['h-1.5 w-1.5 rounded-full', row.enabled ? 'bg-emerald-500' : 'bg-gray-400']"></span>
                {{ row.enabled ? t('common.enabled') : t('common.disabled') }}
              </span>
            </template>

            <template #cell-actions="{ row }">
              <div class="flex items-center justify-end space-x-1">
                <button
                  type="button"
                  class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-gray-200"
                  :title="t('lotteryAdmin.editPrize')"
                  @click="openPrize(row)"
                >
                  <Icon name="edit" size="sm" />
                </button>
                <button
                  type="button"
                  class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:text-gray-400 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                  :title="t('lotteryAdmin.delete')"
                  @click="confirmDelete('prize', row.id, row.name)"
                >
                  <Icon name="trash" size="sm" />
                </button>
              </div>
            </template>
          </DataTable>
        </div>

        <!-- rules -->
        <div
          v-show="tab === 'rules'"
          class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800"
        >
          <DataTable :columns="ruleColumns" :data="rules">
            <template #cell-name="{ row }">
              <span class="font-medium text-gray-900 dark:text-white">{{ row.name }}</span>
            </template>

            <template #cell-event_type="{ row }">
              <span class="inline-flex items-center rounded-md bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700 dark:bg-dark-700 dark:text-gray-300">
                {{ t(`lotteryAdmin.${row.event_type}`) }}
              </span>
            </template>

            <template #cell-beneficiary="{ row }">
              <span class="text-sm text-gray-700 dark:text-gray-300">
                {{ t(`lotteryAdmin.${row.beneficiary}`) }}
              </span>
            </template>

            <template #cell-normal_chances="{ row }">
              <span class="font-mono text-sm font-semibold tabular-nums text-blue-600 dark:text-blue-400">
                +{{ row.normal_chances }}
              </span>
            </template>

            <template #cell-luxury_chances="{ row }">
              <span class="font-mono text-sm font-semibold tabular-nums text-purple-600 dark:text-purple-400">
                +{{ row.luxury_chances }}
              </span>
            </template>

            <template #cell-status="{ row }">
              <span
                :class="[
                  'inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium',
                  row.enabled
                    ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400'
                    : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'
                ]"
              >
                <span :class="['h-1.5 w-1.5 rounded-full', row.enabled ? 'bg-emerald-500' : 'bg-gray-400']"></span>
                {{ row.enabled ? t('common.enabled') : t('common.disabled') }}
              </span>
            </template>

            <template #cell-actions="{ row }">
              <div class="flex items-center justify-end space-x-1">
                <button
                  type="button"
                  class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-gray-200"
                  :title="t('lotteryAdmin.editRule')"
                  @click="openRule(row)"
                >
                  <Icon name="edit" size="sm" />
                </button>
                <button
                  type="button"
                  class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:text-gray-400 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                  :title="t('lotteryAdmin.delete')"
                  @click="confirmDelete('rule', row.id, row.name)"
                >
                  <Icon name="trash" size="sm" />
                </button>
              </div>
            </template>
          </DataTable>
        </div>

        <!-- records -->
        <div
          v-show="tab === 'records'"
          class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800"
        >
          <DataTable v-if="recordMode === 'draws'" :columns="drawsColumns" :data="draws">
            <template #cell-outcome="{ row }">
              <span
                :class="[
                  'inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-xs font-medium',
                  row.outcome === 'win'
                    ? 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400'
                    : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
                ]"
              >
                <Icon :name="row.outcome === 'win' ? 'gift' : 'xCircle'" size="xs" />
                {{ row.outcome === 'win' ? (row.prize?.name || t('lottery.won')) : t('lottery.noPrize') }}
              </span>
            </template>

            <template #cell-pool_key="{ row }">
              <span class="inline-flex items-center rounded-md bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700 dark:bg-dark-700 dark:text-gray-300">
                {{ row.pool_key === 'normal' ? t('lottery.normal') : t('lottery.luxury') }}
              </span>
            </template>

            <template #cell-created_at="{ value }">
              <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
            </template>
          </DataTable>

          <DataTable v-else :columns="ledgerColumns" :data="ledger">
            <template #cell-source="{ row }">
              <span class="text-sm text-gray-700 dark:text-gray-300">
                {{ row.source_type }} / {{ row.source_id }}
              </span>
            </template>

            <template #cell-delta="{ row }">
              <span
                :class="[
                  'font-mono text-sm font-semibold tabular-nums',
                  (row.base_delta || row.extra_delta) > 0
                    ? 'text-emerald-600 dark:text-emerald-400'
                    : 'text-red-600 dark:text-red-400'
                ]"
              >
                {{ (row.base_delta || row.extra_delta) > 0 ? '+' : '' }}{{ row.base_delta || row.extra_delta }}
              </span>
            </template>

            <template #cell-created_at="{ value }">
              <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
            </template>
          </DataTable>
        </div>
      </div>
    </div>

    <LotteryPrizeDialog
      :show="prizeDialog.show"
      :prize="prizeDialog.item"
      :pool-id="selectedPoolId"
      :plans="subscriptionPlans"
      :saving="prizeDialog.saving"
      @cancel="closePrize"
      @save="savePrize"
    />
    <LotteryRuleDialog
      :show="ruleDialog.show"
      :rule="ruleDialog.item"
      :saving="ruleDialog.saving"
      @cancel="closeRule"
      @save="saveRule"
    />
    <ConfirmDialog
      :show="deleteDialog.show"
      :title="t('lotteryAdmin.delete')"
      :message="deleteDialog.name"
      danger
      @cancel="deleteDialog.show = false"
      @confirm="performDelete"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Select from '@/components/common/Select.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import adminLotteryAPI from '@/api/admin/lottery'
import { adminPaymentAPI } from '@/api/admin/payment'
import type { SubscriptionPlan } from '@/types/payment'
import type {
  LotteryChanceLedgerEntry,
  LotteryDraw,
  LotteryPool,
  LotteryPoolInput,
  LotteryPoolKey,
  LotteryPrize,
  LotteryPrizeInput,
  LotteryRule,
  LotteryRuleInput
} from '@/types/lottery'
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

const tabIcons: Record<TabKey, 'gift' | 'sparkles' | 'cog' | 'document'> = {
  pools: 'gift',
  prizes: 'sparkles',
  rules: 'cog',
  records: 'document',
}

const tabs = computed(() => [
  { key: 'pools' as const, label: t('lotteryAdmin.pools') },
  { key: 'prizes' as const, label: t('lotteryAdmin.prizes') },
  { key: 'rules' as const, label: t('lotteryAdmin.rules') },
  { key: 'records' as const, label: t('lotteryAdmin.records') },
])

const prizeColumns = computed(() => [
  { key: 'name', label: t('lotteryAdmin.name') },
  { key: 'prize_type', label: t('lotteryAdmin.prizeType') },
  { key: 'probability', label: t('lotteryAdmin.probability') },
  { key: 'stock', label: t('lotteryAdmin.stock') },
  { key: 'status', label: t('lotteryAdmin.enabled') },
  { key: 'actions', label: '', class: 'w-24 text-right' }
])

const ruleColumns = computed(() => [
  { key: 'name', label: t('lotteryAdmin.name') },
  { key: 'event_type', label: t('lotteryAdmin.event') },
  { key: 'beneficiary', label: t('lotteryAdmin.beneficiary') },
  { key: 'normal_chances', label: t('lottery.normal') },
  { key: 'luxury_chances', label: t('lottery.luxury') },
  { key: 'status', label: t('lotteryAdmin.enabled') },
  { key: 'actions', label: '', class: 'w-24 text-right' }
])

const drawsColumns = computed(() => [
  { key: 'id', label: 'ID' },
  { key: 'user_id', label: t('lotteryAdmin.userId') },
  { key: 'outcome', label: t('lotteryAdmin.outcome') },
  { key: 'pool_key', label: t('lotteryAdmin.name') },
  { key: 'created_at', label: t('lotteryAdmin.time') }
])

const ledgerColumns = computed(() => [
  { key: 'id', label: 'ID' },
  { key: 'user_id', label: t('lotteryAdmin.userId') },
  { key: 'source', label: t('lotteryAdmin.source') },
  { key: 'delta', label: t('lotteryAdmin.delta') },
  { key: 'created_at', label: t('lotteryAdmin.time') }
])

const probabilityTotal = computed(() =>
  Number((prizes.value.filter((item) => item.enabled).reduce((sum, item) => sum + item.probability_ppm, 0) / 10_000).toFixed(4))
)

async function load(): Promise<void> {
  loading.value = true
  try {
    const [poolItems, ruleItems, planRes] = await Promise.all([
      adminLotteryAPI.listPools(),
      adminLotteryAPI.listRules(),
      adminPaymentAPI.getPlans()
    ])
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

async function loadRules(): Promise<void> {
  rules.value = await adminLotteryAPI.listRules()
}

async function selectPool(id: number): Promise<void> {
  selectedPoolId.value = id
  await loadPrizes()
}

function probability(ppm: number): string {
  return `${Number((ppm / 10_000).toFixed(4))}%`
}

function subscriptionPrizeText(prize: LotteryPrize): string {
  const plan = subscriptionPlans.value.find(
    (item) => item.group_id === prize.group_id && item.validity_days === prize.validity_days
  )
  return plan?.name ?? t('lottery.subscriptionPrize', { days: prize.validity_days ?? 0 })
}

async function savePool(key: LotteryPoolKey, input: LotteryPoolInput): Promise<void> {
  savingPool.value = key
  try {
    const saved = await adminLotteryAPI.updatePool(key, input)
    pools.value = pools.value.map((pool) => (pool.key === key ? saved : pool))
    appStore.showSuccess(t('lotteryAdmin.saveSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('lotteryAdmin.operationFailed')))
  } finally {
    savingPool.value = ''
  }
}

function openPrize(item: LotteryPrize | null = null): void {
  prizeDialog.item = item
  prizeDialog.show = true
}

function closePrize(): void {
  if (!prizeDialog.saving) prizeDialog.show = false
}

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
    await loadPrizes()
    prizeDialog.show = false
    appStore.showSuccess(t('lotteryAdmin.saveSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('lotteryAdmin.operationFailed')))
  } finally {
    prizeDialog.saving = false
  }
}

function openRule(item: LotteryRule | null = null): void {
  ruleDialog.item = item
  ruleDialog.show = true
}

function closeRule(): void {
  if (!ruleDialog.saving) ruleDialog.show = false
}

async function saveRule(input: LotteryRuleInput): Promise<void> {
  ruleDialog.saving = true
  try {
    if (ruleDialog.item) await adminLotteryAPI.updateRule(ruleDialog.item.id, input)
    else await adminLotteryAPI.createRule(input)
    rules.value = await adminLotteryAPI.listRules()
    ruleDialog.show = false
    appStore.showSuccess(t('lotteryAdmin.saveSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('lotteryAdmin.operationFailed')))
  } finally {
    ruleDialog.saving = false
  }
}

function confirmDelete(kind: 'prize' | 'rule', id: number, name: string): void {
  Object.assign(deleteDialog, { show: true, kind, id, name })
}

async function performDelete(): Promise<void> {
  try {
    if (deleteDialog.kind === 'prize') {
      await adminLotteryAPI.deletePrize(deleteDialog.id)
      await loadPrizes()
    } else {
      await adminLotteryAPI.deleteRule(deleteDialog.id)
      rules.value = await adminLotteryAPI.listRules()
    }
    deleteDialog.show = false
    appStore.showSuccess(t('lotteryAdmin.deleteSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('lotteryAdmin.operationFailed')))
  }
}

async function loadRecords(): Promise<void> {
  const params = {
    page: 1,
    page_size: 50,
    user_id: recordUserId.value || undefined,
    pool: recordPool.value || undefined
  }
  try {
    if (recordMode.value === 'draws') draws.value = (await adminLotteryAPI.listDraws(params)).items ?? []
    else ledger.value = (await adminLotteryAPI.listLedger(params)).items ?? []
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('lotteryAdmin.operationFailed')))
  }
}

async function switchRecords(mode: 'draws' | 'ledger'): Promise<void> {
  recordMode.value = mode
  await loadRecords()
}

watch(tab, (value) => {
  if (value === 'records') void loadRecords()
})

onMounted(load)
</script>
