import type { PaginationResponse } from '~/models/pagination'
import type { CurrencyResponse } from '~/models/currency'
import type {
  CostCycleTypeToStringDefinition,
  CycleTypeToStringDefinition,
  EmployeeCostDistributionTypeToStringDefinition,
  EmployeeCostTypeToStringDefinition,
} from '~/utils/enum-helper'

// Response
export interface EmployeeResponse {
  id: number
  name: string
  hoursPerMonth: number | null
  salary: number | null
  cycle: CycleTypeToStringDefinition | null
  currency: CurrencyResponse | null
  vacationDaysPerYear?: number | null
  fromDate?: string | null
  toDate?: string | null
  isInFuture: boolean
  historyID: number | null
}

export interface EmployeeHistoryCostLabelResponse {
  id: number
  name: string
}

export interface EmployeeHistoryCostDetailResponse {
  id: number
  month: string
  Amount: number
  Divider: number
  CostID: number
}

export interface EmployeeHistoryCostResponse {
  id: number
  label: EmployeeHistoryCostLabelResponse | null
  cycle: CostCycleTypeToStringDefinition
  amountType: EmployeeCostTypeToStringDefinition
  amount: number
  distributionType: EmployeeCostDistributionTypeToStringDefinition
  relativeOffset: number
  targetDate: string | null
  employeeHistoryID: number
  calculatedAmount: number
  calculatedPreviousExecutionDate: Date
  calculatedNextExecutionDate: Date
  calculatedNextCost: number
  calculatedCostDetails: EmployeeHistoryCostDetailResponse[]
}

export interface EmployeeHistoryResponse {
  id: number
  employeeID: number
  hoursPerMonth: number
  salary: number
  cycle: CycleTypeToStringDefinition
  currency: CurrencyResponse
  vacationDaysPerYear: number
  fromDate: string
  toDate: string | null
  withSeparateCosts: boolean
  nextExecutionDate: string | null
  employeeDeductions: number
  employerCosts: number
}

// Form
export interface EmployeeFormData {
  id: number
  name: string
}

export interface EmployeeHistoryCostLabelFormData {
  id: number
  name: string
}

export interface EmployeeHistoryCostFormData {
  id: number
  labelID?: number
  cycle: CostCycleTypeToStringDefinition
  amountType: EmployeeCostTypeToStringDefinition
  amount: number
  relativeOffset: number
  distributionType: EmployeeCostDistributionTypeToStringDefinition
  targetDate?: Date
}

export interface EmployeeHistoryCostCopyFormData {
  ids: number[]
}

export interface EmployeeHistoryPUTFormData {
  id: number
  hoursPerMonth: number
  salary: number
  cycle: CycleTypeToStringDefinition
  currencyID: number
  vacationDaysPerYear: number
  fromDate: Date
  withSeparateCosts: boolean
}

export interface EmployeeHistoryPATCHFormData {
  id: number
  hoursPerMonth?: number
  salary?: number
  cycle: CycleTypeToStringDefinition
  currencyID?: number
  vacationDaysPerYear?: number
  fromDate?: Date
  withSeparateCosts?: boolean
}

// List Response
export interface ListEmployeeResponse {
  data: EmployeeResponse[]
  pagination: PaginationResponse
}

export interface ListEmployeeHistoryCostResponse {
  data: EmployeeHistoryCostResponse[]
  pagination: PaginationResponse
}

export interface ListEmployeeHistoryCostLabelResponse {
  data: EmployeeHistoryCostLabelResponse[]
  pagination: PaginationResponse
}

export interface ListEmployeeHistoryResponse {
  data: EmployeeHistoryResponse[]
  pagination: PaginationResponse
}
