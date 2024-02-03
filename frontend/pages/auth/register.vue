<template>
  <h1>Register now</h1>
  <form @submit.prevent class="flex flex-col items-center justify-center gap-2">
    <InputText placeholder="E-Mail" v-model.trim="email" :disabled="isLoading"/>
    <InputText placeholder="Passwort" v-model="password" :disabled="isLoading"/>
    <p class="text-sm">Konto vorhanden? Jetzt <NuxtLink :to="Routes.LOGIN" class="underline">einloggen</NuxtLink></p>
    <Button label="Konto erstellen" @click="onRegister" :loading="isLoading" type="submit"/>
  </form>
</template>

<script setup lang="ts">
import useAuth from "~/composables/useAuth";
import {Config} from "~/config/config";
import {Routes} from "~/config/routes";

const {register} = useAuth()
const toast = useToast()

const email = ref('segiralph1@gmail.com')
const password = ref('qwer1234')
const isLoading = ref(false)

const onRegister = async () => {
  isLoading.value = true
  const isRegistered = await register(email.value, password.value)
  if (isRegistered) {
    reloadNuxtApp({force: true})
  } else {
    isLoading.value = false
    toast.add({
      summary: 'Fehler',
      detail: `Die Registrierung ist fehlgeschlagen`,
      severity: 'error',
      life: Config.TOAST_LIFE_TIME,
    })
  }
}
</script>