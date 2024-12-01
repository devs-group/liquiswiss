<template>
  <div class="flex min-h-screen">
    <DesktopMenu v-if="isAuthenticated"/>
    <div class="flex flex-col gap-4 p-4 flex-1 overflow-hidden">
      <MobileMenu v-if="isAuthenticated"/>
      <div class="w-full h-full">
        <NuxtPage/>
      </div>
      <div v-if="serverDateFormatted" class="p-2 bg-zinc-100 dark:bg-zinc-800 self-end">
        <p class="text-xs text-right">Serverdatum: {{ serverDateFormatted }}</p>
      </div>
    </div>
  </div>
  <DynamicDialog/>
  <ConfirmDialog :draggable="false"/>
  <Toast position="bottom-center"/>
  <NuxtLoadingIndicator :height="4" :throttle="1000" color="#10B981" />
</template>

<script setup lang="ts">
import useAuth from "~/composables/useAuth";
import {Config} from "~/config/config";
import {Constants} from "~/utils/constants";
import {DateStringToFormattedDate} from "~/utils/format-helper";

const {isAuthenticated, getAccessToken} = useAuth()
const {useFetchListCurrencies, useFetchListCategories, useFetchListFiatRates, useFetchGetServerTime, serverDate} = useGlobalData()
const toast = useToast()
const hasInitialLoadError = ref(false)

useHead({
  title: 'LIQUISWISS',
  bodyAttrs: () => ({class: 'bg-white dark:bg-zinc-900'})
})

const serverDateFormatted = computed(() => {
  return serverDate.value ? DateStringToFormattedDate(serverDate.value) : ''
})

if (isAuthenticated.value) {
  await Promise.all([useFetchListCurrencies(), useFetchListCategories(), useFetchListFiatRates(), useFetchGetServerTime()])
      .catch(reason => {
        hasInitialLoadError.value = true
      })
}

// This is to ensure users gets an access token if it expires
onMounted(() => {
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
  } else {
    if (localStorage.getItem(Constants.SESSION_EXPIRED_NAME) === 'true') {
      localStorage.removeItem(Constants.SESSION_EXPIRED_NAME)
      toast.add({
        summary: 'Info',
        detail: `Deine Session ist aus Sicherheitsgründen abgelaufen. Bitte logge dich erneut ein.`,
        severity: 'info',
        life: Config.TOAST_LIFE_TIME,
      })
    }
  }
})
</script>
