<template>
  <Card>
    <template #title>
      <div class="flex items-center justify-between">
        <p class="truncate text-base">
          Von {{ fromDateFormatted }}
        </p>
        <div class="flex gap-2 justify-end">
          <Button
            severity="help"
            icon="pi pi-copy"
            outlined
            rounded
            @click="$emit('onClone', employeeHistory)"
          />
          <Button
            icon="pi pi-pencil"
            outlined
            rounded
            @click="$emit('onEdit', employeeHistory)"
          />
        </div>
      </div>
    </template>
    <template #content>
      <div class="flex flex-col gap-2 text-sm">
        <p>{{ employeeHistory.hoursPerMonth }} Arbeitsstunden / Monat</p>
        <p>{{ salaryFormatted }} {{ employeeHistory.currency.code }} / {{ cycle }}</p>
        <p>{{ employeeHistory.vacationDaysPerYear }} Urlaubstage im Jahr</p>
        <p
          v-if="employeeHistory.toDate"
          class="text-orange-500"
        >
          Bis {{ toDateFormatted }}
        </p>
        <p
          v-if="isActive"
          class="bg-liqui-green p-2 font-bold text-center"
        >
          Aktive Historie
        </p>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type { EmployeeHistoryResponse } from '~/models/employee'

const props = defineProps({
  employeeHistory: {
    type: Object as PropType<EmployeeHistoryResponse>,
    required: true,
  },
  isActive: {
    type: Boolean,
    required: true,
  },
})

defineEmits<{
  onEdit: [employeeHistory: EmployeeHistoryResponse]
  onClone: [employeeHistory: EmployeeHistoryResponse]
}>()

const salaryFormatted = computed(() => NumberToFormattedCurrency(AmountToFloat(props.employeeHistory!.salary), props.employeeHistory.currency!.localeCode))
const fromDateFormatted = computed(() => DateStringToFormattedDate(props.employeeHistory.fromDate))
const toDateFormatted = computed(() => props.employeeHistory.toDate ? DateStringToFormattedDate(props.employeeHistory.toDate) : '')
const cycle = computed(() => CycleTypeToOptions().find(ct => ct.value === props.employeeHistory.cycle)?.name ?? '')
</script>
