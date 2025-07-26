<template>
  <form
    class="grid grid-cols-2 gap-2"
    @submit.prevent
  >
    <div class="col-span-2 flex flex-col gap-2">
      <label
        class="text-sm font-bold"
        for="name"
      >Label *</label>
      <InputText
        v-bind="nameProps"
        id="name"
        v-model="name"
        placeholder="BVG, AHV, Quellensteuer, ..."
        :class="{ 'p-invalid': errors['name']?.length }"
        type="text"
      />
      <div class="flex justify-between gap-2">
        <small class="text-liqui-red">{{ errors["name"] }}</small>
      </div>
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
        @click="onSubmit"
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
import type { ISalaryCostLabelFormDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'
import type { SalaryCostLabelFormData } from '~/models/employee'

const dialogRef = inject<ISalaryCostLabelFormDialog>('dialogRef')!

const { createSalaryCostLabel, updateSalaryCostLabel } = useSalaryCostLabels()
const toast = useToast()

const salaryCost = dialogRef.value.data?.employeeCostLabelToEdit
const isCreate = !salaryCost?.id
const isLoading = ref(false)
const errorMessage = ref('')

const { defineField, errors, handleSubmit, meta } = useForm<SalaryCostLabelFormData>({
  validationSchema: yup.object({
    name: yup.string().required('Name wird benötigt').typeError('Ungültiger Wert'),
  }),
  initialValues: {
    id: salaryCost?.id ?? undefined,
    name: isCreate ? undefined : salaryCost?.name ? salaryCost.name : undefined,
  },
})

const [name, nameProps] = defineField('name')

const onSubmit = handleSubmit((values) => {
  isLoading.value = true
  errorMessage.value = ''

  if (isCreate) {
    createSalaryCostLabel(values)
      .then(async (costLabel) => {
        dialogRef?.value.close(costLabel.id)
        toast.add({
          summary: 'Erfolg',
          detail: `Kostenlabel "${costLabel.name}" wurde angelegt`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
      .catch(() => {
        errorMessage.value = 'Kostenlabel konnte nicht angelegt werden'
      })
      .finally(() => {
        isLoading.value = false
      })
  }
  else {
    updateSalaryCostLabel(values)
      .then(async (costLabel) => {
        dialogRef?.value.close(costLabel.id)
        toast.add({
          summary: 'Erfolg',
          detail: `Kostenlabel wurde bearbeitet`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
      .catch(() => {
        errorMessage.value = 'Kostenlabel konnte nicht bearbeitet werden'
      })
      .finally(() => {
        isLoading.value = false
      })
  }
})
</script>
