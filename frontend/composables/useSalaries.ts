import type { ListSalaryResponse, SalaryPATCHFormData, SalaryPUTFormData, SalaryResponse } from '~/models/employee'
import { DefaultListResponse } from '~/models/default-data'

export default function useSalaries() {
  const limitSalaries = useState('limitSalaries', () => 20)
  const pageSalaries = useState('pageSalaries', () => 1)
  const noMoreDataSalaries = useState('noMoreDataSalaries', () => false)
  const salaries = useState<ListSalaryResponse>('salaries', () => DefaultListResponse())

  // Salaries
  const useFetchListSalaries = async (employeeID: number) => {
    const { data, error } = await useFetch<ListSalaryResponse>(`/api/employees/${employeeID}/salary`, {
      method: 'GET',
      query: {
        page: pageSalaries.value,
        limit: limitSalaries.value,
      },
    })
    if (error.value) {
      return Promise.reject('Fehler beim Laden des Lohnverlaufs')
    }
    setSalaries(data.value, false)
  }

  const listSalaries = async (employeeID: number) => {
    try {
      const data = await $fetch<ListSalaryResponse>(`/api/employees/${employeeID}/salary`, {
        method: 'GET',
        query: {
          page: pageSalaries.value,
          limit: limitSalaries.value,
        },
      })
      setSalaries(data, false)
    }
    catch {
      return Promise.reject('Fehler beim Laden des Lohnverlaufs')
    }
  }

  const getSalary = async (salaryID: number) => {
    try {
      return await $fetch<SalaryResponse>(`/api/employees/salary/${salaryID}`, {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject(`Lohn mit ID "${salaryID}" konnte nicht geladen werden`)
    }
  }

  const createSalary = async (employeeID: number, payload: SalaryPUTFormData) => {
    try {
      const response = await $fetch<SalaryResponse>(`/api/employees/${employeeID}/salary`, {
        method: 'POST',
        body: {
          ...payload,
          amount: payload.amount ? AmountToInteger(payload.amount) : undefined,
          fromDate: payload.fromDate ? DateToApiFormat(payload.fromDate) : undefined,
        },
      })
      await listSalaries(employeeID)
      return response
    }
    catch {
      return Promise.reject('Fehler beim Erstellen des Lohns')
    }
  }

  const updateSalary = async (employeeID: number, payload: SalaryPATCHFormData) => {
    try {
      await $fetch<SalaryResponse>(`/api/employees/salary/${payload.id}`, {
        method: 'PATCH',
        body: {
          ...payload,
          amount: payload.amount ? AmountToInteger(payload.amount) : undefined,
          fromDate: payload.fromDate ? DateToApiFormat(payload.fromDate) : undefined,
        },
      })
      await listSalaries(employeeID)
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren des Lohns')
    }
  }

  const deleteSalary = async (employeeID: number, salaryID: number) => {
    try {
      await $fetch(`/api/employees/salary/${salaryID}`, {
        method: 'DELETE',
      })
      await listSalaries(employeeID)
    }
    catch {
      return Promise.reject('Fehler beim Löschen des Lohns')
    }
  }

  const setSalaries = (data: ListSalaryResponse | null, append: boolean) => {
    if (data) {
      if (append) {
        salaries.value!.data = salaries.value!.data.concat(data.data ?? [])
        salaries.value!.pagination = data.pagination
      }
      else {
        salaries.value = data
      }
      noMoreDataSalaries.value = salaries.value.pagination.totalRemaining == 0
    }
    else {
      salaries.value = DefaultListResponse()
    }
  }

  return {
    salaries,
    limitSalaries,
    pageSalaries,
    noMoreDataSalaries,
    useFetchListSalaries,
    listSalaries,
    getSalary,
    createSalary,
    updateSalary,
    deleteSalary,
    setSalaries,
  }
}
