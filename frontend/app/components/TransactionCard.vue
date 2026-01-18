<template>
  <Card :class="{ 'opacity-60': localIsDisabled }">
    <template #title>
      <div class="flex items-center justify-between">
        <a
          v-if="transaction.link"
          :href="normalizedLink"
          target="_blank"
          rel="noopener noreferrer"
          class="truncate text-base text-primary hover:underline"
        >
          {{ transaction.name }}
          <i class="pi pi-external-link text-xs ml-1" />
        </a>
        <p
          v-else
          class="truncate text-base"
        >
          {{ transaction.name }}
        </p>
        <div class="flex items-center gap-2 justify-end">
          <Button
            severity="help"
            icon="pi pi-copy"
            outlined
            rounded
            @click="$emit('onClone', transaction)"
          />
          <Button
            icon="pi pi-pencil"
            outlined
            rounded
            @click="$emit('onEdit', transaction)"
          />
        </div>
      </div>
    </template>
    <template #content>
      <div class="flex flex-col text-sm gap-2">
        <Message
          v-if="localIsDisabled"
          severity="warn"
          size="small"
          :closable="false"
        >
          Diese Transaktion ist deaktiviert und wird nicht berechnet.
        </Message>
        <p>Start: {{ startDate }}</p>
        <p v-if="isRepeating && endDate">
          Ende: {{ endDate }}
        </p>
        <p v-if="nextExecutionDate">
          Nächste {{ getNextLabel }}: {{ nextExecutionDate }}
        </p>
        <p
          v-else
          class="text-liqui-red"
        >
          Abgelaufen
        </p>
        <p class="flex flex-wrap gap-1">
          Betrag: <span
            class="font-bold"
            :class="{ 'text-liqui-red': !isRevenue, 'text-liqui-green': isRevenue }"
          >{{ amountFormatted }} {{ transaction.currency.code }}</span>
        </p>
        <p v-if="isRepeating">
          Wiederkehrend: {{ cycle }}
        </p>
        <p v-else>
          Einmalig
        </p>
        <p class="truncate">
          Kategorie: {{ transaction.category.name }}
        </p>
        <p v-if="transaction.employee">
          Mitarbeiter: {{ transaction.employee.name }}
        </p>
        <div class="flex items-center gap-2">
          <p>Status:</p>
          <ToggleSwitch
            id="transaction-card-disabled"
            class="scale-[0.65] origin-left"
            :model-value="!localIsDisabled"
            :disabled="isUpdating"
            @update:model-value="onToggleDisabled"
          />
        </div>
      </div>
    </template>
  </Card>
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
const isRepeating = computed(() => props.transaction.type === TransactionType.Repeating)
const cycle = computed(() => TransactionCycleTypeToOptions().find(ct => ct.value === props.transaction.cycle)?.name ?? '')
const getNextLabel = computed(() => isRevenue.value ? 'Gutschrift' : 'Belastung')
</script>
