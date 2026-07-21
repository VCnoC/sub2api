/** 管理端团队审核、配置和治理 API。 */

import { apiClient } from '../client'
import type {
  AdminTeamDetail,
  AdminTeamSummary,
  PaginatedResponse,
  TeamAdminStats,
  TeamApplication,
  TeamGovernanceSettings,
} from '@/types'

export async function getStats(): Promise<TeamAdminStats> {
  const { data } = await apiClient.get<TeamAdminStats>('/admin/teams/stats')
  return data
}

export async function listTeams(params: Record<string, unknown>): Promise<PaginatedResponse<AdminTeamSummary>> {
  const { data } = await apiClient.get<PaginatedResponse<AdminTeamSummary>>('/admin/teams', { params })
  return data
}

export async function getTeam(id: number): Promise<AdminTeamDetail> {
  const { data } = await apiClient.get<AdminTeamDetail>(`/admin/teams/${id}`)
  return data
}

export async function listApplications(params: Record<string, unknown>): Promise<PaginatedResponse<TeamApplication>> {
  const { data } = await apiClient.get<PaginatedResponse<TeamApplication>>('/admin/teams/applications', { params })
  return data
}

export async function reviewApplication(id: number, input: { approve: boolean; review_reason: string; waive: boolean; target_limit?: number }): Promise<TeamApplication> {
  const { data } = await apiClient.post<TeamApplication>(`/admin/teams/applications/${id}/review`, input)
  return data
}

export async function getSettings(): Promise<TeamGovernanceSettings> {
  const { data } = await apiClient.get<TeamGovernanceSettings>('/admin/teams/settings')
  return data
}

export async function updateSettings(input: TeamGovernanceSettings): Promise<TeamGovernanceSettings> {
  const { data } = await apiClient.put<TeamGovernanceSettings>('/admin/teams/settings', input)
  return data
}

export async function setStatus(id: number, status: 'active' | 'frozen'): Promise<void> {
  await apiClient.put(`/admin/teams/${id}/status`, { status })
}

export async function setMemberLimit(id: number, memberLimit: number): Promise<void> {
  await apiClient.put(`/admin/teams/${id}/member-limit`, { member_limit: memberLimit })
}

export async function markReviewed(id: number): Promise<void> {
  await apiClient.post(`/admin/teams/${id}/review-complete`)
}

export async function removeMember(teamId: number, memberId: number): Promise<void> {
  await apiClient.delete(`/admin/teams/${teamId}/members/${memberId}`)
}

export const adminTeamsAPI = {
  getStats, listTeams, getTeam, listApplications, reviewApplication, getSettings,
  updateSettings, setStatus, setMemberLimit, markReviewed, removeMember,
}
