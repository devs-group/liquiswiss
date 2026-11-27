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
          size="small"
          class="w-56 max-w-56"
          empty-message="Keine Organisationen gefunden"
          @click.stop
          @change="onChangeOrganisation"
        />
        <div
          v-if="user && scenarios.length"
          class="flex flex-col gap-1 w-56 max-w-56"
        >
          <div class="flex items-center justify-between pl-1">
            <span class="text-xs font-bold text-surface-500">Szenario</span>
            <Button
              v-if="canEdit"
              v-tooltip.top="'Neues Szenario'"
              icon="pi pi-plus"
              severity="secondary"
              size="small"
              text
              rounded
              aria-label="Neues Szenario"
              data-testid="scenario-create"
              @click.stop.prevent="openCreateScenarioDialog()"
              @mousedown.stop
            />
          </div>
          <Select
            v-model="selectedScenarioID"
            :options="scenarios"
            option-label="name"
            option-value="id"
            size="small"
            class="w-full"
            :class="{ '!border-amber-500': currentScenario && !currentScenario.isBase }"
            aria-label="Szenario"
            data-testid="scenario-select"
            @click.stop
            @change="onChangeScenario"
          >
            <template #value="{ value }">
              <span class="flex items-center justify-between gap-2 w-full min-w-0">
                <span class="truncate">{{ scenarios.find(s => s.id === value)?.name ?? 'Szenario' }}</span>
                <!-- Stops the click so the pencil opens the edit dialog instead of the list -->
                <i
                  v-if="canEdit && currentScenario"
                  v-tooltip.top="'Szenario bearbeiten'"
                  class="pi pi-pencil text-xs text-surface-500 hover:text-surface-900 dark:hover:text-surface-0 p-1 -m-1 cursor-pointer"
                  role="button"
                  aria-label="Szenario bearbeiten"
                  data-testid="scenario-edit"
                  @click.stop.prevent="openEditScenarioDialog(currentScenario)"
                  @mousedown.stop
                />
              </span>
            </template>
            <template #option="{ option }">
              <span class="flex items-center gap-2 text-sm">
                <span>{{ option.name }}</span>
                <Tag
                  v-if="option.isBase"
                  value="Basis"
                  severity="secondary"
                />
              </span>
            </template>
          </Select>
        </div>
      </div>
    </template>

    <template #end>
      <div class="flex items-center justify-between p-3 mt-auto">
        <span
          v-if="appVersion"
          class="text-[10px] leading-none tracking-wide text-surface-400/70 dark:text-surface-500/70 pl-2 select-none"
          data-testid="app-version"
        >{{ appVersion }}</span>
        <span v-else />
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
const { organisations, canEdit } = useOrganisations()
const { scenarios, currentScenario, switchScenario, openCreateScenarioDialog, openEditScenarioDialog } = useScenarios()
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

// Build identifiers of the running backend (GET /api/config). Fetched during
// SSR so the version is in the markup rather than appearing a moment later.
const { data: appConfig } = await useFetch<{ sha: string, version: string, appVersion: string }>('/api/config')

// Local, unstamped builds report "dev", which is worth showing as such.
const appVersion = computed(() => appConfig.value?.version ?? '')

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

const selectedScenarioID = ref<number | null>(currentScenario.value?.id ?? null)
watch(currentScenario, (scenario) => {
  selectedScenarioID.value = scenario?.id ?? null
})

// Switching only changes the own view, every page reloads its data afterwards
const onChangeScenario = () => {
  const newID = selectedScenarioID.value
  if (newID == null || newID === currentScenario.value?.id) return
  showGlobalLoadingSpinner.value = true
  switchScenario(newID)
    .then(() => {
      reloadNuxtApp({ force: true })
    })
    .catch(() => {
      showGlobalLoadingSpinner.value = false
      selectedScenarioID.value = currentScenario.value?.id ?? null
      toast.add({
        summary: 'Fehler',
        detail: 'Das Szenario konnte nicht gewechselt werden',
        severity: 'error',
        life: Config.TOAST_LIFE_TIME,
      })
    })
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
