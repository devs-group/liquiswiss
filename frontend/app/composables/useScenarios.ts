import type { ScenarioResponse, ScenarioListItem, ScenarioFormData, ScenarioUpdateFormData } from '~/models/scenario'

export default function useScenarios() {
  const scenarios = useState<ScenarioListItem[]>('scenarios', () => [])
  const activeScenario = useState<ScenarioResponse | null>('activeScenario', () => null)

  const useFetchListScenarios = async () => {
    const { data, error } = await useFetch<ScenarioListItem[]>('/api/scenarios', {
      method: 'GET',
    })
    if (error.value) {
      return Promise.reject('Scenarios could not be loaded')
    }
    setScenarios(data.value ?? [])
  }

  const listScenarios = async () => {
    try {
      const data = await $fetch<ScenarioListItem[]>('/api/scenarios', {
        method: 'GET',
      })
      setScenarios(data ?? [])
    }
    catch {
      return Promise.reject('Error loading scenarios')
    }
  }

  const useFetchGetScenario = async (scenarioID: number) => {
    const { data, error } = await useFetch<ScenarioResponse>(`/api/scenarios/${scenarioID}`, {
      method: 'GET',
    })
    if (error.value) {
      return Promise.reject('Error loading scenario')
    }
    return data.value
  }

  const getScenario = async (scenarioID: number) => {
    try {
      return await $fetch<ScenarioResponse>(`/api/scenarios/${scenarioID}`, {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject('Error loading scenario')
    }
  }

  const getDefaultScenario = async () => {
    try {
      return await $fetch<ScenarioResponse>('/api/scenarios/default', {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject('Error loading default scenario')
    }
  }

  const createScenario = async (payload: ScenarioFormData) => {
    try {
      const scenario = await $fetch<ScenarioResponse>('/api/scenarios', {
        method: 'POST',
        body: payload,
      })
      await listScenarios()
      return scenario
    }
    catch {
      return Promise.reject('Error creating scenario')
    }
  }

  const updateScenario = async (scenarioID: number, payload: ScenarioUpdateFormData) => {
    try {
      await $fetch<ScenarioResponse>(`/api/scenarios/${scenarioID}`, {
        method: 'PATCH',
        body: payload,
      })
      await listScenarios()
    }
    catch {
      return Promise.reject('Error updating scenario')
    }
  }

  const deleteScenario = async (scenarioID: number) => {
    try {
      await $fetch(`/api/scenarios/${scenarioID}`, {
        method: 'DELETE',
      })
      await listScenarios()
      // Force reload the app to ensure all data is fresh after scenario deletion
      reloadNuxtApp({ force: true })
    }
    catch {
      return Promise.reject('Error deleting scenario')
    }
  }

  const setScenarios = (data: ScenarioListItem[] | null) => {
    if (data) {
      scenarios.value = data
    }
    else {
      scenarios.value = []
    }
  }

  const setActiveScenario = (scenario: ScenarioResponse | null) => {
    activeScenario.value = scenario
  }

  const initializeActiveScenario = async () => {
    const { user } = useAuth()

    if (user.value?.currentScenarioID) {
      // Load the user's saved scenario
      try {
        const scenario = await getScenario(user.value.currentScenarioID)
        setActiveScenario(scenario)
      }
      catch {
        // If loading fails, fall back to default scenario
        const defaultScenario = await getDefaultScenario()
        setActiveScenario(defaultScenario)
      }
    }
    else if (!activeScenario.value) {
      // No saved scenario, use default
      const defaultScenario = await getDefaultScenario()
      setActiveScenario(defaultScenario)
    }
  }

  return {
    useFetchListScenarios,
    listScenarios,
    useFetchGetScenario,
    getScenario,
    getDefaultScenario,
    createScenario,
    updateScenario,
    deleteScenario,
    setScenarios,
    setActiveScenario,
    initializeActiveScenario,
    scenarios,
    activeScenario,
  }
}
