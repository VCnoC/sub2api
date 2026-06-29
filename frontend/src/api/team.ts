/**
 * Team API endpoints
 * Handles team creation, invitation, membership, and balance transfer
 */

import { apiClient } from './client'
import type { Team, TeamMember, UsageLog, PaginationParams, PaginatedResponse } from '@/types'

/**
 * Create a new team
 * @param name - Team name
 * @returns Created team information
 */
export async function createTeam(name: string): Promise<Team> {
  const { data } = await apiClient.post<Team>('/user/team', { name })
  return data
}

/**
 * Get current user's team information
 * @returns Team info with role, or null if not in a team
 */
export async function getMyTeam(): Promise<Team | null> {
  const { data } = await apiClient.get<Team | null>('/user/team')
  return data
}

/**
 * Refresh the team invite code
 * @returns New invite code
 */
export async function refreshInviteCode(): Promise<{ invite_code: string }> {
  const { data } = await apiClient.post<{ invite_code: string }>('/user/team/invite-code')
  return data
}

/**
 * Join a team by invite code
 * @param inviteCode - Team invite code
 */
export async function joinTeam(inviteCode: string): Promise<void> {
  await apiClient.post('/user/team/join', { invite_code: inviteCode })
}

/**
 * Leave current team (members only)
 */
export async function leaveTeam(): Promise<void> {
  await apiClient.post('/user/team/leave')
}

/**
 * Remove a member from the team (owner only)
 * @param memberId - Member user ID
 */
export async function removeMember(memberId: number): Promise<void> {
  await apiClient.delete(`/user/team/members/${memberId}`)
}

/**
 * List team members with balance and usage (owner only)
 * @param params - Pagination params
 */
export async function listMembers(params: PaginationParams): Promise<PaginatedResponse<TeamMember>> {
  const { data } = await apiClient.get<PaginatedResponse<TeamMember>>('/user/team/members', {
    params
  })
  return data
}

/**
 * Transfer balance from owner to a member
 * @param memberId - Member user ID
 * @param amount - Amount to transfer
 * @param password - Owner login password for confirmation
 */
export async function transferBalance(
  memberId: number,
  amount: number,
  password: string
): Promise<void> {
  await apiClient.post(`/user/team/members/${memberId}/transfer`, {
    amount,
    password
  })
}

export interface MemberUsageParams extends PaginationParams {
  start_date: string
  end_date: string
}

/**
 * List usage logs for a team member within a date range (owner only)
 * @param memberId - Member user ID
 * @param params - Date range and pagination params
 */
export async function getMemberUsage(
  memberId: number,
  params: MemberUsageParams
): Promise<PaginatedResponse<UsageLog>> {
  const { data } = await apiClient.get<PaginatedResponse<UsageLog>>(`/user/team/members/${memberId}/usage`, {
    params
  })
  return data
}
