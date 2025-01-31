import type { ForecastDetailResponse, ForecastResponse } from '~/models/forecast'

export default function useForecasts() {
  const forecasts = useState<ForecastResponse[]>('forecasts', () => [])
  const forecastDetails = useState<ForecastDetailResponse[]>('forecastDetails', () => [])

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
    useFetchListForecast,
    listForecasts,
    useFetchListForecastDetails,
    listForecastDetails,
    setForecasts,
    calculateForecast,
    excludeForecast,
    includeForecast,
  }
}
