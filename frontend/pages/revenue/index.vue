<template>
  <TabView>
    <TabPanel header="Planungsdaten">
      <div class="flex flex-col gap-4">
        <div class="flex gap-4">
      <span class="p-input-icon-left flex-1">
        <i class="pi pi-search" />
        <InputText class="w-full" v-model="searchText" placeholder="Suchen" />
      </span>
          <Button @click="addRevenue" class="self-end" label="Einnahme hinzufügen" icon="pi pi-angle-double-up"/>
        </div>

        <div v-if="revenues.length" class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          <RevenueCard @on-edit="onEdit" v-for="revenue in revenues" :revenue="revenue"/>
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
               v-tooltip="`${entry.revenue.attributes.amount} ${getCurrencyCodeFromId(entry.revenue.attributes.currency as number)}`"
               v-for="entry of getMonthlyEntries.filter(value => value.months.includes(i))">
              {{entry.revenue.attributes.name}}
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
import RevenueDialog from "~/components/dialogs/RevenueDialog.vue";
import useGlobalData from "~/composables/useGlobalData";
import type {StrapiRevenue} from "~/models/revenue";
import {CycleType, RevenueType} from "~/config/enums";
import {range} from "@antfu/utils";

const dialog = useDialog();
const toast = useToast()
const {fetchCategories, fetchCurrencies, getCurrencyCodeFromId} = useGlobalData()

await fetchCategories()
await fetchCurrencies()
const {data: revenues} = await useFetch('/api/revenue')

const searchText = ref('')
const months = ref([
    'Januar', 'Februar', 'März', 'April', 'Mai', 'Juni', 'Juli', 'August', 'September', 'Oktober', 'November', 'Dezember'
])

const getMonthlyEntries = computed(() => {
  return (revenues.value as StrapiRevenue[]).map(revenue => {
    const start = new Date(revenue.attributes.start)
    const end = revenue.attributes.end ? new Date(revenue.attributes.end) : null
    if (revenue.attributes.type === RevenueType.Single) {
      return {
        revenue: revenue,
        months: [start.getMonth()],
      }
    }
    return {
      revenue: revenue,
      months: revenue.attributes.cycle === CycleType.Monthly ? range(start.getMonth(), end?.getMonth() + 1 ?? 12) : []
    }
  })
})

const onEdit = (revenue: StrapiRevenue) => {
  dialog.open(RevenueDialog, {
    data: {
      revenue: revenue,
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
      const payload = opt.data as StrapiRevenue|'deleted'

      if (payload == 'deleted') {
        revenues.value = await $fetch('/api/revenue')
        toast.add({
          summary: 'Erfolg',
          detail: `Einnahme "${revenue.attributes.name}" wurde gelöscht`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
        return
      }

      $fetch('/api/revenue', {
        method: 'put',
        body: payload,
      }).then(async () => {
        revenues.value = await $fetch('/api/revenue')
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

const addRevenue = () => {
  dialog.open(RevenueDialog, {
    props: {
      header: 'Neue Einnahme anlegen',
      ...ModalConfig,
    },
    onClose: (opt) => {
      if (!opt?.data) {
        // TODO: Handle error?
        return
      }
      const payload = opt.data as StrapiRevenue
      $fetch('/api/revenue', {
        method: 'post',
        body: payload,
      }).then(async (resp) => {
        revenues.value = await $fetch('/api/revenue')
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