import type { MemberRoleType } from '~/utils/types'

export interface InvitationResponse {
  id: number
  organisationId: number
  email: string
  role: MemberRoleType
  invitedBy: number
  invitedByName: string
  expiresAt: string
  createdAt: string
}

export interface CreateInvitationFormData {
  email: string
  role: MemberRoleType
}

export interface CheckInvitationResponse {
  email: string
  organisationName: string
  invitedByName: string
  existingUser: boolean
}

export interface AcceptInvitationFormData {
  token: string
  password?: string
}
