import type { FetchError } from 'ofetch'
import type { BankAccountFormData, BankAccountResponse, ListBankAccountResponse } from '~/models/bank-account'
import { DefaultListResponse } from '~/models/default-data'

export default function useBankAccounts() {
  const limitBankAccounts = useState('limitBankAccounts', () => 20)
  const pageBankAccounts = useState('pageBankAccounts', () => 1)
  const noMoreDataBankAccounts = useState('noMoreDataBankAccounts', () => false)
  const searchBankAccounts = useState('searchBankAccounts', () => '')
  const bankAccounts = useState<ListBankAccountResponse>('bankAccounts', () => DefaultListResponse())

  const { convertAmountToRate } = useGlobalData()
  const { bankAccountSortBy, bankAccountSortOrder } = useSettings()

  const useFetchListBankAccounts = async () => {
    const { data, error } = await useFetch<ListBankAccountResponse>('/api/bank-accounts', {
      method: 'GET',
      query: {
        page: pageBankAccounts.value,
        limit: limitBankAccounts.value,
        sortBy: bankAccountSortBy.value,
        sortOrder: bankAccountSortOrder.value,
        search: searchBankAccounts.value,
      },
    })
    if (error.value) {
      return Promise.reject('Bankkonten konnten nicht geladen werden')
    }
    setBankAccounts(data.value, false)
  }

  const listBankAccounts = async (append: boolean) => {
    try {
      const data = await $fetch<ListBankAccountResponse>('/api/bank-accounts', {
        method: 'GET',
        query: {
          page: pageBankAccounts.value,
          limit: limitBankAccounts.value,
          sortBy: bankAccountSortBy.value,
          sortOrder: bankAccountSortOrder.value,
          search: searchBankAccounts.value,
        },
      })
      setBankAccounts(data, append)
    }
    catch (err: unknown) {
      if (IsAbortedError(err as FetchError)) {
        return Promise.reject('aborted')
      }
      else {
        return Promise.reject('Fehler beim Laden der Bankkonten')
      }
    }
  }

  const getBankAccount = async (bankAccountID: number) => {
    try {
      return await $fetch<BankAccountResponse>(`/api/bank-accounts/${bankAccountID}`, {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject('Fehler beim Laden des Bankkontos')
    }
  }

  const createBankAccount = async (payload: BankAccountFormData) => {
    try {
      await $fetch<BankAccountResponse>(`/api/bank-accounts`, {
        method: 'POST',
        body: {
          ...payload,
          amount: AmountToInteger(payload.amount),
        },
      })
      await listBankAccounts(false)
    }
    catch {
      return Promise.reject('Fehler beim Erstellen des Bankkontos')
    }
  }

  const updateBankAccount = async (payload: BankAccountFormData) => {
    try {
      await $fetch<BankAccountResponse>(`/api/bank-accounts/${payload.id}`, {
        method: 'PATCH',
        body: {
          ...payload,
          amount: AmountToInteger(payload.amount),
        },
      })
      await listBankAccounts(false)
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren des Bankkontos')
    }
  }

  const deleteBankAccount = async (bankAccountID: number) => {
    try {
      await $fetch(`/api/bank-accounts/${bankAccountID}`, {
        method: 'DELETE',
      })
      await listBankAccounts(false)
    }
    catch {
      return Promise.reject('Fehler beim Löschen des Bankkontos')
    }
  }

  const setBankAccounts = (data: ListBankAccountResponse | null, append: boolean) => {
    if (data) {
      if (append) {
        bankAccounts.value!.data = bankAccounts.value!.data.concat(data.data ?? [])
        bankAccounts.value!.pagination = data.pagination
      }
      else {
        bankAccounts.value = data
      }
      noMoreDataBankAccounts.value = bankAccounts.value.pagination.totalRemaining == 0
    }
    else {
      bankAccounts.value = DefaultListResponse()
    }
  }

  const totalBankSaldoInCHF = computed(() => {
    return bankAccounts.value.data.reduce((previousValue, currentValue) => {
      return previousValue + convertAmountToRate(currentValue.amount, currentValue.currency.code)
    }, 0)
  })

  return {
    useFetchListBankAccounts,
    listBankAccounts,
    getBankAccount,
    createBankAccount,
    updateBankAccount,
    deleteBankAccount,
    setBankAccounts,
    bankAccounts,
    limitBankAccounts,
    pageBankAccounts,
    noMoreDataBankAccounts,
    searchBankAccounts,
    totalBankSaldoInCHF,
  }
}
