import type { ScenarioFormData, ScenarioResponse, ScenarioTreeNode } from '~/models/scenario'

export default function useScenarios() {
  const scenarios = useState<ScenarioResponse[]>('scenarios', () => [])

  const buildScenarioTreeNodes = (flatScenarios: ScenarioResponse[]): ScenarioTreeNode[] => {
    const map = new Map<number, ScenarioTreeNode>()
    const roots: ScenarioTreeNode[] = []

    for (const scenario of flatScenarios) {
      map.set(scenario.id, {
        key: String(scenario.id),
        data: { ...scenario, level: 0 },
        children: [],
      })
    }

    for (const node of map.values()) {
      if (node.data.parentScenarioID) {
        map.get(node.data.parentScenarioID)?.children.push(node)
      }
      else {
        roots.push(node)
      }
    }

    const assignLevels = (nodes: ScenarioTreeNode[], level: number) => {
      for (const node of nodes) {
        node.data.level = level
        assignLevels(node.children, level + 1)
      }
    }
    assignLevels(roots, 0)

    return roots
  }

  const scenarioTreeNodes = computed(() => buildScenarioTreeNodes(scenarios.value))

  const useFetchListScenarios = async () => {
    const { data, error } = await useFetch<ScenarioResponse[]>('/api/scenarios', {
      method: 'GET',
    })
    if (error.value || !data.value) {
      return Promise.reject('Szenarien konnten nicht geladen werden')
    }
    setScenarios(data.value, false)
  }

  const listScenarios = async () => {
    try {
      const data = await $fetch<ScenarioResponse[]>('/api/scenarios', {
        method: 'GET',
      })
      setScenarios(data, false)
    }
    catch {
      return Promise.reject('Fehler beim Laden der Szenarien')
    }
  }

  const getScenario = async (scenarioID: number) => {
    try {
      return await $fetch<ScenarioResponse>(`/api/scenarios/${scenarioID}`, {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject('Fehler beim Laden des Szenarios')
    }
  }

  const createScenario = async (payload: ScenarioFormData) => {
    try {
      await $fetch<ScenarioResponse>(`/api/scenarios`, {
        method: 'POST',
        body: {
          ...payload,
        },
      })
      await listScenarios()
    }
    catch {
      return Promise.reject('Fehler beim Erstellen des Szenarios')
    }
  }

  const updateScenario = async (payload: ScenarioFormData) => {
    try {
      await $fetch<ScenarioResponse>(`/api/scenarios/${payload.id}`, {
        method: 'PATCH',
        body: {
          ...payload,
        },
      })
      await listScenarios()
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren des Szenarios')
    }
  }

  const deleteScenario = async (scenarioID: number) => {
    try {
      await $fetch(`/api/scenarios/${scenarioID}`, {
        method: 'DELETE',
      })
      await listScenarios()
    }
    catch {
      return Promise.reject('Fehler beim Löschen des Szenarios')
    }
  }

  const setScenarios = (data: ScenarioResponse[] | null, append: boolean) => {
    if (data) {
      if (append) {
        scenarios.value = scenarios.value.concat(data ?? [])
      }
      else {
        scenarios.value = data
      }
    }
    else {
      scenarios.value = []
    }
  }

  return {
    useFetchListScenarios,
    listScenarios,
    getScenario,
    createScenario,
    updateScenario,
    deleteScenario,
    setScenarios,
    scenarios,
    scenarioTreeNodes,
  }
}
