import type {
  EmployeeHistoryPATCHFormData,
  EmployeeHistoryPUTFormData,
  EmployeeHistoryResponse,
  ListEmployeeHistoryResponse,
} from '~/models/employee'
import { DefaultListResponse } from '~/models/default-data'

export default function useEmployeeHistories() {
  const limitEmployeeHistories = useState('limitEmployeeHistories', () => 20)
  const pageEmployeeHistories = useState('pageEmployeeHistories', () => 1)
  const noMoreDataEmployeeHistories = useState('noMoreDataEmployeeHistories', () => false)
  const employeeHistories = useState<ListEmployeeHistoryResponse>('employeeHistories', () => DefaultListResponse())

  // Employee Histories
  const useFetchListEmployeeHistory = async (employeeID: number) => {
    const { data, error } = await useFetch<ListEmployeeHistoryResponse>(`/api/employees/${employeeID}/history`, {
      method: 'GET',
      query: {
        page: pageEmployeeHistories.value,
        limit: limitEmployeeHistories.value,
      },
    })
    if (error.value) {
      return Promise.reject('Fehler beim Laden der Historie')
    }
    setEmployeeHistories(data.value, false)
  }

  const listEmployeeHistory = async (employeeID: number) => {
    try {
      const data = await $fetch<ListEmployeeHistoryResponse>(`/api/employees/${employeeID}/history`, {
        method: 'GET',
        query: {
          page: pageEmployeeHistories.value,
          limit: limitEmployeeHistories.value,
        },
      })
      setEmployeeHistories(data, false)
    }
    catch {
      return Promise.reject('Fehler beim Laden der Historie')
    }
  }

  const getEmployeeHistory = async (employeeHistoryID: number) => {
    try {
      return await $fetch<EmployeeHistoryResponse>(`/api/employees/history/${employeeHistoryID}`, {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject(`Historie mit ID "${employeeHistoryID}" konnte nicht geladen werden`)
    }
  }

  const createEmployeeHistory = async (employeeID: number, payload: EmployeeHistoryPUTFormData) => {
    try {
      const response = await $fetch<EmployeeHistoryResponse>(`/api/employees/${employeeID}/history`, {
        method: 'POST',
        body: {
          ...payload,
          salary: payload.salary ? AmountToInteger(payload.salary) : undefined,
          fromDate: payload.fromDate ? DateToApiFormat(payload.fromDate) : undefined,
        },
      })
      await listEmployeeHistory(employeeID)
      return response
    }
    catch {
      return Promise.reject('Fehler beim Erstellen der Historie')
    }
  }

  const updateEmployeeHistory = async (employeeID: number, payload: EmployeeHistoryPATCHFormData) => {
    try {
      await $fetch<EmployeeHistoryResponse>(`/api/employees/history/${payload.id}`, {
        method: 'PATCH',
        body: {
          ...payload,
          salary: payload.salary ? AmountToInteger(payload.salary) : undefined,
          fromDate: payload.fromDate ? DateToApiFormat(payload.fromDate) : undefined,
        },
      })
      await listEmployeeHistory(employeeID)
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren der Historie')
    }
  }

  const deleteEmployeeHistory = async (employeeID: number, employeeHistoryID: number) => {
    try {
      await $fetch(`/api/employees/history/${employeeHistoryID}`, {
        method: 'DELETE',
      })
      await listEmployeeHistory(employeeID)
    }
    catch {
      return Promise.reject('Fehler beim Löschen der Historie')
    }
  }

  const setEmployeeHistories = (data: ListEmployeeHistoryResponse | null, append: boolean) => {
    if (data) {
      if (append) {
        employeeHistories.value!.data = employeeHistories.value!.data.concat(data.data ?? [])
        employeeHistories.value!.pagination = data.pagination
      }
      else {
        employeeHistories.value = data
      }
      noMoreDataEmployeeHistories.value = employeeHistories.value.pagination.totalRemaining == 0
    }
    else {
      employeeHistories.value = DefaultListResponse()
    }
  }

  return {
    employeeHistories,
    limitEmployeeHistories,
    pageEmployeeHistories,
    noMoreDataEmployeeHistories,
    useFetchListEmployeeHistory,
    listEmployeeHistory,
    getEmployeeHistory,
    createEmployeeHistory,
    updateEmployeeHistory,
    deleteEmployeeHistory,
    setEmployeeHistories,
  }
}
