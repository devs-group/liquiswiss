import type { ForecastDetailResponse, ForecastResponse } from '~/models/forecast'

type ForecastExclusionChange = {
  key: string
  month: string
  relatedID: number
  relatedTable: string
  originalIsExcluded: boolean
  isExcluded: boolean
}

export default function useForecasts() {
  const forecasts = useState<ForecastResponse[]>('forecasts', () => [])
  const forecastDetails = useState<ForecastDetailResponse[]>('forecastDetails', () => [])
  const forecastExclusionChanges = useState<Record<string, ForecastExclusionChange>>('forecastExclusionChanges', () => ({}))

  const createDraftKey = (month: string, relatedID: number, relatedTable: string) => {
    return `${month}:${relatedTable}:${relatedID}`
  }

  const toggleForecastExclusionChange = (payload: { month: string, relatedID: number, relatedTable: string, originalIsExcluded: boolean }) => {
    const key = createDraftKey(payload.month, payload.relatedID, payload.relatedTable)
    const existing = forecastExclusionChanges.value[key]
    if (!existing) {
      forecastExclusionChanges.value = {
        ...forecastExclusionChanges.value,
        [key]: {
          key,
          month: payload.month,
          relatedID: payload.relatedID,
          relatedTable: payload.relatedTable,
          originalIsExcluded: payload.originalIsExcluded,
          isExcluded: !payload.originalIsExcluded,
        },
      }
      return
    }

    const nextState = !existing.isExcluded
    if (nextState === existing.originalIsExcluded) {
      const { [key]: _, ...rest } = forecastExclusionChanges.value
      forecastExclusionChanges.value = rest
      return
    }

    forecastExclusionChanges.value = {
      ...forecastExclusionChanges.value,
      [key]: {
        ...existing,
        isExcluded: nextState,
      },
    }
  }

  const getForecastExclusionChange = (month: string, relatedID: number, relatedTable: string) => {
    return forecastExclusionChanges.value[createDraftKey(month, relatedID, relatedTable)]
  }

  const clearForecastExclusionChanges = () => {
    forecastExclusionChanges.value = {}
  }

  const applyForecastExclusionChanges = async () => {
    const changes = Object.values(forecastExclusionChanges.value)
    if (!changes.length) {
      return
    }
    try {
      await $fetch('/api/forecasts/exclude', {
        method: 'PUT',
        body: {
          updates: changes.map(change => ({
            month: change.month,
            relatedID: change.relatedID,
            relatedTable: change.relatedTable,
            isExcluded: change.isExcluded,
          })),
        },
      })
    }
    catch {
      throw new Error('Fehler beim Speichern der Prognose-Einstellungen')
    }
  }

  const useFetchListForecast = async (months: number) => {
    const { data, error } = await useFetch<ForecastResponse[]>('/api/forecasts', {
      method: 'GET',
      query: {
        limit: months,
      },
    })
    if (error.value) {
      return Promise.reject('Prognose konnten nicht geladen werden')
    }
    setForecasts(data.value, false)
  }

  const listForecasts = async (months: number) => {
    try {
      const data = await $fetch<ForecastResponse[]>('/api/forecasts', {
        method: 'GET',
        query: {
          limit: months,
        },
      })
      setForecasts(data, false)
    }
    catch {
      return Promise.reject('Fehler beim Laden der Prognose')
    }
  }

  const useFetchListForecastDetails = async (months: number) => {
    const { data, error } = await useFetch<ForecastDetailResponse[]>('/api/forecasts/details', {
      method: 'GET',
      query: {
        limit: months,
      },
    })
    if (error.value) {
      return Promise.reject('Fehler beim Laden der Prognose Details')
    }
    setForecastDetails(data.value, false)
  }

  const listForecastDetails = async (months: number) => {
    try {
      const data = await $fetch<ForecastDetailResponse[]>('/api/forecasts/details', {
        method: 'GET',
        query: {
          limit: months,
        },
      })
      setForecastDetails(data, false)
    }
    catch {
      return Promise.reject('Fehler beim Laden der Prognose Details')
    }
  }

  const calculateForecast = async () => {
    try {
      await $fetch<ForecastDetailResponse[]>('/api/forecasts/calculate', {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject('Fehler beim Berechnen der Prognose')
    }
  }

  const setForecasts = (data: ForecastResponse[] | null, append: boolean) => {
    if (data) {
      if (append) {
        forecasts.value = forecasts.value.concat(data ?? [])
      }
      else {
        forecasts.value = data
      }
    }
    else {
      forecasts.value = []
    }
  }

  const setForecastDetails = (data: ForecastDetailResponse[] | null, append: boolean) => {
    if (data) {
      if (append) {
        forecastDetails.value = forecastDetails.value.concat(data ?? [])
      }
      else {
        forecastDetails.value = data
      }
    }
    else {
      forecastDetails.value = []
    }
  }

  const excludeForecast = async (month: string, relatedID: string | number, relatedTable: string) => {
    try {
      await $fetch<ForecastDetailResponse[]>('/api/forecasts/exclude', {
        method: 'POST',
        body: {
          month,
          relatedID,
          relatedTable,
        },
      })
    }
    catch {
      return Promise.reject('Fehler beim Ausschliessen der Prognose')
    }
  }

  const includeForecast = async (month: string, relatedID: string | number, relatedTable: string) => {
    try {
      await $fetch<ForecastDetailResponse[]>('/api/forecasts/exclude', {
        method: 'DELETE',
        body: {
          month,
          relatedID,
          relatedTable,
        },
      })
    }
    catch {
      return Promise.reject('Fehler beim Berücksichtigen der Prognose')
    }
  }

  return {
    forecasts,
    forecastDetails,
    forecastExclusionChanges,
    useFetchListForecast,
    listForecasts,
    useFetchListForecastDetails,
    listForecastDetails,
    setForecasts,
    calculateForecast,
    excludeForecast,
    includeForecast,
    toggleForecastExclusionChange,
    getForecastExclusionChange,
    clearForecastExclusionChanges,
    applyForecastExclusionChanges,
  }
}
