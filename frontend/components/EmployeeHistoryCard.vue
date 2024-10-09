<template>
  <Card>
    <template #title>
      <div class="flex items-center justify-between">
        <p class="truncate text-base">Von {{fromDateFormatted}}</p>
        <div class="flex gap-2 justify-end">
          <Button @click="$emit('onClone', employeeHistory)" severity="help" icon="pi pi-copy" outlined rounded />
          <Button @click="$emit('onEdit', employeeHistory)" icon="pi pi-pencil" outlined rounded />
        </div>
      </div>
    </template>
    <template #content>
      <div class="flex flex-col gap-2 text-sm">
        <p>{{employeeHistory.hoursPerMonth}} max. Stunden pro Monat</p>
        <p>{{salaryFormatted}} {{employeeHistory.salaryCurrency.code}} pro Monat</p>
        <p>{{employeeHistory.vacationDaysPerYear}} Urlaubstage im Jahr</p>
        <p v-if="employeeHistory.toDate">Bis {{toDateFormatted}}</p>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type {EmployeeHistoryResponse} from "~/models/employee";

const props = defineProps({
  employeeHistory: {
    type: Object as PropType<EmployeeHistoryResponse>,
    required: true,
  }
})

defineEmits<{
  'onEdit': [employeeHistory: EmployeeHistoryResponse]
  'onClone': [employeeHistory: EmployeeHistoryResponse]
}>()

const salaryFormatted = computed(() => NumberToFormattedCurrency(AmountToFloat(props.employeeHistory!.salaryPerMonth), props.employeeHistory.salaryCurrency!.localeCode))
const fromDateFormatted = computed(() => DateStringToFormattedDate(props.employeeHistory.fromDate))
const toDateFormatted = computed(() => props.employeeHistory.toDate ? DateStringToFormattedDate(props.employeeHistory.toDate) : '')
</script>
