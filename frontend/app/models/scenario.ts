import type { ForecastDetailResponse, ForecastResponse } from '~/models/forecast'
import type { BankAccountResponse } from '~/models/bank-account'

export interface ScenarioResponse {
  id: number
  name: string
  isBase: boolean
  parentId: number | null
  isCurrent: boolean
  createdAt: string
}

// Stored forecast of another scenario, read-only for the side-by-side comparison
export interface ScenarioForecastResponse {
  scenario: ScenarioResponse
  forecasts: ForecastResponse[]
  forecastDetails: ForecastDetailResponse[]
  bankAccounts: BankAccountResponse[]
}

export interface ScenarioFormData {
  id?: number
  name: string
  sourceScenarioId?: number | null
}
