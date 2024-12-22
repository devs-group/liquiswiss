<template>
  <form
    class="grid grid-cols-2 gap-2"
    @submit.prevent
  >
    <div class="col-span-2 flex flex-col gap-2">
      <label
        class="text-sm font-bold"
        for="name"
      >Name*</label>
      <InputText
        v-bind="nameProps"
        id="name"
        v-model="name"
        :class="{ 'p-invalid': errors['name']?.length }"
        type="text"
      />
      <small class="text-liqui-red">{{ errors["name"] || '&nbsp;' }}</small>
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
        label="Speichern"
        icon="pi pi-save"
        type="submit"
        @click="onCreateOrganisation"
      />
      <Button
        :loading="isLoading"
        label="Abbrechen"
        severity="secondary"
        @click="dialogRef?.close()"
      />
    </div>
  </form>
</template>

<script setup lang="ts">
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import type { IOrganisationFormDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'
import type { OrganisationFormData } from '~/models/organisation'
import { RouteNames } from '~/config/routes'

const dialogRef = inject<IOrganisationFormDialog>('dialogRef')!

const { createOrganisation } = useOrganisations()
const toast = useToast()

const isLoading = ref(false)
const errorMessage = ref('')

const { defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
  }),
  initialValues: {
    name: '',
  } as OrganisationFormData,
})

const [name, nameProps] = defineField('name')

const onCreateOrganisation = handleSubmit((values) => {
  isLoading.value = true
  errorMessage.value = ''
  createOrganisation(values)
    .then(async (organisation) => {
      dialogRef?.value.close()
      navigateTo({ name: RouteNames.SETTINGS_ORGANISATION_EDIT, params: { id: organisation.id } })
      toast.add({
        summary: 'Erfolg',
        detail: `Organisation "${organisation.name}" wurde angelegt`,
        severity: 'success',
        life: Config.TOAST_LIFE_TIME,
      })
    })
    .catch(() => {
      errorMessage.value = 'Organisation konnte nicht angelegt werden'
    })
    .finally(() => {
      isLoading.value = false
    })
})
</script>
