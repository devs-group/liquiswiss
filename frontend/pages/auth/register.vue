<template>
  <div class="flex flex-col gap-2">
    <Logo/>

    <form @submit.prevent class="flex flex-col items-center justify-center gap-2">
      <InputText placeholder="E-Mail" type="email" name="email" autocomplete="email" v-model.trim="email" :disabled="isLoading"/>
      <InputText placeholder="Passwort" type="password" name="password" autocomplete="new-password" v-model="password" :disabled="isLoading"/>
      <p class="text-sm">Konto vorhanden? Jetzt <NuxtLink :to="{name: RouteNames.LOGIN}" class="underline">einloggen</NuxtLink></p>
      <Button label="Konto erstellen" @click="onRegister" :disabled="isLoading" :loading="isLoading" type="submit"/>
    </form>
  </div>
</template>

<script setup lang="ts">
import useAuth from "~/composables/useAuth";
import {Config} from "~/config/config";
import {RouteNames} from "~/config/routes";

const {register} = useAuth()
const toast = useToast()

const email = ref('')
const password = ref('')
const isLoading = ref(false)

const onRegister = async () => {
  isLoading.value = true
  const isRegistered = await register(email.value, password.value)
  if (isRegistered) {
    reloadNuxtApp({force: true}) // , path: RoutePaths.HOME
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