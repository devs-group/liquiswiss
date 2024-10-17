<template>
  <div class="flex flex-col gap-4">
    <Button @click="onCreateBankAccount" class="self-end" label="Bankkonto hinzufügen" icon="pi pi-building"/>

    <p class="text-sm font-bold text-right">Gesamtvermögen: ~ {{totalSaldo}} CHF</p>

    <Message v-if="bankAccountsErrorMessage.length" severity="error" :closable="false" class="col-span-full">{{bankAccountsErrorMessage}}</Message>
    <div v-else-if="bankAccounts?.length" class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <BankAccountCard @on-edit="onEditBankAccount" v-for="bankAccount in bankAccounts" :bankAccount="bankAccount"/>
    </div>
    <p v-else>Es gibt noch keine Bankkonten</p>
  </div>
</template>

<script setup lang="ts">
import {ModalConfig} from "~/config/dialog-props";
import BankAccountDialog from "~/components/dialogs/BankAccountDialog.vue";
import type {BankAccountResponse} from "~/models/bank-account";
import BankAccountCard from "~/components/BankAccountCard.vue";
import useBankAccounts from "~/composables/useBankAccounts";
import {Constants} from "~/utils/constants";

const dialog = useDialog();
const {bankAccounts, listBankAccounts} = useBankAccounts()
const {convertAmountToRate} = useGlobalData()

const bankAccountsErrorMessage = ref('')

const totalSaldo = computed(() => {
  const totalBankSaldo = bankAccounts.value.reduce((previousValue, currentValue) => {
    return previousValue + convertAmountToRate(currentValue.amount ,currentValue.currency.code)
  }, 0)
  return NumberToFormattedCurrency(AmountToFloat(totalBankSaldo), Constants.BASE_LOCALE_CODE)
})

await listBankAccounts(false)
    .catch(() => {
      bankAccountsErrorMessage.value = 'Bankkonten konnten nicht geladen werden'
    })

const onCreateBankAccount = () => {
  dialog.open(BankAccountDialog, {
    props: {
      header: 'Neues Bankkonto anlegen',
      ...ModalConfig,
    },
  })
}

const onEditBankAccount = (bankAccount: BankAccountResponse) => {
  dialog.open(BankAccountDialog, {
    data: {
      bankAccount: bankAccount,
    },
    props: {
      header: 'Bankkonto bearbeiten',
      ...ModalConfig,
    },
  })
}
</script>