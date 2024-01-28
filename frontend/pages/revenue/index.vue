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
      <DataTable :value="[]" tableStyle="min-width: 50rem">
        <Column field="code" header="Code"></Column>
        <Column field="name" header="Name"></Column>
        <Column field="category" header="Category"></Column>
        <Column field="quantity" header="Quantity"></Column>
      </DataTable>
    </TabPanel>
  </TabView>
</template>

<script setup lang="ts">
import {ModalConfig} from "~/config/dialog-props";
import {Config} from "~/config/config";
import RevenueDialog from "~/components/dialogs/RevenueDialog.vue";
import useGlobalData from "~/composables/useGlobalData";
import type {StrapiRevenue} from "~/models/revenue";

const dialog = useDialog();
const toast = useToast()
const {fetchCategories, fetchCurrencies} = useGlobalData()

await fetchCategories()
await fetchCurrencies()
const {data: revenues} = await useFetch('/api/revenue')

const searchText = ref('')
const currentDay = ref(new Date())

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