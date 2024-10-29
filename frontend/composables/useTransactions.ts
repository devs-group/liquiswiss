import {ref} from 'vue';
import {DefaultListResponse} from "~/models/classes";
import type {ListTransactionResponse, TransactionFormData, TransactionResponse} from "~/models/transaction";
import {IsAbortedError} from "~/utils/error-helper";

const limitTransactions = ref(20)
const pageTransactions = ref(1)
const noMoreDataTransactions = ref(false)
const transactions = ref<ListTransactionResponse>(new DefaultListResponse());
const abortController = ref<AbortController|null>(null)

export default function useTransactions() {
    const {transactionSortBy, transactionSortOrder} = useSettings()

    const listTransactions = async (append: boolean)  => {
        if (abortController.value) {
            abortController.value.abort()
        }
        abortController.value = new AbortController()

        const {data, status, error} = await useFetch<ListTransactionResponse>('/api/transactions', {
            method: 'GET',
            query: {
                page: pageTransactions.value,
                limit: limitTransactions.value,
                sortBy: transactionSortBy.value,
                sortOrder: transactionSortOrder.value,
            },
            signal: abortController.value.signal,
        });

        if (IsAbortedError(error.value)) {
            return Promise.reject('aborted')
        } else if (status.value === 'error') {
            return Promise.reject('Fehler beim Laden der Transaktionen')
        } else {
            if (data.value) {
                if (append) {
                    transactions.value!.data = transactions.value!.data.concat(data.value?.data ?? [])
                    transactions.value!.pagination = data.value?.pagination
                } else {
                    transactions.value = data.value
                }
                noMoreDataTransactions.value = transactions.value.pagination.totalRemaining == 0
            } else {
                transactions.value = new DefaultListResponse()
            }
        }
        return Promise.resolve()
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
        const {data, status} = await useFetch<TransactionResponse>(`/api/transactions/${transactionID}`, {
            method: 'GET',
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Laden der Transaktion')
        } else {
        }
        return Promise.resolve(data.value)
    }

    const createTransaction = async (payload: TransactionFormData) => {
        let id = 0

        const {data, status} = await useFetch<TransactionResponse>(`/api/transactions`, {
            method: 'POST',
            body: {
                ...payload,
                amount: AmountToInteger(payload.amount),
                startDate: DateToApiFormat(payload.startDate),
                endDate: payload.endDate ? DateToApiFormat(payload.endDate) : null,
            },
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Erstellen der Transaktion')
        } else {
            await listTransactions(false)
            if (data.value) {
                id = data.value.id
            }
        }
        return Promise.resolve(id)
    }

    const updateTransaction = async (payload: TransactionFormData) => {
        const {status} = await useFetch<TransactionResponse>(`/api/transactions/${payload.id}`, {
            method: 'PATCH',
            body: {
                ...payload,
                amount: AmountToInteger(payload.amount),
                startDate: DateToApiFormat(payload.startDate),
                endDate: payload.endDate ? DateToApiFormat(payload.endDate) : null,
            },
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Aktualisieren der Transaktion')
        } else {
            await listTransactions(false)
        }
        return Promise.resolve()
    }

    const deleteTransaction = async (transactionID: number) => {
        const {status} = await useFetch(`/api/transactions/${transactionID}`, {
            method: 'DELETE',
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Aktualisieren der Transaktion')
        } else {
            await listTransactions(false)
        }
        return Promise.resolve()
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
