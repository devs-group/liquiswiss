import type { CategoryResponse } from '~/models/category'
import type { CurrencyResponse } from '~/models/currency'
import type { PaginationResponse } from '~/models/pagination'
import type { VatResponse } from '~/models/vat'
import type { TransactionCycleTypeToStringDefinition } from '~/utils/enum-helper'

export interface TransactionResponse {
  id: number
  name: string
  amount: number
  vat: VatResponse | null
  vatAmount: number
  vatIncluded: boolean
  cycle: TransactionCycleTypeToStringDefinition | null
  type: TransactionTypeToStringDefinition
  startDate: string
  endDate: string | null
  nextExecutionDate: string | null
  category: CategoryResponse
  currency: CurrencyResponse
  employee: TransactionEmployeeResponse | null
}

export interface TransactionEmployeeResponse {
  id: number
  name: string
}

export interface ListTransactionResponse {
  data: TransactionResponse[]
  pagination: PaginationResponse
}

export interface TransactionFormData {
  id: number
  name: string
  amount: number
  vat?: number
  vatIncluded: boolean
  cycle?: TransactionCycleTypeToStringDefinition
  type: TransactionTypeToStringDefinition
  startDate: Date
  endDate?: Date
  category: number
  currency: number
  employee: number
}
