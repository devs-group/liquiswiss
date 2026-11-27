<template>
  <form
    id="scenario-form"
    class="grid grid-cols-1 gap-2"
    @submit.prevent
  >
    <Message
      v-if="externalChange"
      severity="warn"
      :closable="false"
    >
      <div class="flex items-center justify-between gap-4 w-full">
        <span v-if="externalChange === 'deleted'">
          Dieses Szenario wurde soeben extern gelöscht.
        </span>
        <span v-else>
          Dieses Szenario wurde soeben extern geändert. Beim Speichern werden die externen Änderungen überschrieben.
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

    <p
      v-if="isCreate"
      class="text-sm text-surface-500"
    >
      Das neue Szenario ist eine vollständige Kopie der Vorlage: Mitarbeitende, Löhne, Transaktionen, Bankkonten und Ausschlüsse.
      Änderungen daran wirken sich nicht auf die Vorlage aus. Kategorien, MWST und Lohnkosten-Labels gelten für alle Szenarien.
    </p>

    <div class="flex flex-col gap-2">
      <label
        class="text-sm font-bold"
        for="scenario-name"
      >Name *</label>
      <InputText
        v-bind="nameProps"
        id="scenario-name"
        v-model="name"
        :class="{ 'p-invalid': errors['name']?.length }"
        type="text"
      />
      <small class="text-liqui-red">{{ errors["name"] || '&nbsp;' }}</small>
    </div>

    <div
      v-if="isCreate"
      class="flex flex-col gap-2"
    >
      <label
        class="text-sm font-bold"
        for="scenario-source"
      >Vorlage *</label>
      <Select
        v-bind="sourceScenarioIdProps"
        id="scenario-source"
        v-model="sourceScenarioId"
        :options="scenarios"
        option-label="name"
        option-value="id"
        empty-message="Keine Szenarien gefunden"
        :class="{ 'p-invalid': errors['sourceScenarioId']?.length }"
      />
      <small class="text-liqui-red">{{ errors["sourceScenarioId"] || '&nbsp;' }}</small>
    </div>

    <hr class="my-4">

    <Message
      v-if="errorMessage.length"
      severity="error"
      :closable="false"
    >
      {{ errorMessage }}
    </Message>

    <div class="flex items-center justify-end gap-2">
      <Button
        :disabled="(!isDirty && !isCreate) || !meta.valid || isLoading"
        :loading="isLoading"
        :label="isCreate ? 'Anlegen' : 'Speichern'"
        icon="pi pi-save"
        type="submit"
        @click="onSubmit"
      />
      <Button
        :loading="isLoading"
        label="Abbrechen"
        severity="contrast"
        @click="requestClose()"
      />
      <Button
        v-if="!isCreate && !scenario?.isBase"
        v-tooltip.top="'Szenario löschen'"
        :disabled="isLoading"
        severity="danger"
        outlined
        rounded
        icon="pi pi-trash"
        @click="onDeleteScenario"
      />
    </div>
  </form>
</template>

<script setup lang="ts">
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import type { IScenarioFormDialog } from '~/interfaces/dialog-interfaces'
import type { ScenarioFormData } from '~/models/scenario'
import { Config } from '~/config/config'

const dialogRef = inject<IScenarioFormDialog>('dialogRef')!

const { scenarios, currentScenario, getScenario, createScenario, updateScenario, deleteScenario } = useScenarios()
const confirm = useConfirm()
const toast = useToast()

const isLoading = ref(false)
const scenario = ref(dialogRef.value.data?.scenario)
const isCreate = !scenario.value?.id
const errorMessage = ref('')

const { defineField, errors, handleSubmit, meta, resetForm, values: formValues, setValues } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt').max(100, 'Höchstens 100 Zeichen'),
    sourceScenarioId: isCreate
      ? yup.number().required('Vorlage wird benötigt').typeError('Ungültige Vorlage')
      : yup.number().nullable(),
  }),
  initialValues: {
    id: scenario.value?.id ?? undefined,
    name: scenario.value?.name ?? '',
    sourceScenarioId: isCreate ? currentScenario.value?.id ?? null : null,
  } as ScenarioFormData,
})

const isDirty = useTrimmedDirty(meta, formValues)

const { requestClose, close } = useDialogGuard({
  dirty: () => isDirty.value,
  close: payload => dialogRef.value.close(payload),
  draft: {
    key: `scenario:${isCreate ? 'new' : scenario.value?.id}`,
    capture: () => ({ ...formValues }),
    restore: saved => setValues(saved),
  },
})

const { useExternalChangeBanner } = useRealtimeChanges()
const externalChange = useExternalChangeBanner('scenario', () => isCreate ? undefined : scenario.value?.id)

const onReloadExternalChange = async () => {
  if (!scenario.value?.id) return
  try {
    const fresh = await getScenario(scenario.value.id)
    scenario.value = fresh
    resetForm({ values: { id: fresh.id, name: fresh.name, sourceScenarioId: null } as ScenarioFormData })
    externalChange.value = null
  }
  catch {
    // Keep the banner; reloading failed
  }
}

const [name, nameProps] = defineField('name')
const [sourceScenarioId, sourceScenarioIdProps] = defineField('sourceScenarioId')

const showError = (message: string) => {
  errorMessage.value = message
  nextTick(() => {
    scrollToParentBottom('scenario-form')
  })
}

const onSubmit = handleSubmit((rawValues) => {
  const values = trimStringValues(rawValues)
  isLoading.value = true
  errorMessage.value = ''

  if (isCreate) {
    createScenario(values)
      .then(() => {
        toast.add({
          summary: 'Erfolg',
          detail: `Szenario "${values.name}" wurde angelegt und ist jetzt aktiv`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
        close({ reload: true })
      })
      .catch(showError)
      .finally(() => {
        isLoading.value = false
      })
  }
  else {
    updateScenario(values)
      .then(() => {
        toast.add({
          summary: 'Erfolg',
          detail: `Szenario "${values.name}" wurde bearbeitet`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
        close()
      })
      .catch(showError)
      .finally(() => {
        isLoading.value = false
      })
  }
})

const onDeleteScenario = () => {
  if (!scenario.value) return
  const deleted = scenario.value
  confirm.require({
    header: 'Löschen',
    message: `Szenario "${deleted.name}" mit allen Daten darin löschen? Das lässt sich nicht rückgängig machen.`,
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      isLoading.value = true
      deleteScenario(deleted.id)
        .then(() => {
          toast.add({
            summary: 'Erfolg',
            detail: `Szenario "${deleted.name}" wurde gelöscht`,
            severity: 'success',
            life: Config.TOAST_LIFE_TIME,
          })
          close({ reload: true })
        })
        .catch(showError)
        .finally(() => {
          isLoading.value = false
        })
    },
    reject: () => {
    },
  })
}
</script>
