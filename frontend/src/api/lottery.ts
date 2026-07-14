/** 登录用户的抽奖摘要、执行和历史 API。 */

import { apiClient } from './client'
import type { PaginatedResponse } from '@/types'
import type { LotteryDraw, LotteryPoolKey, LotterySummary } from '@/types/lottery'

function requestKey(): string {
  return globalThis.crypto?.randomUUID?.() ?? `lottery-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

export async function getLotterySummary(): Promise<LotterySummary> {
  const { data } = await apiClient.get<LotterySummary>('/lottery')
  return data
}

export async function drawLottery(pool: LotteryPoolKey): Promise<LotteryDraw> {
  const { data } = await apiClient.post<LotteryDraw>(
    `/lottery/pools/${pool}/draw`,
    {},
    { headers: { 'Idempotency-Key': requestKey() } },
  )
  return data
}

export async function listLotteryHistory(
  page = 1,
  pageSize = 20,
  pool?: LotteryPoolKey,
): Promise<PaginatedResponse<LotteryDraw>> {
  const { data } = await apiClient.get<PaginatedResponse<LotteryDraw>>('/lottery/history', {
    params: { page, page_size: pageSize, pool },
  })
  return data
}

export const lotteryAPI = {
  summary: getLotterySummary,
  draw: drawLottery,
  history: listLotteryHistory,
}

export default lotteryAPI
