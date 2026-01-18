import type { FetchError } from 'ofetch'
import type { EmployeeFormData, EmployeeResponse, ListEmployeeResponse } from '~/models/employee'
import { DefaultListResponse } from '~/models/default-data'

export default function useEmployees() {
  const limitEmployees = useState('limitEmployees', () => 20)
  const pageEmployees = useState('pageEmployees', () => 1)
  const noMoreDataEmployees = useState('noMoreDataEmployees', () => false)
  const searchEmployees = useState('searchEmployees', () => '')
  const employees = useState<ListEmployeeResponse>('employees', () => DefaultListResponse())

  const { employeeSortBy, employeeSortOrder } = useSettings()

  // Employees
  const useFetchListEmployees = async () => {
    const { data, error } = await useFetch<ListEmployeeResponse>('/api/employees', {
      method: 'GET',
      query: {
        page: pageEmployees.value,
        limit: limitEmployees.value,
        sortBy: employeeSortBy.value,
        sortOrder: employeeSortOrder.value,
        search: searchEmployees.value,
      },
    })
    if (error.value) {
      return Promise.reject('Mitarbeiter konnten nicht geladen werden')
    }
    setEmployees(data.value, false)
  }

  const listEmployees = async (append: boolean) => {
    try {
      const data = await $fetch<ListEmployeeResponse>('/api/employees', {
        method: 'GET',
        query: {
          page: pageEmployees.value,
          limit: limitEmployees.value,
          sortBy: employeeSortBy.value,
          sortOrder: employeeSortOrder.value,
          search: searchEmployees.value,
        },
      })
      setEmployees(data, append)
    }
    catch (err: unknown) {
      if (IsAbortedError(err as FetchError)) {
        return Promise.reject('aborted')
      }
      else {
        return Promise.reject('Fehler beim Laden der Mitarbeiter')
      }
    }
  }

  const useFetchGetEmployee = async (employeeID: number) => {
    const { data, error } = await useFetch<EmployeeResponse>(`/api/employees/${employeeID}`, {
      method: 'GET',
    })
    if (error.value) {
      return Promise.reject(`Mitarbeiter mit ID "${employeeID}" konnte nicht geladen werden`)
    }
    return data.value
  }

  const getEmployee = async (employeeID: number) => {
    try {
      return await $fetch<EmployeeResponse>(`/api/employees/${employeeID}`, {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject(`Mitarbeiter mit ID "${employeeID}" konnte nicht geladen werden`)
    }
  }

  const createEmployee = async (payload: EmployeeFormData) => {
    let id = 0

    try {
      const data = await $fetch<EmployeeResponse>(`/api/employees`, {
        method: 'POST',
        body: payload,
      })
      id = data.id
    }
    catch {
      return Promise.reject('Fehler beim Erstellen des Mitarbeiters')
    }

    return Promise.resolve(id)
  }

  const updateEmployee = async (payload: EmployeeFormData) => {
    try {
      return await $fetch<EmployeeResponse>(`/api/employees/${payload.id}`, {
        method: 'PATCH',
        body: payload,
      })
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren des Mitarbeiters')
    }
  }

  const deleteEmployee = async (employeeID: number) => {
    try {
      await $fetch(`/api/employees/${employeeID}`, {
        method: 'DELETE',
      })
      employees.value!.data = employees.value!.data.filter(employee => employee.id !== employeeID)
    }
    catch {
      return Promise.reject('Fehler beim Löschen des Mitarbeiters')
    }
  }

  const setEmployees = (data: ListEmployeeResponse | null, append: boolean) => {
    if (data) {
      if (append) {
        employees.value!.data = employees.value!.data.concat(data.data ?? [])
        employees.value!.pagination = data.pagination
      }
      else {
        employees.value = data
      }
      noMoreDataEmployees.value = employees.value.pagination.totalRemaining == 0
    }
    else {
      employees.value = DefaultListResponse()
    }
  }

  return {
    employees,
    limitEmployees,
    pageEmployees,
    noMoreDataEmployees,
    searchEmployees,
    useFetchListEmployees,
    listEmployees,
    useFetchGetEmployee,
    getEmployee,
    createEmployee,
    updateEmployee,
    deleteEmployee,
    setEmployees,
  }
}
