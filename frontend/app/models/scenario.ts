export interface ScenarioResponse {
  id: number
  name: string
  isBase: boolean
  parentId: number | null
  isCurrent: boolean
  createdAt: string
}

export interface ScenarioFormData {
  id?: number
  name: string
  sourceScenarioId?: number | null
}
