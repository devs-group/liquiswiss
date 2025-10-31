<template>
  <div class="grid grid-cols-salary-costs-copy items-center w-full *:bg-zinc-100 *:dark:bg-zinc-800 *:border *:border-r-0 *:border-b-0 *:last:border-b *:border-gray-600 *:p-1 *:text-sm *:truncate">
    <div class="flex justify-center">
      <Checkbox
        v-model="isSelected"
        binary
      />
    </div>
    <div class="flex items-center gap-2 justify-end">
      <p class="truncate">
        {{ SalaryCostUtils.title(salaryCost) }}
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
import type { SalaryCostResponse } from '~/models/employee'
import type { CurrencyResponse } from '~/models/currency'
import { SalaryCostUtils } from '~/utils/models/salary-cost-utils'

const props = defineProps({
  salaryCost: {
    type: Object as PropType<SalaryCostResponse>,
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
  emits('onSelection', props.salaryCost, value)
})

const emits = defineEmits<{
  onSelection: [salaryCost: SalaryCostResponse, isSelected: boolean]
}>()

const isFixed = computed(
  () => SalaryCostUtils.isFixed(props.salaryCost),
)
const amountFormatted = computed(
  () => SalaryCostUtils.amountFormatted(props.salaryCost, props.currency),
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
  () => SalaryCostUtils.unit(props.salaryCost, props.currency),
)
</script>
