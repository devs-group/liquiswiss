<template>
  <Card>
    <template #title>
      <div class="flex items-center justify-between">
        <p class="truncate text-base">
          {{ SalaryCostUtils.title(salaryCost) }}
        </p>
        <div class="flex items-center gap-2 justify-end">
          <Button
            icon="pi pi-copy"
            severity="help"
            outlined
            rounded
            @click="$emit('onClone', salaryCost)"
          />
          <Button
            icon="pi pi-pencil"
            outlined
            rounded
            @click="$emit('onEdit', salaryCost)"
          />
          <Button
            severity="danger"
            icon="pi pi-trash"
            outlined
            rounded
            @click="$emit('onDelete', salaryCost)"
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
            {{ costCycle }}: {{ amountFormatted }}{{ unit }} ({{ calculatedAmountFormatted }} {{
              salary.currency.code
            }})
          </p>
        </div>
        <template v-if="!isInactive">
          <p>
            Gesamt: {{ nextCostFormatted }} {{ salary.currency.code }} {{ getNextCostExecutionDateHint }}
          </p>
        </template>
        <p v-else>
          Inaktiv
        </p>
        <div class="flex flex-col gap-1">
          <p>Zu Kosten von</p>
          <div class="flex flex-wrap items-center gap-1">
            <span
              class="flex-1 select-none truncate rounded-md px-2 py-1 text-[10px] sm:text-xs font-semibold text-center"
              :class="employerBadgeClasses"
            >
              {{ employerLabel }}
            </span>
            <span
              class="flex-1 select-none truncate rounded-md px-2 py-1 text-[10px] sm:text-xs font-semibold text-center"
              :class="employeeBadgeClasses"
            >
              {{ employeeLabel }}
            </span>
          </div>
        </div>
        <hr>
        <p class="text-xs">
          {{ getNextSalaryPaymentHint }}
        </p>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type { PropType } from 'vue'
import type { SalaryCostResponse, SalaryResponse } from '~/models/employee'
import { SalaryCostUtils } from '~/utils/models/salary-cost-utils'
import { DateStringToFormattedDate, DateStringToFormattedWordDate } from '~/utils/format-helper'
import { EmployeeCostDistributionType } from '~/config/enums'
import { EmployeeCostDistributionTypeToOptions } from '~/utils/enum-helper'

const props = defineProps({
  salaryCost: {
    type: Object as PropType<SalaryCostResponse>,
    required: true,
  },
  salary: {
    type: Object as PropType<SalaryResponse>,
    required: true,
  },
})

defineEmits<{
  onClone: [salaryCost: SalaryCostResponse]
  onEdit: [salaryCost: SalaryCostResponse]
  onDelete: [salaryCost: SalaryCostResponse]
}>()

const isFixed = computed(
  () => SalaryCostUtils.isFixed(props.salaryCost),
)
const isInactive = computed(
  () => SalaryCostUtils.isInactive(props.salaryCost),
)
const amountFormatted = computed(
  () => SalaryCostUtils.amountFormatted(props.salaryCost, props.salary.currency),
)
const calculatedAmountFormatted = computed(
  () => SalaryCostUtils.calculatedAmountFormatted(props.salaryCost, props.salary.currency),
)
const nextCostFormatted = computed(
  () => SalaryCostUtils.nextCostFormatted(props.salaryCost, props.salary.currency),
)
const amountType = computed(
  () => SalaryCostUtils.amountType(props.salaryCost),
)
const distributionOptions = EmployeeCostDistributionTypeToOptions()
const employerLabel
  = distributionOptions.find(option => option.value === EmployeeCostDistributionType.Employer)?.name
    ?? 'Arbeitgeber (AG)'
const employeeLabel
  = distributionOptions.find(option => option.value === EmployeeCostDistributionType.Employee)?.name
    ?? 'Arbeitnehmer (AN)'
const activeDistributionClasses = 'bg-liqui-green text-green-900 dark:text-green-100'
const inactiveDistributionClasses = 'bg-zinc-200 text-zinc-600 dark:bg-zinc-700 dark:text-zinc-300 opacity-50'
const distributionTypeValue = computed(() => props.salaryCost.distributionType)
const isBothDistribution = computed(() => distributionTypeValue.value === EmployeeCostDistributionType.Both)
const employerBadgeClasses = computed(() => {
  if (isBothDistribution.value) {
    return activeDistributionClasses
  }
  return distributionTypeValue.value === EmployeeCostDistributionType.Employer
    ? activeDistributionClasses
    : inactiveDistributionClasses
})
const employeeBadgeClasses = computed(() => {
  if (isBothDistribution.value) {
    return activeDistributionClasses
  }
  return distributionTypeValue.value === EmployeeCostDistributionType.Employee
    ? activeDistributionClasses
    : inactiveDistributionClasses
})
const costCycle = computed(
  () => SalaryCostUtils.costCycle(props.salaryCost),
)
const unit = computed(
  () => SalaryCostUtils.unit(props.salaryCost, props.salary.currency),
)
const getNextCostExecutionDateHint = computed(() => {
  return isInactive.value ? '-' : `im ${DateStringToFormattedWordDate(props.salaryCost?.calculatedNextExecutionDate)}`
})
const getNextSalaryPaymentHint = computed(() => {
  if (props.salary.nextExecutionDate) {
    return `Lohnzahlung am: ${DateStringToFormattedDate(props.salary.nextExecutionDate)}`
  }
  return 'Keine weitere Lohnzahlung'
})
</script>
