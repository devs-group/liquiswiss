<template>
  <div class="grid grid-cols-transactions items-center w-full *:bg-gray-100 *:border *:border-r-0 *:border-b-0 *:last:border-b *:border-gray-600 *:p-1 *:text-sm *:truncate">
    <div class="flex items-center gap-2 justify-end">
      <p class="truncate">{{transaction.name}}</p>
      <span @click="$emit('onClone', transaction)" class="pi pi-copy cursor-pointer text-help"></span>
      <span @click="$emit('onEdit', transaction)" class="pi pi-pencil cursor-pointer text-primary"></span>
    </div>
    <p>{{ startDate }}</p>
    <p>{{ endDate || '-' }}</p>
    <p>{{ nextExecutionDate || '-' }}</p>
    <p class="flex flex-wrap gap-1">
      <span class="font-bold" :class="{'text-red-500': !isRevenue, 'text-green-500': isRevenue}">{{amountFormatted}} {{transaction.currency.code}}</span>
    </p>
    <p v-if="isRepeating">{{cycle}}</p>
    <p v-else>Einmalig</p>
    <p>{{transaction.category.name}}</p>
    <p class="!border-r">{{transaction.employee?.name ?? '-'}}</p>
  </div>
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
  'onClone': [transaction: TransactionResponse]
}>()

const isRevenue = computed(() => props.transaction.amount >= 0)
const startDate = computed(() => DateStringToFormattedDate(props.transaction.startDate))
const endDate = computed(() => props.transaction.endDate ? DateStringToFormattedDate(props.transaction.endDate) : '')
const nextExecutionDate = computed(() => props.transaction.nextExecutionDate ? DateStringToFormattedDate(props.transaction.nextExecutionDate) : '')
const amountFormatted = computed(() => NumberToFormattedCurrency(AmountToFloat(props.transaction.amount), props.transaction.currency.localeCode))
const isRepeating = computed(() => props.transaction.type === TransactionType.Repeating)
const cycle = computed(() => CycleTypeToOptions().find((ct) => ct.value === props.transaction.cycle)?.name ?? '')
</script>
