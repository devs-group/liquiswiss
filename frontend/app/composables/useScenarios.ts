import type { ScenarioFormData, ScenarioResponse } from '~/models/scenario'
import ScenarioDialog from '~/components/dialogs/ScenarioDialog.vue'
import { ModalConfig } from '~/config/dialog-props'

export default function useScenarios() {
  const scenarios = useState<ScenarioResponse[]>('scenarios', () => [])
  const currentScenario = computed(() => scenarios.value.find(s => s.isCurrent) ?? null)

  const useFetchListScenarios = async () => {
    const { data, error } = await useFetch<ScenarioResponse[]>('/api/scenarios', {
      method: 'GET',
    })
    if (error.value) {
      return Promise.reject('Szenarien konnten nicht geladen werden')
    }
    scenarios.value = data.value ?? []
  }

  const listScenarios = async () => {
    try {
      scenarios.value = await $fetch<ScenarioResponse[]>('/api/scenarios', { method: 'GET' })
    }
    catch {
      return Promise.reject('Szenarien konnten nicht geladen werden')
    }
  }

  const getScenario = async (scenarioID: number) => {
    try {
      return await $fetch<ScenarioResponse>(`/api/scenarios/${scenarioID}`, { method: 'GET' })
    }
    catch {
      return Promise.reject('Szenario konnte nicht geladen werden')
    }
  }

  // The backend switches into the new scenario right away
  const createScenario = async (payload: ScenarioFormData) => {
    try {
      return await $fetch<ScenarioResponse>('/api/scenarios', {
        method: 'POST',
        body: { name: payload.name, sourceScenarioId: payload.sourceScenarioId ?? null },
      })
    }
    catch {
      return Promise.reject('Szenario konnte nicht angelegt werden')
    }
  }

  const updateScenario = async (payload: ScenarioFormData) => {
    try {
      await $fetch<ScenarioResponse>(`/api/scenarios/${payload.id}`, {
        method: 'PATCH',
        body: { name: payload.name },
      })
      await listScenarios()
    }
    catch {
      return Promise.reject('Szenario konnte nicht bearbeitet werden')
    }
  }

  const deleteScenario = async (scenarioID: number) => {
    try {
      await $fetch(`/api/scenarios/${scenarioID}`, { method: 'DELETE' })
    }
    catch {
      return Promise.reject('Szenario konnte nicht gelöscht werden')
    }
  }

  const switchScenario = async (scenarioID: number) => {
    try {
      return await $fetch<ScenarioResponse>(`/api/scenarios/${scenarioID}/switch`, { method: 'POST' })
    }
    catch {
      return Promise.reject('Szenario konnte nicht gewechselt werden')
    }
  }

  // Dialog: the open dialog lives in the URL (?scenario=new | <id>) so it survives reloads
  const route = useRoute()
  const router = useRouter()
  const dialog = useDialog()

  const setScenarioQuery = async (value: string | null) => {
    const query = { ...route.query }
    if (value === null) delete query.scenario
    else query.scenario = value
    await router.replace({ query })
  }

  const onClosedScenarioDialog = async (options?: { data?: { reload?: boolean } }) => {
    await setScenarioQuery(null)
    // Create, delete and switch change which data every page shows, so start fresh
    if (options?.data?.reload) {
      reloadNuxtApp({ force: true })
    }
  }

  const openCreateScenarioDialog = () => {
    setScenarioQuery('new')
    dialog.open(ScenarioDialog, {
      props: {
        header: 'Neues Szenario anlegen',
        ...ModalConfig,
      },
      onClose: onClosedScenarioDialog,
    })
  }

  const openEditScenarioDialog = (scenario: ScenarioResponse) => {
    setScenarioQuery(String(scenario.id))
    dialog.open(ScenarioDialog, {
      data: { scenario },
      props: {
        header: 'Szenario bearbeiten',
        ...ModalConfig,
      },
      onClose: onClosedScenarioDialog,
    })
  }

  const restoreScenarioDialog = () => {
    const param = route.query.scenario
    if (typeof param !== 'string' || !param.length) return
    if (param === 'new') {
      openCreateScenarioDialog()
      return
    }
    const scenario = scenarios.value.find(s => s.id === Number(param))
    if (!scenario) {
      setScenarioQuery(null)
      return
    }
    openEditScenarioDialog(scenario)
  }

  return {
    scenarios,
    currentScenario,
    useFetchListScenarios,
    listScenarios,
    getScenario,
    createScenario,
    updateScenario,
    deleteScenario,
    switchScenario,
    openCreateScenarioDialog,
    openEditScenarioDialog,
    restoreScenarioDialog,
  }
}
