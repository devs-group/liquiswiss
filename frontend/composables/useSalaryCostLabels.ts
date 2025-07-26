import type { FetchError } from 'ofetch'
import type { ListSalaryCostLabelResponse, SalaryCostLabelFormData, SalaryCostLabelResponse } from '~/models/employee'
import { DefaultListResponse } from '~/models/default-data'

export default function useSalaryCostLabels() {
  const limitSalaryCostsLabels = useState('limitSalaryCostsLabels', () => 20)
  const pageSalaryCostsLabels = useState('pageSalaryCostsLabels', () => 1)
  const noMoreDataSalaryCostsLabels = useState('noMoreDataSalaryCostsLabels', () => false)
  const salaryCostsLabels = useState<ListSalaryCostLabelResponse>('salaryCostsLabels', () => DefaultListResponse())

  // Employees
  const useFetchListSalaryCostsLabels = async (append: boolean) => {
    const { data, error } = await useFetch<ListSalaryCostLabelResponse>(`/api/employees/salary/costs/labels`, {
      method: 'GET',
      query: {
        page: pageSalaryCostsLabels.value,
        limit: limitSalaryCostsLabels.value,
      },
    })
    if (error.value) {
      return Promise.reject('Lohnkosten Labels konnten nicht geladen werden')
    }
    setSalaryCostsLabels(data.value, append)
  }

  const listSalaryCostsLabels = async (append: boolean) => {
    try {
      const data = await $fetch<ListSalaryCostLabelResponse>(`/api/employees/salary/costs/labels`, {
        method: 'GET',
        query: {
          page: pageSalaryCostsLabels.value,
          limit: limitSalaryCostsLabels.value,
        },
      })
      setSalaryCostsLabels(data, append)
    }
    catch (err: unknown) {
      if (IsAbortedError(err as FetchError)) {
        return Promise.reject('aborted')
      }
      else {
        return Promise.reject('Fehler beim Laden der Lohnkosten Labels')
      }
    }
  }

  const useFetchGetSalaryCostLabel = async (salaryLabelID: number) => {
    const { data, error } = await useFetch<SalaryCostLabelResponse>(`/api/employees/salary/costs/labels/${salaryLabelID}`, {
      method: 'GET',
    })
    if (error.value) {
      return Promise.reject(`Lohnkosten Label mit ID "${salaryLabelID}" konnte nicht geladen werden`)
    }
    return data.value
  }

  const getSalaryCostLabel = async (salaryCostLabelID: number) => {
    try {
      return await $fetch<SalaryCostLabelResponse>(`/api/employees/salary/costs/labels/${salaryCostLabelID}`, {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject(`Lohnkosten Label mit ID "${salaryCostLabelID}" konnte nicht geladen werden`)
    }
  }

  const createSalaryCostLabel = async (payload: SalaryCostLabelFormData) => {
    try {
      const data = await $fetch<SalaryCostLabelResponse>(`/api/employees/salary/costs/labels`, {
        method: 'POST',
        body: payload,
      })
      return Promise.resolve(data)
    }
    catch {
      return Promise.reject('Fehler beim Erstellen des Lohnkosten Labels')
    }
  }

  const updateSalaryCostLabel = async (payload: SalaryCostLabelFormData) => {
    try {
      return await $fetch<SalaryCostLabelResponse>(`/api/employees/salary/costs/labels/${payload.id}`, {
        method: 'PATCH',
        body: payload,
      })
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren des Lohnkosten Labels')
    }
  }

  const deleteSalaryCostLabel = async (salaryCostLabelID: number) => {
    try {
      await $fetch(`/api/employees/salary/costs/labels/${salaryCostLabelID}`, {
        method: 'DELETE',
      })
      salaryCostsLabels.value!.data = salaryCostsLabels.value!.data.filter(salaryCostLabel => salaryCostLabel.id !== salaryCostLabelID)
    }
    catch {
      return Promise.reject('Fehler beim Löschen des Lohnkosten Labels')
    }
  }

  const setSalaryCostsLabels = (data: ListSalaryCostLabelResponse | null, append: boolean) => {
    if (data) {
      if (append) {
        salaryCostsLabels.value!.data = salaryCostsLabels.value!.data.concat(data.data ?? [])
        salaryCostsLabels.value!.pagination = data.pagination
      }
      else {
        salaryCostsLabels.value = data
      }
      noMoreDataSalaryCostsLabels.value = salaryCostsLabels.value.pagination.totalRemaining == 0
    }
    else {
      salaryCostsLabels.value = DefaultListResponse()
    }
  }

  return {
    salaryCostsLabels,
    useFetchListSalaryCostsLabels,
    listSalaryCostsLabels,
    useFetchGetSalaryCostLabel,
    getSalaryCostLabel,
    createSalaryCostLabel,
    updateSalaryCostLabel,
    deleteSalaryCostLabel,
  }
}
