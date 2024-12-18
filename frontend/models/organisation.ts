import type { OrganisationRoleType } from '~/utils/types'
import type { PaginationResponse } from '~/models/pagination'

export interface OrganisationResponse {
  id: number
  name: string
  memberCount: number
  role: OrganisationRoleType
  isDefault: boolean
}

export interface ListOrganisationResponse {
  data: OrganisationResponse[]
  pagination: PaginationResponse
}

export interface OrganisationFormData {
  name: string
}
