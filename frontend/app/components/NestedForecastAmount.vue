<template>
  <div class="flex flex-1 justify-between items-center gap-2 px-1">
    <p
      class="truncate"
      :class="[getColumnTextSize]"
    >
      {{ amountFormatted(displayedCategoryAmount) }} {{ currencyCode }}
    </p>
    <i
      v-if="relatedID"
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

const onExcludeForecast = () => {
  if (relatedID.value && relatedTable.value) {
    toggleForecastExclusionChange({
      month: props.forecastDetail.month,
      relatedID: Number(relatedID.value),
      relatedTable: relatedTable.value as string,
      originalIsExcluded: Boolean(originalIsExcluded.value),
    })
  }
}

const getCategoryRelatedValue = (
  response: ForecastDetailResponse,
  categoryName: string,
  type: 'revenue' | 'expense',
  value: 'relatedID' | 'relatedTable' | 'isExcluded',
): string | number | boolean | null => {
  const data: ForecastDetailRevenueExpenseResponse[] = response[type]

  const findValueRecursively = (
    items: ForecastDetailRevenueExpenseResponse[],
    targetName: string,
  ): string | number | boolean | null => {
    for (const item of items) {
      if (item.name === targetName) {
        return item[value] ?? null
      }

      if (item.children) {
        const childValue = findValueRecursively(item.children, targetName)
        if (childValue) {
          return childValue
        }
      }
    }

    return null
  }

  return findValueRecursively(data, categoryName)
}

const originalIsExcluded = computed(() => {
  return getCategoryRelatedValue(props.forecastDetail, props.category.name, props.forecastType, 'isExcluded')
})
const relatedID = computed(() => {
  return getCategoryRelatedValue(props.forecastDetail, props.category.name, props.forecastType, 'relatedID')
})
const relatedTable = computed(() => {
  return getCategoryRelatedValue(props.forecastDetail, props.category.name, props.forecastType, 'relatedTable')
})
const draftChange = computed(() => {
  if (!relatedID.value || !relatedTable.value) {
    return undefined
  }
  return getForecastExclusionChange(props.forecastDetail.month, Number(relatedID.value), relatedTable.value as string)
})
const effectiveIsExcluded = computed(() => {
  if (draftChange.value) {
    return draftChange.value.isExcluded
  }
  return Boolean(originalIsExcluded.value)
})
const getExclusionIcon = computed(() => {
  return effectiveIsExcluded.value ? 'pi-history text-liqui-blue' : 'pi-check-square text-liqui-green'
})
const getExclusionTooltip = computed(() => {
  return effectiveIsExcluded.value ? 'Zur Prognose hinzufügen' : 'Von der Prognose ausschliessen'
})

const categoryAmount = computed((): number => {
  const data: ForecastDetailRevenueExpenseResponse[] = props.forecastDetail[props.forecastType]

  const findAmountRecursively = (
    items: ForecastDetailRevenueExpenseResponse[],
    targetName: string,
  ): number => {
    for (const item of items) {
      if (item.name === targetName) {
        const childrenAmount = item.children
          ? item.children.reduce(
              (sum, child) => sum + findAmountRecursively([child], child.name),
              0,
            )
          : 0
        return (item.amount ?? 0) + childrenAmount
      }

      if (item.children) {
        const childAmount = findAmountRecursively(item.children, targetName)
        if (childAmount !== 0) {
          return childAmount
        }
      }
    }

    return 0
  }

  return AmountToFloat(findAmountRecursively(data, props.category.name))
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
