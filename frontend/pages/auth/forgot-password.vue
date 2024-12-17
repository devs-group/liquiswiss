<template>
  <div class="flex flex-col gap-4">
    <Logo class="!text-3xl"/>

    <h1 class="text-2xl font-bold text-center">Passwort vergessen</h1>

    <div class="flex flex-col items-center gap-2 w-full max-w-lg mx-auto">
      <template v-if="!isFinished">
        <Message severity="secondary">
          Sie haben Ihr Passwort vergessen? Kein Problem! Geben Sie die E-Mail Adresse ein mit der Sie sich
          registriert hatten und wir senden Ihnen einen Link zum Zurücksetzen des Passworts zu.
        </Message>

        <form @submit.prevent class="grid grid-cols-1 gap-2 w-full">
          <div class="flex flex-col gap-2 col-span-full">
            <label class="text-sm font-bold" for="email">E-Mail*</label>
            <InputText v-model="email" v-bind="emailProps"
                       :class="{'p-invalid': errors['email']?.length}"
                       id="email" type="email" autocomplete="email"/>
            <small class="text-liqui-red">{{ errors["email"] }}</small>
          </div>

          <p class="w-full text-sm text-right">
            <NuxtLink :to="{name: RouteNames.AUTH_LOGIN}" class="underline">Zurück zum Login</NuxtLink>
          </p>

          <div class="flex justify-end gap-2 col-span-full">
            <Button @click="onForgotPassword" :disabled="!meta.valid || (meta.valid && !meta.dirty) || isSubmitting" :loading="isSubmitting" label="Passwort zurücksetzen" icon="pi pi-lock" type="submit"/>
          </div>
        </form>
      </template>
      <div v-else class="flex flex-col gap-2">
        <Message severity="success">
          Vielen Dank. Bitte prüfen Sie nun Ihre Mailbox für den Link zum Zurücksetzen des Passworts.
          Es kann vorkommen, dass E-Mails auch im "Spam" Ordner landen. Schauen sie ggf. auch dort nach.
          Sollte diese E-Mail nicht in unserem System existieren, weisen wir Sie aus Sicherheitsgründen nicht darauf hin.
        </Message>

        <p class="w-full text-sm text-right">
          <NuxtLink :to="{name: RouteNames.AUTH_LOGIN}" class="underline">Zurück zum Login</NuxtLink>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import useAuth from "~/composables/useAuth";
import {Config} from "~/config/config";
import {useForm} from "vee-validate";
import * as yup from "yup";
import type {ForgotPasswordFormData} from "~/models/auth";
import {RouteNames} from "~/config/routes";

useHead({
  title: 'Passwort vergessen',
  meta: [
    {name: 'robots', content: 'noindex, nofollow'}
  ]
})

const {forgotPassword} = useAuth()
const toast = useToast()

const isFinished = ref(false)
const isSubmitting = ref(false)

const { defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    email: yup.string().email('Ungültiges E-Mail Format').trim().required('E-Mail wird benötigt'),
  }),
  initialValues: {
    email: '',
  } as ForgotPasswordFormData
});

const [email, emailProps] = defineField('email')

const onForgotPassword = handleSubmit(async (values) => {
  isSubmitting.value = true
  forgotPassword(values)
      .then(() => {
        isFinished.value = true
      })
      .catch(() => {
        toast.add({
          summary: 'Fehler',
          detail: `Das Zurücksetzen des Passworts ist fehlgeschlagen`,
          severity: 'error',
          life: Config.TOAST_LIFE_TIME,
        })
      })
      .finally(() => {
        isSubmitting.value = false
      })
})
</script>