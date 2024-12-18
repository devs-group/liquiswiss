import type { PaginationResponse } from '~/models/pagination'
import type { CurrencyResponse } from '~/models/currency'

export interface EmployeeFormData {
  id: number
  name: string
}

export interface EmployeeResponse {
  id: number
  name: string
  hoursPerMonth: number | null
  salaryPerMonth: number | null
  currency: CurrencyResponse | null
  vacationDaysPerYear?: number | null
  fromDate?: string | null
  toDate?: string | null
  isInFuture: boolean
  historyID: number | null
}

export interface ListEmployeeResponse {
  data: EmployeeResponse[]
  pagination: PaginationResponse
}

export interface EmployeeHistoryFormData {
  id: number
  hoursPerMonth: number
  salaryPerMonth: number
  currencyID: number
  vacationDaysPerYear: number
  fromDate: Date
  toDate?: Date
}

export interface EmployeeHistoryResponse {
  id: number
  employeeID: string
  hoursPerMonth: number
  salaryPerMonth: number
  currency: CurrencyResponse
  vacationDaysPerYear: number
  fromDate: string
  toDate: string | null
}

export interface ListEmployeeHistoryResponse {
  data: EmployeeHistoryResponse[]
  pagination: PaginationResponse
}
