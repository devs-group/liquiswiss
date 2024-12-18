import type { CategoryResponse, ListCategoryResponse } from '~/models/category'
import type { CurrencyResponse, ListCurrencyResponse } from '~/models/currency'
import type { FiatRateResponse } from '~/models/fiat-rate'
import type { ServerTimeResponse } from '~/models/server-time'

export default function useGlobalData() {
  const categories = useState<CategoryResponse[]>('catgegories', () => [])
  const currencies = useState<CurrencyResponse[]>('currencies', () => [])
  const fiatRates = useState<FiatRateResponse[]>('fiatRates', () => [])
  const serverDate = useState<Date | null>('serverDate', () => null)
  const showGlobalLoadingSpinner = useState<boolean>('showGlobalLoadingSpinner', () => false)

  const useFetchListCategories = async () => {
    const { data, error } = await useFetch<ListCategoryResponse>('/api/categories', {
      method: 'GET',
      query: {
        page: 1,
        limit: 100,
      },
    })
    if (error.value) {
      return Promise.reject('Kategorien konnten nicht geladen werden')
    }
    categories.value = data.value?.data ?? []
  }

  const useFetchListCurrencies = async () => {
    const { data, error } = await useFetch<ListCurrencyResponse>('/api/currencies', {
      method: 'GET',
      query: {
        page: 1,
        limit: 100,
      },
    })
    if (error.value) {
      return Promise.reject('Währungen konnten nicht geladen werden')
    }
    currencies.value = data.value?.data ?? []
  }

  const useFetchListFiatRates = async () => {
    const { data, error } = await useFetch<FiatRateResponse[]>('/api/fiat-rates', {
      method: 'GET',
    })
    if (error.value) {
      return Promise.reject('Wechselkurse konnten nicht geladen werden')
    }
    fiatRates.value = data.value ?? []
  }

  const useFetchGetServerTime = async () => {
    const { data, error } = await useFetch<ServerTimeResponse>('/api/global/time', {
      method: 'GET',
    })
    if (error.value) {
      return Promise.reject('Serverzeit konnte nicht geladen werden')
    }
    serverDate.value = data.value?.date ? new Date(data.value!.date) : null
  }

  const convertAmountToRate = (amount: number, currency: string) => {
    const fiatRate = fiatRates.value.find(fr => fr.target == currency)
    if (fiatRate) {
      return amount / fiatRate.rate
    }
    return amount
  }

  return {
    categories,
    useFetchListCategories,
    currencies,
    useFetchListCurrencies,
    fiatRates,
    useFetchListFiatRates,
    convertAmountToRate,
    useFetchGetServerTime,
    serverDate,
    showGlobalLoadingSpinner,
  }
}
