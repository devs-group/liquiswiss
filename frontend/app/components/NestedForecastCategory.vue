<template>
  <div
    class="w-full flex items-center col-span-full"
  >
    <div
      class="group flex gap-1 border-b border-l border-zinc-600 dark:border-zinc-400 p-1 min-w-28"
      :class="[getColumnColor, { 'cursor-pointer': hasChildren }]"
      @click="onToggleChildren(category.name)"
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
      :depth="depth+1"
    />
  </template>
</template>

<script setup lang="ts">
import type { ForecastDetailResponse, ForecastDetailRevenueExpenseResponse } from '~/models/forecast'

const { forecastShowChildDetails } = useSettings()

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

const onToggleChildren = (categoryName: string) => {
  if (hasChildren.value) {
    if (forecastShowChildDetails.value.includes(categoryName)) {
      forecastShowChildDetails.value = forecastShowChildDetails.value.filter(n => n != categoryName)
    }
    else {
      forecastShowChildDetails.value.push(categoryName)
    }
  }
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
