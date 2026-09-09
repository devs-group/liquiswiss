<template>
  <div class="flex flex-1 justify-between items-center gap-2 px-1">
    <p
      class="truncate"
      :class="[getColumnTextSize]"
    >
      {{ amountFormatted(displayedCategoryAmount) }} {{ currencyCode }}
    </p>
    <i
      v-if="relatedIDs.length && canEdit"
      class="pi !text-2xs cursor-pointer hover:scale-125 transition-transform"
      :class="[getExclusionIcon]"
      :title="getExclusionTooltip"
      @click="onExcludeForecast()"
    />
  </div>
</template>

<script lang="ts" setup>
import type { ForecastDetailResponse, ForecastDetailRevenueExpenseResponse } from '~/models/forecast'

const { getOrganisationCurrencyLocaleCode } = useAuth()
const { toggleForecastExclusionChange, getForecastExclusionChange } = useForecasts()
const { canEdit } = useOrganisations()

const props = defineProps({
  category: {
    type: Object as PropType<ForecastDetailRevenueExpenseResponse>,
    required: true,
  },
  forecastDetail: {
    type: Object as PropType<ForecastDetailResponse>,
    required: true,
  },
  currencyCode: {
    type: String,
    required: true,
  },
  forecastType: {
    type: String as PropType<'revenue' | 'expense'>,
    required: true,
  },
  performanceFactor: {
    type: Number,
    default: 1,
  },
  depth: {
    type: Number,
    default: 0,
  },
})

const findNodeByName = (
  items: ForecastDetailRevenueExpenseResponse[],
  targetName: string,
): ForecastDetailRevenueExpenseResponse | undefined => {
  for (const item of items) {
    if (item.name === targetName) {
      return item
    }

    if (item.children) {
      const child = findNodeByName(item.children, targetName)
      if (child) {
        return child
      }
    }
  }

  return undefined
}

const detailNode = computed(() => {
  return findNodeByName(props.forecastDetail[props.forecastType] ?? [], props.category.name)
})

const relatedTable = computed(() => detailNode.value?.relatedTable ?? '')

// A row can stand for more than one entity, e.g. two transactions with the same
// name in one category or a salary cost label shared by several employees. Older
// forecast details only carry the single relatedID, hence the fallback.
const relatedIDs = computed(() => {
  const node = detailNode.value
  if (!node) {
    return []
  }
  if (node.relatedIDs?.length) {
    return node.relatedIDs
  }
  return node.relatedID ? [node.relatedID] : []
})

const originalIsExcluded = computed(() => Boolean(detailNode.value?.isExcluded))

const draftFor = (relatedID: number) => {
  return getForecastExclusionChange(props.forecastDetail.month, relatedID, relatedTable.value)
}

// The backend only reports whether the whole row is excluded, so a row that is
// only partly excluded reads as included until it gets toggled once
const isRelatedExcluded = (relatedID: number) => {
  const draft = draftFor(relatedID)
  return draft ? draft.isExcluded : originalIsExcluded.value
}

const effectiveIsExcluded = computed(() => {
  return relatedIDs.value.length > 0 && relatedIDs.value.every(isRelatedExcluded)
})

const onExcludeForecast = () => {
  if (!relatedIDs.value.length || !relatedTable.value) {
    return
  }
  // Toggling the row has to move every entity behind it to the same state
  const shouldExclude = !effectiveIsExcluded.value
  for (const relatedID of relatedIDs.value) {
    if (isRelatedExcluded(relatedID) === shouldExclude) {
      continue
    }
    toggleForecastExclusionChange({
      month: props.forecastDetail.month,
      relatedID,
      relatedTable: relatedTable.value,
      originalIsExcluded: originalIsExcluded.value,
    })
  }
}
const getExclusionIcon = computed(() => {
  return effectiveIsExcluded.value ? 'pi-history text-liqui-blue' : 'pi-check-square text-liqui-green'
})
const getExclusionTooltip = computed(() => {
  return effectiveIsExcluded.value ? 'Zur Prognose hinzufügen' : 'Von der Prognose ausschliessen'
})

const categoryAmount = computed((): number => {
  const data: ForecastDetailRevenueExpenseResponse[] = props.forecastDetail[props.forecastType]

  return AmountToFloat(FindForecastAmount(data, props.category.name))
})

const isVATCategory = computed(() => {
  return props.category.name === 'Mwst.'
})

const shouldApplyPerformanceFactor = computed(() => {
  // Apply performance factor to:
  // 1. All revenue items
  // 2. VAT (Mwst.) in expenses
  return props.forecastType === 'revenue' || isVATCategory.value
})

const displayedCategoryAmount = computed(() => {
  if (effectiveIsExcluded.value) {
    return 0
  }

  const baseAmount = categoryAmount.value
  return shouldApplyPerformanceFactor.value ? baseAmount * props.performanceFactor : baseAmount
})

const amountFormatted = (amount: number) => {
  return NumberToFormattedCurrency(amount, getOrganisationCurrencyLocaleCode.value)
}

const getColumnTextSize = computed(() => {
  switch (props.depth) {
    // Case not supported, just in case someone adds it they will immediately notice it in the frontend
    case 0:
      return 'text-xs'
    default:
      return 'text-2xs'
  }
})
</script>
