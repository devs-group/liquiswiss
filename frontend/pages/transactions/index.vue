<template>
  <div class="flex flex-col gap-4 relative">
    <div class="flex flex-col sm:flex-row gap-2 justify-between items-center">
      <div class="flex items-center gap-2 w-full sm:w-auto">
        <InputText v-model="search" placeholder="Suchen"/>
        <Button @click="toggleTransactionDisplayType" :icon="getDisplayIcon"/>
      </div>
      <Button @click="onCreateTransaction" class="self-end" label="Transaktion hinzufügen" icon="pi pi-money-bill"/>
    </div>

    <Message v-if="transactionsErrorMessage.length" severity="error" :closable="false" class="col-span-full">{{transactionsErrorMessage}}</Message>
    <template v-else-if="filterTransactions.length">
      <div v-if="transactionDisplay == 'grid'" class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        <TransactionCard @on-edit="onEditTransaction" @on-clone="onCloneTransaction" v-for="transaction in filterTransactions" :transaction="transaction"/>
      </div>
      <div v-else class="flex flex-col overflow-x-auto pb-2">
        <TransactionHeaders @on-sort="onSort"/>
        <TransactionRow @on-edit="onEditTransaction" @on-clone="onCloneTransaction" v-for="transaction in filterTransactions" :transaction="transaction"/>
      </div>
    </template>
    <p v-else>Es gibt noch keine Transaktionen</p>

    <div v-if="transactions?.data.length" class="self-center">
      <Button v-if="!noMoreDataTransactions" severity="info" label="Mehr anzeigen" @click="onLoadMoreTransactions" :loading="isLoadingMore"/>
      <p v-else class="text-xs opacity-60">Keine weiteren Transaktionen ...</p>
    </div>

    <FullProgressSpinner :show="isLoading"/>
  </div>
</template>

<script setup lang="ts">
import {ModalConfig} from "~/config/dialog-props";
import TransactionDialog from "~/components/dialogs/TransactionDialog.vue";
import type {TransactionResponse} from "~/models/transaction";
import TransactionCard from "~/components/TransactionCard.vue";
import TransactionRow from "~/components/TransactionRow.vue";
import FullProgressSpinner from "~/components/FullProgressSpinner.vue";
import TransactionHeaders from "~/components/TransactionHeaders.vue";

const dialog = useDialog();
const {transactions, noMoreDataTransactions, pageTransactions, useFetchListTransactions, listTransactions} = useTransactions()
const {toggleTransactionDisplayType, transactionDisplay} = useSettings()

const isLoading = ref(false)
const isLoadingMore = ref(false)
const transactionsErrorMessage = ref('')
const search = ref('')

// Computed
const getDisplayIcon = computed(() => transactionDisplay.value == 'list' ? 'pi pi-microsoft' : 'pi pi-list')
const filterTransactions = computed(() => {
  return transactions.value.data
      .filter(t => t.name.toLowerCase().includes(search.value.toLowerCase()))
})

// Init
await useFetchListTransactions()
    .catch(reason => {
      transactionsErrorMessage.value = reason
    })

// Functions
const onSort = () => {
  isLoading.value = false
  nextTick(() => {
    isLoading.value = true
    listTransactions(false)
        .then(() => {
          isLoading.value = false
        })
        .catch((err) => {
          if (err !== 'aborted') {
            isLoading.value = false
            transactionsErrorMessage.value = err
          }
        })
  })
}

const onCreateTransaction = () => {
  dialog.open(TransactionDialog, {
    props: {
      header: 'Neue Transaktion anlegen',
      ...ModalConfig,
    },
  })
}

const onEditTransaction = (transaction: TransactionResponse) => {
  dialog.open(TransactionDialog, {
    data: {
      transaction: transaction,
    },
    props: {
      header: 'Transaktion bearbeiten',
      ...ModalConfig,
    },
  })
}

const onCloneTransaction = (transaction: TransactionResponse) => {
  dialog.open(TransactionDialog, {
    data: {
      transaction: transaction,
      isClone: true,
    },
    props: {
      header: 'Transaktion klonen',
      ...ModalConfig,
    },
  })
}

const onLoadMoreTransactions = async (event: MouseEvent) => {
  isLoadingMore.value = true
  pageTransactions.value += 1
  listTransactions(false)
      .catch(reason => {
        transactionsErrorMessage.value = reason
      })
      .finally(() => isLoadingMore.value = false)
}
</script>