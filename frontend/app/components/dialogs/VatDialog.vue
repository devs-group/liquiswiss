<template>
  <form
    class="grid grid-cols-2 gap-2"
    @submit.prevent
  >
    <Message
      v-if="externalChange"
      severity="warn"
      class="col-span-full"
      :closable="false"
    >
      <div class="flex items-center justify-between gap-4 w-full">
        <span v-if="externalChange === 'deleted'">
          Diese Mehrwertsteuer wurde soeben extern gelöscht.
        </span>
        <span v-else>
          Diese Mehrwertsteuer wurde soeben extern geändert. Beim Speichern werden die externen Änderungen überschrieben.
        </span>
        <Button
          v-if="externalChange === 'updated'"
          label="Neu laden"
          icon="pi pi-refresh"
          size="small"
          severity="warn"
          @click="onReloadExternalChange"
        />
      </div>
    </Message>
    <div class="col-span-2 flex flex-col gap-2">
      <label
        class="text-sm font-bold"
        for="value"
      >Prozentualer Wert *</label>
      <InputNumber
        v-bind="valueProps"
        id="value"
        v-model="value"
        suffix=" %"
        :class="{ 'p-invalid': errors['value']?.length }"
        mode="decimal"
        :allow-empty="false"
        placeholder="Beispiel: 8.1"
        fluid
        :max-fraction-digits="2"
        @paste="onParseAmount"
        @input="event => value = event.value"
        @focus="selectAllOnFocus"
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
        @click="requestClose()"
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
import { selectAllOnFocus } from '~/utils/element-helper'

const dialogRef = inject<IVatFormDialog>('dialogRef')!

const { vats, getVat, createVat, updateVat } = useVat()
const toast = useToast()

const vat = ref(dialogRef.value.data?.vatToEdit)
const isCreate = !vat.value?.id
const isLoading = ref(false)
const errorMessage = ref('')

const { defineField, errors, handleSubmit, meta, resetForm, values: formValues, setValues } = useForm<VatFormData>({
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
    id: vat.value?.id ?? undefined,
    value: isCreate ? undefined : vat.value?.value ? AmountToFloat(vat.value.value) : undefined,
  },
})

// Real-time: warn when the VAT being edited was changed or deleted
// externally (form values stay untouched, saving would overwrite)
// Escape/X close with dirty-confirm; unsaved entries survive reloads
const { requestClose, close } = useDialogGuard({
  dirty: () => meta.value.dirty,
  close: payload => dialogRef.value.close(payload),
  draft: {
    key: `vat:${vat.value?.id ?? 'new'}`,
    capture: () => ({ ...formValues }),
    restore: saved => setValues(saved),
  },
})

const { useExternalChangeBanner } = useRealtimeChanges()
const externalChange = useExternalChangeBanner('vat', () => isCreate ? undefined : vat.value?.id)

const onReloadExternalChange = async () => {
  if (!vat.value?.id) return
  try {
    const fresh = await getVat(vat.value.id)
    vat.value = fresh
    resetForm({
      values: {
        id: fresh.id,
        value: fresh.value ? AmountToFloat(fresh.value) : undefined,
      },
    })
    externalChange.value = null
  }
  catch {
    // Keep the banner; reloading failed
  }
}

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
        close(vat.id)
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
        close(values.id)
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
  if (event instanceof ClipboardEvent) {
    const pastedText = event.clipboardData?.getData('text') ?? ''
    const parsedAmount = parseCurrency(pastedText, false)
    value.value = parsedAmount.length > 0 ? parseFloat(parsedAmount) : 0
  }
}
</script>
