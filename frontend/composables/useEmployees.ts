import type {
    EmployeeFormData,
    EmployeeHistoryFormData, EmployeeHistoryResponse,
    EmployeeResponse, ListEmployeeHistoryResponse,
    ListEmployeeResponse
} from "~/models/employee";
import {DefaultListResponse} from "~/models/default-data";
import {IsAbortedError} from "~/utils/error-helper";

export default function useEmployees() {
    const limitEmployees = useState('limitEmployees', () => 20)
    const pageEmployees = useState('pageEmployees', () => 1)
    const noMoreDataEmployees = useState('noMoreDataEmployees', () => false)
    const limitEmployeeHistories = useState('limitEmployeeHistories', () => 20)
    const pageEmployeeHistories = useState('pageEmployeeHistories', () => 1)
    const noMoreDataEmployeeHistories = useState('noMoreDataEmployeeHistories', () => false)
    const employees = useState<ListEmployeeResponse>('employees', () => DefaultListResponse());
    const employeeHistories = useState<ListEmployeeHistoryResponse>('employeeHistories', () => DefaultListResponse())

    const {employeeSortBy, employeeSortOrder} = useSettings()

    const useFetchListEmployees = async () => {
        const {data, error} = await useFetch<ListEmployeeResponse>('/api/employees', {
            method: 'GET',
            query: {
                page: pageEmployees.value,
                limit: limitEmployees.value,
                sortBy: employeeSortBy.value,
                sortOrder: employeeSortOrder.value,
            }
        });
        if (error.value) {
            return Promise.reject('Mitarbeiter konnten nicht geladen werden')
        }
        setEmployees(data.value, false)
    }

    const listEmployees = async (append: boolean)  => {
        try {
            const data = await $fetch<ListEmployeeResponse>('/api/employees', {
                method: 'GET',
                query: {
                    page: pageEmployees.value,
                    limit: limitEmployees.value,
                    sortBy: employeeSortBy.value,
                    sortOrder: employeeSortOrder.value,
                }
            });
            setEmployees(data, append)
        } catch (err: any) {
            if (IsAbortedError(err)) {
                return Promise.reject('aborted')
            } else {
                return Promise.reject('Fehler beim Laden der Mitarbeiter')
            }
        }
    }

    const useFetchGetEmployee = async (employeeID: number) => {
        const {data, error} = await useFetch<EmployeeResponse>(`/api/employees/${employeeID}`, {
            method: 'GET',
        });
        if (error.value) {
            return Promise.reject(`Mitarbeiter mit ID "${employeeID}" konnte nicht geladen werden`)
        }
        return data.value
    }

    const getEmployee = async (employeeID: number) => {
        try {
            return await $fetch<EmployeeResponse>(`/api/employees/${employeeID}`, {
                method: 'GET',
            });
        } catch (err) {
            return Promise.reject(`Mitarbeiter mit ID "${employeeID}" konnte nicht geladen werden`)
        }
    }

    const createEmployee = async (payload: EmployeeFormData) => {
        let id = 0

        try {
            const data = await $fetch<EmployeeResponse>(`/api/employees`, {
                method: 'POST',
                body: payload,
            });
            id = data.id
        } catch (err) {
            return Promise.reject('Fehler beim Erstellen des Mitarbeiters')
        }

        return Promise.resolve(id)
    }

    const updateEmployee = async (payload: EmployeeFormData) => {
        try {
            return await $fetch<EmployeeResponse>(`/api/employees/${payload.id}`, {
                method: 'PATCH',
                body: payload,
            });
        } catch (err) {
            return Promise.reject('Fehler beim Aktualisieren des Mitarbeiters')
        }
    }

    const deleteEmployee = async (employeeID: number) => {
        try {
            await $fetch(`/api/employees/${employeeID}`, {
                method: 'DELETE',
            });
            employees.value!.data = employees.value!.data.filter(employee => employee.id !== employeeID)
        } catch (err) {
            return Promise.reject('Fehler beim Löschen des Mitarbeiters')
        }
    }

    const useFetchListEmployeeHistory = async (employeeID: number) => {
        const {data, error} = await useFetch<ListEmployeeHistoryResponse>(`/api/employees/${employeeID}/history`, {
            method: 'GET',
            query: {
                page: pageEmployeeHistories.value,
                limit: limitEmployeeHistories.value,
            }
        });
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
                }
            });
            setEmployeeHistories(data, false)
        } catch (err) {
            return Promise.reject('Fehler beim Laden der Historie')
        }
    }

    const createEmployeeHistory = async (employeeID: number, payload: EmployeeHistoryFormData) => {
        try {
            await $fetch<EmployeeHistoryResponse>(`/api/employees/${employeeID}/history`, {
                method: 'POST',
                body: {
                    ...payload,
                    salaryPerMonth: AmountToInteger(payload.salaryPerMonth),
                    fromDate: DateToApiFormat(payload.fromDate),
                    toDate: payload.toDate ? DateToApiFormat(payload.toDate) : undefined,
                },
            });
            await listEmployeeHistory(employeeID)
        } catch (err) {
            return Promise.reject('Fehler beim Erstellen der Historie')
        }
    }

    const updateEmployeeHistory = async (employeeID: number, payload: EmployeeHistoryFormData) => {
        try {
            await $fetch<EmployeeHistoryResponse>(`/api/employees/history/${payload.id}`, {
                method: 'PATCH',
                body: {
                    ...payload,
                    salaryPerMonth: AmountToInteger(payload.salaryPerMonth),
                    fromDate: DateToApiFormat(payload.fromDate),
                    toDate: payload.toDate ? DateToApiFormat(payload.toDate) : undefined,
                },
            });
            await listEmployeeHistory(employeeID)
        } catch (err) {
            return Promise.reject('Fehler beim Aktualisieren der Historie')
        }
    }

    const deleteEmployeeHistory = async (employeeID: number, employeeHistoryID: number) => {
        try {
            await $fetch(`/api/employees/history/${employeeHistoryID}`, {
                method: 'DELETE',
            });
            await listEmployeeHistory(employeeID)
        } catch (err) {
            return Promise.reject('Fehler beim Löschen der Historie')
        }
    }

    const setEmployees = (data: ListEmployeeResponse|null, append: boolean) => {
        if (data) {
            if (append) {
                employees.value!.data = employees.value!.data.concat(data.data ?? [])
                employees.value!.pagination = data.pagination
            } else {
                employees.value = data
            }
            noMoreDataEmployees.value = employees.value.pagination.totalRemaining == 0
        } else {
            employees.value = DefaultListResponse()
        }
    }

    const setEmployeeHistories = (data: ListEmployeeHistoryResponse|null, append: boolean) => {
        if (data) {
            if (append) {
                employeeHistories.value!.data = employeeHistories.value!.data.concat(data.data ?? [])
                employeeHistories.value!.pagination = data.pagination
            } else {
                employeeHistories.value = data
            }
            noMoreDataEmployeeHistories.value = employeeHistories.value.pagination.totalRemaining == 0
        } else {
            employeeHistories.value = DefaultListResponse()
        }
    }

    return {
        employees,
        limitEmployees,
        pageEmployees,
        noMoreDataEmployees,
        employeeHistories,
        limitEmployeeHistories,
        pageEmployeeHistories,
        noMoreDataEmployeeHistories,
        useFetchListEmployees,
        listEmployees,
        // getEmployeesPagination,
        useFetchGetEmployee,
        getEmployee,
        createEmployee,
        updateEmployee,
        deleteEmployee,
        useFetchListEmployeeHistory,
        listEmployeeHistory,
        createEmployeeHistory,
        updateEmployeeHistory,
        deleteEmployeeHistory,
        setEmployees,
        setEmployeeHistories,
    };
}
