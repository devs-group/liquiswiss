<template>
  <form
    id="scenario-form"
    class="grid grid-cols-2 gap-2"
    @submit.prevent
  >
    <div class="flex flex-col gap-2 col-span-full">
      <label
        class="text-sm font-bold"
        for="name"
      >Name *</label>
      <InputText
        v-bind="nameProps"
        id="name"
        v-model="name"
        :class="{ 'p-invalid': errors['name']?.length }"
        type="text"
      />
      <small class="text-liqui-red">{{ errors["name"] || '&nbsp;' }}</small>
    </div>

    <div
      v-if="isCreate"
      class="flex flex-col gap-2 col-span-full"
    >
      <label
        class="text-sm font-bold"
        for="type"
      >Typ *</label>
      <Select
        v-bind="typeProps"
        id="type"
        v-model="type"
        :options="scenarioTypeOptions"
        option-label="label"
        option-value="value"
        placeholder="Bitte wählen"
        :class="{ 'p-invalid': errors['type']?.length }"
      />
      <small class="text-liqui-red">{{ errors["type"] || '&nbsp;' }}</small>
    </div>

    <div
      v-if="isCreate && type === 'vertical'"
      class="flex flex-col gap-2 col-span-full"
    >
      <label
        class="text-sm font-bold"
        for="parentScenarioId"
      >Eltern-Szenario *</label>
      <Select
        v-bind="parentScenarioIdProps"
        id="parentScenarioId"
        v-model="parentScenarioId"
        :options="availableParentScenarios"
        option-label="name"
        option-value="id"
        placeholder="Bitte wählen"
        :class="{ 'p-invalid': errors['parentScenarioId']?.length }"
      />
      <small class="text-liqui-red">{{ errors["parentScenarioId"] || '&nbsp;' }}</small>
    </div>

    <div
      v-if="isCreate && type === 'horizontal'"
      class="flex flex-col gap-2 col-span-full"
    >
      <label
        class="text-sm font-bold"
        for="sourceScenarioId"
      >Daten kopieren von (optional)</label>
      <Select
        v-bind="sourceScenarioIdProps"
        id="sourceScenarioId"
        v-model="sourceScenarioId"
        :options="availableSourceScenarios"
        option-label="name"
        option-value="id"
        placeholder="Leeres Szenario erstellen"
        show-clear
      />
      <small class="text-gray-500">Wenn leer, wird ein leeres Szenario erstellt</small>
    </div>

    <hr class="my-4 col-span-full">

    <Message
      v-if="errorMessage.length"
      severity="error"
      :closable="false"
      class="col-span-full"
    >
      {{ errorMessage }}
    </Message>

    <div class="flex items-center justify-end gap-2 col-span-full">
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
        severity="contrast"
        @click="dialogRef?.close()"
      />
      <Button
        v-if="!isCreate && !scenario?.isDefault"
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
import { Config } from '~/config/config'
import type { ScenarioFormData, ScenarioUpdateFormData } from '~/models/scenario'

const dialogRef = inject<IScenarioFormDialog>('dialogRef')!

const { createScenario, updateScenario, deleteScenario, scenarios } = useScenarios()
const confirm = useConfirm()
const toast = useToast()

// Data
const isLoading = ref(false)
const scenario = dialogRef.value.data?.scenario
const isCreate = !scenario?.id
const errorMessage = ref('')

const scenarioTypeOptions = [
  { label: 'Horizontal (Unabhängige Kopie)', value: 'horizontal' },
  { label: 'Vertikal (Vererbt von Eltern)', value: 'vertical' },
]

const availableParentScenarios = computed(() => {
  return scenarios.value.filter(s => s.type === 'horizontal')
})

const availableSourceScenarios = computed(() => {
  return scenarios.value
})

const { defineField, errors, handleSubmit, meta, values } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
    type: yup.string().when('$isCreate', {
      is: true,
      then: (schema) => schema.required('Typ wird benötigt'),
      otherwise: (schema) => schema.notRequired(),
    }),
    parentScenarioId: yup.number().when(['type', '$isCreate'], {
      is: (type: string, isCreate: boolean) => type === 'vertical' && isCreate,
      then: (schema) => schema.required('Eltern-Szenario wird benötigt').typeError('Eltern-Szenario wird benötigt'),
      otherwise: (schema) => schema.notRequired(),
    }),
    sourceScenarioId: yup.number().notRequired(),
  }),
  initialValues: {
    name: scenario?.name ?? '',
    type: scenario?.type ?? 'horizontal',
    parentScenarioId: scenario?.parentScenarioId ?? undefined,
    sourceScenarioId: undefined,
  } as ScenarioFormData & { sourceScenarioId?: number },
  context: {
    isCreate,
  },
})

const [name, nameProps] = defineField('name')
const [type, typeProps] = defineField('type')
const [parentScenarioId, parentScenarioIdProps] = defineField('parentScenarioId')
const [sourceScenarioId, sourceScenarioIdProps] = defineField('sourceScenarioId')

const onSubmit = handleSubmit((values) => {
  isLoading.value = true
  errorMessage.value = ''

  if (isCreate) {
    const payload: ScenarioFormData = {
      name: values.name,
      type: values.type as 'horizontal' | 'vertical',
    }

    // For vertical scenarios, set parent
    if (payload.type === 'vertical' && values.parentScenarioId) {
      payload.parentScenarioId = values.parentScenarioId
    }
    // For horizontal scenarios, set parent only if copying from another scenario
    else if (payload.type === 'horizontal' && values.sourceScenarioId) {
      payload.parentScenarioId = values.sourceScenarioId
    }

    createScenario(payload)
      .then(() => {
        dialogRef.value.close()
        toast.add({
          summary: 'Erfolg',
          detail: `Szenario "${values.name}" wurde angelegt`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
      .catch(() => {
        errorMessage.value = 'Szenario konnte nicht angelegt werden'
        nextTick(() => {
          scrollToParentBottom('scenario-form')
        })
      })
      .finally(() => {
        isLoading.value = false
      })
  }
  else {
    const payload: ScenarioUpdateFormData = {
      name: values.name,
    }

    updateScenario(scenario!.id, payload)
      .then(() => {
        dialogRef.value.close()
        toast.add({
          summary: 'Erfolg',
          detail: `Szenario "${values.name}" wurde bearbeitet`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
      .catch(() => {
        errorMessage.value = 'Szenario konnte nicht bearbeitet werden'
        nextTick(() => {
          scrollToParentBottom('scenario-form')
        })
      })
      .finally(() => {
        isLoading.value = false
      })
  }
})

const onDeleteScenario = () => {
  confirm.require({
    header: 'Löschen',
    message: 'Szenario vollständig löschen?',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      if (scenario) {
        isLoading.value = true
        deleteScenario(scenario.id)
          .then(() => {
            toast.add({
              summary: 'Erfolg',
              detail: `Szenario "${scenario.name}" wurde gelöscht`,
              severity: 'success',
              life: Config.TOAST_LIFE_TIME,
            })
            dialogRef.value.close()
          })
          .catch(() => {
            errorMessage.value = 'Szenario konnte nicht gelöscht werden'
            nextTick(() => {
              scrollToParentBottom('scenario-form')
            })
          })
          .finally(() => {
            isLoading.value = false
          })
      }
    },
    reject: () => {
    },
  })
}
</script>
