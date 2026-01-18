<template>
  <div
    class="w-full flex items-center col-span-full"
  >
    <div
      v-tooltip="isAutoCategory ? 'MwSt. ist eine automatisch berechnete Ausgabe basierend auf den MwSt.-Einstellungen.' : undefined"
      class="group flex gap-1 border-b border-l border-zinc-600 dark:border-zinc-400 p-1 min-w-28"
      :class="[getColumnColor, { 'cursor-pointer': hasChildren && !isAutoCategory, 'cursor-default': isAutoCategory }]"
      @click="onToggleChildren(category.name)"
    >
      <p
        class="w-full truncate"
        :class="[getColumnTextAlignment, getColumnTextSize]"
      >
        {{ isAutoCategory.value ? `${category.name} (auto)` : category.name }}
      </p>
      <i
        v-if="hasChildren && !isAutoCategory.value"
        class="pi opacity-0 group-hover:opacity-100 transition-opacity"
        :class="{ 'pi-sort-down': forecastShowChildDetails.includes(category.name), 'pi-sort-up': !forecastShowChildDetails.includes(category.name) }"
      />
    </div>
    <div
      v-for="data in forecastDetails"
      :key="data.forecastID"
      class="w-full border-b border-l last:border-r border-zinc-600 dark:border-zinc-400 p-1 min-w-40"
      :class="[getColumnColor]"
    >
      <NestedForecastAmount
        :category="category"
        :forecast-detail="data"
        :currency-code="currencyCode"
        :forecast-type="forecastType"
        :performance-factor="performanceFactor"
      />
    </div>
  </div>

  <template v-if="forecastShowChildDetails.includes(category.name)">
    <NestedForecastCategory
      v-for="child in childCategories"
      :key="child"
      :category="{ name: child, children: [], amount: 0, relatedID: 0, relatedTable: '', isExcluded: false }"
      :forecast-type="forecastType"
      :forecast-details="forecastDetails"
      :currency-code="currencyCode"
      :performance-factor="performanceFactor"
      :depth="depth+1"
      :visited="[...visited, category.name]"
    />
  </template>
</template>

<script setup lang="ts">
import type { ForecastDetailResponse, ForecastDetailRevenueExpenseResponse } from '~/models/forecast'

const { forecastShowChildDetails, setForecastShowChildDetails } = useSettings()

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
  performanceFactor: {
    type: Number,
    default: 1,
  },
  depth: {
    type: Number,
    default: 0,
  },
  visited: {
    type: Array as PropType<string[]>,
    default: () => [],
  },
})

const onToggleChildren = (categoryName: string) => {
  if (hasChildren.value) {
    if (forecastShowChildDetails.value.includes(categoryName)) {
      setForecastShowChildDetails(forecastShowChildDetails.value.filter(n => n !== categoryName))
    }
    else {
      setForecastShowChildDetails([...forecastShowChildDetails.value, categoryName])
    }
  }
}

const hasChildren = computed(() => {
  return !isAutoCategory.value && props.category.children && props.category.children.length > 0
})

const childCategories = computed(() => {
  const categories: string[] = []
  props.forecastDetails
    .map(f => (props.forecastType == 'revenue' ? f.revenue : f.expense).find(e => e.name === props.category.name)?.children)
    .filter(f => f)
    .flat()
    .forEach((e) => {
      if (e && e.name !== props.category.name && !props.visited.includes(e.name) && !categories.includes(e.name)) {
        categories.push(e.name)
      }
    })
  return categories.sort((a, b) => a.localeCompare(b))
})

const getColumnColor = computed(() => {
  if (isAutoCategory.value) {
    return 'bg-amber-50 dark:bg-amber-900/40'
  }
  switch (props.depth) {
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

const isAutoCategory = computed(() => props.category.name === 'Mwst.')
</script>
