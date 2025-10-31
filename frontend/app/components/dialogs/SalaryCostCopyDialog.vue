<template>
  <div
    class="flex flex-col gap-2"
  >
    <Message
      v-if="salaryCostsErrorMessage.length"
      severity="error"
      size="small"
      class="col-span-full"
      closable
    >
      {{ salaryCostsErrorMessage }}
    </Message>

    <div
      v-if="isLoadingSalaries"
      class="flex justify-center items-center"
    >
      <ProgressSpinner />
    </div>
    <div
      v-else
      class="flex flex-col overflow-x-auto pb-2"
    >
      <SalaryCostCopyHeaders v-model="selectAll" />
      <SalaryCostCopyRow
        v-for="cost of salaryCosts"
        :key="cost.id"
        :salary-cost="cost"
        :currency="salary.currency"
        :select-all="selectAll"
        @on-selection="onSalaryCostSelection"
      />
    </div>

    <div class="flex flex-col gap-2">
      <p>Kopiere zu Lohn:</p>
      <Select
        v-model="selectedSalaryID"
        :loading="isLoadingSalaries"
        :options="filteredSalaries"
        placeholder="Bitte wählen"
        :option-label="(data) => SalaryUtils.title(data)"
        option-value="id"
      />
    </div>

    <hr class="my-4 col-span-full">

    <div class="flex justify-end gap-2 col-span-full">
      <Button
        label="Kopieren"
        :disabled="!canCopy || isCopying"
        @click="onCopy"
      />
      <Button
        :disabled="isCopying"
        label="Schliessen"
        severity="contrast"
        @click="dialogRef.close()"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SalaryCostResponse } from '~/models/employee'
import type { ISalaryCostCopyDialog } from '~/interfaces/dialog-interfaces'
import { SalaryUtils } from '~/utils/models/salary-utils'
import { Config } from '~/config/config'
import SalaryCostCopyHeaders from '~/components/SalaryCostCopyHeaders.vue'

const dialogRef = inject<ISalaryCostCopyDialog>('dialogRef')!

const { salaries, listSalaries } = useSalaries()
const { listSalaryCosts, copySalaryCost } = useSalaryCosts()
const toast = useToast()

// Data
const isCopying = ref(false)
const salary = ref(dialogRef.value.data!.salary)
const isLoadingSalaries = ref(false)
const isLoadingSalaryCosts = ref(false)
const salariesErrorMessage = ref('')
const salaryCostsErrorMessage = ref('')
const salaryCosts = ref<SalaryCostResponse[]>([])
const selectedSalaryCosts = ref<SalaryCostResponse[]>([])
// Select all by default
const selectAll = ref(true)
const selectedSalaryID = ref<number>()

const filteredSalaries = computed(() => {
  return salaries.value.data.filter(s => s.id !== salary.value.id)
})
const canCopy = computed(() => {
  return selectedSalaryCosts.value.length > 0 && !!selectedSalaryID.value
})

const onListSalaryCosts = () => {
  if (salary.value) {
    salaryCostsErrorMessage.value = ''
    isLoadingSalaryCosts.value = true
    listSalaryCosts(salary.value.id)
      .then((resp) => {
        salaryCosts.value = resp.data
        selectedSalaryCosts.value = resp.data
      })
      .catch(() => {
        salaryCostsErrorMessage.value = 'Es gab einen Fehler beim Laden der Lohnkosten'
      })
      .finally(() => {
        isLoadingSalaryCosts.value = false
      })
  }
}

onMounted(() => {
  onListSalaryCosts()
  isLoadingSalaries.value = true
  listSalaries(salary.value.employeeID)
    .catch(() => {
      salariesErrorMessage.value = 'Es gab einen Fehler beim Laden des Lohnverlaufs'
    })
    .finally(() => {
      isLoadingSalaries.value = false
    })
})

const onCopy = () => {
  isCopying.value = true
  const idsToCopy = selectedSalaryCosts.value.map(hc => hc.id)
  if (!idsToCopy.length || !selectedSalaryID.value) {
    return
  }
  copySalaryCost(selectedSalaryID.value, {
    ids: idsToCopy,
  })
    .then(() => {
      dialogRef.value.close(true)
      toast.add({
        summary: 'Erfolg',
        detail: `Lohnkosten wurden erfolgreich kopiert`,
        severity: 'success',
        life: Config.TOAST_LIFE_TIME,
      })
    })
    .catch(() => {
      toast.add({
        summary: 'Fehler',
        detail: `Konnte Lohnkosten nicht kopieren`,
        severity: 'error',
        life: Config.TOAST_LIFE_TIME,
      })
    })
    .finally(() => {
      isCopying.value = false
    })
}

const onSalaryCostSelection = (salaryCost: SalaryCostResponse, isSelected: boolean) => {
  if (isSelected) {
    const exists = selectedSalaryCosts.value.find(hc => hc.id == salaryCost.id)
    if (!exists) {
      selectedSalaryCosts.value.push(salaryCost)
    }
  }
  else {
    selectedSalaryCosts.value = selectedSalaryCosts.value?.filter(hc => hc.id != salaryCost.id)
  }
}
</script>
