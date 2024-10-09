import {ref} from 'vue';
import type {BankAccountFormData, BankAccountResponse} from "~/models/bank-account";

const bankAccounts = ref<BankAccountResponse[]>([]);

export default function useBankAccounts() {
    const listBankAccounts = async ()  => {
        const {data, status} = await useFetch<BankAccountResponse[]>('/api/bank-accounts', {
            method: 'GET',
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Laden der Bankkonten')
        } else {
            if (data.value) {
                bankAccounts.value = data.value

            } else {
                bankAccounts.value = []
            }
        }
        return Promise.resolve()
    }

    const getBankAccount = async (bankAccountID: number) => {
        const {data, status} = await useFetch<BankAccountResponse>(`/api/bank-accounts/${bankAccountID}`, {
            method: 'GET',
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Laden des Bankkontos')
        } else {
        }
        return Promise.resolve(data.value)
    }

    const createBankAccount = async (payload: BankAccountFormData) => {
        let id = 0

        const {data, status} = await useFetch<BankAccountResponse>(`/api/bank-accounts`, {
            method: 'POST',
            body: {
                ...payload,
                amount: AmountToInteger(payload.amount),
            },
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Erstellen des Bankkontos')
        } else {
            await listBankAccounts(false)
            if (data.value) {
                id = data.value.id
            }
        }
        return Promise.resolve(id)
    }

    const updateBankAccount = async (payload: BankAccountFormData) => {
        const {status} = await useFetch<BankAccountResponse>(`/api/bank-accounts/${payload.id}`, {
            method: 'PATCH',
            body: {
                ...payload,
                amount: AmountToInteger(payload.amount),
            },
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Aktualisieren des Bankkontos')
        } else {
            await listBankAccounts(false)
        }
        return Promise.resolve()
    }

    const deleteBankAccount = async (bankAccountID: number) => {
        const {status} = await useFetch(`/api/bank-accounts/${bankAccountID}`, {
            method: 'DELETE',
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Löschen des Bankkontos')
        } else {
            await listBankAccounts(false)
        }
        return Promise.resolve()
    }

    return {
        bankAccounts,
        listBankAccounts,
        getBankAccount,
        createBankAccount,
        updateBankAccount,
        deleteBankAccount,
    };
}
