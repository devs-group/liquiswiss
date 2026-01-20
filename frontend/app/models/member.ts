import type { OrganisationRoleType } from '~/utils/types'

export interface MemberPermission {
  id: number
  userId: number
  organisationId: number
  entityType: string | null
  canView: boolean
  canEdit: boolean
  canDelete: boolean
  createdAt: string
  updatedAt?: string
}

export interface OrganisationMemberResponse {
  userId: number
  name: string
  email: string
  role: OrganisationRoleType
  isDefault: boolean
  permission?: MemberPermission
}

export interface UpdateMemberFormData {
  role?: string
  canView?: boolean
  canEdit?: boolean
  canDelete?: boolean
}
