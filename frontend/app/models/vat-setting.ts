export interface VatSettingFormData {
  enabled: boolean
  billingDate: string // YYYY-MM-DD format - Rechnungszeitpunkt
  transactionMonthOffset: number // Months after billing date (0-12)
  interval: 'monthly' | 'quarterly' | 'biannually' | 'yearly'
}

export interface VatSettingResponse {
  id: number
  organisationId: number
  enabled: boolean
  billingDate: string
  transactionMonthOffset: number
  interval: 'monthly' | 'quarterly' | 'biannually' | 'yearly'
  createdAt: string
  updatedAt: string
}
