<template>
  <form
    id="scenario-form"
    class="grid grid-cols-2 gap-2"
    @submit.prevent
  >
    <div class="flex flex-col gap-2 col-span-full sm:col-span-1">
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

    <div class="flex flex-col gap-2 col-span-full sm:col-span-1">
      <label
        class="text-sm font-bold"
        for="name"
      >Szenariotyp *</label>
      <SelectButton
        v-bind="scenarioTypeProps"
        v-model="scenarioType"
        option-label="label"
        option-value="type"
        :options="options"
        :disabled="!isCreate"
      />
      <small class="text-liqui-red">{{ errors["scenarioType"] || '&nbsp;' }}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full sm:col-span-1">
      <label
        class="text-sm font-bold"
        for="name"
      >Hilfe</label>
      <Message
        severity="warn"
        size="small"
      >
        <template v-if="scenarioType === 'based_on'">
          Ein <strong>"basierend auf"</strong> Szenario basiert auf einem Elternszenario.
          Die Daten dieses Szenarios benutzen solange den Wert des Elternszenarios,
          bis diese überschrieben werden.
        </template>
        <template v-else-if="scenarioType === 'independent'">
          Ein <strong>"unabhängiges"</strong> Szenario basiert nicht auf einem Elternszenario.
          Es können jedoch Daten von einem anderen Szenario <strong>kopiert</strong> werden.
        </template>
      </Message>
    </div>

    <div class="flex flex-col gap-2 col-span-full sm:col-span-1">
      <label
        class="text-sm font-bold"
        for="name"
      >{{ scenarioType == 'based_on' ? 'Elternszenario wählen *' : 'Kopieren von Szenario (Optional)' }}</label>
      <Select
        v-bind="parentScenarioIDProps"
        id="name"
        v-model="parentScenarioID"
        empty-message="Keine Elternszenarien gefunden"
        filter
        auto-filter-focus
        show-clear
        empty-filter-message="Keine Resultate gefunden"
        :options="filteredScenarios"
        option-label="name"
        option-value="id"
        placeholder="Bitte wählen"
        :class="{ 'p-invalid': errors['parentScenarioID']?.length }"
        type="text"
        :disabled="!isCreate && scenarioType !== 'based_on'"
      />
      <small class="text-liqui-red">{{ errors["parentScenarioID"] || '&nbsp;' }}</small>
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
        :label="isClone ? 'Klonen' : 'Speichern'"
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
        v-if="!isCreate && !isDefaultScenario"
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
import { Config } from '~/config/config'
import type { IScenarioFormDialog } from '~/interfaces/dialog-interfaces'
import type { ScenarioFormData, ScenarioType } from '~/models/scenario'

const dialogRef = inject<IScenarioFormDialog>('dialogRef')!

const { scenarios, createScenario, updateScenario, deleteScenario } = useScenarios()
const confirm = useConfirm()
const toast = useToast()

const options = ref([{
  label: 'Basierend auf',
  type: 'based_on',
}, {
  label: 'Unabhängig',
  type: 'independent',
}])

// Data
const isLoading = ref(false)
const scenario = dialogRef.value.data?.scenario
const isClone = dialogRef.value.data?.isClone
const isCreate = isClone || !scenario?.id
const isDefaultScenario = scenario?.isDefault ?? false
const errorMessage = ref('')

const getDescendantIds = (parentId: number): Set<number> => {
  const descendants = new Set<number>()
  const children = scenarios.value.filter(s => s.parentScenarioID === parentId)
  for (const child of children) {
    descendants.add(child.id)
    for (const id of getDescendantIds(child.id)) {
      descendants.add(id)
    }
  }
  return descendants
}

const filteredScenarios = computed(() => {
  if (!scenario?.id) return scenarios.value
  const excludeIds = getDescendantIds(scenario.id)
  excludeIds.add(scenario.id)
  return scenarios.value.filter(s => !excludeIds.has(s.id))
})

const { defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
    scenarioType: yup.string<ScenarioType>().required('Szenariotyp wird benötigt'),
    parentScenarioID: yup.number().nullable().typeError('Ungültiges Elternszenario')
      .test('required-for-based-on', 'Elternszenario wird benötigt', (value) => {
        if (scenarioType.value === 'based_on') {
          return value != null
        }
        return true
      }),
  }),
  initialValues: {
    id: isClone ? undefined : scenario?.id ?? undefined,
    name: scenario?.name ?? '',
    scenarioType: scenario?.parentScenarioID ? 'based_on' : 'independent',
    parentScenarioID: scenario?.parentScenarioID ?? undefined,
  } as ScenarioFormData,
})

const [name, nameProps] = defineField('name')
const [scenarioType, scenarioTypeProps] = defineField('scenarioType')
const [parentScenarioID, parentScenarioIDProps] = defineField('parentScenarioID')

const onSubmit = handleSubmit((values) => {
  isLoading.value = true
  errorMessage.value = ''

  if (isCreate) {
    createScenario({
      ...values,
      scenarioType: scenarioType.value,
    })
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
    updateScenario({
      ...values,
      scenarioType: undefined,
    })
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
