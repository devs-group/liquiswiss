export interface ScenarioResponse {
  id: number
  name: string
  isDefault: boolean
  createdAt: number
  parentScenarioID?: number
}

export interface ScenarioFormData {
  id: number
  name: string
  parentScenarioID?: number
}

export interface ScenarioTreeNode {
  key: string
  data: ScenarioResponse & { level: number }
  children: ScenarioTreeNode[]
}
