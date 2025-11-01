<template>
  <div class="flex flex-col gap-4">
    <div class="flex flex-col sm:flex-row gap-2 justify-between items-center">
      <div class="flex items-center gap-2 w-full sm:w-auto">
        <InputText
          v-model="search"
          placeholder="Suchen"
        />
        <Button
          :icon="getDisplayIcon"
          @click="toggleBankAccountDisplayType"
        />
      </div>
      <Button
        class="self-end"
        label="Bankkonto hinzufügen"
        icon="pi pi-building"
        @click="onCreateBankAccount"
      />
    </div>

    <p class="text-sm font-bold text-right">
      Gesamtvermögen: ~ {{ totalSaldo }} {{ getOrganisationCurrencyCode }}
    </p>

    <Message
      v-if="bankAccountsErrorMessage.length"
      severity="error"
      :closable="false"
      class="col-span-full"
    >
      {{ bankAccountsErrorMessage }}
    </Message>
    <template v-else-if="filterBankAccounts.length">
      <div
        v-if="bankAccountDisplay == 'grid'"
        class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4"
      >
        <BankAccountCard
          v-for="bankAccount in filterBankAccounts"
          :key="bankAccount.id"
          :bank-account="bankAccount"
          @on-edit="onEditBankAccount"
          @on-clone="onCloneBankAccount"
        />
      </div>
      <div
        v-else
        class="flex flex-col overflow-x-auto pb-2"
      >
        <BankAccountHeaders @on-sort="onSort" />
        <BankAccountRow
          v-for="bankAccount in filterBankAccounts"
          :key="bankAccount.id"
          :bank-account="bankAccount"
          @on-edit="onEditBankAccount"
          @on-clone="onCloneBankAccount"
        />
      </div>
    </template>
    <p v-else>
      Es gibt noch keine Bankkonten
    </p>
  </div>
</template>

<script setup lang="ts">
import { ModalConfig } from '~/config/dialog-props'
import BankAccountDialog from '~/components/dialogs/BankAccountDialog.vue'
import type { BankAccountResponse } from '~/models/bank-account'
import BankAccountCard from '~/components/BankAccountCard.vue'
import useBankAccounts from '~/composables/useBankAccounts'

useHead({
  title: 'Bankkonten',
})

const dialog = useDialog()
const { getOrganisationCurrencyCode, getOrganisationCurrencyLocaleCode } = useAuth()
const { bankAccounts, totalBankSaldoInCHF, useFetchListBankAccounts, listBankAccounts } = useBankAccounts()
const { toggleBankAccountDisplayType, bankAccountDisplay } = useSettings()

const isLoading = ref(false)
const search = ref('')
const bankAccountsErrorMessage = ref('')

const totalSaldo = computed(() => {
  return NumberToFormattedCurrency(AmountToFloat(totalBankSaldoInCHF.value), getOrganisationCurrencyLocaleCode.value)
})
const getDisplayIcon = computed(() => bankAccountDisplay.value == 'list' ? 'pi pi-microsoft' : 'pi pi-list')
const filterBankAccounts = computed(() => {
  return bankAccounts.value
    .filter(t => t.name.toLowerCase().includes(search.value.toLowerCase()))
})

await useFetchListBankAccounts()
  .catch((reason) => {
    bankAccountsErrorMessage.value = reason
  })

const onSort = () => {
  isLoading.value = false
  nextTick(() => {
    isLoading.value = true
    listBankAccounts()
      .then(() => {
        isLoading.value = false
      })
      .catch((err) => {
        if (err !== 'aborted') {
          isLoading.value = false
          bankAccountsErrorMessage.value = err
        }
      })
  })
}

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

const onCloneBankAccount = (bankAccount: BankAccountResponse) => {
  dialog.open(BankAccountDialog, {
    data: {
      bankAccount: bankAccount,
      isClone: true,
    },
    props: {
      header: 'Bankkonto klonen',
      ...ModalConfig,
    },
  })
}
</script>
