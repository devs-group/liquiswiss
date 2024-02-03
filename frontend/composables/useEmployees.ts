import { ref } from 'vue';
import type {EmployeeFormData, EmployeeResponse, ListEmployeeResponse} from "~/models/employee";
import type {PaginationResponse} from "~/models/pagination";
import {DateToApiFormat} from "~/utils/format-helper";

const limit = ref(20)
const page = ref(1)
const noMoreData = ref(false)
const employees = ref<ListEmployeeResponse>({
    data: [],
    pagination: {
        currentPage: 1,
        totalCount: 0,
        totalPages: 0,
        totalRemaining: 0,
    }
});

export default function useEmployees() {
    const getEmployees = async (append: boolean)  => {
        try {
            const {data} = await useFetch<ListEmployeeResponse>('/api/employees', {
                method: 'GET',
                query: {
                    page: page.value,
                    limit: limit.value,
                }
            });
            if (data.value) {
                if (append) {
                    employees.value!.data = employees.value!.data.concat(data.value?.data ?? [])
                    employees.value!.pagination = data.value?.pagination
                } else {
                    employees.value = data.value
                }
                noMoreData.value = employees.value.pagination.totalRemaining == 0
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
                    limit: limit.value,
                }
            });
            if (data.value) {
                employees.value!.pagination = data.value
            }
        } catch (error) {
            console.error('Error loading employees pagination:', error);
        }
    }

    const getEmployee = async (id: string) => {
        try {
            const {data} = await useFetch(`/api/employees/${id}`, {
                method: 'GET',
            });
            return data.value
        } catch (error) {
            console.error('Error getting employee:', error);
        }
    }

    const createEmployee = async (payload: EmployeeFormData) => {
        try {
            const {data} = await useFetch<EmployeeResponse>(`/api/employees`, {
                method: 'POST',
                body: {
                    ...payload,
                    entryDate: DateToApiFormat(payload.entryDate),
                    exitDate: payload.exitDate ? DateToApiFormat(payload.exitDate) : undefined,
                },
            });
            // Update data list in frontend
            if (data.value) {
                employees.value!.data.push(data.value)
            }
            // Update Pagination from backend
            await getEmployeesPagination()
        } catch (error) {
            console.error('Error creating employee:', error);
        }
    }

    const updateEmployee = async (payload: EmployeeFormData) => {
        try {
            const {data} = await useFetch<EmployeeResponse>(`/api/employees/${payload.id}`, {
                method: 'PATCH',
                body: {
                    ...payload,
                    entryDate: DateToApiFormat(payload.entryDate),
                    exitDate: payload.exitDate ? DateToApiFormat(payload.exitDate) : undefined,
                },
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

    return {
        employees,
        limit,
        page,
        noMoreData,
        getEmployees,
        getEmployeesPagination,
        getEmployee,
        createEmployee,
        updateEmployee,
        deleteEmployee,
    };
}
