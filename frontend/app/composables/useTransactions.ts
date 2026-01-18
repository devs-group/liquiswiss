import type { FetchError } from 'ofetch'
import type { ListTransactionResponse, TransactionFormData, TransactionPatchData, TransactionResponse } from '~/models/transaction'
import { DefaultListResponse } from '~/models/default-data'

export default function useTransactions() {
  const limitTransactions = useState('limitTransactions', () => 50)
  const pageTransactions = useState('pageTransactions', () => 1)
  const noMoreDataTransactions = useState('noMoreDataTransactions', () => false)
  const transactions = useState<ListTransactionResponse>('transactions', () => DefaultListResponse())
  const searchTransactions = useState('searchTransactions', () => '')

  const abortController = ref<AbortController | null>(null)

  const { transactionSortBy, transactionSortOrder } = useSettings()

  const useFetchListTransactions = async () => {
    const { data, error } = await useFetch<ListTransactionResponse>('/api/transactions', {
      method: 'GET',
      query: {
        page: pageTransactions.value,
        limit: limitTransactions.value,
        sortBy: transactionSortBy.value,
        sortOrder: transactionSortOrder.value,
        search: searchTransactions.value,
      },
    })
    if (error.value) {
      return Promise.reject('Transaktionen konnten nicht geladen werden')
    }
    setTransactions(data.value, false)
  }

  const listTransactions = async (append: boolean) => {
    if (abortController.value) {
      abortController.value.abort()
    }
    abortController.value = new AbortController()

    try {
      const data = await $fetch<ListTransactionResponse>('/api/transactions', {
        method: 'GET',
        query: {
          page: pageTransactions.value,
          limit: limitTransactions.value,
          sortBy: transactionSortBy.value,
          sortOrder: transactionSortOrder.value,
          search: searchTransactions.value,
        },
        signal: abortController.value.signal,
      })
      setTransactions(data, append)
    }
    catch (err: unknown) {
      if (IsAbortedError(err as FetchError)) {
        return Promise.reject('aborted')
      }
      else {
        return Promise.reject('Fehler beim Laden der Transaktionen')
      }
    }
  }

  const createTransaction = async (payload: TransactionFormData) => {
    try {
      await $fetch<TransactionResponse>(`/api/transactions`, {
        method: 'POST',
        body: {
          ...payload,
          amount: AmountToInteger(payload.amount),
          startDate: DateToApiFormat(payload.startDate),
          endDate: payload.endDate ? DateToApiFormat(payload.endDate) : null,
        },
      })
      await listTransactions(false)
    }
    catch {
      return Promise.reject('Fehler beim Erstellen der Transaktion')
    }
  }

  const updateTransaction = async (payload: TransactionFormData) => {
    try {
      await $fetch<TransactionResponse>(`/api/transactions/${payload.id}`, {
        method: 'PATCH',
        body: {
          ...payload,
          amount: AmountToInteger(payload.amount),
          startDate: DateToApiFormat(payload.startDate),
          endDate: payload.endDate ? DateToApiFormat(payload.endDate) : null,
        },
      })
      await listTransactions(false)
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren der Transaktion')
    }
  }

  const deleteTransaction = async (transactionID: number) => {
    try {
      await $fetch(`/api/transactions/${transactionID}`, {
        method: 'DELETE',
      })
      await listTransactions(false)
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren der Transaktion')
    }
  }

  const patchTransaction = async (payload: TransactionPatchData) => {
    try {
      const body: Record<string, unknown> = { ...payload }
      if (body.amount !== undefined) {
        body.amount = AmountToInteger(body.amount as number)
      }
      if (body.startDate instanceof Date) {
        body.startDate = DateToApiFormat(body.startDate)
      }
      if (body.endDate instanceof Date) {
        body.endDate = DateToApiFormat(body.endDate)
      }
      await $fetch<TransactionResponse>(`/api/transactions/${payload.id}`, {
        method: 'PATCH',
        body: {
          ...body,
        },
      })
      await listTransactions(false)
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren der Transaktion')
    }
  }

  const setTransactions = (data: ListTransactionResponse | null, append: boolean) => {
    if (data) {
      if (append) {
        transactions.value!.data = transactions.value!.data.concat(data.data ?? [])
        transactions.value!.pagination = data.pagination
      }
      else {
        transactions.value = data
      }
      noMoreDataTransactions.value = transactions.value.pagination.totalRemaining == 0
    }
    else {
      transactions.value = DefaultListResponse()
    }
  }

  return {
    transactions,
    limitTransactions,
    pageTransactions,
    noMoreDataTransactions,
    searchTransactions,
    useFetchListTransactions,
    listTransactions,
    createTransaction,
    updateTransaction,
    patchTransaction,
    deleteTransaction,
    setTransactions,
  }
}
