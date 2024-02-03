<template>
  <Card>
    <template #title>
      <div class="flex items-center justify-between">
        <p class="truncate text-base">{{employee.name}}</p>
        <div class="flex justify-end">
          <Button @click="$emit('onEdit', employee)" icon="pi pi-pencil" outlined rounded />
        </div>
      </div>
    </template>
    <template #content>
      <div class="flex flex-col gap-2 text-sm">
        <p>
          {{employee.hoursPerMonth}} max. Stunden pro Monat
        </p>
        <p>
          {{employee.vacationDaysPerYear}} Urlaubstage im Jahr
        </p>
        <p v-if="employee.entryDate">
          Zugehörig seit {{DateStringToFormattedDate(employee.entryDate)}}
        </p>
        <p v-if="employee.exitDate">
          Austritt am {{DateStringToFormattedDate(employee.exitDate)}}
        </p>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import {DateStringToFormattedDate} from "~/utils/format-helper";
import type {EmployeeResponse} from "~/models/employee";

defineProps({
  employee: {
    type: Object as PropType<EmployeeResponse>,
    required: true,
  }
})

defineEmits<{
  'onEdit': [employee: EmployeeResponse]
}>()
</script>
