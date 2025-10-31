<template>
  <div class="flex flex-col gap-4 w-full max-w-lg mx-auto">
    <Logo class="!text-3xl" />

    <h1 class="text-2xl font-bold text-center">
      Registrierung
    </h1>

    <div class="flex flex-col items-center gap-2">
      <template v-if="!isFinished">
        <Message severity="secondary">
          Willkommen, bitte geben Sie Ihre E-Mail Adresse ein, mit welcher Sie sich registrieren wollen und wir senden Ihnen einen Link
          zur Bestätigung zu.
        </Message>

        <form
          class="grid grid-cols-1 gap-2 w-full"
          @submit.prevent
        >
          <div class="flex flex-col gap-2 col-span-full">
            <label
              class="text-sm font-bold"
              for="email"
            >E-Mail*</label>
            <InputText
              v-bind="emailProps"
              id="email"
              v-model="email"
              :class="{ 'p-invalid': errorsEmail['email']?.length }"
              type="email"
              autocomplete="email"
            />
            <small class="text-liqui-red">{{ errorsEmail["email"] }}</small>
          </div>

          <p class="w-full text-sm text-right">
            Konto vorhanden? Jetzt <NuxtLink
              :to="{ name: RouteNames.AUTH_LOGIN }"
              class="underline"
            >anmelden</NuxtLink>
          </p>

          <div class="flex justify-end gap-2 col-span-full">
            <Button
              :disabled="!metaPassword.valid || (metaPassword.valid && !metaPassword.dirty) || isSubmitting"
              :loading="isSubmitting"
              label="Registrierung starten"
              icon="pi pi-user"
              type="submit"
              @click="onRegistration"
            />
          </div>
        </form>
      </template>
      <Message
        v-else
        severity="success"
      >
        Vielen Dank für Ihre Registrierung. Bitte prüfen Sie nun Ihre Mailbox. Es kann vorkommen, dass E-Mails auch im "Spam" Ordner landen. Schauen sie ggf. auch dort nach.
      </Message>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import useAuth from '~/composables/useAuth'
import { Config } from '~/config/config'
import type { RegistrationFormData } from '~/models/auth'
import { RouteNames } from '~/config/routes'

useHead({
  title: 'Registrierung',
})

const { registration } = useAuth()
const toast = useToast()

const isFinished = ref(false)
const isSubmitting = ref(false)

const { defineField: defineFieldEmail, errors: errorsEmail, handleSubmit: handleSubmitEmail, meta: metaPassword } = useForm({
  validationSchema: yup.object({
    email: yup.string().email('Ungültiges E-Mail Format').trim().required('E-Mail wird benötigt'),
  }),
  initialValues: {
    email: '',
  } as RegistrationFormData,
})

const [email, emailProps] = defineFieldEmail('email')

const onRegistration = handleSubmitEmail(async (values) => {
  isSubmitting.value = true
  registration(values)
    .then(() => {
      isFinished.value = true
    })
    .catch(() => {
      toast.add({
        summary: 'Fehler',
        detail: `Die Registrierung ist fehlgeschlagen. Für diese E-Mail besteht evtl. bereits ein Konto`,
        severity: 'error',
        life: Config.TOAST_LIFE_TIME,
      })
    })
    .finally(() => {
      isSubmitting.value = false
    })
})
</script>
