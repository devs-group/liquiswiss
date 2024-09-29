<template>
  <div class="flex flex-col gap-4">
    <Button @click="onCreateTransaction" class="self-end" label="Transaktion hinzufügen" icon="pi pi-money-bill"/>

    <div v-if="transactions?.data.length" class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <TransactionCard @on-edit="onEditTransaction" v-for="transaction in transactions.data" :transaction="transaction"/>
    </div>
    <p v-else>Es gibt noch keine Transaktionen</p>

    <div v-if="transactions?.data.length" class="self-center">
      <Button v-if="!noMoreDataTransactions" severity="info" label="Mehr anzeigen" @click="onLoadMoreEmployees" :loading="isLoadingMore"/>
      <p v-else class="text-xs opacity-60">Keine weiteren Mitarbeiter ...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ModalConfig} from "~/config/dialog-props";
import {Config} from "~/config/config";
import TransactionDialog from "~/components/dialogs/TransactionDialog.vue";
import useGlobalData from "~/composables/useGlobalData";
import type {StrapiTransaction} from "~/models/transaction";
import {CycleType, TransactionType} from "~/config/enums";
import {range} from "@antfu/utils";
import TransactionCard from "~/components/TransactionCard.vue";

const dialog = useDialog();
const toast = useToast()
const {transactions, noMoreDataTransactions, pageTransactions, listTransactions} = useTransactions()

const isLoadingMore = ref(false)

await listTransactions(false)

const onCreateTransaction = () => {
  dialog.open(TransactionDialog, {
    props: {
      header: 'Neue Transaktion anlegen',
      ...ModalConfig,
    },
  })
}

const onEditTransaction = (transaction: StrapiTransaction) => {
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

const onLoadMoreEmployees = async (event: MouseEvent) => {
  isLoadingMore.value = true
  pageTransactions.value += 1
  await listTransactions(false)
  isLoadingMore.value = false
}
</script>