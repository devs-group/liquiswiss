<template>
  <Card>
    <template #title>
      <div class="flex items-center justify-between">
        <p class="truncate text-base">
          {{ EmployeeHistoryCostUtils.title(employeeHistoryCost) }}
        </p>
        <div class="flex gap-2 justify-end">
          <Button
            icon="pi pi-copy"
            severity="help"
            outlined
            rounded
            @click="$emit('onClone', employeeHistoryCost)"
          />
          <Button
            icon="pi pi-pencil"
            outlined
            rounded
            @click="$emit('onEdit', employeeHistoryCost)"
          />
          <Button
            severity="danger"
            icon="pi pi-trash"
            outlined
            rounded
            @click="$emit('onDelete', employeeHistoryCost)"
          />
        </div>
      </div>
    </template>
    <template #content>
      <div
        class="flex flex-col gap-2 text-sm"
        :class="{ 'opacity-50 cursor-default': isInactive }"
      >
        <p>{{ amountType }}</p>
        <div>
          <p v-if="isFixed">
            {{ costCycle }}: {{ amountFormatted }} {{ unit }}
          </p>
          <p v-else>
            {{ costCycle }}: {{ amountFormatted }}{{ unit }} ({{ calculatedAmountFormatted }} {{ employeeHistory.currency.code }})
          </p>
        </div>
        <template v-if="!isInactive">
          <p v-if="isOnce">
            Gesamt: {{ nextCostFormatted }} {{ employeeHistory.currency.code }} {{ getNextCostExecutionDateHint }}
          </p>
        </template>
        <p v-else>
          Inaktiv
        </p>
        <p>Zu Kosten von <strong>{{ distributionType }}</strong></p>
        <hr>
        <p class="text-xs">
          {{ getNextHistoryPaymentHint }}
        </p>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type { PropType } from 'vue'
import type { EmployeeHistoryCostResponse, EmployeeHistoryResponse } from '~/models/employee'
import { EmployeeHistoryCostUtils } from '~/utils/models/employee-history-cost-utils'

const props = defineProps({
  employeeHistoryCost: {
    type: Object as PropType<EmployeeHistoryCostResponse>,
    required: true,
  },
  employeeHistory: {
    type: Object as PropType<EmployeeHistoryResponse>,
    required: true,
  },
})

defineEmits<{
  onClone: [employeeHistoryCost: EmployeeHistoryCostResponse]
  onEdit: [employeeHistoryCost: EmployeeHistoryCostResponse]
  onDelete: [employeeHistoryCost: EmployeeHistoryCostResponse]
}>()

const isFixed = computed(
  () => EmployeeHistoryCostUtils.isFixed(props.employeeHistoryCost),
)
const isOnce = computed(
  () => EmployeeHistoryCostUtils.isOnce(props.employeeHistoryCost),
)
const isInactive = computed(
  () => EmployeeHistoryCostUtils.isInactive(props.employeeHistoryCost),
)
const amountFormatted = computed(
  () => EmployeeHistoryCostUtils.amountFormatted(props.employeeHistoryCost, props.employeeHistory.currency),
)
const calculatedAmountFormatted = computed(
  () => EmployeeHistoryCostUtils.calculatedAmountFormatted(props.employeeHistoryCost, props.employeeHistory.currency),
)
const nextCostFormatted = computed(
  () => EmployeeHistoryCostUtils.nextCostFormatted(props.employeeHistoryCost, props.employeeHistory.currency),
)
const amountType = computed(
  () => EmployeeHistoryCostUtils.amountType(props.employeeHistoryCost),
)
const distributionType = computed(
  () => EmployeeHistoryCostUtils.distributionType(props.employeeHistoryCost),
)
const costCycle = computed(
  () => EmployeeHistoryCostUtils.costCycle(props.employeeHistoryCost),
)
const unit = computed(
  () => EmployeeHistoryCostUtils.unit(props.employeeHistoryCost, props.employeeHistory.currency),
)
const getNextCostExecutionDateHint = computed(() => {
  return isInactive.value ? '-' : `am ${DateStringToFormattedDate(props.employeeHistoryCost.nextExecutionDate)}`
})
const getNextHistoryPaymentHint = computed(() => {
  if (props.employeeHistory.nextExecutionDate) {
    return `Lohnzahlung am: ${DateStringToFormattedDate(props.employeeHistory.nextExecutionDate)}`
  }
  return 'Keine weitere Lohnzahlung'
})
</script>
