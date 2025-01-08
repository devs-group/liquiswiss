import type { OrganisationRoleType } from '~/utils/types'
import type { PaginationResponse } from '~/models/pagination'
import type { CurrencyResponse } from '~/models/currency'

export interface OrganisationResponse {
  id: number
  name: string
  currency: CurrencyResponse
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
  currencyID: number
}
