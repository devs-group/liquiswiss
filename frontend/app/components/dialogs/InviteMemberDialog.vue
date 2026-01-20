<template>
  <form
    class="grid grid-cols-2 gap-2"
    @submit.prevent
  >
    <div class="col-span-2 flex flex-col gap-2">
      <label
        class="text-sm font-bold"
        for="email"
      >E-Mail*</label>
      <InputText
        v-bind="emailProps"
        id="email"
        v-model="email"
        :class="{ 'p-invalid': errors['email']?.length }"
        type="email"
      />
      <small class="text-liqui-red">{{ errors["email"] || '&nbsp;' }}</small>
    </div>

    <div class="col-span-2 flex flex-col gap-2">
      <label
        class="text-sm font-bold"
        for="role"
      >Rolle*</label>
      <Select
        v-bind="roleProps"
        id="role"
        v-model="role"
        :options="roleOptions"
        option-label="label"
        option-value="value"
        placeholder="Bitte wählen"
        :class="{ 'p-invalid': errors['role']?.length }"
      />
      <small class="text-liqui-red">{{ errors["role"] || '&nbsp;' }}</small>
    </div>

    <Message
      v-if="errorMessage.length"
      severity="error"
      :closable="false"
      class="col-span-full"
    >
      {{ errorMessage }}
    </Message>

    <div class="flex justify-end gap-2 col-span-full">
      <Button
        :disabled="!meta.valid || isLoading"
        :loading="isLoading"
        label="Einladen"
        icon="pi pi-send"
        type="submit"
        @click="onInvite"
      />
      <Button
        :loading="isLoading"
        label="Abbrechen"
        severity="contrast"
        @click="dialogRef?.close()"
      />
    </div>
  </form>
</template>

<script setup lang="ts">
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import type { IInviteMemberDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'
import type { CreateInvitationFormData } from '~/models/invitation'

const dialogRef = inject<IInviteMemberDialog>('dialogRef')!

const { createInvitation } = useInvitations()
const toast = useToast()

const isLoading = ref(false)
const errorMessage = ref('')

const roleOptions = [
  { label: 'Admin', value: 'admin' },
  { label: 'Editor', value: 'editor' },
  { label: 'Nur Lesen', value: 'read-only' },
]

const { defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    email: yup.string().email('Ungültige E-Mail-Adresse').trim().required('E-Mail wird benötigt'),
    role: yup.string().required('Rolle wird benötigt'),
  }),
  initialValues: {
    email: '',
    role: 'read-only',
  } as CreateInvitationFormData,
})

const [email, emailProps] = defineField('email')
const [role, roleProps] = defineField('role')

const onInvite = handleSubmit((values) => {
  isLoading.value = true
  errorMessage.value = ''
  createInvitation(dialogRef.value.data.organisationId, values)
    .then(() => {
      dialogRef?.value.close()
      toast.add({
        summary: 'Erfolg',
        detail: `Einladung an "${values.email}" wurde gesendet`,
        severity: 'success',
        life: Config.TOAST_LIFE_TIME,
      })
    })
    .catch((err) => {
      errorMessage.value = err
    })
    .finally(() => {
      isLoading.value = false
    })
})
</script>
