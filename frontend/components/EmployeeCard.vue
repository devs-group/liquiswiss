<template>
  <Card>
    <template #title>
      <div class="flex items-center justify-between">
        <p class="truncate text-base">
          {{ employee.name }}
        </p>
        <div class="flex justify-end">
          <Button
            icon="pi pi-cog"
            outlined
            rounded
            @click="$emit('onEdit', employee)"
          />
        </div>
      </div>
    </template>
    <template #content>
      <div
        v-if="employee.hoursPerMonth !== null"
        class="flex flex-col gap-2 text-sm"
      >
        <p
          v-if="!employee.isInFuture"
          class="text-green-500 font-bold"
        >
          Aktuelle Daten:
        </p>
        <p
          v-else
          class="text-orange-500 font-bold"
        >
          Kommende Daten:
        </p>
        <p>{{ employee.hoursPerMonth }} Arbeitsstunden / Monat</p>
        <p>
          {{ salaryFormatted }} {{ employee.currency?.code }} / {{ cycle }}
        </p>
        <p>
          {{ employee.vacationDaysPerYear }} Urlaubstage / Jahr
        </p>
        <p v-if="employee.fromDate">
          <strong>Gültig {{ employee.isInFuture ? 'ab' : 'seit' }}</strong> {{ DateStringToFormattedDate(employee.fromDate) }}
        </p>
        <p v-if="employee.toDate">
          <strong>Gültig bis</strong> {{ DateStringToFormattedDate(employee.toDate) }}
        </p>
      </div>
      <div
        v-else
        class="flex flex-col gap-2 text-sm"
      >
        <p class="text-sm text-liqui-red">
          Der Mitarbeiter hat aktuell keine Daten
        </p>
      </div>
    </template>
  </Card>
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
  onEdit: [employee: EmployeeResponse]
}>()

const salaryFormatted = computed(() => NumberToFormattedCurrency(AmountToFloat(props.employee.salary!), props.employee.currency!.localeCode))
const cycle = computed(() => CycleTypeToOptions().find(ct => ct.value === props.employee.cycle)?.name ?? '')
</script>
