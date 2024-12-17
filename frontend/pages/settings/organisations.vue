<template>
  <div class="flex flex-col gap-4">
    <div class="flex justify-between items-center gap-2">
      <hr class="h-0.5 bg-black flex-1"/>
      <p class="text-xl">Ihre Organisationen</p>
      <hr class="h-0.5 bg-black flex-1"/>
    </div>

    <Button @click="onCreateOrganisation" label="Organisation hinzufügen" class="self-end" icon="pi pi-building"/>

    <div class="flex flex-col sm:flex-row gap-2">
      <Menu class="sm:!rounded-none sm:!border-t-0 sm:!border-b-0" :model="items">
        <template #start>
          <p class="p-4 pb-2 text-sm opacity-40 cursor-default">Organisation wählen</p>
        </template>

        <template #item="{ item, props }">
          <router-link v-if="item.routeName" v-slot="{ href, navigate, isActive }" :to="{name: item.routeName, params: item.params}" custom>
            <a v-ripple :href="href" v-bind="props.action" @click="navigate">
              <span class="truncate w-48 max-w-48" :class="{'text-liqui-green': isActive}">{{ item.label }}</span>
            </a>
          </router-link>
          <a v-else v-ripple :href="item.url" :target="item.target" v-bind="props.action">
            <span :class="item.icon" />
            <span class="ml-2">{{ item.label }}</span>
          </a>
        </template>
      </Menu>

      <NuxtPage/>
    </div>
  </div>
</template>

<script setup lang="ts">
import type {MenuItem} from "primevue/menuitem";
import {RouteNames} from "~/config/routes";
import {ModalConfig} from "~/config/dialog-props";
import OrganisationDialog from "~/components/dialogs/OrganisationDialog.vue";

useHead({
  title: 'Organisationen',
})

const dialog = useDialog()
const {organisations} = useOrganisations()

const items = computed<MenuItem[]>(() => organisations.value.map(o => {
  return {
    label: o.name,
    routeName: RouteNames.SETTINGS_ORGANISATION_EDIT,
    params: {id: o.id},
  }
}))

const onCreateOrganisation = () => {
  dialog.open(OrganisationDialog, {
    props: {
      header: 'Neue Organisation anlegen',
      ...ModalConfig,
    },
  })
}
</script>
