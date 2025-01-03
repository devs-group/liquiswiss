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
            {{ cycle }}: {{ amountFormatted }} {{ unit }}
          </p>
          <p v-else>
            {{ cycle }}: {{ amountFormatted }}{{ unit }} ({{ calculatedAmountFormatted }} {{ currency.code }})
          </p>
        </div>
        <p v-if="isOnce">
          Betrag: {{ nextCostFormatted }} {{ currency.code }}
        </p>
        <p v-else>
          Betrag: {{ nextCostFormatted }} {{ currency.code }} {{ getAmountOffset }}
        </p>
        <p>Nächste Belastung: {{ isInactive ? '-' : DateStringToFormattedDate(employeeHistoryCost.nextExecutionDate) }} </p>
        <p>Zu Kosten von <strong>{{ distributionType }}</strong></p>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type { PropType } from 'vue'
import type { EmployeeHistoryCostResponse } from '~/models/employee'
import type { CurrencyResponse } from '~/models/currency'
import { EmployeeHistoryCostUtils } from '~/utils/models/employee-history-cost-utils'

const props = defineProps({
  employeeHistoryCost: {
    type: Object as PropType<EmployeeHistoryCostResponse>,
    required: true,
  },
  currency: {
    type: Object as PropType<CurrencyResponse>,
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
  () => EmployeeHistoryCostUtils.amountFormatted(props.employeeHistoryCost, props.currency),
)
const calculatedAmountFormatted = computed(
  () => EmployeeHistoryCostUtils.calculatedAmountFormatted(props.employeeHistoryCost, props.currency),
)
const nextCostFormatted = computed(
  () => EmployeeHistoryCostUtils.nextCostFormatted(props.employeeHistoryCost, props.currency),
)
const amountType = computed(
  () => EmployeeHistoryCostUtils.amountType(props.employeeHistoryCost),
)
const distributionType = computed(
  () => EmployeeHistoryCostUtils.distributionType(props.employeeHistoryCost),
)
const cycle = computed(
  () => EmployeeHistoryCostUtils.cycle(props.employeeHistoryCost),
)
const unit = computed(
  () => EmployeeHistoryCostUtils.unit(props.employeeHistoryCost, props.currency),
)
const getAmountOffset = computed(
  () => EmployeeHistoryCostUtils.getAmountOffset(props.employeeHistoryCost),
)
</script>
