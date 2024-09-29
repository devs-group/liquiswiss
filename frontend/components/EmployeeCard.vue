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
      <div v-if="employee.hoursPerMonth !== null" class="flex flex-col gap-2 text-sm">
        <p class="font-bold">Aktuelle Daten:</p>
        <p>{{employee.hoursPerMonth}} Arbeitsstunden / Monat</p>
        <p v-if="employee.salaryPerMonth && employee.salaryCurrency">
          {{salaryFormatted}} {{employee.salaryCurrency.code}} / Monat
        </p>
        <p v-if="employee.vacationDaysPerYear">
          {{employee.vacationDaysPerYear}} Urlaubstage / Jahr
        </p>
        <p v-if="employee.fromDate">
          Gültig seit {{DateStringToFormattedDate(employee.fromDate)}}
        </p>
        <p v-if="employee.toDate">
          Gültig bis {{DateStringToFormattedDate(employee.toDate)}}
        </p>
      </div>
      <div v-else class="flex flex-col gap-2 text-sm">
        <p class="text-sm text-red-400">Der Mitarbeiter hat noch keine Daten</p>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type {EmployeeResponse} from "~/models/employee";

const props = defineProps({
  employee: {
    type: Object as PropType<EmployeeResponse>,
    required: true,
  }
})

defineEmits<{
  'onEdit': [employee: EmployeeResponse]
}>()

const salaryFormatted = computed(() => NumberToFormattedCurrency(AmountToFloat(props.employee!.salaryPerMonth), props.employee.salaryCurrency!.localeCode))
</script>
