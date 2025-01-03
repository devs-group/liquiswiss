<template>
  <Card>
    <template #title>
      <div class="relative flex items-center justify-between">
        <p class="truncate text-base">
          Von {{ fromDateFormatted }}
        </p>
        <div class="flex gap-2 justify-end">
          <Button
            severity="help"
            icon="pi pi-copy"
            outlined
            rounded
            @click="$emit('onClone', employeeHistory)"
          />
          <Button
            icon="pi pi-pencil"
            outlined
            rounded
            @click="$emit('onEdit', employeeHistory)"
          />
        </div>
        <p
          v-if="isActive"
          class="absolute -top-9 left-0 whitespace-nowrap text-sm bg-liqui-green p-2 rounded-xl font-bold text-center"
        >
          Aktive Historie
        </p>
      </div>
    </template>
    <template #content>
      <div class="flex flex-col gap-2 text-sm">
        <p>{{ employeeHistory.hoursPerMonth }} Arbeitsstunden / Monat</p>
        <div class="flex flex-col">
          <p v-if="withSeparateCosts">
            {{ totalSalaryCostFormatted }} {{ employeeHistory.currency.code }} Gesamtkosten
          </p>
          <p v-else>
            {{ grossSalaryFormatted }} {{ employeeHistory.currency.code }} / {{ cycle }}
          </p>
          <div
            v-if="withSeparateCosts"
            class="flex flex-col text-xs"
          >
            <p>Brutto: {{ grossSalaryFormatted }} {{ employeeHistory.currency.code }} / {{ cycle }}</p>
            <p>Netto: {{ netSalaryFormatted }} {{ employeeHistory.currency.code }} / {{ cycle }}</p>
          </div>
        </div>
        <p>{{ employeeHistory.vacationDaysPerYear }} Urlaubstage im Jahr</p>
        <p
          v-if="employeeHistory.toDate"
          class="text-orange-500"
        >
          Bis {{ toDateFormatted }}
        </p>
        <p
          v-else
          class="text-orange-500"
        >
          Dauerhaft
        </p>

        <div class="flex items-center gap-2">
          <label
            class="text-sm font-bold"
            for="with-separate-costs"
          >Lohnkosten separat erfassen</label>
          <div class="flex items-center">
            <ToggleSwitch
              id="with-separate-costs"
              v-model="withSeparateCosts"
            />
          </div>
        </div>
        <div
          v-if="withSeparateCosts"
          class="flex items-center gap-2"
        >
          <Button
            v-if="hasCosts"
            v-tooltip.top="'Lohnkosten in andere Historie kopieren'"
            icon="pi pi-copy"
            severity="help"
            @click="onCopyAllCosts"
          />
          <Button
            icon="pi pi-pencil"
            @click="onShowCostOverview"
          />
        </div>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type { EmployeeHistoryResponse } from '~/models/employee'
import { Config } from '~/config/config'
import { ModalConfig } from '~/config/dialog-props'
import EmployeeHistoryCostOverviewDialog from '~/components/dialogs/EmployeeHistoryCostOverviewDialog.vue'
import EmployeeHistoryCostCopyDialog from '~/components/dialogs/EmployeeHistoryCostCopyDialog.vue'
import { EmployeeHistoryUtils } from '~/utils/models/employee-history-utils'

const toast = useToast()
const dialog = useDialog()

const { updateEmployeeHistory, listEmployeeHistory } = useEmployeeHistories()

const props = defineProps({
  employeeHistory: {
    type: Object as PropType<EmployeeHistoryResponse>,
    required: true,
  },
  isActive: {
    type: Boolean,
    required: true,
  },
})

defineEmits<{
  onEdit: [employeeHistory: EmployeeHistoryResponse]
  onClone: [employeeHistory: EmployeeHistoryResponse]
}>()

const withSeparateCosts = ref(props.employeeHistory.withSeparateCosts)

watch(withSeparateCosts, (value) => {
  updateEmployeeHistory(props.employeeHistory.employeeID, {
    id: props.employeeHistory.id,
    cycle: props.employeeHistory.cycle,
    withSeparateCosts: value,
  })
    .then(() => {
      toast.add({
        summary: 'Erfolg',
        detail: `Änderung gespeichert`,
        severity: 'info',
        life: Config.TOAST_LIFE_TIME_SHORT,
      })
    })
})

const grossSalaryFormatted = computed(
  () => EmployeeHistoryUtils.grossSalaryFormatted(props.employeeHistory),
)
const netSalaryFormatted = computed(
  () => EmployeeHistoryUtils.netSalaryFormatted(props.employeeHistory),
)
const totalSalaryCostFormatted = computed(
  () => EmployeeHistoryUtils.totalSalaryCostFormatted(props.employeeHistory),
)
const fromDateFormatted = computed(
  () => EmployeeHistoryUtils.fromDateFormatted(props.employeeHistory),
)
const toDateFormatted = computed(
  () => EmployeeHistoryUtils.toDateFormatted(props.employeeHistory),
)
const cycle = computed(
  () => EmployeeHistoryUtils.cycle(props.employeeHistory),
)
const hasCosts = computed(
  () => EmployeeHistoryUtils.hasCosts(props.employeeHistory),
)

const onShowCostOverview = () => {
  dialog.open(EmployeeHistoryCostOverviewDialog, {
    props: {
      header: `Lohnkostenübersicht`,
      ...ModalConfig,
    },
    data: {
      employeeHistory: props.employeeHistory,
    },
    onClose: (options) => {
      if (options?.data) {
        listEmployeeHistory(props.employeeHistory.employeeID)
      }
    },
  })
}

const onCopyAllCosts = () => {
  dialog.open(EmployeeHistoryCostCopyDialog, {
    props: {
      header: `Lohnkosten kopieren`,
      ...ModalConfig,
    },
    data: {
      employeeHistory: props.employeeHistory,
    },
    onClose: (options) => {
      console.log(options?.data)
      if (options?.data) {
        listEmployeeHistory(props.employeeHistory.employeeID)
      }
    },
  })
}
</script>
