<template>
  <div class="flex flex-col gap-4">
    <Logo/>

    <div v-if="isValid" class="flex flex-col items-center gap-2 w-full max-w-xl mx-auto">
      <Message severity="secondary">
        Fast geschafft! Geben Sie lediglich ein sicheres Passwort ein (mind. 8 Zeichen) um Ihre Registrierung abzuschliessen
      </Message>

      <form @submit.prevent class="grid grid-cols-1 gap-2 w-full max-w-sm">
        <div class="flex flex-col gap-2 col-span-full">
          <label class="text-sm font-bold" for="password">Passwort ändern*</label>
          <InputText v-model="password" v-bind="passwordProps"
                     :class="{'p-invalid': errorsPassword['password']?.length}"
                     id="password" type="password"/>
          <small class="text-liqui-red">{{errorsPassword["password"] || '&nbsp;'}}</small>
        </div>
        <div class="flex flex-col gap-2 col-span-full">
          <label class="text-sm font-bold" for="passwordRepeat">Passwort ändern wiederholen*</label>
          <InputText v-model="passwordRepeat" v-bind="passwordRepeatProps"
                     :class="{'p-invalid': errorsPassword['passwordRepeat']?.length}"
                     id="passwordRepeat" type="password"/>
          <small class="text-liqui-red">{{errorsPassword["passwordRepeat"] || '&nbsp;'}}</small>
        </div>

        <div class="flex justify-end gap-2 col-span-full">
          <Button @click="onFinishRegistration" severity="info" :disabled="!metaPassword.valid || (metaPassword.valid && !metaPassword.dirty) || isSubmitting" :loading="isSubmitting" label="Konto erstellen" icon="pi pi-save" type="submit"/>
        </div>
      </form>
    </div>
    <div v-else class="flex flex-col items-center gap-2 w-full max-w-2xl mx-auto">
      <Message class="w-full" severity="error">
        Dieser Registrierungslink ist nicht mehr gültig. Bitte registrieren Sie sich erneut.
      </Message>
      <NuxtLink :to="{name: RouteNames.AUTH_REGISTRATION}" replace>
        <Button label="Zur Registrierung" severity="info"/>
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import useAuth from "~/composables/useAuth";
import {Config} from "~/config/config";
import {RouteNames} from "~/config/routes";
import {useForm} from "vee-validate";
import * as yup from "yup";
import type {UserPasswordFormData} from "~/models/auth";

useHead({
  title: 'Validierung',
  meta: [
    {name: 'robots', content: 'noindex, nofollow'}
  ]
})

const {registrationCheckCode, registrationFinish} = useAuth()
const toast = useToast()
const route = useRoute()

const isValid = ref(true)
const isSubmitting = ref(false)

const email = ref(route.query.email as string ?? '')
const code = ref(route.query.code as string ?? '')

isValid.value = await registrationCheckCode({
  email: email.value,
  code: code.value,
})
    .catch(() => isValid.value = false)

const { defineField: defineFieldPassword, errors: errorsPassword, handleSubmit: handleSubmitPassword, meta: metaPassword } = useForm({
  validationSchema: yup.object({
    password: yup.string().when( {
      is: (val: string) => val.length > 0,
      then: (schema) => schema.min(8, 'Bitte geben Sie mind. 8 Zeichen ein').required('Passwort wird benötigt'),
      otherwise: (schema) => schema.notRequired(),
    }),
    passwordRepeat: yup.string().test('passwords-match', 'Passwörter stimmen nicht überein',(value, context) => {
      return context.parent.password === value
    }),
  }),
  initialValues: {
    password: '',
    passwordRepeat: '',
  } as UserPasswordFormData
});

const [password, passwordProps] = defineFieldPassword('password')
const [passwordRepeat, passwordRepeatProps] = defineFieldPassword('passwordRepeat')

const onFinishRegistration = handleSubmitPassword(async (values) => {
  isSubmitting.value = true
  const isRegistered = await registrationFinish({
    email: email.value,
    code: code.value,
    password: values.password,
  })
  if (isRegistered) {
    reloadNuxtApp({force: true})
  } else {
    isSubmitting.value = false
    toast.add({
      summary: 'Fehler',
      detail: `Die Registrierung ist fehlgeschlagen`,
      severity: 'error',
      life: Config.TOAST_LIFE_TIME,
    })
  }
})
</script>