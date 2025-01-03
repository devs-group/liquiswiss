<template>
  <Menu
    class="sm:!rounded-none sm:!border-t-0 sm:!border-b-0"
    :model="items"
  >
    <template #start>
      <div class="flex flex-col gap-2 p-4">
        <Logo class="hidden sm:block" />
        <Select
          v-if="user"
          v-model="selectedOrganisationID"
          :options="organisations"
          option-label="name"
          option-value="id"
          class="w-56 max-w-56"
          empty-message="Keine Organisationen gefunden"
          @click.stop
          @change="onChangeOrganisation"
        />
      </div>
    </template>

    <template #item="{ item, props }">
      <router-link
        v-if="item.routeName"
        v-slot="{ href, navigate, isActive }"
        :to="{ name: item.routeName }"
        custom
      >
        <a
          v-ripple
          :href="href"
          v-bind="props.action"
          @click="navigate"
        >
          <span :class="item.icon" />
          <span
            class="ml-2"
            :class="{ 'text-liqui-green': isActive }"
          >{{ item.label }}</span>
        </a>
      </router-link>
      <a
        v-else
        v-ripple
        :href="item.url"
        :target="item.target"
        v-bind="props.action"
      >
        <span :class="item.icon" />
        <span class="ml-2">{{ item.label }}</span>
      </a>
    </template>
  </Menu>
</template>

<script setup lang="ts">
import type { MenuItem } from 'primevue/menuitem'
import type { SelectChangeEvent } from 'primevue'
import { RouteNames } from '~/config/routes'
import useAuth from '~/composables/useAuth'
import { Config } from '~/config/config'

const { logout, user, updateCurrentOrganisation } = useAuth()
const { organisations } = useOrganisations()
const { showGlobalLoadingSpinner } = useGlobalData()
const { skipOrganisationSwitchQuestion } = useSettings()
const confirm = useConfirm()
const toast = useToast()

const selectedOrganisationID = ref<number | null>(user.value?.currentOrganisationID ?? null)

const items = ref<MenuItem[]>([
  { label: 'Prognose', icon: 'pi pi-chart-line', routeName: RouteNames.HOME },
  { label: 'Mitarbeitende', icon: 'pi pi-users', routeName: RouteNames.EMPLOYEES },
  { label: 'Transaktionen', icon: 'pi pi-money-bill', routeName: RouteNames.TRANSACTIONS },
  { label: 'Bankkonten', icon: 'pi pi-building', routeName: RouteNames.BANK_ACCOUNTS },
  { label: 'Einstellungen', icon: 'pi pi-cog', routeName: RouteNames.SETTINGS },
  { label: 'Abmelden', icon: 'pi pi-sign-out', command: async () => {
    confirm.require({
      header: 'Abmelden',
      message: 'Möchten Sie sich wirklich abmelden?',
      icon: 'pi pi-exclamation-triangle',
      rejectLabel: 'Nein',
      acceptLabel: 'Ja',
      accept: async () => {
        await logout()
        reloadNuxtApp({ force: true }) // , path: RoutePaths.AUTH
      },
      reject: () => {
      },
    })
  } },
])

const onChangeOrganisation = (event: SelectChangeEvent) => {
  const currentSelectedOrganisationID = user.value?.currentOrganisationID ?? null
  const newSelectedOrganisationID = selectedOrganisationID.value
  if (newSelectedOrganisationID === currentSelectedOrganisationID || newSelectedOrganisationID == null) {
    // Selection hasn't changed
    return
  }
  const newOrganisation = organisations.value.find(o => o.id === event.value)
  if (skipOrganisationSwitchQuestion.value) {
    updateOrganisation(newSelectedOrganisationID)
  }
  else {
    confirm.require({
      header: 'Organisation wechseln',
      message: `Möchten Sie die Organisation auf "${newOrganisation!.name}" wechseln?`,
      icon: 'pi pi-question-circle',
      rejectLabel: 'Nein',
      acceptLabel: 'Ja',
      accept: () => updateOrganisation(newSelectedOrganisationID),
      reject: () => {
        selectedOrganisationID.value = currentSelectedOrganisationID
      },
    })
  }
}

const updateOrganisation = (newSelectedOrganisationID: number) => {
  showGlobalLoadingSpinner.value = true
  updateCurrentOrganisation({ organisationId: newSelectedOrganisationID })
    .then(() => {
      reloadNuxtApp({ force: true })
    })
    .catch(() => {
      showGlobalLoadingSpinner.value = false
      toast.add({
        summary: 'Fehler',
        detail: `Die Organisation konnte nicht geändert werden. Dies ist ein Systemfehler`,
        severity: 'error',
        life: Config.TOAST_LIFE_TIME,
      })
    })
}
</script>
