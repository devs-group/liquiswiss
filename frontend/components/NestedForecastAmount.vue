<template>
  <div class="flex flex-1 justify-between items-center gap-2 px-1">
    <p
      class="truncate"
      :class="[getColumnTextSize]"
    >
      {{ amountFormatted(getCategoryAmount) }} {{ currencyCode }}
    </p>
    <i
      v-if="getCategoryRelatedValue(forecastDetail, category.name, forecastType, 'relatedID') && getCategoryAmount"
      class="pi !text-2xs cursor-pointer hover:scale-125 transition-transform"
      :class="[getExclusionIcon]"
      @click="onExcludeForecast(forecastDetail, category.name, forecastType)"
    />
  </div>
</template>

<script lang="ts" setup>
import type { ForecastDetailResponse, ForecastDetailRevenueExpenseResponse } from '~/models/forecast'
import { Config } from '~/config/config'

const { getOrganisationCurrencyLocaleCode } = useAuth()
const { excludeForecast, includeForecast } = useForecasts()
const toast = useToast()

const emits = defineEmits(['onRecalculateForecasts'])

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
  depth: {
    type: Number,
    default: 0,
  },
})

const onExcludeForecast = (
  data: ForecastDetailResponse,
  categoryName: string,
  type: 'revenue' | 'expense') => {
  const relatedID = getCategoryRelatedValue(data, categoryName, type, 'relatedID')
  const relatedTable = getCategoryRelatedValue(data, categoryName, type, 'relatedTable')
  const isExcluded = getCategoryRelatedValue(data, categoryName, type, 'isExcluded')
  if (relatedID && relatedTable) {
    if (!isExcluded) {
      excludeForecast(data.month, relatedID as number, relatedTable as string)
        .then(() => {
          toast.add({
            summary: 'Erfolg',
            detail: `Prognose wird für Monat "${data.month}" ausgeschlossen"`,
            severity: 'success',
            life: Config.TOAST_LIFE_TIME_MEDIUM,
          })
        })
        .catch(() => {
          toast.add({
            summary: 'Fehler',
            detail: `Ausschliessen der Prognose für Monat "${data.month}" fehlgeschlagen`,
            severity: 'error',
            life: Config.TOAST_LIFE_TIME_MEDIUM,
          })
        })
        .finally(() => {
          emits('onRecalculateForecasts')
        })
    }
    else {
      includeForecast(data.month, relatedID as number, relatedTable as string)
        .then(() => {
          toast.add({
            summary: 'Erfolg',
            detail: `Prognose wird für Monat "${data.month}" berücksichtigt`,
            severity: 'success',
            life: Config.TOAST_LIFE_TIME_MEDIUM,
          })
        })
        .catch(() => {
          toast.add({
            summary: 'Fehler',
            detail: `Berücksichtigen der Prognose für Monat "${data.month}" fehlgeschlagen`,
            severity: 'error',
            life: Config.TOAST_LIFE_TIME_MEDIUM,
          })
        })
        .finally(() => {
          emits('onRecalculateForecasts')
        })
    }
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

const getExclusionIcon = computed(() => {
  const isExcluded = getCategoryRelatedValue(props.forecastDetail, props.category.name, props.forecastType, 'isExcluded')
  return isExcluded ? 'pi-history text-liqui-blue' : 'pi-check-square text-liqui-green'
})

const getCategoryAmount = computed((): number => {
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
