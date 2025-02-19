<template>
  <div class="flex min-h-screen">
    <DesktopMenu v-if="isAuthenticated" />
    <div class="flex flex-col gap-4 p-4 flex-1 overflow-hidden">
      <MobileMenu v-if="isAuthenticated" />
      <div class="w-full h-full">
        <NuxtPage />
      </div>
      <div
        v-if="serverDateFormatted"
        class="p-2 bg-zinc-100 dark:bg-zinc-800 self-end"
      >
        <p class="text-xs text-right">
          Serverdatum: {{ serverDateFormatted }}
        </p>
      </div>
    </div>
  </div>
  <div
    :class="{ '!translate-y-0 opacity-100 pointer-events-auto': updateAvailable }"
    class="flex items-center justify-center gap-2 bg-zinc-800 dark:bg-zinc-600 text-white
       !bg-opacity-80 rounded-xl backdrop-blur-sm fixed bottom-2 right-2 left-2 sm:left-auto p-4
       pointer-events-none transform translate-y-full opacity-0 transition-all duration-300"
  >
    <p class="text-sm flex-1 cursor-default">
      Es gibt eine neue Version dieser Webseite
    </p>
    <Button
      label="Neu laden"
      severity="help"
      size="small"
      @click="reloadNuxtApp({ force: true })"
    />
  </div>
  <DynamicDialog />
  <ConfirmDialog
    :draggable="false"
    :breakpoints="confirmBreakpoints"
  />
  <Toast position="bottom-center" />
  <NuxtLoadingIndicator
    :height="4"
    :throttle="1000"
    color="#10B981"
  />
  <FullProgressSpinner :show="showGlobalLoadingSpinner" />
</template>

<script setup lang="ts">
import type { ConfirmDialogBreakpoints } from 'primevue'
import useAuth from '~/composables/useAuth'
import { Config } from '~/config/config'

const { isAuthenticated, getAccessToken, getOrganisationCurrencyCode } = useAuth()
const { useFetchListCurrencies, useFetchListCategories, useFetchListFiatRates, useFetchGetServerTime, serverDate, showGlobalLoadingSpinner } = useGlobalData()
const { useFetchListOrganisations } = useOrganisations()
const toast = useToast()
const { hook } = useNuxtApp()
const hasInitialLoadError = ref(false)
const updateAvailable = ref(false)

useHead({
  titleTemplate: title => title ? `${title} - LiquiSwiss` : 'LiquiSwiss',
  bodyAttrs: () => ({ class: 'bg-white dark:bg-zinc-900' }),
})

const confirmBreakpoints = { '639px': '90vw' } as ConfirmDialogBreakpoints

const serverDateFormatted = computed(() => {
  return serverDate.value ? DateStringToFormattedDate(serverDate.value) : ''
})

if (isAuthenticated.value) {
  await Promise.all([useFetchListCurrencies(), useFetchListCategories(), useFetchListFiatRates(getOrganisationCurrencyCode.value), useFetchGetServerTime()])
    .catch(() => {
      hasInitialLoadError.value = true
    })
  await useFetchListOrganisations()
    .catch(() => {
      toast.add({
        summary: 'Fehler',
        detail: `Wir konnten Ihre Organisationen nicht laden. Dies scheint ein Systemfehler zu sein`,
        severity: 'error',
        life: Config.TOAST_LIFE_TIME,
      })
    })
}

// This is to ensure users gets an access token if it expires
onMounted(() => {
  hook('app:manifest:update', () => {
    updateAvailable.value = true
  })

  if (isAuthenticated.value) {
    getAccessToken()
    if (hasInitialLoadError.value) {
      toast.add({
        summary: 'Fehler',
        detail: `Es scheint aktuell technische Probleme zu geben.`,
        severity: 'warn',
        life: Config.TOAST_LIFE_TIME,
      })
    }
  }
  else {
    if (localStorage.getItem(Constants.SESSION_EXPIRED_NAME) === 'true') {
      localStorage.removeItem(Constants.SESSION_EXPIRED_NAME)
      toast.add({
        summary: 'Info',
        detail: `Ihre Session ist aus Sicherheitsgründen abgelaufen. Bitte loggen Sie sich erneut ein.`,
        severity: 'info',
        life: Config.TOAST_LIFE_TIME,
      })
    }
  }
})
</script>
