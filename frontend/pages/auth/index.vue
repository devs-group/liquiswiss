<template>
  <div class="flex flex-col gap-4">
    <Logo class="!text-3xl"/>

    <h1 class="text-2xl font-bold text-center">Login</h1>

    <div class="flex flex-col items-center gap-2 w-full max-w-lg mx-auto">
      <Message severity="secondary">
        Willkommen zurück! ❤️
      </Message>

      <form @submit.prevent class="grid grid-cols-1 gap-2 w-full">
        <div class="flex flex-col gap-2 col-span-full">
          <label class="text-sm font-bold" for="email">E-Mail*</label>
          <InputText v-model="email" v-bind="emailProps"
                     :class="{'p-invalid': errors['email']?.length}"
                     id="email" type="email" autocomplete="email"/>
          <small class="text-liqui-red">{{ errors["email"] }}</small>
        </div>

        <div class="flex flex-col gap-2 col-span-full">
          <label class="text-sm font-bold" for="password">Passwort*</label>
          <InputText v-model="password" v-bind="passwordProps"
                     :class="{'p-invalid': errors['password']?.length}"
                     id="password" type="password" autocomplete="current-password"/>
          <small class="text-liqui-red">{{ errors["password"] }}</small>
        </div>

        <p class="w-full text-sm text-left">
          <NuxtLink :to="{name: RouteNames.FORGOT_PASSWORD}" class="underline">Passwort vergessen?</NuxtLink>
        </p>

        <p class="w-full text-sm text-right">
          Kein Konto? Jetzt <NuxtLink :to="{name: RouteNames.REGISTRATION}" class="underline">registrieren</NuxtLink>
        </p>

        <div class="flex justify-end gap-2 col-span-full">
          <Button @click="onLogin" :disabled="!meta.valid || (meta.valid && !meta.dirty) || isLoading" :loading="isLoading" label="Login" icon="pi pi-sign-in" type="submit"/>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import useAuth from "~/composables/useAuth";
import {Config} from "~/config/config";
import {RouteNames} from "~/config/routes";
import {useForm} from "vee-validate";
import * as yup from "yup";
import type {LoginFormData} from "~/models/auth";

useHead({
  title: 'Login',
})

const {login} = useAuth()
const toast = useToast()

const isLoading = ref(false)

const { defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    email: yup.string().email('Ungültiges E-Mail Format').trim().required('E-Mail wird benötigt'),
    password: yup.string().required('Password wird benötigt'),
  }),
  initialValues: {
    email: '',
    password: '',
  } as LoginFormData
});

const [email, emailProps] = defineField('email')
const [password, passwordProps] = defineField('password')

const onLogin = handleSubmit(async (values) => {
  isLoading.value = true
  login(values)
      .then(() => {
        reloadNuxtApp({force: true})
      })
      .catch(() => {
        isLoading.value = false
        toast.add({
          summary: 'Fehler',
          detail: `Login fehlgeschlagen, bitte prüfe deine Zugangsdaten`,
          severity: 'error',
          life: Config.TOAST_LIFE_TIME,
        })
      })

})
</script>