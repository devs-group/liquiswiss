import { ref } from 'vue';
import type {
    EmployeeFormData,
    EmployeeHistoryFormData, EmployeeHistoryResponse,
    EmployeeResponse, ListEmployeeHistoryResponse,
    ListEmployeeResponse
} from "~/models/employee";
import type {PaginationResponse} from "~/models/pagination";
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
        try {
            const {data} = await useFetch<ListEmployeeResponse>('/api/employees', {
                method: 'GET',
                query: {
                    page: pageEmployees.value,
                    limit: limitEmployees.value,
                }
            });
            if (data.value) {
                if (append) {
                    employees.value!.data = employees.value!.data.concat(data.value?.data ?? [])
                    employees.value!.pagination = data.value?.pagination
                } else {
                    employees.value = data.value
                }
                noMoreDataEmployees.value = employees.value.pagination.totalRemaining == 0
            }
        } catch (error) {
            console.error('Error listing employees:', error);
        }
    }

    const getEmployeesPagination = async ()  => {
        try {
            const {data} = await useFetch<PaginationResponse>('/api/employees/pagination', {
                method: 'GET',
                query: {
                    // Can always be one
                    page: 1,
                    limit: limitEmployees.value,
                }
            });
            if (data.value) {
                employees.value!.pagination = data.value
            }
        } catch (error) {
            console.error('Error loading employees pagination:', error);
        }
    }

    const getEmployee = async (id: number) => {
        try {
            const {data} = await useFetch<EmployeeResponse>(`/api/employees/${id}`, {
                method: 'GET',
            });
            return data.value
        } catch (error) {
            console.error('Error getting employee:', error);
        }
        return null
    }

    const getEmployeeHistory = async (employeeID: number) => {
        try {
            const {data} = await useFetch<ListEmployeeHistoryResponse>(`/api/employees/${employeeID}/history`, {
                method: 'GET',
                query: {
                    page: pageEmployeeHistories.value,
                    limit: limitEmployeeHistories.value,
                }
            });
            if (data.value) {
                employeeHistories.value = data.value
            } else {
                employeeHistories.value = new DefaultListResponse()
            }
            noMoreDataEmployeeHistories.value = employeeHistories.value.pagination.totalRemaining == 0
        } catch (error) {
            console.error('Error getting employee:', error);
        }
    }

    const createEmployee = async (payload: EmployeeFormData) => {
        let id = 0
        try {
            const {data} = await useFetch<EmployeeResponse>(`/api/employees`, {
                method: 'POST',
                body: payload,
            });
            // Update data list in frontend
            if (data.value) {
                employees.value!.data.push(data.value)
                id = data.value.id
            }
            // Update Pagination from backend
            await getEmployeesPagination()
        } catch (error) {
            console.error('Error creating employee:', error);
        }
        return id
    }

    const createEmployeeHistory = async (employeeID: number, payload: EmployeeHistoryFormData) => {
        try {
            await useFetch<EmployeeHistoryResponse>(`/api/employees/${employeeID}/history`, {
                method: 'POST',
                body: {
                    ...payload,
                    salaryPerMonth: payload.salaryPerMonth * 100,
                    fromDate: DateToApiFormat(payload.fromDate),
                    toDate: payload.toDate ? DateToApiFormat(payload.toDate) : undefined,
                },
            });
            await getEmployeeHistory(employeeID)
        } catch (error) {
            console.error('Error creating employee:', error);
        }
        return []
    }

    const updateEmployee = async (payload: EmployeeFormData) => {
        try {
            const {data} = await useFetch<EmployeeResponse>(`/api/employees/${payload.id}`, {
                method: 'PATCH',
                body: payload,
            });
            // Update data list in frontend
            if (data.value) {
                employees.value!.data = employees.value!.data.map(
                    employee => employee.id === data.value?.id ? data.value : employee
                )
            }
            // Update Pagination from backend
            await getEmployeesPagination()
        } catch (error) {
            console.error('Error updating employee:', error);
        }
    }

    const updateEmployeeHistory = async (employeeID: number, payload: EmployeeHistoryFormData) => {
        try {
            await useFetch<EmployeeHistoryResponse>(`/api/employees/history/${payload.id}`, {
                method: 'PATCH',
                body: {
                    ...payload,
                    salaryPerMonth: payload.salaryPerMonth * 100,
                    fromDate: DateToApiFormat(payload.fromDate),
                    toDate: payload.toDate ? DateToApiFormat(payload.toDate) : undefined,
                },
            });
            await getEmployeeHistory(employeeID)
        } catch (error) {
            console.error('Error updating employee history:', error);
        }
    }

    const deleteEmployee = async (id: number) => {
        try {
            const {data} = await useFetch(`/api/employees/${id}`, {
                method: 'DELETE',
            });
            // Update data list in frontend
            employees.value!.data = employees.value!.data.filter(employee => employee.id !== id)
            // Update Pagination from backend
            await getEmployeesPagination()
        } catch (error) {
            console.error('Error deleting employee:', error);
        }
    }

    const deleteEmployeeHistory = async (employeeID: number, id: number) => {
        try {
            await useFetch(`/api/employees/history/${id}`, {
                method: 'DELETE',
            });
            await getEmployeeHistory(employeeID)
        } catch (error) {
            console.error('Error deleting employee history:', error);
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
        getEmployees,
        getEmployeesPagination,
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
