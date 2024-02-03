<template>
  <Card>
    <template #title>
      <div class="flex items-center justify-between">
        <p class="truncate text-base">{{transaction.attributes.name}}</p>
        <div class="flex justify-end">
          <Button @click="$emit('onEdit', transaction)" icon="pi pi-pencil" outlined rounded />
        </div>
      </div>
    </template>
    <template #content>
      <div class="flex flex-col text-sm">
        <p>Start: {{start}}</p>
        <p v-if="isRepeating && end">Ende: {{end}}</p>
        <p class="flex flex-wrap gap-1">Betrag: <span>{{amountFormatted}} {{currencyCode}}</span></p>
        <p v-if="isRepeating">Wiederkehrend: {{cycle}}</p>
        <p v-else>Einmalig</p>
        <p class="truncate">Kategorie: {{category}}</p>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import {DateStringToFormattedDate, NumberToFormattedCurrency} from "~/utils/format-helper";
import type {StrapiTransaction} from "~/models/transaction";
import useGlobalData from "~/composables/useGlobalData";

const {getCategoryNameFromId, getCurrencyCodeFromId, getCurrencyLocaleCodeFromId} = useGlobalData()

const props = defineProps({
  transaction: {
    type: Object as PropType<StrapiTransaction>,
    required: true,
  }
})

defineEmits<{
  'onEdit': [transaction: StrapiTransaction]
}>()

const start = computed(() => DateStringToFormattedDate(props.transaction.attributes.start))
const end = computed(() => props.transaction.attributes.end ? DateStringToFormattedDate(props.transaction.attributes.end) : '')
const currencyCode = computed(() => getCurrencyCodeFromId(props.transaction.attributes.currency as number))
const currencyLocaleCode = computed(() => getCurrencyLocaleCodeFromId(props.transaction.attributes.currency as number))
const amountFormatted = computed(() => NumberToFormattedCurrency(props.transaction.attributes.amount, currencyLocaleCode.value))
const type = computed(() => props.transaction.attributes.type)
const isRepeating = computed(() => type.value === 'repeating')
const cycle = computed(() => CycleTypeToOptions().find((ct) => ct.value === props.transaction.attributes.cycle)?.name ?? '')
const category = computed(() => getCategoryNameFromId(props.transaction.attributes.category as number))
</script>
