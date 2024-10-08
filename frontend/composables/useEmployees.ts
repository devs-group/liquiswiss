import { ref } from 'vue';
import type {
    EmployeeFormData,
    EmployeeHistoryFormData, EmployeeHistoryResponse,
    EmployeeResponse, ListEmployeeHistoryResponse,
    ListEmployeeResponse
} from "~/models/employee";
import {DefaultListResponse} from "~/models/classes";

const limitEmployees = ref(20)
const pageEmployees = ref(1)
const noMoreDataEmployees = ref(false)
const limitEmployeeHistories = ref(20)
const pageEmployeeHistories = ref(1)
const noMoreDataEmployeeHistories = ref(false)
const employees = ref<ListEmployeeResponse>(new DefaultListResponse());
const employeeHistories = ref<ListEmployeeHistoryResponse>(new DefaultListResponse())

export default function useEmployees() {
    const getEmployees = async (append: boolean)  => {
        const {data, status} = await useFetch<ListEmployeeResponse>('/api/employees', {
            method: 'GET',
            query: {
                page: pageEmployees.value,
                limit: limitEmployees.value,
            }
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Laden der Mitarbeiter')
        } else {
            if (data.value) {
                if (append) {
                    employees.value!.data = employees.value!.data.concat(data.value?.data ?? [])
                    employees.value!.pagination = data.value?.pagination
                } else {
                    employees.value = data.value
                }
                noMoreDataEmployees.value = employees.value.pagination.totalRemaining == 0
            } else {
                employees.value = new DefaultListResponse()
            }
            noMoreDataEmployees.value = employees.value.pagination.totalRemaining == 0
        }
        return Promise.resolve()
    }

    // const getEmployeesPagination = async ()  => {
    //     try {
    //         const {data} = await useFetch<PaginationResponse>('/api/employees/pagsination', {
    //             method: 'GET',
    //             query: {
    //                 // Can always be one
    //                 page: 1,
    //                 limit: limitEmployees.value,
    //             }
    //         });
    //         if (data.value) {
    //             employees.value!.pagination = data.value
    //         }
    //     } catch (error) {
    //         console.error('Error loading employees pagination:', error);
    //     }
    // }

    const getEmployee = async (id: number) => {
        const {data, status} = await useFetch<EmployeeResponse>(`/api/employees/${id}`, {
            method: 'GET',
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Laden des Mitarbeiters')
        } else {
        }
        return Promise.resolve(data.value)
    }

    const getEmployeeHistory = async (employeeID: number) => {
        const {data, status} = await useFetch<ListEmployeeHistoryResponse>(`/api/employees/${employeeID}/history`, {
            method: 'GET',
            query: {
                page: pageEmployeeHistories.value,
                limit: limitEmployeeHistories.value,
            }
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Laden der Historie')
        } else {
            if (data.value) {
                employeeHistories.value = data.value
            } else {
                employeeHistories.value = new DefaultListResponse()
            }
            noMoreDataEmployeeHistories.value = employeeHistories.value.pagination.totalRemaining == 0
        }
        return Promise.resolve()
    }

    const createEmployee = async (payload: EmployeeFormData) => {
        let id = 0

        const {data, status} = await useFetch<EmployeeResponse>(`/api/employees`, {
            method: 'POST',
            body: payload,
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Erstellen des Mitarbeiters')
        } else {
            if (data.value) {
                id = data.value.id
            }
        }
        return Promise.resolve(id)
    }

    const createEmployeeHistory = async (employeeID: number, payload: EmployeeHistoryFormData) => {
        const {status} = await useFetch<EmployeeHistoryResponse>(`/api/employees/${employeeID}/history`, {
            method: 'POST',
            body: {
                ...payload,
                salaryPerMonth: AmountToInteger(payload.salaryPerMonth),
                fromDate: DateToApiFormat(payload.fromDate),
                toDate: payload.toDate ? DateToApiFormat(payload.toDate) : undefined,
            },
        });
        if (status.value === 'error') {
            return Promise.reject('Fehler beim Erstellen der Historie')
        } else {
            await getEmployeeHistory(employeeID)
        }
        return Promise.resolve()
    }

    const updateEmployee = async (payload: EmployeeFormData) => {
        const {data, status} = await useFetch<EmployeeResponse>(`/api/employees/${payload.id}`, {
            method: 'PATCH',
            body: payload,
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Aktualisieren der Historie')
        }
        return Promise.resolve(data.value)
    }

    const updateEmployeeHistory = async (employeeID: number, payload: EmployeeHistoryFormData) => {
        const {status} = await useFetch<EmployeeHistoryResponse>(`/api/employees/history/${payload.id}`, {
            method: 'PATCH',
            body: {
                ...payload,
                salaryPerMonth: AmountToInteger(payload.salaryPerMonth),
                fromDate: DateToApiFormat(payload.fromDate),
                toDate: payload.toDate ? DateToApiFormat(payload.toDate) : undefined,
            },
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Aktualisieren der Historie')
        } else {
            await getEmployeeHistory(employeeID)
        }
        return Promise.resolve()
    }

    const deleteEmployee = async (id: number) => {
        const {data, status} = await useFetch(`/api/employees/${id}`, {
            method: 'DELETE',
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Löschen der Historie')
        } else {
            employees.value!.data = employees.value!.data.filter(employee => employee.id !== id)
        }
        return Promise.resolve()
    }

    const deleteEmployeeHistory = async (employeeID: number, id: number) => {
        const {status} = await useFetch(`/api/employees/history/${id}`, {
            method: 'DELETE',
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Löschen der Historie')
        } else {
            await getEmployeeHistory(employeeID)
        }
        return Promise.resolve()
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
        getEmployees,
        // getEmployeesPagination,
        getEmployee,
        getEmployeeHistory,
        createEmployee,
        createEmployeeHistory,
        updateEmployee,
        updateEmployeeHistory,
        deleteEmployee,
        deleteEmployeeHistory,
    };
}
