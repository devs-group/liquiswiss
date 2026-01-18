export type DisplayType = 'grid' | 'list'
export type SortOrderType = 'ASC' | 'DESC'

export interface UserOrganisationSettingResponse {
  id: number
  userId: number
  organisationId: number
  forecastMonths: number
  forecastPerformance: number
  forecastRevenueDetails: boolean
  forecastExpenseDetails: boolean
  forecastChildDetails: string[]
  employeeDisplay: DisplayType
  employeeSortBy: string
  employeeSortOrder: SortOrderType
  employeeHideTerminated: boolean
  transactionDisplay: DisplayType
  transactionSortBy: string
  transactionSortOrder: SortOrderType
  transactionHideDisabled: boolean
  bankAccountDisplay: DisplayType
  bankAccountSortBy: string
  bankAccountSortOrder: SortOrderType
  createdAt: string
  updatedAt: string
}

export interface UpdateUserOrganisationSetting {
  forecastMonths?: number
  forecastPerformance?: number
  forecastRevenueDetails?: boolean
  forecastExpenseDetails?: boolean
  forecastChildDetails?: string[]
  employeeDisplay?: DisplayType
  employeeSortBy?: string
  employeeSortOrder?: SortOrderType
  employeeHideTerminated?: boolean
  transactionDisplay?: DisplayType
  transactionSortBy?: string
  transactionSortOrder?: SortOrderType
  transactionHideDisabled?: boolean
  bankAccountDisplay?: DisplayType
  bankAccountSortBy?: string
  bankAccountSortOrder?: SortOrderType
}
