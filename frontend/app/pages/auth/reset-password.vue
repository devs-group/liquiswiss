<template>
  <div class="flex flex-col gap-4">
    <Logo />

    <div
      v-if="isValid"
      class="flex flex-col items-center gap-2 w-full max-w-xl mx-auto"
    >
      <Message severity="secondary">
        Geben Sie Ihr neues, sicheres Passwort ein (mind. 8 Zeichen) um das Passwort zurückzusetzen
      </Message>

      <form
        class="grid grid-cols-1 gap-2 w-full max-w-sm"
        @submit.prevent
      >
        <div class="flex flex-col gap-2 col-span-full">
          <label
            class="text-sm font-bold"
            for="password"
          >Passwort ändern*</label>
          <InputText
            v-bind="passwordProps"
            id="password"
            v-model="password"
            :class="{ 'p-invalid': errorsPassword['password']?.length }"
            type="password"
          />
          <small class="text-liqui-red">{{ errorsPassword["password"] || '&nbsp;' }}</small>
        </div>
        <div class="flex flex-col gap-2 col-span-full">
          <label
            class="text-sm font-bold"
            for="passwordRepeat"
          >Passwort ändern wiederholen*</label>
          <InputText
            v-bind="passwordRepeatProps"
            id="passwordRepeat"
            v-model="passwordRepeat"
            :class="{ 'p-invalid': errorsPassword['passwordRepeat']?.length }"
            type="password"
          />
          <small class="text-liqui-red">{{ errorsPassword["passwordRepeat"] || '&nbsp;' }}</small>
        </div>

        <div class="flex justify-end gap-2 col-span-full">
          <Button
            severity="info"
            :disabled="!metaPassword.valid || (metaPassword.valid && !metaPassword.dirty) || isSubmitting"
            :loading="isSubmitting"
            label="Passwort setzen"
            icon="pi pi-save"
            type="submit"
            @click="onFinishResetPassword"
          />
        </div>
      </form>
    </div>
    <div
      v-else
      class="flex flex-col items-center gap-2 w-full max-w-2xl mx-auto"
    >
      <Message
        class="w-full"
        severity="error"
      >
        Der Link zum Zurücksetzen des Passworts ist nicht mehr gültig.
        Bitte setzen Sie dass Passwort erneut zurück.
      </Message>
      <NuxtLink
        :to="{ name: RouteNames.AUTH_LOGIN }"
        replace
      >
        <Button
          label="Zum Login"
          severity="info"
        />
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import useAuth from '~/composables/useAuth'
import { Config } from '~/config/config'
import { RouteNames } from '~/config/routes'
import type { UserPasswordFormData } from '~/models/auth'

useHead({
  title: 'Passwort zurücksetzen',
  meta: [
    { name: 'robots', content: 'noindex, nofollow' },
  ],
})

const { resetPassword, resetPasswordCheckCode } = useAuth()
const toast = useToast()
const route = useRoute()

const isValid = ref(true)
const isSubmitting = ref(false)

const email = ref(route.query.email as string ?? '')
const code = ref(route.query.code as string ?? '')

isValid.value = await resetPasswordCheckCode({
  email: email.value,
  code: code.value,
})
  .catch(() => isValid.value = false)

const { defineField: defineFieldPassword, errors: errorsPassword, handleSubmit: handleSubmitPassword, meta: metaPassword } = useForm({
  validationSchema: yup.object({
    password: yup.string().when({
      is: (val: string) => val.length > 0,
      then: schema => schema.min(8, 'Bitte geben Sie mind. 8 Zeichen ein').required('Passwort wird benötigt'),
      otherwise: schema => schema.notRequired(),
    }),
    passwordRepeat: yup.string().test('passwords-match', 'Passwörter stimmen nicht überein', (value, context) => {
      return context.parent.password === value
    }),
  }),
  initialValues: {
    password: '',
    passwordRepeat: '',
  } as UserPasswordFormData,
})

const [password, passwordProps] = defineFieldPassword('password')
const [passwordRepeat, passwordRepeatProps] = defineFieldPassword('passwordRepeat')

const onFinishResetPassword = handleSubmitPassword(async (values) => {
  isSubmitting.value = true
  await resetPassword({
    email: email.value,
    code: code.value,
    password: values.password,
  })
    .then(async () => {
      toast.add({
        summary: 'Erfolg',
        detail: `Das Passwort wurde zurückgesetzt. Sie können sich nun anmelden`,
        severity: 'success',
        life: Config.TOAST_LIFE_TIME,
      })
      await navigateTo({ name: RouteNames.AUTH_LOGIN, replace: true })
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
