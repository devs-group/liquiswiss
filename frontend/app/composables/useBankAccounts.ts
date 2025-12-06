import type { BankAccountFormData, BankAccountResponse } from '~/models/bank-account'

export default function useBankAccounts() {
  const bankAccounts = useState<BankAccountResponse[]>('bankAccounts', () => [])

  const { convertAmountToRate } = useGlobalData()

  const useFetchListBankAccounts = async () => {
    const { data, error } = await useFetch<BankAccountResponse[]>('/api/bank-accounts', {
      method: 'GET',
    })
    if (error.value || !data.value) {
      return Promise.reject('Bankkonten konnten nicht geladen werden')
    }
    setBankAccounts(data.value, false)
  }

  const listBankAccounts = async () => {
    try {
      const data = await $fetch<BankAccountResponse[]>('/api/bank-accounts', {
        method: 'GET',
      })
      setBankAccounts(data, false)
    }
    catch {
      return Promise.reject('Fehler beim Laden der Bankkonten')
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
      await listBankAccounts()
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
      await listBankAccounts()
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
      await listBankAccounts()
    }
    catch {
      return Promise.reject('Fehler beim Löschen des Bankkontos')
    }
  }

  const setBankAccounts = (data: BankAccountResponse[] | null, append: boolean) => {
    if (data) {
      if (append) {
        bankAccounts.value = bankAccounts.value.concat(data ?? [])
      }
      else {
        bankAccounts.value = data
      }
    }
    else {
      bankAccounts.value = []
    }
  }

  const totalBankSaldoInCHF = computed(() => {
    return bankAccounts.value.reduce((previousValue, currentValue) => {
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
    totalBankSaldoInCHF,
  }
}
