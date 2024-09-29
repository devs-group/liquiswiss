import {ref} from 'vue';
import {DefaultListResponse} from "~/models/classes";
import type {ListTransactionResponse, TransactionFormData, TransactionResponse} from "~/models/transaction";

const limitTransactions = ref(20)
const pageTransactions = ref(1)
const noMoreDataTransactions = ref(false)
const transactions = ref<ListTransactionResponse>(new DefaultListResponse());

export default function useTransactions() {
    const listTransactions = async (append: boolean)  => {
        try {
            const {data} = await useFetch<ListTransactionResponse>('/api/transactions', {
                method: 'GET',
                query: {
                    page: pageTransactions.value,
                    limit: limitTransactions.value,
                }
            });
            if (data.value) {
                if (append) {
                    transactions.value!.data = transactions.value!.data.concat(data.value?.data ?? [])
                    transactions.value!.pagination = data.value?.pagination
                } else {
                    transactions.value = data.value
                }
                noMoreDataTransactions.value = transactions.value.pagination.totalRemaining == 0
            }
        } catch (error) {
            console.error('Error listing transactions:', error);
        }
    }

    // const getTransactionPagination = async ()  => {
    //     try {
    //         const {data} = await useFetch<PaginationResponse>('/api/employees/pagination', {
    //             method: 'GET',
    //             query: {
    //                 // Can always be one
    //                 page: 1,
    //                 limit: limitTransactions.value,
    //             }
    //         });
    //         if (data.value) {
    //             employees.value!.pagination = data.value
    //         }
    //     } catch (error) {
    //         console.error('Error loading employees pagination:', error);
    //     }
    // }

    const getTransaction = async (transactionID: number) => {
        try {
            const {data} = await useFetch<TransactionResponse>(`/api/transactions/${transactionID}`, {
                method: 'GET',
            });
            return data.value
        } catch (error) {
            console.error('Error getting transaction:', error);
        }
        return null
    }

    const createTransaction = async (payload: TransactionFormData) => {
        let id = 0
        try {
            const {data} = await useFetch<TransactionResponse>(`/api/transactions`, {
                method: 'POST',
                body: {
                    ...payload,
                    amount: AmountToInteger(payload.amount),
                    startDate: DateToApiFormat(payload.startDate),
                    endDate: payload.endDate ? DateToApiFormat(payload.endDate) : null,
                },
            });
            await listTransactions(false)
            // Update data list in frontend
            if (data.value) {
                id = data.value.id
            }
            // Update Pagination from backend
            // await getEmployeesPagination()
        } catch (error) {
            console.error('Error creating transaction:', error);
        }
        return id
    }

    const updateTransaction = async (payload: TransactionFormData) => {
        try {
            const {data} = await useFetch<TransactionResponse>(`/api/transactions/${payload.id}`, {
                method: 'PATCH',
                body: {
                    ...payload,
                    amount: AmountToInteger(payload.amount),
                    startDate: DateToApiFormat(payload.startDate),
                    endDate: payload.endDate ? DateToApiFormat(payload.endDate) : null,
                },
            });
            await listTransactions(false)
            // Update Pagination from backend
            // await getEmployeesPagination()
        } catch (error) {
            console.error('Error updating transaction:', error);
        }
    }

    const deleteTransaction = async (transactionID: number) => {
        try {
            const {data} = await useFetch(`/api/transactions/${transactionID}`, {
                method: 'DELETE',
            });
            await listTransactions(false)
            // Update Pagination from backend
            // await getEmployeesPagination()
        } catch (error) {
            console.error('Error deleting employee:', error);
        }
    }

    return {
        transactions,
        limitTransactions,
        pageTransactions,
        noMoreDataTransactions,
        listTransactions,
        getTransaction,
        createTransaction,
        updateTransaction,
        deleteTransaction,
    };
}
