export interface ScenarioResponse {
  id: number
  name: string
  isDefault: boolean
  createdAt: number
  parentScenarioID?: number
}

export type ScenarioType = 'based_on' | 'independent'

export interface ScenarioFormData {
  id: number
  name: string
  scenarioType?: ScenarioType
  parentScenarioID?: number
}

export interface ScenarioTreeNode {
  key: string
  data: ScenarioResponse & { level: number }
  children: ScenarioTreeNode[]
}
