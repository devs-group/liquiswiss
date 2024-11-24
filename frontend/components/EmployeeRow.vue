<template>
  <div class="grid grid-cols-employees items-center w-full *:bg-gray-100 *:border *:border-r-0 *:border-b-0 *:last:border-b *:border-gray-600 *:p-1 *:text-sm *:truncate">
    <div class="flex items-center gap-2 justify-end">
      <p class="truncate">{{employee.name}}</p>
      <span @click="$emit('onEdit', employee)" class="pi pi-pencil cursor-pointer text-primary"></span>
    </div>
    <p>{{ employee.hoursPerMonth || '-' }}</p>
    <p>{{ salaryFormatted || '-' }}</p>
    <p>{{ employee.vacationDaysPerYear || '-' }}</p>
    <p>{{fromDate}}</p>
    <p class="!border-r">{{toDate}}</p>
  </div>
</template>

<script setup lang="ts">
import type {EmployeeResponse} from "~/models/employee";
import {DateStringToFormattedDate} from "~/utils/format-helper";

const props = defineProps({
  employee: {
    type: Object as PropType<EmployeeResponse>,
    required: true,
  }
})

defineEmits<{
  'onEdit': [transaction: EmployeeResponse]
  'onClone': [transaction: EmployeeResponse]
}>()

const salaryFormatted = computed(() => props.employee.salaryPerMonth ? NumberToFormattedCurrency(AmountToFloat(props.employee.salaryPerMonth ?? 0), props.employee.salaryCurrency!.localeCode) : '-')
const fromDate = computed(() => props.employee.fromDate ? DateStringToFormattedDate(props.employee.fromDate) : '-')
const toDate = computed(() => props.employee.toDate ? DateStringToFormattedDate(props.employee.toDate) : '-')
</script>
