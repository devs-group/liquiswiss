export interface ForecastResponse {
  data: {
    month: string
    revenue: number
    expense: number
    cashflow: number
  }
  updatedAt: string
}

export interface ForecastDetailRevenueExpenseResponse {
  name: string
  amount: number
  // Kept for forecast details that were stored before relatedIDs existed
  relatedID: number
  // Every entity merged into this row, e.g. two transactions sharing a name
  relatedIDs?: number[]
  relatedTable: string
  // Only true when every entity of the row is excluded
  isExcluded: boolean
  children?: ForecastDetailRevenueExpenseResponse[]
}

export interface ForecastDetailResponse {
  month: string
  revenue: ForecastDetailRevenueExpenseResponse[]
  expense: ForecastDetailRevenueExpenseResponse[]
  forecastID: number
}
