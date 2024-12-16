<template>
  <div class="grid grid-cols-bank-accounts items-center w-full *:bg-zinc-100 *:dark:bg-zinc-800 *:border *:border-r-0 *:border-b-0 *:last:border-b *:border-gray-600 *:p-1 *:text-sm *:truncate">
    <div class="flex items-center gap-2 justify-end">
      <p class="truncate">{{bankAccount.name}}</p>
      <span @click="$emit('onClone', bankAccount)" class="pi pi-copy cursor-pointer text-help"></span>
      <span @click="$emit('onEdit', bankAccount)" class="pi pi-pencil cursor-pointer text-primary"></span>
    </div>
    <p class="flex flex-wrap gap-1 !border-r">
      <span class="font-bold" :class="{'text-red-500': !isRevenue, 'text-green-500': isRevenue}">{{amountFormatted}} {{bankAccount.currency.code}}</span>
    </p>
  </div>
</template>

<script setup lang="ts">
import type {BankAccountResponse} from "~/models/bank-account";

const props = defineProps({
  bankAccount: {
    type: Object as PropType<BankAccountResponse>,
    required: true,
  }
})

defineEmits<{
  'onEdit': [bankAccount: BankAccountResponse]
  'onClone': [bankAccount: BankAccountResponse]
}>()

const isRevenue = computed(() => props.bankAccount.amount >= 0)
const amountFormatted = computed(() => NumberToFormattedCurrency(AmountToFloat(props.bankAccount.amount), props.bankAccount.currency.localeCode))
</script>
