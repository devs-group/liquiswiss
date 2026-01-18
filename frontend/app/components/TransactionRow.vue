<template>
  <div
    :class="[
      'grid grid-cols-transactions items-center w-full *:bg-zinc-100 *:dark:bg-zinc-800 *:border *:border-r-0 *:border-b-0 *:last:border-b *:border-gray-600 *:p-1 *:text-sm *:truncate',
      transaction.isDisabled ? 'opacity-60 italic' : '',
    ]"
  >
    <div class="flex items-center gap-2 justify-end min-w-[160px]">
      <a
        v-if="transaction.link"
        :href="normalizedLink"
        target="_blank"
        rel="noopener noreferrer"
        class="truncate text-primary hover:underline"
      >
        {{ transaction.name }}
        <i class="pi pi-external-link text-xs" />
      </a>
      <p
        v-else
        class="truncate"
      >
        {{ transaction.name }}
      </p>
      <span
        v-if="transaction.isDisabled"
        class="text-xs uppercase tracking-wide text-orange-500"
      >Deaktiviert</span>
      <span
        class="pi pi-copy cursor-pointer text-help"
        @click="$emit('onClone', transaction)"
      />
      <span
        class="pi pi-pencil cursor-pointer text-primary"
        @click="$emit('onEdit', transaction)"
      />
    </div>
    <div class="flex items-center justify-center">
      <ToggleSwitch
        class="scale-[0.65] origin-center"
        :model-value="!localIsDisabled"
        :disabled="isUpdating"
        @update:model-value="onToggleDisabled"
      />
    </div>
    <p>{{ startDate }}</p>
    <p>{{ endDate || '-' }}</p>
    <p v-if="nextExecutionDate">
      {{ nextExecutionDate }}
    </p>
    <p
      v-else
      class="text-liqui-red"
    >
      Abgelaufen
    </p>
    <p class="flex flex-wrap gap-1">
      <span
        class="font-bold"
        :class="{ 'text-red-500': !isRevenue, 'text-green-500': isRevenue }"
      >{{ amountFormatted }} {{ transaction.currency.code }}</span>
    </p>
    <p>{{ vatFormatted }}</p>
    <p v-if="isRepeating">
      {{ cycle }}
    </p>
    <p v-else>
      Einmalig
    </p>
    <p>{{ transaction.category.name }}</p>
    <p class="!border-r">
      {{ transaction.employee?.name ?? '-' }}
    </p>
  </div>
</template>

<script setup lang="ts">
import type { TransactionResponse } from '~/models/transaction'
import { TransactionType } from '~/config/enums'
import { Config } from '~/config/config'

const props = defineProps({
  transaction: {
    type: Object as PropType<TransactionResponse>,
    required: true,
  },
})

defineEmits<{
  onEdit: [transaction: TransactionResponse]
  onClone: [transaction: TransactionResponse]
}>()

const toast = useToast()
const { patchTransaction } = useTransactions()
const localIsDisabled = ref(props.transaction.isDisabled)
const isUpdating = ref(false)

watch(() => props.transaction.isDisabled, (value) => {
  localIsDisabled.value = value
})

const onToggleDisabled = (isActive: boolean) => {
  const previous = localIsDisabled.value
  const nextDisabled = !isActive
  localIsDisabled.value = nextDisabled
  isUpdating.value = true
  patchTransaction({
    id: props.transaction.id,
    isDisabled: nextDisabled,
  })
    .then(() => {
      toast.add({
        summary: 'Erfolg',
        detail: nextDisabled ? 'Transaktion deaktiviert' : 'Transaktion aktiviert',
        severity: 'info',
        life: Config.TOAST_LIFE_TIME_SHORT,
      })
    })
    .catch(() => {
      localIsDisabled.value = previous
      toast.add({
        summary: 'Fehler',
        detail: 'Status konnte nicht geändert werden',
        severity: 'error',
        life: Config.TOAST_LIFE_TIME_SHORT,
      })
    })
    .finally(() => {
      isUpdating.value = false
    })
}

const isRevenue = computed(() => props.transaction.amount >= 0)
const normalizedLink = computed(() => props.transaction.link ? NormalizeUrl(props.transaction.link) : '')
const startDate = computed(() => DateStringToFormattedDate(props.transaction.startDate))
const endDate = computed(() => props.transaction.endDate ? DateStringToFormattedDate(props.transaction.endDate) : '')
const nextExecutionDate = computed(() => props.transaction.nextExecutionDate ? DateStringToFormattedDate(props.transaction.nextExecutionDate) : '')
const amountFormatted = computed(() => NumberToFormattedCurrency(AmountToFloat(props.transaction.amount), props.transaction.currency.localeCode))
const vatFormatted = computed(() => {
  if (!props.transaction.vat) {
    return '-'
  }
  const amount = NumberToFormattedCurrency(AmountToFloat(props.transaction.vatAmount), props.transaction.currency.localeCode)
  return props.transaction.vatIncluded ? `inkl. ${amount} ${props.transaction.currency.code}` : `zzgl. ${amount} ${props.transaction.currency.code}`
})
const isRepeating = computed(() => props.transaction.type === TransactionType.Repeating)
const cycle = computed(() => TransactionCycleTypeToOptions().find(ct => ct.value === props.transaction.cycle)?.name ?? '')
</script>
