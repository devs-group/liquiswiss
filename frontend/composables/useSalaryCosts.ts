import type { FetchError } from 'ofetch'
import type {
  ListSalaryCostResponse,
  SalaryCostCopyFormData,
  SalaryCostFormData,
  SalaryCostResponse,
} from '~/models/employee'
import { EmployeeCostType } from '~/config/enums'

export default function useSalaryCosts() {
  const limitSalaryCosts = useState('limitSalaryCosts', () => 20)
  const pageSalaryCosts = useState('pageSalaryCosts', () => 1)

  // Employees
  const useFetchListSalaryCosts = async (salaryID: number) => {
    const { data, error } = await useFetch<ListSalaryCostResponse>(`/api/employees/salary/${salaryID}/costs`, {
      method: 'GET',
      query: {
        page: pageSalaryCosts.value,
        limit: limitSalaryCosts.value,
      },
    })
    if (error.value) {
      return Promise.reject('Lohnnebenkosten konnten nicht geladen werden')
    }
    return data.value
  }

  const listSalaryCosts = async (salaryID: number) => {
    try {
      return await $fetch<ListSalaryCostResponse>(`/api/employees/salary/${salaryID}/costs`, {
        method: 'GET',
        query: {
          page: pageSalaryCosts.value,
          limit: limitSalaryCosts.value,
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

  const useFetchGetSalaryCost = async (salaryCostID: number) => {
    const { data, error } = await useFetch<SalaryCostResponse>(`/api/employees/salary/costs/${salaryCostID}`, {
      method: 'GET',
    })
    if (error.value) {
      return Promise.reject(`Lohnkosten mit ID "${salaryCostID}" konnte nicht geladen werden`)
    }
    return data.value
  }

  const getSalaryCost = async (salaryCostID: number) => {
    try {
      return await $fetch<SalaryCostResponse>(`/api/employees/salary/costs/${salaryCostID}`, {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject(`Lohnkosten mit ID "${salaryCostID}" konnte nicht geladen werden`)
    }
  }

  const createSalaryCost = async (salaryID: number, payload: SalaryCostFormData) => {
    try {
      return await $fetch<SalaryCostResponse>(`/api/employees/salary/${salaryID}/costs`, {
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

  const copySalaryCost = async (salaryID: number, payload: SalaryCostCopyFormData) => {
    try {
      return await $fetch<SalaryCostResponse>(`/api/employees/salary/${salaryID}/costs/copy`, {
        method: 'POST',
        body: payload,
      })
    }
    catch {
      return Promise.reject('Fehler beim Kopieren der Lohnnebenkosten')
    }
  }

  const updateSalaryCost = async (payload: SalaryCostFormData) => {
    try {
      return await $fetch<SalaryCostResponse>(`/api/employees/salary/costs/${payload.id}`, {
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

  const deleteSalaryCost = async (salaryCostID: number) => {
    try {
      await $fetch(`/api/employees/salary/costs/${salaryCostID}`, {
        method: 'DELETE',
      })
    }
    catch {
      return Promise.reject('Fehler beim Löschen der Lohnnebenkosten')
    }
  }

  return {
    useFetchListSalaryCosts,
    listSalaryCosts,
    useFetchGetSalaryCost,
    getSalaryCost,
    createSalaryCost,
    copySalaryCost,
    updateSalaryCost,
    deleteSalaryCost,
  }
}
