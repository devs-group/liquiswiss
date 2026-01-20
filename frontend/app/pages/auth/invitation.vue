<template>
  <div class="flex flex-col gap-4">
    <Logo />

    <!-- Loading state -->
    <div
      v-if="isChecking"
      class="flex flex-col items-center gap-4"
    >
      <ProgressSpinner />
      <p>Einladung wird überprüft...</p>
    </div>

    <!-- Valid invitation -->
    <div
      v-else-if="invitationData"
      class="flex flex-col items-center gap-4 w-full max-w-xl mx-auto"
    >
      <Message severity="success">
        Sie wurden eingeladen, der Organisation <strong>{{ invitationData.organisationName }}</strong> beizutreten.
      </Message>

      <div class="text-center text-sm text-gray-500">
        Eingeladen von: {{ invitationData.invitedByName }}
      </div>

      <!-- Existing user - just accept -->
      <template v-if="invitationData.existingUser">
        <Message severity="info">
          Sie haben bereits ein Konto. Klicken Sie auf "Annehmen", um der Organisation beizutreten.
        </Message>

        <Button
          label="Einladung annehmen"
          icon="pi pi-check"
          :loading="isSubmitting"
          @click="onAcceptExistingUser"
        />
      </template>

      <!-- New user - needs password -->
      <template v-else>
        <Message severity="secondary">
          Geben Sie ein sicheres Passwort ein (mind. 8 Zeichen), um Ihr Konto zu erstellen und der Organisation beizutreten.
        </Message>

        <form
          class="grid grid-cols-1 gap-2 w-full max-w-sm"
          @submit.prevent
        >
          <div class="flex flex-col gap-2 col-span-full">
            <label
              class="text-sm font-bold"
              for="email"
            >E-Mail</label>
            <InputText
              id="email"
              :model-value="invitationData.email"
              disabled
              type="email"
            />
          </div>

          <div class="flex flex-col gap-2 col-span-full">
            <label
              class="text-sm font-bold"
              for="password"
            >Passwort*</label>
            <InputText
              v-bind="passwordProps"
              id="password"
              v-model="password"
              :class="{ 'p-invalid': errors['password']?.length }"
              type="password"
            />
            <small class="text-liqui-red">{{ errors["password"] || '&nbsp;' }}</small>
          </div>

          <div class="flex flex-col gap-2 col-span-full">
            <label
              class="text-sm font-bold"
              for="passwordRepeat"
            >Passwort wiederholen*</label>
            <InputText
              v-bind="passwordRepeatProps"
              id="passwordRepeat"
              v-model="passwordRepeat"
              :class="{ 'p-invalid': errors['passwordRepeat']?.length }"
              type="password"
            />
            <small class="text-liqui-red">{{ errors["passwordRepeat"] || '&nbsp;' }}</small>
          </div>

          <div class="flex justify-end gap-2 col-span-full">
            <Button
              severity="info"
              :disabled="!meta.valid || !meta.dirty || isSubmitting"
              :loading="isSubmitting"
              label="Konto erstellen & beitreten"
              icon="pi pi-user-plus"
              type="submit"
              @click="onAcceptNewUser"
            />
          </div>
        </form>
      </template>
    </div>

    <!-- Invalid/expired invitation -->
    <div
      v-else
      class="flex flex-col items-center gap-2 w-full max-w-2xl mx-auto"
    >
      <Message
        class="w-full"
        severity="error"
      >
        {{ errorMessage || 'Dieser Einladungslink ist nicht gültig oder abgelaufen.' }}
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
import type { CheckInvitationResponse } from '~/models/invitation'
import { Config } from '~/config/config'
import { RouteNames } from '~/config/routes'

useHead({
  title: 'Einladung',
  meta: [
    { name: 'robots', content: 'noindex, nofollow' },
  ],
})

const route = useRoute()
const toast = useToast()
const { checkInvitation, acceptInvitation } = useInvitations()

const token = ref(route.query.token as string ?? '')
const isChecking = ref(true)
const isSubmitting = ref(false)
const invitationData = ref<CheckInvitationResponse | null>(null)
const errorMessage = ref('')

// Check invitation validity
if (token.value) {
  invitationData.value = await checkInvitation(token.value)
    .catch((err) => {
      errorMessage.value = err
      return null
    })
    .finally(() => {
      isChecking.value = false
    })
}
else {
  isChecking.value = false
  errorMessage.value = 'Kein Einladungstoken gefunden'
}

// Form for new users
const { defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    password: yup.string()
      .min(8, 'Bitte geben Sie mind. 8 Zeichen ein')
      .required('Passwort wird benötigt'),
    passwordRepeat: yup.string()
      .test('passwords-match', 'Passwörter stimmen nicht überein', (value, context) => {
        return context.parent.password === value
      }),
  }),
  initialValues: {
    password: '',
    passwordRepeat: '',
  },
})

const [password, passwordProps] = defineField('password')
const [passwordRepeat, passwordRepeatProps] = defineField('passwordRepeat')

// Accept for existing user
const onAcceptExistingUser = async () => {
  isSubmitting.value = true
  await acceptInvitation({ token: token.value })
    .then(() => {
      toast.add({
        summary: 'Erfolg',
        detail: 'Sie sind der Organisation beigetreten',
        severity: 'success',
        life: Config.TOAST_LIFE_TIME,
      })
      reloadNuxtApp({ force: true })
    })
    .catch((err) => {
      toast.add({
        summary: 'Fehler',
        detail: err,
        severity: 'error',
        life: Config.TOAST_LIFE_TIME,
      })
      isSubmitting.value = false
    })
}

// Accept for new user (with password)
const onAcceptNewUser = handleSubmit(async (values) => {
  isSubmitting.value = true
  await acceptInvitation({ token: token.value, password: values.password })
    .then(() => {
      toast.add({
        summary: 'Erfolg',
        detail: 'Ihr Konto wurde erstellt und Sie sind der Organisation beigetreten',
        severity: 'success',
        life: Config.TOAST_LIFE_TIME,
      })
      reloadNuxtApp({ force: true })
    })
    .catch((err) => {
      toast.add({
        summary: 'Fehler',
        detail: err,
        severity: 'error',
        life: Config.TOAST_LIFE_TIME,
      })
      isSubmitting.value = false
    })
})
</script>
