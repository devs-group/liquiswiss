<template>
  <div class="grid grid-cols-employees items-center w-full *:bg-zinc-100 *:dark:bg-zinc-800 *:border *:border-r-0 *:border-b-0 *:last:border-b *:border-gray-600 *:p-1 *:text-sm *:truncate">
    <div class="flex items-center gap-2 justify-end">
      <p class="truncate">
        {{ employee.name }}
      </p>
      <span
        class="pi pi-pencil cursor-pointer text-primary"
        @click="$emit('onEdit', employee)"
      />
    </div>
    <p>{{ employee.hoursPerMonth || '-' }}</p>
    <p v-if="hasHistoryData">
      {{ salaryFormatted || '-' }} {{ employee.currency.code }} / {{ cycle }}
    </p>
    <p v-else>
      -
    </p>
    <p>
      {{ employee.vacationDaysPerYear || '-' }}
    </p>
    <p>{{ fromDate }}</p>
    <p class="!border-r">
      {{ toDate }}
    </p>
  </div>
</template>

<script setup lang="ts">
import type { EmployeeResponse } from '~/models/employee'
import { EmployeeUtils } from '~/utils/models/employee-utils'

const props = defineProps({
  employee: {
    type: Object as PropType<EmployeeResponse>,
    required: true,
  },
})

defineEmits<{
  onEdit: [transaction: EmployeeResponse]
  onClone: [transaction: EmployeeResponse]
}>()

const salaryFormatted = computed(
  () => EmployeeUtils.salaryFormatted(props.employee),
)
const fromDate = computed(
  () => EmployeeUtils.fromDate(props.employee),
)
const toDate = computed(
  () => EmployeeUtils.toDate(props.employee),
)
const cycle = computed(
  () => EmployeeUtils.cycle(props.employee),
)
const hasHistoryData = computed(
  () => EmployeeUtils.hasHistoryData(props.employee),
)
</script>
