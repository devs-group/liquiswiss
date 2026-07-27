<template>
  <form
    class="grid grid-cols-1 gap-2"
    @submit.prevent
  >
    <Message
      v-if="externalChange === 'deleted'"
      severity="warn"
      :closable="false"
    >
      Diese Kategorie wurde soeben extern gelöscht.
    </Message>

    <p class="text-sm">
      <span class="font-bold">"{{ category?.name }}"</span> wird noch von Transaktionen verwendet.
      Diese werden auf die Ersatzkategorie übertragen, danach wird die Kategorie gelöscht.
    </p>

    <div class="flex flex-col gap-2">
      <label
        class="text-sm font-bold"
        for="target-category"
      >Ersatzkategorie *</label>
      <Select
        id="target-category"
        v-model="targetID"
        empty-message="Keine Kategorien gefunden"
        :options="targetOptions"
        option-label="name"
        option-value="id"
        placeholder="Bitte wählen"
        filter
        auto-filter-focus
        empty-filter-message="Keine Resultate gefunden"
      />
    </div>

    <Message
      v-if="errorMessage.length"
      severity="error"
      :closable="false"
    >
      {{ errorMessage }}
    </Message>

    <div class="flex justify-end gap-2">
      <Button
        :disabled="!targetID || isLoading || externalChange === 'deleted'"
        :loading="isLoading"
        label="Übertragen und löschen"
        icon="pi pi-arrow-right-arrow-left"
        severity="danger"
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
import type { ICategoryReassignDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'

const dialogRef = inject<ICategoryReassignDialog>('dialogRef')!

// Escape/X close (no form: nothing to guard)
const { requestClose } = useDialogGuard({
  dirty: () => false,
  close: payload => dialogRef.value.close(payload),
})

const { categories, deleteCategory, reassignCategoryTransactions } = useCategories()
const toast = useToast()

const category = ref(dialogRef.value.data?.categoryToDelete)
const targetID = ref<number | null>(null)
const isLoading = ref(false)
const errorMessage = ref('')

// Any other visible category qualifies as target, system presets included
const targetOptions = computed(() => categories.value.filter(c => c.id !== category.value?.id))

// Real-time: when the category vanished externally there is nothing left to do
const { useExternalChangeBanner } = useRealtimeChanges()
const externalChange = useExternalChangeBanner('category', () => category.value?.id)

const onSubmit = async () => {
  if (!category.value || !targetID.value) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const affected = await reassignCategoryTransactions(category.value.id, targetID.value)
    await deleteCategory(category.value.id)
    toast.add({
      summary: 'Erfolg',
      detail: `${affected} Transaktion(en) übertragen, Kategorie "${category.value.name}" wurde gelöscht`,
      severity: 'success',
      life: Config.TOAST_LIFE_TIME,
    })
    dialogRef.value.close(true)
  }
  catch (err: unknown) {
    errorMessage.value = typeof err === 'string' ? err : (err as { message?: string }).message ?? 'Vorgang fehlgeschlagen'
  }
  finally {
    isLoading.value = false
  }
}
</script>
