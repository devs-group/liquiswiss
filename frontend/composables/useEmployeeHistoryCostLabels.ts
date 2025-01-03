import type { FetchError } from 'ofetch'
import type {
  EmployeeHistoryCostLabelFormData,
  EmployeeHistoryCostLabelResponse,
  ListEmployeeHistoryCostLabelResponse,
} from '~/models/employee'
import { DefaultListResponse } from '~/models/default-data'

export default function useEmployeeHistoryCostLabels() {
  const limitEmployeeHistoryCostsLabels = useState('limitEmployeeHistoryCostsLabels', () => 20)
  const pageEmployeeHistoryCostsLabels = useState('pageEmployeeHistoryCostsLabels', () => 1)
  const noMoreDataEmployeeHistoryCostsLabels = useState('noMoreDataEmployeeHistoryCostsLabels', () => false)
  const employeeHistoryCostsLabels = useState<ListEmployeeHistoryCostLabelResponse>('employeeHistoryCostsLabels', () => DefaultListResponse())

  // Employees
  const useFetchListEmployeeHistoryCostsLabels = async (append: boolean) => {
    const { data, error } = await useFetch<ListEmployeeHistoryCostLabelResponse>(`/api/employees/history/costs/labels`, {
      method: 'GET',
      query: {
        page: pageEmployeeHistoryCostsLabels.value,
        limit: limitEmployeeHistoryCostsLabels.value,
      },
    })
    if (error.value) {
      return Promise.reject('Lohnkosten Labels konnten nicht geladen werden')
    }
    setEmployeeHistoryCostsLabels(data.value, append)
  }

  const listEmployeeHistoryCostsLabels = async (append: boolean) => {
    try {
      const data = await $fetch<ListEmployeeHistoryCostLabelResponse>(`/api/employees/history/costs/labels`, {
        method: 'GET',
        query: {
          page: pageEmployeeHistoryCostsLabels.value,
          limit: limitEmployeeHistoryCostsLabels.value,
        },
      })
      setEmployeeHistoryCostsLabels(data, append)
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

  const useFetchGetEmployeeHistoryCostLabel = async (historyCostLabelID: number) => {
    const { data, error } = await useFetch<EmployeeHistoryCostLabelResponse>(`/api/employees/history/costs/labels/${historyCostLabelID}`, {
      method: 'GET',
    })
    if (error.value) {
      return Promise.reject(`Lohnkosten Label mit ID "${historyCostLabelID}" konnte nicht geladen werden`)
    }
    return data.value
  }

  const getEmployeeHistoryCostLabel = async (historyCostLabelID: number) => {
    try {
      return await $fetch<EmployeeHistoryCostLabelResponse>(`/api/employees/history/costs/labels/${historyCostLabelID}`, {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject(`Lohnkosten Label mit ID "${historyCostLabelID}" konnte nicht geladen werden`)
    }
  }

  const createEmployeeHistoryCostLabel = async (payload: EmployeeHistoryCostLabelFormData) => {
    try {
      const data = await $fetch<EmployeeHistoryCostLabelResponse>(`/api/employees/history/costs/labels`, {
        method: 'POST',
        body: payload,
      })
      return Promise.resolve(data)
    }
    catch {
      return Promise.reject('Fehler beim Erstellen des Lohnkosten Labels')
    }
  }

  const updateEmployeeHistoryCostLabel = async (payload: EmployeeHistoryCostLabelFormData) => {
    try {
      return await $fetch<EmployeeHistoryCostLabelResponse>(`/api/employees/history/costs/labels/${payload.id}`, {
        method: 'PATCH',
        body: payload,
      })
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren des Lohnkosten Labels')
    }
  }

  const deleteEmployeeHistoryCostLabel = async (historyCostLabelID: number) => {
    try {
      await $fetch(`/api/employees/history/costs/labels/${historyCostLabelID}`, {
        method: 'DELETE',
      })
      employeeHistoryCostsLabels.value!.data = employeeHistoryCostsLabels.value!.data.filter(historyCost => historyCost.id !== historyCostLabelID)
    }
    catch {
      return Promise.reject('Fehler beim Löschen des Lohnkosten Labels')
    }
  }

  const setEmployeeHistoryCostsLabels = (data: ListEmployeeHistoryCostLabelResponse | null, append: boolean) => {
    if (data) {
      if (append) {
        employeeHistoryCostsLabels.value!.data = employeeHistoryCostsLabels.value!.data.concat(data.data ?? [])
        employeeHistoryCostsLabels.value!.pagination = data.pagination
      }
      else {
        employeeHistoryCostsLabels.value = data
      }
      noMoreDataEmployeeHistoryCostsLabels.value = employeeHistoryCostsLabels.value.pagination.totalRemaining == 0
    }
    else {
      employeeHistoryCostsLabels.value = DefaultListResponse()
    }
  }

  return {
    employeeHistoryCostsLabels,
    useFetchListEmployeeHistoryCostsLabels,
    listEmployeeHistoryCostsLabels,
    useFetchGetEmployeeHistoryCostLabel,
    getEmployeeHistoryCostLabel,
    createEmployeeHistoryCostLabel,
    updateEmployeeHistoryCostLabel,
    deleteEmployeeHistoryCostLabel,
  }
}
