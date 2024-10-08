<template>
  <Card>
    <template #title>
      <div class="flex items-center justify-between">
        <p class="truncate text-base">{{transaction.name}}</p>
        <div class="flex justify-end">
          <Button @click="$emit('onEdit', transaction)" icon="pi pi-pencil" outlined rounded />
        </div>
      </div>
    </template>
    <template #content>
      <div class="flex flex-col text-sm">
        <p>Start: {{ startDate }}</p>
        <p v-if="isRepeating && endDate">Ende: {{ endDate }}</p>
        <p class="flex flex-wrap gap-1">
          Betrag: <span class="font-bold" :class="{'text-red-500': !isRevenue, 'text-green-500': isRevenue}">{{amountFormatted}} {{transaction.currency.code}}</span>
        </p>
        <p v-if="isRepeating">Wiederkehrend: {{cycle}}</p>
        <p v-else>Einmalig</p>
        <p class="truncate">Kategorie: {{transaction.category.name}}</p>
        <p v-if="transaction.employee">Mitarbeiter: {{transaction.employee.name}}</p>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type {TransactionResponse} from "~/models/transaction";
import {TransactionType} from "~/config/enums";

const props = defineProps({
  transaction: {
    type: Object as PropType<TransactionResponse>,
    required: true,
  }
})

defineEmits<{
  'onEdit': [transaction: TransactionResponse]
}>()

const isRevenue = computed(() => props.transaction.amount >= 0)
const startDate = computed(() => DateStringToFormattedDate(props.transaction.startDate))
const endDate = computed(() => props.transaction.endDate ? DateStringToFormattedDate(props.transaction.endDate) : '')
const amountFormatted = computed(() => NumberToFormattedCurrency(AmountToFloat(props.transaction.amount), props.transaction.currency.localeCode))
const isRepeating = computed(() => props.transaction.type === TransactionType.Repeating)
const cycle = computed(() => CycleTypeToOptions().find((ct) => ct.value === props.transaction.cycle)?.name ?? '')
</script>
