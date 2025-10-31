<template>
  <Card>
    <template #title>
      <div class="flex items-center justify-between">
        <p class="truncate text-base">
          {{ bankAccount.name }}
        </p>
        <div class="flex gap-2 justify-end">
          <Button
            severity="help"
            icon="pi pi-copy"
            outlined
            rounded
            @click="$emit('onClone', bankAccount)"
          />
          <Button
            icon="pi pi-pencil"
            outlined
            rounded
            @click="$emit('onEdit', bankAccount)"
          />
        </div>
      </div>
    </template>
    <template #content>
      <div class="flex flex-col text-sm">
        <p class="flex flex-wrap gap-1">
          Kontostand: <span
            class="font-bold"
            :class="{ 'text-red-500': bankAccount.amount < 0, 'text-green-500': bankAccount.amount > 0 }"
          >
            {{ amountFormatted }} {{ bankAccount.currency.code }}
          </span>
        </p>
        <p
          v-if="bankAccount.currency.code != getOrganisationCurrencyCode"
          class="flex flex-wrap gap-1"
        >
          <span class="text-xs">
            ~ {{ amountToBaseFormatted }} {{ getOrganisationCurrencyCode }}
          </span>
        </p>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type { BankAccountResponse } from '~/models/bank-account'

const { getOrganisationCurrencyCode } = useAuth()
const { convertAmountToRate } = useGlobalData()

const props = defineProps({
  bankAccount: {
    type: Object as PropType<BankAccountResponse>,
    required: true,
  },
})

defineEmits<{
  onEdit: [bankAccount: BankAccountResponse]
  onClone: [bankAccount: BankAccountResponse]
}>()

const amountFormatted = computed(() => NumberToFormattedCurrency(AmountToFloat(props.bankAccount.amount), props.bankAccount.currency.localeCode))
const amountToBaseFormatted = computed(() => NumberToFormattedCurrency(
  convertAmountToRate(AmountToFloat(props.bankAccount.amount), props.bankAccount.currency.code),
  getOrganisationCurrencyCode.value,
))
</script>
