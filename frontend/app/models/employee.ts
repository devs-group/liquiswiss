import type { PaginationResponse } from '~/models/pagination'
import type { CurrencyResponse } from '~/models/currency'
import type {
  CostCycleTypeToStringDefinition,
  EmployeeCostDistributionTypeToStringDefinition,
  EmployeeCostTypeToStringDefinition,
  SalaryCycleTypeToStringDefinition,
} from '~/utils/enum-helper'

// Response
export interface EmployeeResponse {
  id: number
  name: string
  hoursPerMonth: number | null
  salaryAmount: number | null
  cycle: SalaryCycleTypeToStringDefinition | null
  currency: CurrencyResponse | null
  vacationDaysPerYear?: number | null
  fromDate?: string | null
  toDate?: string | null
  isInFuture: boolean
  withSeparateCosts: boolean
  isTerminated: boolean
  willBeTerminated: boolean
  salaryID: number | null
}

export interface SalaryCostLabelResponse {
  id: number
  name: string
}

export interface SalaryCostDetailResponse {
  id: number
  month: string
  Amount: number
  Divider: number
  CostID: number
}

export interface SalaryCostResponse {
  id: number
  label: SalaryCostLabelResponse | null
  cycle: CostCycleTypeToStringDefinition
  amountType: EmployeeCostTypeToStringDefinition
  amount: number
  distributionType: EmployeeCostDistributionTypeToStringDefinition
  relativeOffset: number
  targetDate: string | null
  salaryID: number
  baseSalaryCostIDs: number[]
  calculatedAmount: number
  calculatedPreviousExecutionDate: Date
  calculatedNextExecutionDate: Date
  calculatedNextCost: number
  calculatedCostDetails: SalaryCostDetailResponse[]
}

export interface SalaryResponse {
  id: number
  employeeID: number
  hoursPerMonth: number
  amount: number
  cycle: SalaryCycleTypeToStringDefinition
  currency: CurrencyResponse
  vacationDaysPerYear: number
  fromDate: string
  toDate: string | null
  withSeparateCosts: boolean
  hasSeparateCostsDefined: boolean
  isTermination: boolean
  nextExecutionDate: string | null
  employeeDeductions: number
  employerCosts: number
  isDisabled: boolean
}

// Form
export interface EmployeeFormData {
  id: number
  name: string
}

export interface SalaryCostLabelFormData {
  id: number
  name: string
}

export interface SalaryCostFormData {
  id: number
  labelID?: number
  cycle: CostCycleTypeToStringDefinition
  amountType: EmployeeCostTypeToStringDefinition
  amount: number
  relativeOffset: number
  distributionType: EmployeeCostDistributionTypeToStringDefinition
  targetDate?: Date
  baseSalaryCostIDs?: number[]
}

export interface SalaryCostCopyFormData {
  ids: number[]
  sourceSalaryID?: number
}

export interface SalaryPUTFormData {
  id: number
  hoursPerMonth: number
  amount: number
  cycle: SalaryCycleTypeToStringDefinition
  currencyID: number
  vacationDaysPerYear: number
  fromDate: Date
  withSeparateCosts: boolean
  isTermination: boolean
}

export interface SalaryPATCHFormData {
  id: number
  hoursPerMonth?: number
  amount?: number
  cycle: SalaryCycleTypeToStringDefinition
  currencyID?: number
  vacationDaysPerYear?: number
  fromDate?: Date
  withSeparateCosts?: boolean
  isDisabled?: boolean
}

// List Response
export interface ListEmployeeResponse {
  data: EmployeeResponse[]
  pagination: PaginationResponse
}

export interface ListSalaryCostResponse {
  data: SalaryCostResponse[]
  pagination: PaginationResponse
}

export interface ListSalaryCostLabelResponse {
  data: SalaryCostLabelResponse[]
  pagination: PaginationResponse
}

export interface ListSalaryResponse {
  data: SalaryResponse[]
  pagination: PaginationResponse
}
