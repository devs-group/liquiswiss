<template>
  <div
    class="flex flex-col gap-2"
  >
    <Message
      v-if="costsErrorMessage.length"
      severity="error"
      size="small"
      class="col-span-full"
      closable
    >
      {{ costsErrorMessage }}
    </Message>

    <div
      v-if="isLoadingHistories"
      class="flex justify-center items-center"
    >
      <ProgressSpinner />
    </div>
    <div
      v-else
      class="flex flex-col overflow-x-auto pb-2"
    >
      <EmployeeHistoryCostCopyHeaders v-model="selectAll" />
      <EmployeeHistoryCostCopyRow
        v-for="cost of costs"
        :key="cost.id"
        :employee-history-cost="cost"
        :currency="employeeHistory.currency"
        :select-all="selectAll"
        @on-selection="onCostSelection"
      />
    </div>

    <div class="flex flex-col gap-2">
      <p>Kopiere zu Historie:</p>
      <Select
        v-model="selectedEmployeeHistoryID"
        :loading="isLoadingHistories"
        :options="filteredHistories"
        placeholder="Bitte wählen"
        :option-label="(data) => EmployeeHistoryUtils.title(data)"
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
import type { EmployeeHistoryCostResponse } from '~/models/employee'
import type { IHistoryCostCopyDialog } from '~/interfaces/dialog-interfaces'
import { EmployeeHistoryUtils } from '~/utils/models/employee-history-utils'
import { Config } from '~/config/config'
import EmployeeHistoryCostCopyHeaders from '~/components/EmployeeHistoryCostCopyHeaders.vue'

const dialogRef = inject<IHistoryCostCopyDialog>('dialogRef')!

const { employeeHistories, listEmployeeHistory } = useEmployeeHistories()
const { listEmployeeHistoryCosts, copyEmployeeHistoryCost } = useEmployeeHistoryCosts()
const toast = useToast()

// Data
const isCopying = ref(false)
const employeeHistory = ref(dialogRef.value.data!.employeeHistory)
const isLoadingHistories = ref(false)
const isLoadingCosts = ref(false)
const historiesErrorMessage = ref('')
const costsErrorMessage = ref('')
const costs = ref<EmployeeHistoryCostResponse[]>([])
const selectedCosts = ref<EmployeeHistoryCostResponse[]>([])
// Select all by default
const selectAll = ref(true)
const selectedEmployeeHistoryID = ref<number>()

const filteredHistories = computed(() => {
  return employeeHistories.value.data.filter(h => h.id !== employeeHistory.value.id)
})
const canCopy = computed(() => {
  return selectedCosts.value.length > 0 && !!selectedEmployeeHistoryID.value
})

const onListEmployeeHistoryCosts = () => {
  if (employeeHistory.value) {
    costsErrorMessage.value = ''
    isLoadingCosts.value = true
    listEmployeeHistoryCosts(employeeHistory.value.id)
      .then((resp) => {
        costs.value = resp.data
        selectedCosts.value = resp.data
      })
      .catch(() => {
        costsErrorMessage.value = 'Es gab einen Fehler beim Laden der Lohnkosten'
      })
      .finally(() => {
        isLoadingCosts.value = false
      })
  }
}

onMounted(() => {
  onListEmployeeHistoryCosts()
  isLoadingHistories.value = true
  listEmployeeHistory(employeeHistory.value.employeeID)
    .catch(() => {
      historiesErrorMessage.value = 'Es gab einen Fehler beim Laden der Historien'
    })
    .finally(() => {
      isLoadingHistories.value = false
    })
})

const onCopy = () => {
  isCopying.value = true
  const idsToCopy = selectedCosts.value.map(hc => hc.id)
  if (!idsToCopy.length || !selectedEmployeeHistoryID.value) {
    return
  }
  copyEmployeeHistoryCost(selectedEmployeeHistoryID.value, {
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
}

const onCostSelection = (historyCost: EmployeeHistoryCostResponse, isSelected: boolean) => {
  if (isSelected) {
    const exists = selectedCosts.value.find(hc => hc.id == historyCost.id)
    if (!exists) {
      selectedCosts.value.push(historyCost)
    }
  }
  else {
    selectedCosts.value = selectedCosts.value?.filter(hc => hc.id != historyCost.id)
  }
}
</script>
