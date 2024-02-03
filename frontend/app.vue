<template>
  <div class="flex flex-col gap-4 p-4">
    <MainMenu v-if="user"/>
    <p v-if="user">Willkommen {{user.email}}</p>
    <div class="w-full h-full">
      <NuxtPage/>
    </div>
  </div>
  <DynamicDialog/>
  <ConfirmPopup/>
  <Toast position="bottom-right"/>
</template>

<script setup lang="ts">
import useAuth from "~/composables/useAuth";
import {Config} from "~/config/config";
import {Constants} from "~/utils/constants";

const {user, getAccessToken, getProfile} = useAuth()
const route = useRoute()
const toast = useToast()

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
}

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

useHead({
  title: 'LiquiSwiss'
})
</script>
