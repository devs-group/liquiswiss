export interface VatSettingFormData {
  enabled: boolean
  billingDate: string // YYYY-MM-DD format - Rechnungszeitpunkt
  transactionDate: string // YYYY-MM-DD format - Transaktionszeitpunkt
  interval: 'monthly' | 'quarterly' | 'biannually' | 'yearly'
}

export interface VatSettingResponse {
  id: number
  organisationId: number
  enabled: boolean
  billingDate: string
  transactionDate: string
  interval: 'monthly' | 'quarterly' | 'biannually' | 'yearly'
  createdAt: string
  updatedAt: string
}
