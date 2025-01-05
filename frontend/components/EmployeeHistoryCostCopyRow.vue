<template>
  <div class="grid grid-cols-employee-history-costs-copy items-center w-full *:bg-zinc-100 *:dark:bg-zinc-800 *:border *:border-r-0 *:border-b-0 *:last:border-b *:border-gray-600 *:p-1 *:text-sm *:truncate">
    <div class="flex justify-center">
      <Checkbox
        v-model="isSelected"
        binary
      />
    </div>
    <div class="flex items-center gap-2 justify-end">
      <p class="truncate">
        {{ EmployeeHistoryCostUtils.title(employeeHistoryCost) }}
      </p>
    </div>
    <p>{{ amountType }}</p>
    <div>
      <p v-if="isFixed">
        {{ costCycle }}: {{ amountFormatted }} {{ unit }}
      </p>
      <p v-else>
        {{ costCycle }}: {{ amountFormatted }}{{ unit }}
      </p>
    </div>
    <p class="!border-r">
      <strong>{{ distributionType }}</strong>
    </p>
  </div>
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
  selectAll: {
    type: Boolean,
    default: true,
    required: true,
  },
})

const isSelected = ref(props.selectAll)

watch(() => props.selectAll, (value) => {
  isSelected.value = value
})
watch(isSelected, (value) => {
  emits('onSelection', props.employeeHistoryCost, value)
})

const emits = defineEmits<{
  onSelection: [employeeHistoryCost: EmployeeHistoryCostResponse, isSelected: boolean]
}>()

const isFixed = computed(
  () => EmployeeHistoryCostUtils.isFixed(props.employeeHistoryCost),
)
const amountFormatted = computed(
  () => EmployeeHistoryCostUtils.amountFormatted(props.employeeHistoryCost, props.currency),
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
  () => EmployeeHistoryCostUtils.unit(props.employeeHistoryCost, props.currency),
)
</script>
