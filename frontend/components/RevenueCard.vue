<template>
  <Card>
    <template #title>
      <div class="flex items-center justify-between">
        <p class="truncate text-base">{{revenue.attributes.name}}</p>
        <div class="flex justify-end">
          <Button @click="$emit('onEdit', revenue)" icon="pi pi-pencil" outlined rounded />
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
import type {StrapiRevenue} from "~/models/revenue";
import useGlobalData from "~/composables/useGlobalData";

const {getCategoryNameFromId, getCurrencyCodeFromId, getCurrencyLocaleCodeFromId} = useGlobalData()

const props = defineProps({
  revenue: {
    type: Object as PropType<StrapiRevenue>,
    required: true,
  }
})

defineEmits<{
  'onEdit': [revenue: StrapiRevenue]
}>()

const start = computed(() => DateStringToFormattedDate(props.revenue.attributes.start))
const end = computed(() => props.revenue.attributes.end ? DateStringToFormattedDate(props.revenue.attributes.end) : '')
const currencyCode = computed(() => getCurrencyCodeFromId(props.revenue.attributes.currency as number))
const currencyLocaleCode = computed(() => getCurrencyLocaleCodeFromId(props.revenue.attributes.currency as number))
const amountFormatted = computed(() => NumberToFormattedCurrency(props.revenue.attributes.amount, currencyLocaleCode.value))
const type = computed(() => props.revenue.attributes.type)
const isRepeating = computed(() => type.value === 'repeating')
const cycle = computed(() => CycleTypeToOptions().find((ct) => ct.value === props.revenue.attributes.cycle)?.name ?? '')
const category = computed(() => getCategoryNameFromId(props.revenue.attributes.category as number))
</script>
