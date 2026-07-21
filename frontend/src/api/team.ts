/**
 * Team API endpoints
 * Handles team creation, invitation, membership, and balance transfer
 */

import { apiClient } from './client'
import type { Team, TeamApplication, TeamCreationEligibility, TeamGovernanceState, TeamJoinRequest, TeamMember, UsageLog, PaginationParams, PaginatedResponse } from '@/types'

/**
 * Create a new team
 * @param name - Team name
 * @returns Created team information
 */
export async function createTeam(name: string, reason = '', additionalInfo = ''): Promise<TeamApplication> {
  const { data } = await apiClient.post<TeamApplication>('/user/team', { name, reason, additional_info: additionalInfo })
  return data
}

export async function getMyCreateApplication(): Promise<TeamApplication | null> {
  const { data } = await apiClient.get<TeamApplication | null>('/user/team/application')
  return data
}

export async function getCreationEligibility(): Promise<TeamCreationEligibility> {
  const { data } = await apiClient.get<TeamCreationEligibility>('/user/team/eligibility')
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
export async function joinTeam(inviteCode: string, message = ''): Promise<TeamJoinRequest> {
  const { data } = await apiClient.post<TeamJoinRequest>('/user/team/join', { invite_code: inviteCode, message })
  return data
}

export async function listJoinRequests(status = 'pending'): Promise<TeamJoinRequest[]> {
  const { data } = await apiClient.get<TeamJoinRequest[]>('/user/team/join-requests', { params: { status } })
  return data
}

export async function reviewJoinRequest(id: number, approve: boolean, reason = ''): Promise<TeamJoinRequest> {
  const { data } = await apiClient.post<TeamJoinRequest>(`/user/team/join-requests/${id}/review`, { approve, reason })
  return data
}

export async function getTeamGovernance(): Promise<TeamGovernanceState> {
  const { data } = await apiClient.get<TeamGovernanceState>('/user/team/governance')
  return data
}

export async function upgradeTeam(): Promise<TeamGovernanceState> {
  const { data } = await apiClient.post<TeamGovernanceState>('/user/team/upgrade')
  return data
}

export async function requestTeamExpansion(targetLimit: number, reason: string): Promise<TeamApplication> {
  const { data } = await apiClient.post<TeamApplication>('/user/team/expand', { target_limit: targetLimit, reason })
  return data
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

/**
 * Deposit current user's balance into the team fund
 * @param amount - Amount to deposit
 * @param password - Login password for confirmation
 */
export async function depositFund(amount: number, password: string): Promise<void> {
  await apiClient.post('/user/team/fund/deposit', { amount, password })
}

/**
 * Allocate team fund to a member (owner only)
 * @param memberId - Member user ID
 * @param amount - Amount to allocate
 * @param password - Owner login password for confirmation
 */
export async function allocateFund(
  memberId: number,
  amount: number,
  password: string
): Promise<void> {
  await apiClient.post(`/user/team/members/${memberId}/allocate`, {
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
