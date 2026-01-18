import type { CurrencyResponse } from '~/models/currency'
import type { ListResponse } from '~/models/list-response'

export interface BankAccountResponse {
  id: number
  name: string
  amount: number
  currency: CurrencyResponse
}

export type ListBankAccountResponse = ListResponse<BankAccountResponse>

export interface BankAccountFormData {
  id: number
  name: string
  amount: number
  currency: number
}
