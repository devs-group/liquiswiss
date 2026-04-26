<template>
  <Menu
    class="nav-menu"
    :model="items"
    :pt="{
      root: { class: '!flex !flex-col !h-full !bg-transparent !border-0 !rounded-none !shadow-none' },
      list: { class: '!flex-1' },
    }"
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

    <template #end>
      <div class="flex justify-end p-3 mt-auto">
        <Button
          v-tooltip.top="`Farbmodus: ${currentDarkModeOption.title} (klicken zum Wechseln)`"
          :icon="currentDarkModeOption.icon"
          :aria-label="`Farbmodus: ${currentDarkModeOption.title}`"
          severity="secondary"
          text
          rounded
          data-testid="color-mode-toggle"
          @click.stop.prevent="cycleColorMode"
          @mousedown.stop
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
            class="ml-2 flex-1"
            :class="{ 'text-liqui-green': isActive }"
          >{{ item.label }}</span>
          <Badge
            v-if="item.badge"
            :value="item.badge"
            severity="warn"
          />
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
        <span class="ml-2 flex-1">{{ item.label }}</span>
        <Badge
          v-if="item.badge"
          :value="item.badge"
          severity="warn"
        />
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
const { myPendingInvitations } = useInvitations()
const confirm = useConfirm()
const toast = useToast()
const colorMode = useColorMode()

const darkModeMeta: Record<DarkModeType, { icon: string, title: string }> = {
  system: { icon: 'pi pi-desktop', title: 'System' },
  dark: { icon: 'pi pi-moon', title: 'Dunkel' },
  light: { icon: 'pi pi-sun', title: 'Hell' },
}

const currentDarkModeOption = computed(() => {
  const pref = (colorMode.preference as DarkModeType) ?? 'system'
  return darkModeMeta[pref] ?? darkModeMeta.system
})

const cycleColorMode = () => {
  const order: DarkModeType[] = ['system', 'dark', 'light']
  const current = (colorMode.preference as DarkModeType) ?? 'system'
  const idx = order.indexOf(current)
  colorMode.preference = order[(idx + 1) % order.length]!
}

const selectedOrganisationID = ref<number | null>(user.value?.currentOrganisationID ?? null)

const pendingInvitationCount = computed(() => myPendingInvitations.value.length)

const items = computed<MenuItem[]>(() => [
  { label: 'Prognose', icon: 'pi pi-chart-line', routeName: RouteNames.HOME },
  { label: 'Mitarbeitende', icon: 'pi pi-users', routeName: RouteNames.EMPLOYEES },
  { label: 'Transaktionen', icon: 'pi pi-money-bill', routeName: RouteNames.TRANSACTIONS },
  { label: 'Bankkonten', icon: 'pi pi-building', routeName: RouteNames.BANK_ACCOUNTS },
  { label: 'Organisation', icon: 'pi pi-sitemap', routeName: RouteNames.ORGANISATION },
  {
    label: 'Einstellungen',
    icon: 'pi pi-cog',
    routeName: RouteNames.SETTINGS,
    badge: pendingInvitationCount.value > 0 ? pendingInvitationCount.value.toString() : undefined,
  },
  { label: 'Abmelden', icon: 'pi pi-sign-out', command: async () => {
    confirm.require({
      header: 'Abmelden',
      message: 'Möchten Sie sich wirklich abmelden?',
      icon: 'pi pi-exclamation-triangle',
      rejectLabel: 'Nein',
      acceptLabel: 'Ja',
      accept: async () => {
        await logout()
        reloadNuxtApp({ force: true })
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
