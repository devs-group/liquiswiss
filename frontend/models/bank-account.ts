import type { CurrencyResponse } from '~/models/currency'

export interface BankAccountResponse {
  id: number
  name: string
  amount: number
  currency: CurrencyResponse
}

export interface BankAccountFormData {
  id: number
  name: string
  amount: number
  currency: number
}
