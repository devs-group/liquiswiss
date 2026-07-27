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
          Diese Kategorie wurde soeben extern gelöscht.
        </span>
        <span v-else>
          Diese Kategorie wurde soeben extern geändert. Beim Speichern werden die externen Änderungen überschrieben.
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
      >Name *</label>
      <InputText
        v-bind="nameProps"
        id="name"
        v-model="name"
        placeholder="Miete, Marketing, Software, ..."
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
import type { ICategoryFormDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'
import type { CategoryFormData } from '~/models/category'

const dialogRef = inject<ICategoryFormDialog>('dialogRef')!

const { getCategory, createCategory, updateCategory } = useCategories()
const toast = useToast()

const category = ref(dialogRef.value.data?.categoryToEdit)
const isCreate = !category.value?.id
const isLoading = ref(false)
const errorMessage = ref('')

const { defineField, errors, handleSubmit, meta, resetForm, values, setValues } = useForm<CategoryFormData>({
  validationSchema: yup.object({
    name: yup.string().required('Name wird benötigt').typeError('Ungültiger Wert'),
  }),
  initialValues: {
    id: category.value?.id ?? undefined,
    name: isCreate ? undefined : category.value?.name,
  },
})

const isDirty = useTrimmedDirty(meta, values)

// Escape/X close with dirty-confirm; unsaved entries survive reloads
const { requestClose, close } = useDialogGuard({
  dirty: () => isDirty.value,
  close: payload => dialogRef.value.close(payload),
  draft: {
    key: `category:${category.value?.id ?? 'new'}`,
    capture: () => ({ ...values }),
    restore: saved => setValues(saved),
  },
})

// Real-time: warn when the category being edited was changed or deleted
// externally (form values stay untouched, saving would overwrite)
const { useExternalChangeBanner } = useRealtimeChanges()
const externalChange = useExternalChangeBanner('category', () => isCreate ? undefined : category.value?.id)

const onReloadExternalChange = async () => {
  if (!category.value?.id) return
  try {
    const fresh = await getCategory(category.value.id)
    category.value = fresh
    resetForm({
      values: {
        id: fresh.id,
        name: fresh.name,
      },
    })
    externalChange.value = null
  }
  catch {
    // Keep the banner; reloading failed
  }
}

const [name, nameProps] = defineField('name')

const onSubmit = handleSubmit((rawValues) => {
  const values = trimStringValues(rawValues)
  isLoading.value = true
  errorMessage.value = ''

  if (isCreate) {
    createCategory(values)
      .then((created) => {
        close(created.id)
        toast.add({
          summary: 'Erfolg',
          detail: `Kategorie "${created.name}" wurde angelegt`,
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
  }
  else {
    updateCategory(values)
      .then((updated) => {
        close(updated.id)
        toast.add({
          summary: 'Erfolg',
          detail: 'Kategorie wurde bearbeitet',
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
  }
})
</script>
