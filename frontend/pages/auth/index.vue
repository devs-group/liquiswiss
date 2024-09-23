<template>
  <form @submit.prevent class="flex flex-col items-center justify-center gap-2">
    <InputText placeholder="E-Mail" v-model.trim="email" :disabled="isLoading"/>
    <InputText placeholder="Passwort" v-model="password" :disabled="isLoading"/>
    <p class="text-sm">Kein Konto? Jetzt <NuxtLink :to="Routes.REGISTER" class="underline">registrieren</NuxtLink></p>
    <Button label="Login" @click="onLogin" :loading="isLoading" type="submit"/>
  </form>
</template>

<script setup lang="ts">
import useAuth from "~/composables/useAuth";
import {Config} from "~/config/config";
import {Routes} from "~/config/routes";

const {login} = useAuth()
const toast = useToast()

const email = ref('')
const password = ref('')
const isLoading = ref(false)

const onLogin = async () => {
  isLoading.value = true
  const isLoggedIn = await login(email.value, password.value)
  if (isLoggedIn) {
    reloadNuxtApp({force: true})
  } else {
    isLoading.value = false
    toast.add({
      summary: 'Fehler',
      detail: `Login fehlgeschlagen, bitte prüfe deine Zugangsdaten`,
      severity: 'error',
      life: Config.TOAST_LIFE_TIME,
    })
  }
}
</script>