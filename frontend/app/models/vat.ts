export interface VatFormData {
  id?: number
  value: number | string
}

export interface VatResponse {
  id: number
  value: number
  formattedValue: string
  canEdit: boolean
}
