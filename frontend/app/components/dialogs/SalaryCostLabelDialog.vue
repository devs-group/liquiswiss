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
          Dieses Kostenlabel wurde soeben extern gelöscht.
        </span>
        <span v-else>
          Dieses Kostenlabel wurde soeben extern geändert. Beim Speichern werden die externen Änderungen überschrieben.
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
        :disabled="!isDirty || !meta.valid || isLoading"
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
import type { ISalaryCostLabelFormDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'
import type { SalaryCostLabelFormData } from '~/models/employee'

const dialogRef = inject<ISalaryCostLabelFormDialog>('dialogRef')!

const { getSalaryCostLabel, createSalaryCostLabel, updateSalaryCostLabel } = useSalaryCostLabels()
const toast = useToast()

const salaryCost = ref(dialogRef.value.data?.employeeCostLabelToEdit)
const isCreate = !salaryCost.value?.id
const isLoading = ref(false)
const errorMessage = ref('')

const { defineField, errors, handleSubmit, meta, resetForm, values: formValues, setValues } = useForm<SalaryCostLabelFormData>({
  validationSchema: yup.object({
    name: yup.string().required('Name wird benötigt').typeError('Ungültiger Wert'),
  }),
  initialValues: {
    id: salaryCost.value?.id ?? undefined,
    name: isCreate ? undefined : salaryCost.value?.name ? salaryCost.value.name : undefined,
  },
})

// Real-time: warn when the cost label being edited was changed or deleted
// externally (form values stay untouched, saving would overwrite)
const isDirty = useTrimmedDirty(meta, formValues)

// Escape/X close with dirty-confirm; unsaved entries survive reloads
const { requestClose, close } = useDialogGuard({
  dirty: () => isDirty.value,
  close: payload => dialogRef.value.close(payload),
  draft: {
    key: `salary_cost_label:${salaryCost.value?.id ?? 'new'}`,
    capture: () => ({ ...formValues }),
    restore: saved => setValues(saved),
  },
})

const { useExternalChangeBanner } = useRealtimeChanges()
const externalChange = useExternalChangeBanner('salary_cost_label', () => isCreate ? undefined : salaryCost.value?.id)

const onReloadExternalChange = async () => {
  if (!salaryCost.value?.id) return
  try {
    const fresh = await getSalaryCostLabel(salaryCost.value.id)
    salaryCost.value = fresh
    resetForm({
      values: {
        id: fresh.id,
        name: fresh.name ?? undefined,
      },
    })
    externalChange.value = null
  }
  catch {
    // Keep the banner; reloading failed
  }
}

const [name, nameProps] = defineField('name')

const onSubmit = handleSubmit((values) => {
  isLoading.value = true
  errorMessage.value = ''

  if (isCreate) {
    createSalaryCostLabel(values)
      .then(async (costLabel) => {
        close(costLabel.id)
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
        close(costLabel.id)
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
