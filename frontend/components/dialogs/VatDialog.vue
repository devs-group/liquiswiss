<template>
  <form
    class="grid grid-cols-2 gap-2"
    @submit.prevent
  >
    <div class="col-span-2 flex flex-col gap-2">
      <label
        class="text-sm font-bold"
        for="name"
      >Prozentualer Wert *</label>
      <InputText
        v-bind="valueProps"
        id="value"
        v-model="value"
        :class="{ 'p-invalid': errors['value']?.length }"
        type="text"
        placeholder="Beispiel: 8.1"
        @input="onParseAmount"
      />
      <div class="flex justify-between gap-2">
        <small class="text-liqui-red">{{ errors["value"] || '&nbsp;' }}</small>
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
        severity="info"
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
import type { IVatFormDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'
import type { VatFormData } from '~/models/vat'

const dialogRef = inject<IVatFormDialog>('dialogRef')!

const { vats, createVat, updateVat } = useVat()
const toast = useToast()

const vat = dialogRef.value.data?.vat
const isCreate = !vat?.id
const isLoading = ref(false)
const errorMessage = ref('')

const { defineField, errors, handleSubmit, meta } = useForm<VatFormData>({
  validationSchema: yup.object({
    value: yup.number()
      .required('Prozentualer Wert wird benötigt')
      .min(0.01, 'Mindestwert: 0.01')
      .typeError('Ungültiger Wert')
      .test('unique', 'Mehrwertsteuer existiert bereits', (value) => {
        const valueAsInt = AmountToInteger(value)
        return !vats.value.find(vat => vat.value === valueAsInt)
      }),
  }),
  initialValues: {
    id: vat?.id ?? undefined,
    value: isCreate ? undefined : vat?.value ? AmountToFloat(vat.value) : undefined,
  },
})

const [value, valueProps] = defineField('value')

const onSubmit = handleSubmit((values) => {
  const val = parseFloat(values.value as string)
  isLoading.value = true
  errorMessage.value = ''

  if (isCreate) {
    createVat({
      value: val,
    })
      .then(async (vat) => {
        dialogRef?.value.close(vat.id)
        toast.add({
          summary: 'Erfolg',
          detail: `Mehrwertsteuer "${vat.formattedValue}" wurde angelegt`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
      .catch(() => {
        errorMessage.value = 'Mehrwertsteuer konnte nicht angelegt werden'
      })
      .finally(() => {
        isLoading.value = false
      })
  }
  else {
    updateVat({
      id: values.id,
      value: val,
    })
      .then(async () => {
        dialogRef?.value.close(values.id)
        toast.add({
          summary: 'Erfolg',
          detail: `Mehrwertsteuer wurde bearbeitet`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
      .catch(() => {
        errorMessage.value = 'Mehrwertsteuer konnte nicht bearbeitet werden'
      })
      .finally(() => {
        isLoading.value = false
      })
  }
})

const onParseAmount = (event: Event) => {
  if (event instanceof InputEvent) {
    parseNumberInput(event, value as Ref<number>, false)
  }
}
</script>
