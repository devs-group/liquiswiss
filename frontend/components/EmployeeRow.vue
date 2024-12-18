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
    <p>{{ salaryFormatted || '-' }}</p>
    <p>{{ employee.vacationDaysPerYear || '-' }}</p>
    <p>{{ fromDate }}</p>
    <p class="!border-r">
      {{ toDate }}
    </p>
  </div>
</template>

<script setup lang="ts">
import type { EmployeeResponse } from '~/models/employee'

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

const salaryFormatted = computed(() => props.employee.salaryPerMonth ? NumberToFormattedCurrency(AmountToFloat(props.employee.salaryPerMonth ?? 0), props.employee.currency!.localeCode) : '-')
const fromDate = computed(() => props.employee.fromDate ? DateStringToFormattedDate(props.employee.fromDate) : '-')
const toDate = computed(() => props.employee.toDate ? DateStringToFormattedDate(props.employee.toDate) : '-')
</script>
