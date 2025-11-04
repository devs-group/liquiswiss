export type ScenarioType = 'horizontal' | 'vertical'

export interface ScenarioResponse {
  id: number
  name: string
  type: ScenarioType
  isDefault: boolean
  parentScenarioId: number | null
  organisationId: number
  createdAt: string
  updatedAt: string
}

export interface ScenarioListItem {
  id: number
  name: string
  type: ScenarioType
  isDefault: boolean
  parentScenarioId: number | null
  createdAt: string
  updatedAt: string
}

export interface ScenarioFormData {
  name: string
  type: ScenarioType
  parentScenarioId?: number
}

export interface ScenarioUpdateFormData {
  name?: string
}
