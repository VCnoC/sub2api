/** 双奖池抽奖的用户与管理端数据结构。 */

export type LotteryPoolKey = 'normal' | 'luxury'
export type LotteryCycleType = 'daily' | 'weekly'
export type LotteryPrizeType = 'balance' | 'subscription'
export type LotteryEventType = 'signup' | 'redeem' | 'recharge'

export interface LotteryPool {
  id: number
  key: LotteryPoolKey
  name: string
  enabled: boolean
  cycle_type: LotteryCycleType
  cycle_chances: number
  starts_at?: string | null
  ends_at?: string | null
  created_at: string
  updated_at: string
}

export interface LotteryPrize {
  id: number
  pool_id: number
  name: string
  description: string
  image_data?: string
  prize_type: LotteryPrizeType
  balance_amount?: number | null
  group_id?: number | null
  validity_days?: number | null
  probability_ppm: number
  stock_total?: number | null
  stock_used: number
  enabled: boolean
  sort_order: number
}

export interface LotteryPoolSummary {
  pool: LotteryPool
  prizes: LotteryPrize[]
  base_remaining: number
  extra_remaining: number
  period_key: string
  active: boolean
}

export interface LotterySummary {
  pools: LotteryPoolSummary[]
}

export interface LotteryPrizeSnapshot {
  id?: number
  name?: string
  description?: string
  prize_type?: LotteryPrizeType
  balance_amount?: number | null
  group_id?: number | null
  validity_days?: number | null
  probability_ppm?: number
}

export interface LotteryDraw {
  id: number
  user_id?: number
  pool_id: number
  pool_key?: LotteryPoolKey
  outcome: 'win' | 'none'
  chance_source: 'base' | 'extra'
  prize_id?: number | null
  prize?: LotteryPrizeSnapshot
  base_remaining?: number
  extra_remaining?: number
  created_at: string
}

export interface LotteryRule {
  id: number
  name: string
  event_type: LotteryEventType
  beneficiary: 'inviter' | 'invitee'
  normal_chances: number
  luxury_chances: number
  recharge_mode?: 'single' | 'cumulative' | null
  recharge_threshold?: number | null
  repeatable: boolean
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface LotteryChanceLedgerEntry {
  id: number
  user_id: number
  pool_id: number
  pool_key?: LotteryPoolKey
  action: 'grant' | 'refund_reversal' | 'draw'
  base_delta: number
  extra_delta: number
  rule_id?: number | null
  source_type: string
  source_id: string
  source_user_id?: number | null
  tier_no: number
  metadata?: Record<string, unknown>
  created_at: string
}

export interface LotteryPoolInput {
  name: string
  enabled: boolean
  cycle_type: LotteryCycleType
  cycle_chances: number
  starts_at?: string | null
  ends_at?: string | null
}

export interface LotteryPrizeInput {
  pool_id: number
  name: string
  description: string
  image_data: string
  prize_type: LotteryPrizeType
  balance_amount?: number | null
  group_id?: number | null
  validity_days?: number | null
  probability_ppm: number
  stock_total?: number | null
  enabled: boolean
  sort_order: number
}

export interface LotteryRuleInput {
  name: string
  event_type: LotteryEventType
  beneficiary: 'inviter' | 'invitee'
  normal_chances: number
  luxury_chances: number
  recharge_mode?: 'single' | 'cumulative' | null
  recharge_threshold?: number | null
  repeatable: boolean
  enabled: boolean
}
