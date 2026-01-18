<template>
  <div class="flex flex-col gap-4 relative">
    <div class="flex flex-col sm:flex-row gap-2 justify-between items-center">
      <div class="flex items-center gap-2 w-full sm:w-auto">
        <SearchInput
          :model-value="searchBankAccounts"
          @update:model-value="onSearch"
          @clear="onClearSearch"
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
    <template v-else-if="bankAccounts.data.length">
      <div
        v-if="bankAccountDisplay == 'grid'"
        class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4"
      >
        <BankAccountCard
          v-for="bankAccount in bankAccounts.data"
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
          v-for="bankAccount in bankAccounts.data"
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

    <div
      v-if="bankAccounts?.data.length"
      class="self-center"
    >
      <Button
        v-if="!noMoreDataBankAccounts"
        severity="info"
        label="Mehr anzeigen"
        :loading="isLoadingMore"
        @click="onLoadMoreBankAccounts"
      />
      <p
        v-else
        class="text-xs opacity-60"
      >
        Keine weiteren Bankkonten ...
      </p>
    </div>

    <FullProgressSpinner :show="isLoading" />
  </div>
</template>

<script setup lang="ts">
import { ModalConfig } from '~/config/dialog-props'
import BankAccountDialog from '~/components/dialogs/BankAccountDialog.vue'
import type { BankAccountResponse } from '~/models/bank-account'
import BankAccountCard from '~/components/BankAccountCard.vue'
import useBankAccounts from '~/composables/useBankAccounts'
import FullProgressSpinner from '~/components/FullProgressSpinner.vue'

useHead({
  title: 'Bankkonten',
})

const dialog = useDialog()
const { getOrganisationCurrencyCode, getOrganisationCurrencyLocaleCode } = useAuth()
const { bankAccounts, noMoreDataBankAccounts, pageBankAccounts, searchBankAccounts, totalBankSaldoInCHF, useFetchListBankAccounts, listBankAccounts } = useBankAccounts()
const { toggleBankAccountDisplayType, bankAccountDisplay } = useSettings()

const isLoading = ref(false)
const isLoadingMore = ref(false)
const bankAccountsErrorMessage = ref('')

// Computed
const totalSaldo = computed(() => {
  return NumberToFormattedCurrency(AmountToFloat(totalBankSaldoInCHF.value), getOrganisationCurrencyLocaleCode.value)
})
const getDisplayIcon = computed(() => bankAccountDisplay.value == 'list' ? 'pi pi-microsoft' : 'pi pi-list')

// Debounced search
let searchTimeout: ReturnType<typeof setTimeout> | null = null
const onSearch = (value: string) => {
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
  const trimmedValue = value.trim()
  if (trimmedValue === searchBankAccounts.value) {
    return
  }
  searchTimeout = setTimeout(() => {
    searchBankAccounts.value = trimmedValue
    pageBankAccounts.value = 1
    isLoading.value = true
    listBankAccounts(false)
      .catch((err) => {
        if (err !== 'aborted') {
          bankAccountsErrorMessage.value = err
        }
      })
      .finally(() => {
        isLoading.value = false
      })
  }, 300)
}

const onClearSearch = () => {
  if (searchBankAccounts.value === '') return
  searchBankAccounts.value = ''
  pageBankAccounts.value = 1
  isLoading.value = true
  listBankAccounts(false)
    .catch((err) => {
      if (err !== 'aborted') {
        bankAccountsErrorMessage.value = err
      }
    })
    .finally(() => {
      isLoading.value = false
    })
}

// Init
await useFetchListBankAccounts()
  .catch((reason) => {
    bankAccountsErrorMessage.value = reason
  })

// Functions
const onSort = () => {
  isLoading.value = false
  nextTick(() => {
    isLoading.value = true
    listBankAccounts(false)
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

const onLoadMoreBankAccounts = async () => {
  isLoadingMore.value = true
  pageBankAccounts.value += 1
  listBankAccounts(true)
    .catch((err) => {
      if (err !== 'aborted') {
        bankAccountsErrorMessage.value = err
      }
    })
    .finally(() => isLoadingMore.value = false)
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
