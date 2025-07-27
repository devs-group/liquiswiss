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
          v-if="employee.isTerminated"
          class="text-orange-500 font-bold"
        >
          Arbeitsverhältniss aufgelöst
        </p>
        <p
          v-else-if="!employee.isInFuture"
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
        <template v-if="!employee.isTerminated">
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
        </template>
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

const salaryFormatted = computed(() => NumberToFormattedCurrency(AmountToFloat(props.employee.salaryAmount!), props.employee.currency!.localeCode))
const cycle = computed(() => SalaryCycleTypeToOptions().find(ct => ct.value === props.employee.cycle)?.name ?? '')
</script>
