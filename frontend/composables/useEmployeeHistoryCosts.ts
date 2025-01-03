import type { FetchError } from 'ofetch'
import type {
  EmployeeHistoryCostCopyFormData,
  EmployeeHistoryCostFormData,
  EmployeeHistoryCostResponse,
  ListEmployeeHistoryCostResponse,
} from '~/models/employee'
import { EmployeeCostType } from '~/config/enums'

export default function useEmployeeHistoryCosts() {
  const limitEmployeeHistoryCosts = useState('limitEmployeeHistoryCosts', () => 20)
  const pageEmployeeHistoryCosts = useState('pageEmployeeHistoryCosts', () => 1)

  // Employees
  const useFetchListEmployeeHistoryCosts = async (historyID: number) => {
    const { data, error } = await useFetch<ListEmployeeHistoryCostResponse>(`/api/employees/history/${historyID}/costs`, {
      method: 'GET',
      query: {
        page: pageEmployeeHistoryCosts.value,
        limit: limitEmployeeHistoryCosts.value,
      },
    })
    if (error.value) {
      return Promise.reject('Lohnnebenkosten konnten nicht geladen werden')
    }
    return data.value
  }

  const listEmployeeHistoryCosts = async (historyID: number) => {
    try {
      return await $fetch<ListEmployeeHistoryCostResponse>(`/api/employees/history/${historyID}/costs`, {
        method: 'GET',
        query: {
          page: pageEmployeeHistoryCosts.value,
          limit: limitEmployeeHistoryCosts.value,
        },
      })
    }
    catch (err: unknown) {
      if (IsAbortedError(err as FetchError)) {
        return Promise.reject('aborted')
      }
      else {
        return Promise.reject('Fehler beim Laden der Lohnnebenkosten')
      }
    }
  }

  const useFetchGetEmployeeHistoryCost = async (historyCostID: number) => {
    const { data, error } = await useFetch<EmployeeHistoryCostResponse>(`/api/employees/history/costs/${historyCostID}`, {
      method: 'GET',
    })
    if (error.value) {
      return Promise.reject(`Lohnkosten mit ID "${historyCostID}" konnte nicht geladen werden`)
    }
    return data.value
  }

  const getEmployeeHistoryCost = async (historyCostID: number) => {
    try {
      return await $fetch<EmployeeHistoryCostResponse>(`/api/employees/history/costs/${historyCostID}`, {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject(`Lohnkosten mit ID "${historyCostID}" konnte nicht geladen werden`)
    }
  }

  const createEmployeeHistoryCost = async (historyID: number, payload: EmployeeHistoryCostFormData) => {
    try {
      return await $fetch<EmployeeHistoryCostResponse>(`/api/employees/history/${historyID}/costs`, {
        method: 'POST',
        body: {
          ...payload,
          amount: payload.amountType == EmployeeCostType.Fixed
            ? AmountToInteger(payload.amount, 2)
            : AmountToInteger(payload.amount, 3),
          targetDate: payload.targetDate ? DateToApiFormat(payload.targetDate) : undefined,
        },
      })
    }
    catch {
      return Promise.reject('Fehler beim Erstellen der Lohnnebenkosten')
    }
  }

  const copyEmployeeHistoryCost = async (historyID: number, payload: EmployeeHistoryCostCopyFormData) => {
    try {
      return await $fetch<EmployeeHistoryCostResponse>(`/api/employees/history/${historyID}/costs/copy`, {
        method: 'POST',
        body: payload,
      })
    }
    catch {
      return Promise.reject('Fehler beim Kopieren der Lohnnebenkosten')
    }
  }

  const updateEmployeeHistoryCost = async (payload: EmployeeHistoryCostFormData) => {
    try {
      return await $fetch<EmployeeHistoryCostResponse>(`/api/employees/history/costs/${payload.id}`, {
        method: 'PATCH',
        body: {
          ...payload,
          amount: payload.amountType == EmployeeCostType.Fixed
            ? AmountToInteger(payload.amount, 2)
            : AmountToInteger(payload.amount, 3),
          targetDate: payload.targetDate ? DateToApiFormat(payload.targetDate) : undefined,
        },
      })
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren der Lohnnebenkosten')
    }
  }

  const deleteEmployeeHistoryCost = async (historyCostID: number) => {
    try {
      await $fetch(`/api/employees/history/costs/${historyCostID}`, {
        method: 'DELETE',
      })
    }
    catch {
      return Promise.reject('Fehler beim Löschen der Lohnnebenkosten')
    }
  }

  return {
    useFetchListEmployeeHistoryCosts,
    listEmployeeHistoryCosts,
    useFetchGetEmployeeHistoryCost,
    getEmployeeHistoryCost,
    createEmployeeHistoryCost,
    copyEmployeeHistoryCost,
    updateEmployeeHistoryCost,
    deleteEmployeeHistoryCost,
  }
}
