<template>
  <div class="flex flex-col gap-4 p-4">
    <MainMenu v-if="user"/>
    <div class="w-full h-full">
      <NuxtPage/>
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

const {user, getAccessToken, getProfile} = useAuth()
const {fetchCurrencies, fetchCategories} = useGlobalData()
const route = useRoute()
const toast = useToast()

// This is to ensure users gets an access token if it expires
onMounted(() => {
  if (!!user.value) {
    getAccessToken()
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

// Initial check if user is authenticated or not
await getProfile(true)

if (!user.value) {
  if (!route.path.includes('/auth')) {
    await navigateTo('/auth', {replace: true})
  }
} else {
  if (route.path.includes('/auth')) {
    await navigateTo('/', {replace: true})
  }
  await Promise.all([fetchCurrencies(), fetchCategories()])
}

useHead({
  title: 'LiquiSwiss'
})
</script>
