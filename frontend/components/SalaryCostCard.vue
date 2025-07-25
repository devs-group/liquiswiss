<template>
  <Card>
    <template #title>
      <div class="flex items-center justify-between">
        <p class="truncate text-base">
          {{ SalaryCostUtils.title(salaryCost) }}
        </p>
        <div class="flex gap-2 justify-end">
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
        <p>Zu Kosten von <strong>{{ distributionType }}</strong></p>
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
import { DateStringToFormattedWordDate } from '~/utils/format-helper'

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
const distributionType = computed(
  () => SalaryCostUtils.distributionType(props.salaryCost),
)
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
