/** 管理端奖池、奖品、规则与审计 API。 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'
import type {
  LotteryChanceLedgerEntry,
  LotteryDraw,
  LotteryPool,
  LotteryPoolInput,
  LotteryPoolKey,
  LotteryPrize,
  LotteryPrizeInput,
  LotteryRule,
  LotteryRuleInput,
} from '@/types/lottery'

function key(scope: string): string {
  return `${scope}-${globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`}`
}

const writeConfig = (scope: string) => ({ headers: { 'Idempotency-Key': key(scope) } })

export async function listPools(): Promise<LotteryPool[]> {
  const { data } = await apiClient.get<LotteryPool[]>('/admin/lottery/pools')
  return data
}

export async function updatePool(pool: LotteryPoolKey, input: LotteryPoolInput): Promise<LotteryPool> {
  const { data } = await apiClient.patch<LotteryPool>(`/admin/lottery/pools/${pool}`, input, writeConfig(`pool-${pool}`))
  return data
}

export async function listPrizes(poolId: number): Promise<LotteryPrize[]> {
  const { data } = await apiClient.get<LotteryPrize[]>('/admin/lottery/prizes', { params: { pool_id: poolId } })
  return data
}

export async function createPrize(input: LotteryPrizeInput): Promise<LotteryPrize> {
  const { data } = await apiClient.post<LotteryPrize>('/admin/lottery/prizes', input, writeConfig('prize-create'))
  return data
}

export async function updatePrize(id: number, input: LotteryPrizeInput): Promise<LotteryPrize> {
  const { data } = await apiClient.patch<LotteryPrize>(`/admin/lottery/prizes/${id}`, input, writeConfig(`prize-${id}`))
  return data
}

export async function deletePrize(id: number): Promise<void> {
  await apiClient.delete(`/admin/lottery/prizes/${id}`, writeConfig(`prize-delete-${id}`))
}

export async function listRules(): Promise<LotteryRule[]> {
  const { data } = await apiClient.get<LotteryRule[]>('/admin/lottery/rules')
  return data
}

export async function createRule(input: LotteryRuleInput): Promise<LotteryRule> {
  const { data } = await apiClient.post<LotteryRule>('/admin/lottery/rules', input, writeConfig('rule-create'))
  return data
}

export async function updateRule(id: number, input: LotteryRuleInput): Promise<LotteryRule> {
  const { data } = await apiClient.patch<LotteryRule>(`/admin/lottery/rules/${id}`, input, writeConfig(`rule-${id}`))
  return data
}

export async function deleteRule(id: number): Promise<void> {
  await apiClient.delete(`/admin/lottery/rules/${id}`, writeConfig(`rule-delete-${id}`))
}

export async function listDraws(params: Record<string, unknown> = {}): Promise<PaginatedResponse<LotteryDraw>> {
  const { data } = await apiClient.get<PaginatedResponse<LotteryDraw>>('/admin/lottery/draws', { params })
  return data
}

export async function listLedger(params: Record<string, unknown> = {}): Promise<PaginatedResponse<LotteryChanceLedgerEntry>> {
  const { data } = await apiClient.get<PaginatedResponse<LotteryChanceLedgerEntry>>('/admin/lottery/chance-ledger', { params })
  return data
}

export const adminLotteryAPI = {
  listPools, updatePool, listPrizes, createPrize, updatePrize, deletePrize,
  listRules, createRule, updateRule, deleteRule, listDraws, listLedger,
}

export default adminLotteryAPI
