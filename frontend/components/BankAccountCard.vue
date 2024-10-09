<template>
  <Card>
    <template #title>
      <div class="flex items-center justify-between">
        <p class="truncate text-base">{{ bankAccount.name }}</p>
        <div class="flex justify-end">
          <Button @click="$emit('onEdit', bankAccount)" icon="pi pi-pencil" outlined rounded />
        </div>
      </div>
    </template>
    <template #content>
      <div class="flex flex-col text-sm">
        <p class="flex flex-wrap gap-1">
          Kontostand: <span class="font-bold"
                        :class="{'text-red-500': bankAccount.amount < 0, 'text-green-500': bankAccount.amount > 0}">
          {{ amountFormatted }} {{ bankAccount.currency.code }}
        </span>
        </p>
      </div>
    </template>
  </Card>
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
}>()

const amountFormatted = computed(() => NumberToFormattedCurrency(AmountToFloat(props.bankAccount.amount), props.bankAccount.currency.localeCode))
</script>
