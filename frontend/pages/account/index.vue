<template>
  <div class="flex flex-col gap-4">
    <div class="flex justify-between items-center gap-2">
      <hr class="h-0.5 bg-black flex-1"/>
      <p class="text-xl">Dein Profil</p>
      <hr class="h-0.5 bg-black flex-1"/>
    </div>

    <form @submit.prevent class="grid grid-cols-2 gap-2">
      <div class="flex flex-col gap-2 col-span-full md:col-span-1">
        <label class="text-sm font-bold" for="name">Name*</label>
        <InputText v-model="name" v-bind="nameProps"
                   :class="{'p-invalid': errorsProfile['name']?.length}"
                   id="name" type="text"/>
        <small class="text-liqui-red">{{errorsProfile["name"] || '&nbsp;'}}</small>
      </div>
      <div class="flex flex-col gap-2 col-span-full md:col-span-1">
        <label class="text-sm font-bold" for="email">E-Mail*</label>
        <InputText v-model="email" v-bind="emailProps"
                   :class="{'p-invalid': errorsProfile['email']?.length}"
                   id="email" type="email"/>
        <small class="text-liqui-red">{{errorsProfile["email"] || '&nbsp;'}}</small>
      </div>

      <Message v-if="profileUpdateMessage.length" severity="success" :life="Config.MESSAGE_LIFE_TIME" :sticky="false" :closable="false" class="col-span-full">{{profileUpdateMessage}}</Message>
      <Message v-if="profileUpdateErrorMessage.length" severity="error" :life="Config.MESSAGE_LIFE_TIME" :sticky="false" :closable="false" class="col-span-full">{{profileUpdateErrorMessage}}</Message>

      <div class="flex justify-end gap-2 col-span-full">
        <Button @click="onUpdateProfile" severity="info" :disabled="!metaProfile.valid || (metaProfile.valid && !metaProfile.dirty) || isSubmittingProfile" :loading="isSubmittingProfile" label="Profil aktualisieren" icon="pi pi-save" type="submit"/>
      </div>
    </form>

    <div class="flex justify-between items-center gap-2">
      <hr class="h-0.5 bg-black flex-1"/>
      <p class="text-xl">Sicherheit</p>
      <hr class="h-0.5 bg-black flex-1"/>
    </div>

    <form @submit.prevent class="grid grid-cols-2 gap-2">
      <div class="flex flex-col gap-2 col-span-full md:col-span-1">
        <label class="text-sm font-bold" for="password">Passwort ändern*</label>
        <InputText v-model="password" v-bind="passwordProps"
                   :class="{'p-invalid': errorsPassword['password']?.length}"
                   id="password" type="password"/>
        <small class="text-liqui-red">{{errorsPassword["password"] || '&nbsp;'}}</small>
      </div>
      <div class="flex flex-col gap-2 col-span-full md:col-span-1">
        <label class="text-sm font-bold" for="passwordRepeat">Passwort ändern wiederholen*</label>
        <InputText v-model="passwordRepeat" v-bind="passwordRepeatProps"
                   :class="{'p-invalid': errorsPassword['passwordRepeat']?.length}"
                   id="passwordRepeat" type="password"/>
        <small class="text-liqui-red">{{errorsPassword["passwordRepeat"] || '&nbsp;'}}</small>
      </div>

      <Message v-if="passwordUpdateMessage.length" severity="success" :life="Config.MESSAGE_LIFE_TIME" :sticky="false" :closable="false" class="col-span-full">{{passwordUpdateMessage}}</Message>
      <Message v-if="passwordUpdateErrorMessage.length" severity="error" :life="Config.MESSAGE_LIFE_TIME" :sticky="false" :closable="false" class="col-span-full">{{passwordUpdateErrorMessage}}</Message>

      <div class="flex justify-end gap-2 col-span-full">
        <Button @click="onUpdatePassword" severity="info" :disabled="!metaPassword.valid || (metaPassword.valid && !metaPassword.dirty) || isSubmittingPassword" :loading="isSubmittingPassword" label="Neues Passwort setzen" icon="pi pi-save" type="submit"/>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import {useForm} from "vee-validate";
import * as yup from "yup";
import type {UserProfileFormData, UserPasswordFormData} from "~/models/auth";
import {Config} from "~/config/config";

const {user, updateProfile, updatePassword} = useAuth()

const isSubmittingProfile = ref(false)
const profileUpdateMessage = ref('')
const profileUpdateErrorMessage = ref('')
const isSubmittingPassword = ref(false)
const passwordUpdateMessage = ref('')
const passwordUpdateErrorMessage = ref('')

const { defineField: defineFieldProfile, errors: errorsProfile, handleSubmit: handleSubmitProfile, meta: metaProfile, resetForm: resetFormProfile } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
    email: yup.string().email('Ungültiges E-Mail Format').trim().required('Name wird benötigt'),
  }),
  initialValues: {
    id: user.value?.id ?? undefined,
    name: user.value?.name ?? '',
    email: user.value?.email ?? '',
  } as UserProfileFormData
});

const { defineField: defineFieldPassword, errors: errorsPassword, handleSubmit: handleSubmitPassword, meta: metaPassword, resetForm: resetFormPassword } = useForm({
  validationSchema: yup.object({
    password: yup.string().when( {
      is: (val: string) => val.length > 0,
      then: (schema) => schema.min(8, 'Bitte gib mind. 8 Zeichen ein').required('Passwort wird benötigt'),
      otherwise: (schema) => schema.notRequired(),
    }),
    passwordRepeat: yup.string().test('passwords-match', 'Passwörter stimmen nicht überin',(value, context) => {
      return context.parent.password === value
    }),
  }),
  initialValues: {
    password: '',
    passwordRepeat: '',
  } as UserPasswordFormData
});

const [name, nameProps] = defineFieldProfile('name')
const [email, emailProps] = defineFieldProfile('email')

const [password, passwordProps] = defineFieldPassword('password')
const [passwordRepeat, passwordRepeatProps] = defineFieldPassword('passwordRepeat')

const onUpdateProfile = handleSubmitProfile((values) => {
  profileUpdateMessage.value = ''
  profileUpdateErrorMessage.value = ''
  isSubmittingProfile.value = true
  updateProfile(values)
      .then(async () => {
        resetFormProfile({values})
        profileUpdateMessage.value = 'Profil wurde bearbeitet'
      })
      .catch(() => {
        profileUpdateErrorMessage.value = 'Profil konnte nicht bearbeitet werden'
      })
      .finally(() => {
        isSubmittingProfile.value = false
      })
})

const onUpdatePassword = handleSubmitPassword((values) => {
  passwordUpdateMessage.value = ''
  passwordUpdateErrorMessage.value = ''
  isSubmittingPassword.value = true
  updatePassword(values)
      .then(async () => {
        resetFormPassword()
        passwordUpdateMessage.value = 'Password wurde geändert'
      })
      .catch(() => {
        passwordUpdateErrorMessage.value = 'Passwort konnte nicht geändert werden'
      })
      .finally(() => {
        isSubmittingPassword.value = false
      })
})
</script>
