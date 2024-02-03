<template>
  <TabView>
    <TabPanel header="Planungsdaten">
      <div class="flex flex-col gap-4">
        <div class="flex gap-4">
      <span class="p-input-icon-left flex-1">
        <i class="pi pi-search" />
        <InputText class="w-full" v-model="searchText" placeholder="Suchen" />
      </span>
          <Button @click="addTransaction" class="self-end" label="Einnahme hinzufügen" icon="pi pi-angle-double-up"/>
        </div>

        <div v-if="transactions?.length" class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          <TransactionCard @on-edit="onEdit" v-for="transaction in transactions" :transaction="transaction"/>
        </div>
        <p v-else>Es gibt noch keine Einnahmen</p>
      </div>
    </TabPanel>
    <TabPanel header="Kalender">
      <div class="flex flex-col gap-4 overflow-x-auto pb-4">
        <div class="flex">
          <div class="flex flex-col flex-1 min-w-28 border-l border-t border-b last:border-r text-center" v-for="(month, i) of months">
            <p class="p-2 font-bold bg-gray-100">{{month.substring(0, 3)}}</p>
            <p class="border-t border-1 p-2 text-xs hover:bg-green-300 cursor-pointer"
               v-tooltip="`${entry.transaction.attributes.amount} ${getCurrencyCodeFromId(entry.transaction.attributes.currency as number)}`"
               v-for="entry of getMonthlyEntries.filter(value => value.months.includes(i))">
              {{entry.transaction.attributes.name}}
            </p>
          </div>
        </div>
      </div>
    </TabPanel>
  </TabView>
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
const {fetchCategories, fetchCurrencies, getCurrencyCodeFromId} = useGlobalData()

await fetchCategories()
await fetchCurrencies()
const {data: transactions} = await useFetch('/api/transactions')

const searchText = ref('')
const months = ref([
    'Januar', 'Februar', 'März', 'April', 'Mai', 'Juni', 'Juli', 'August', 'September', 'Oktober', 'November', 'Dezember'
])

const getMonthlyEntries = computed(() => {
  return transactions?.value ? (transactions.value).map((transaction: StrapiTransaction) => {
    const start = new Date(transaction.attributes.start)
    const end = transaction.attributes.end ? new Date(transaction.attributes.end) : null
    if (transaction.attributes.type === TransactionType.Single) {
      return {
        transaction: transaction,
        months: [start.getMonth()],
      }
    }
    return {
      transaction: transaction,
      months: transaction.attributes.cycle === CycleType.Monthly ? range(start.getMonth(), end?.getMonth() + 1 ?? 12) : []
    }
  }) : []
})

const onEdit = (transaction: StrapiTransaction) => {
  dialog.open(TransactionDialog, {
    data: {
      transaction: transaction,
    },
    props: {
      header: 'Einnahme bearbeiten',
      ...ModalConfig,
    },
    onClose: async (opt) => {
      if (!opt?.data) {
        // TODO: Handle error?
        return
      }
      const payload = opt.data as StrapiTransaction|'deleted'

      if (payload == 'deleted') {
        transactions.value = await $fetch('/api/transaction')
        toast.add({
          summary: 'Erfolg',
          detail: `Einnahme "${transaction.attributes.name}" wurde gelöscht`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
        return
      }

      $fetch('/api/transaction', {
        method: 'put',
        body: payload,
      }).then(async () => {
        transactions.value = await $fetch('/api/transaction')
        toast.add({
          summary: 'Erfolg',
          detail: `Einnahme wurde aktualisiert`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
    }
  })
}

const addTransaction = () => {
  dialog.open(TransactionDialog, {
    props: {
      header: 'Neue Einnahme anlegen',
      ...ModalConfig,
    },
    onClose: (opt) => {
      if (!opt?.data) {
        // TODO: Handle error?
        return
      }
      const payload = opt.data as StrapiTransaction
      $fetch('/api/transaction', {
        method: 'post',
        body: payload,
      }).then(async (resp) => {
        transactions.value = await $fetch('/api/transaction')
        toast.add({
          summary: 'Erfolg',
          detail: `Einnahme "${resp.data.attributes.name}" wurde angelegt`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
    }
  })
}
</script>