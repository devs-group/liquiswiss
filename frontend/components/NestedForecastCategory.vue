<template>
  <div
    class="w-full flex items-center col-span-full"
  >
    <div
      class="group flex gap-1 border-b border-l border-zinc-600 dark:border-zinc-400 p-1 min-w-28"
      :class="[getColumnColor, { 'cursor-pointer': hasChildren }]"
      @click="onToggleChildren"
    >
      <p
        class="w-full truncate"
        :class="[getColumnTextAlignment, getColumnTextSize]"
      >
        {{ category.name }}
      </p>
      <i
        v-if="hasChildren"
        class="pi opacity-0 group-hover:opacity-100 transition-opacity"
        :class="{ 'pi-sort-down': forecastShowChildDetails, 'pi-sort-up': !forecastShowChildDetails }"
      />
    </div>
    <div
      v-for="data in forecastDetails"
      :key="data.forecastID"
      class="w-full border-b border-l last:border-r border-zinc-600 dark:border-zinc-400 p-1 min-w-40"
      :class="[getColumnColor]"
    >
      <div class="flex flex-1 justify-between items-center gap-2 px-1">
        <p
          class="truncate"
          :class="[getColumnTextSize]"
        >
          {{ getCategoryAmount(data, category.name, forecastType) }} {{ currencyCode }}
        </p>
        <i
          v-if="getCategoryRelatedValue(data, category.name, forecastType, 'relatedID')"
          class="pi !text-2xs cursor-pointer hover:scale-125 transition-transform"
          :class="[getExclusionIcon(data, category.name, forecastType)]"
          @click="onExcludeForecast(data, category.name, forecastType)"
        />
      </div>
    </div>
  </div>

  <template v-if="forecastShowChildDetails">
    <NestedForecastCategory
      v-for="child in childCategories"
      :key="child"
      :category="{ name: child, children: [], amount: 0, relatedID: 0, relatedTable: '', isExcluded: false }"
      :forecast-type="forecastType"
      :forecast-details="forecastDetails"
      :currency-code="currencyCode"
      :depth="depth+1"
      @on-recalculate-forecasts="onRecalculateForecasts"
    />
  </template>
</template>

<script setup lang="ts">
import type { ForecastDetailResponse, ForecastDetailRevenueExpenseResponse } from '~/models/forecast'
import { Config } from '~/config/config'

const { getOrganisationCurrencyLocaleCode } = useAuth()
const { forecastShowChildDetails } = useSettings()
const { excludeForecast, includeForecast } = useForecasts()
const toast = useToast()

const props = defineProps({
  category: {
    type: Object as PropType<ForecastDetailRevenueExpenseResponse>,
    required: true,
  },
  forecastDetails: {
    type: Array as PropType<ForecastDetailResponse[]>,
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

const emits = defineEmits(['onRecalculateForecasts'])

const onToggleChildren = () => {
  if (hasChildren.value) {
    forecastShowChildDetails.value = !forecastShowChildDetails.value
  }
}

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

const getCategoryAmount = (
  response: ForecastDetailResponse,
  categoryName: string,
  type: 'revenue' | 'expense',
): string => {
  const data: ForecastDetailRevenueExpenseResponse[] = response[type]

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

  return NumberToFormattedCurrency(AmountToFloat(findAmountRecursively(data, categoryName)), getOrganisationCurrencyLocaleCode.value)
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

const getExclusionIcon = (
  data: ForecastDetailResponse,
  categoryName: string,
  type: 'revenue' | 'expense') => {
  const isExcluded = getCategoryRelatedValue(data, categoryName, type, 'isExcluded')
  return isExcluded ? 'pi-history text-liqui-blue' : 'pi-check-square text-liqui-green'
}

const onRecalculateForecasts = () => {
  emits('onRecalculateForecasts')
}

const hasChildren = computed(() => {
  return props.category.children && props.category.children.length > 0
})

const childCategories = computed(() => {
  const categories: string[] = []
  props.forecastDetails
    .map(f => (props.forecastType == 'revenue' ? f.revenue : f.expense).find(e => e.name === props.category.name)?.children)
    .filter(f => f)
    .flat()
    .forEach((e) => {
      if (e && !categories.includes(e.name)) {
        categories.push(e.name)
      }
    })
  return categories.sort((a, b) => a.localeCompare(b))
})

const getColumnColor = computed(() => {
  switch (props.depth) {
    // Case not supported, just in case someone adds it they will immediately notice it in the frontend
    case 2:
      return 'bg-white dark:bg-black'
    case 1:
      return 'bg-zinc-100 dark:bg-zinc-900'
    default:
      return 'bg-zinc-300 dark:bg-zinc-800'
  }
})

const getColumnTextAlignment = computed(() => {
  switch (props.depth) {
    // Case not supported, just in case someone adds it they will immediately notice it in the frontend
    case 0:
      return 'text-left'
    default:
      return 'text-right'
  }
})

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
