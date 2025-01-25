<template>
  <div
    class="w-full flex items-center col-span-full"
  >
    <div
      class="flex gap-1 border-b border-l border-zinc-600 dark:border-zinc-400 p-1 min-w-28"
      :class="[getColumnColor, { 'cursor-pointer': hasChildren }]"
      @click="toggleChildren"
    >
      <p class="w-full text-xs text-right truncate">
        {{ category.name }}
      </p>
      <i
        v-if="hasChildren"
        class="pi"
        :class="{ 'pi-sort-down': forecastShowSalaryCostDetails, 'pi-sort-up': !forecastShowSalaryCostDetails }"
      />
    </div>
    <div
      v-for="data in forecastDetails"
      :key="data.forecastID"
      class="w-full border-b border-l last:border-r border-zinc-600 dark:border-zinc-400 bg-zinc-100 dark:bg-zinc-800 p-1 min-w-40"
    >
      <p class="text-xs text-center">
        {{ getCategoryAmount(data, category.name, 'expense') }} {{ currencyCode }}
      </p>
    </div>
  </div>

  <!-- Special case -->
  <template v-if="forecastShowSalaryCostDetails">
    <NestedForecastCategory
      v-for="child in childCategories"
      :key="child"
      :category="{ name: child, children: [] }"
      :forecast-details="forecastDetails"
      :currency-code="currencyCode"
      :depth="depth+1"
    />
  </template>
</template>

<script setup lang="ts">
import type { ForecastDetailResponse, ForecastDetailRevenueExpenseResponse } from '~/models/forecast'

const { getOrganisationCurrencyLocaleCode } = useAuth()
const { forecastShowSalaryCostDetails } = useSettings()

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
  depth: {
    type: Number,
    default: 0,
  },
})

const toggleChildren = () => {
  if (hasChildren.value) {
    forecastShowSalaryCostDetails.value = !forecastShowSalaryCostDetails.value
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

const hasChildren = computed(() => {
  return props.category.children && props.category.children.length > 0
})

const childCategories = computed(() => {
  const categories: string[] = []
  props.forecastDetails
    .map(f => f.expense.find(e => e.name === props.category.name)?.children)
    .filter(f => f)
    .flat()
    .forEach((e) => {
      if (!categories.includes(e.name)) {
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
      return 'bg-zinc-400 dark:bg-zinc-600'
    default:
      return 'bg-zinc-300 dark:bg-zinc-700'
  }
})
</script>
