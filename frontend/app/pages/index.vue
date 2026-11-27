<template>
  <div
    class="flex flex-col gap-2"
  >
    <Message
      v-if="bankAccountsErrorMessage.length"
      severity="error"
      :closable="false"
      class="col-span-full"
    >
      {{ bankAccountsErrorMessage }}
    </Message>
    <Message
      v-if="forecastDetailsErrorMessage.length"
      severity="error"
      :closable="false"
      class="col-span-full"
    >
      {{ forecastDetailsErrorMessage }}
    </Message>
    <Message
      v-if="forecastCalculateErrorMessage.length"
      severity="error"
      :closable="false"
      class="col-span-full"
    >
      {{ forecastCalculateErrorMessage }}
    </Message>
    <Message
      v-if="forecastErrorMessage.length"
      severity="error"
      :closable="false"
      class="col-span-full"
    >
      {{ forecastErrorMessage }}
    </Message>
    <div
      v-else
      class="flex flex-col gap-4"
    >
      <div class="flex flex-col gap-4">
        <!-- Wrapping row instead of a 3 column grid: the three controls need
             far more than the sm breakpoint offers and used to overlap -->
        <div class="flex flex-wrap items-center gap-x-6 gap-y-2">
          <div class="flex shrink-0 items-center gap-2">
            <p class="text-sm whitespace-nowrap">
              Zeitraum:
            </p>
            <InputNumber
              :model-value="forecastMonths"
              show-buttons
              button-layout="horizontal"
              :step="1"
              :min="1"
              :allow-empty="false"
              size="small"
              mode="decimal"
              input-class="w-24"
              :suffix="forecastMonths == 1 ? ' Monat' : ' Monate'"
              :max="36"
              @update:model-value="setForecastMonths"
            >
              <template #incrementbuttonicon>
                <span class="pi pi-plus" />
              </template>
              <template #decrementbuttonicon>
                <span class="pi pi-minus" />
              </template>
            </InputNumber>
          </div>
          <div class="flex shrink-0 items-center gap-2">
            <p class="text-sm whitespace-nowrap">
              Performance ({{ forecastPerformance }}%):
            </p>
            <Slider
              :model-value="forecastPerformance"
              class="w-40 xl:w-56"
              @update:model-value="setForecastPerformance"
            />
          </div>
          <div class="flex shrink-0 items-center gap-2 sm:ml-auto">
            <Button
              v-tooltip.bottom="forecastShowZeroRows ? 'Zeilen ohne Beträge ausblenden' : 'Zeilen ohne Beträge anzeigen'"
              :icon="forecastShowZeroRows ? 'pi pi-eye' : 'pi pi-eye-slash'"
              :severity="forecastShowZeroRows ? 'contrast' : 'secondary'"
              size="small"
              :label="forecastShowZeroRows ? 'Nullzeilen sichtbar' : 'Nullzeilen ausgeblendet'"
              @click="toggleForecastShowZeroRows"
            />
          </div>
        </div>

        <div class="flex flex-col items-end gap-1">
          <Button
            v-if="canEdit"
            :disabled="isLoading"
            label="Neu berechnen"
            size="small"
            @click="onCalculateForecast"
          />
          <p class="text-xs">
            Zuletzt berechnet am
            <span>
              {{ latestUpdate }}
            </span>
          </p>
        </div>

        <ClientOnly>
          <div v-if="hasNoDataInCurrentMonth">
            <Message size="small">
              Hinweis: Sie befinden sich noch im Monat "{{ localMonth }}" aufgrund Ihrer Zeitzone, für diesen Monat werden keine Prognosedaten mehr dargestellt
            </Message>
          </div>
        </ClientOnly>

        <div
          ref="forecastScroll"
          data-testid="forecast-scroll"
          class="relative flex flex-col overflow-x-auto pb-2"
          @scroll.passive="onForecastScroll"
        >
          <!-- min-w-max lets the grid grow past the scroll container so the
               sticky label column stays pinned across the whole scroll range,
               while still stretching to full width for few months -->
          <div class="grid grid-cols-12 items-center min-w-max">
            <div class="flex items-center col-span-full">
              <div
                class="sticky left-0 z-10 shrink-0 w-28 border-t border-b border-l border-zinc-600 dark:border-zinc-400 bg-zinc-300 dark:bg-zinc-700 p-2"
                :class="stickyDividerClass"
              >
                <p class="text-xs">
                  &nbsp;
                </p>
              </div>
              <div
                v-for="month in months"
                :key="month"
                class="w-full border-t border-b border-l last:border-r border-zinc-600 dark:border-zinc-400 bg-zinc-300 dark:bg-zinc-700 p-2 min-w-40"
              >
                <p class="text-xs text-center font-bold">
                  {{ month }}
                </p>
              </div>
            </div>

            <div class="flex items-center col-span-full">
              <div
                v-tooltip.top="detailToggleTooltip"
                class="group sticky left-0 z-10 shrink-0 w-28 flex gap-1 border-b border-l border-zinc-600 dark:border-zinc-400 bg-zinc-400 dark:bg-zinc-600 p-2"
                :class="[stickyDividerClass, isComparing ? 'cursor-default' : 'cursor-pointer']"
                @click="onToggleRevenueDetails"
              >
                <p class="text-xs font-bold">
                  Einnahmen
                </p>
                <i
                  v-if="!isComparing"
                  class="pi opacity-0 group-hover:opacity-100 transition-opacity"
                  :class="{ 'pi-sort-down': showRevenueDetails, 'pi-sort-up': !showRevenueDetails }"
                />
              </div>
              <div
                v-for="(revenue, i) in revenues"
                :key="i"
                class="w-full border-b border-l last:border-r border-zinc-600 dark:border-zinc-400 text-liqui-green p-2 min-w-40"
              >
                <p class="text-xs text-center">
                  {{ revenue.formatted }} {{ getOrganisationCurrencyCode }}
                </p>
              </div>
            </div>

            <CollapsibleRows
              v-if="forecastShowRevenueDetails"
              :open="showRevenueDetails"
            >
              <NestedForecastCategory
                v-for="category in revenueCategories"
                :key="category.name"
                :category="category"
                forecast-type="revenue"
                :forecast-details="forecastDetails"
                :currency-code="getOrganisationCurrencyCode"
                :performance-factor="forecastPerformance / 100"
                :sticky-divider-class="stickyDividerClass"
              />
            </CollapsibleRows>

            <div class="flex items-center col-span-full">
              <div
                v-tooltip.top="detailToggleTooltip"
                class="group sticky left-0 z-10 shrink-0 w-28 flex gap-1 border-b border-l border-zinc-600 dark:border-zinc-400 bg-zinc-400 dark:bg-zinc-600 p-2"
                :class="[stickyDividerClass, isComparing ? 'cursor-default' : 'cursor-pointer']"
                @click="onToggleExpenseDetails"
              >
                <p class="text-xs font-bold">
                  Ausgaben
                </p>
                <i
                  v-if="!isComparing"
                  class="pi opacity-0 group-hover:opacity-100 transition-opacity"
                  :class="{ 'pi-sort-down': showExpenseDetails, 'pi-sort-up': !showExpenseDetails }"
                />
              </div>
              <div
                v-for="(expense, i) in expenses"
                :key="i"
                class="w-full border-b border-l last:border-r border-zinc-600 dark:border-zinc-400 text-liqui-red p-2 min-w-40"
              >
                <p class="text-xs text-center">
                  {{ expense.formatted }} {{ getOrganisationCurrencyCode }}
                </p>
              </div>
            </div>

            <CollapsibleRows
              v-if="forecastShowExpenseDetails"
              :open="showExpenseDetails"
            >
              <NestedForecastCategory
                v-for="category in expenseCategories"
                :key="category.name"
                :category="category"
                forecast-type="expense"
                :forecast-details="forecastDetails"
                :currency-code="getOrganisationCurrencyCode"
                :performance-factor="forecastPerformance / 100"
                :sticky-divider-class="stickyDividerClass"
              />
            </CollapsibleRows>

            <div class="flex items-center col-span-full">
              <div
                class="sticky left-0 z-10 shrink-0 w-28 cursor-default border-b border-l border-zinc-600 dark:border-zinc-400 bg-zinc-300 dark:bg-zinc-700 p-2"
                :class="stickyDividerClass"
              >
                <p class="text-xs">
                  Cashflow
                </p>
              </div>
              <div
                v-for="(cashflow, i) in cashflows"
                :key="i"
                class="w-full border-b border-l last:border-r border-zinc-600 dark:border-zinc-400 p-2 min-w-40"
                :class="{ 'text-liqui-red': cashflow.amount < 0, 'text-liqui-green': cashflow.amount > 0 }"
              >
                <p class="text-xs text-center">
                  {{ cashflow.formatted }} {{ getOrganisationCurrencyCode }}
                </p>
              </div>
            </div>

            <div class="flex items-center col-span-full">
              <div
                class="sticky left-0 z-10 shrink-0 w-28 cursor-default border-b border-l border-zinc-600 dark:border-zinc-400 bg-zinc-300 dark:bg-zinc-700 p-2"
                :class="stickyDividerClass"
              >
                <p class="text-xs">
                  Endsaldo
                </p>
              </div>
              <div
                v-for="(saldo, i) in saldos"
                :key="i"
                class="w-full border-b border-l last:border-r border-zinc-600 dark:border-zinc-400 p-2 min-w-40"
                :class="{ 'bg-liqui-red': saldo.amount < 0, 'bg-liqui-green': saldo.amount > 0 }"
              >
                <p class="text-xs text-center font-bold truncate">
                  {{ saldo.formatted }} {{ getOrganisationCurrencyCode }}
                </p>
              </div>
            </div>

            <!-- Grid rows 0fr to 1fr animates the height without knowing it -->
            <div
              v-if="shownCompareForecast"
              class="col-span-full grid transition-[grid-template-rows] duration-300 ease-in-out"
              :class="isCompareOpen ? 'grid-rows-[1fr]' : 'grid-rows-[0fr]'"
              @transitionend="onCompareTransitionEnd"
            >
              <!-- clip, not hidden: hidden would become the scroll box and break the sticky labels -->
              <div class="min-h-0 overflow-y-clip">
                <!-- Title sets the comparison apart from the active scenario -->
                <div class="pt-4">
                  <p class="sticky left-0 w-max pb-1 text-sm font-bold">
                    Vergleich mit «{{ shownCompareForecast.scenario.name }}»
                  </p>
                  <div
                    class="flex items-center"
                    data-testid="forecast-compare-saldo"
                  >
                    <div
                      v-tooltip.top="{ value: `Endsaldo «${shownCompareForecast.scenario.name}»`, showDelay: 0 }"
                      class="sticky left-0 z-10 shrink-0 w-28 cursor-default border-t border-b border-l border-zinc-600 dark:border-zinc-400 bg-zinc-300 dark:bg-zinc-700 p-2"
                      :class="stickyDividerClass"
                    >
                      <p class="text-xs truncate">
                        Endsaldo
                      </p>
                    </div>
                    <div
                      v-for="(cell, i) in compareCells"
                      :key="i"
                      class="w-full border-t border-b border-l last:border-r border-zinc-600 dark:border-zinc-400 p-2 min-w-40"
                      :class="{ 'text-liqui-green': cell && cell.diff.amount > 0, 'text-liqui-red': cell && cell.diff.amount < 0 }"
                    >
                      <p class="flex items-center justify-center gap-1 text-xs truncate">
                        <i
                          v-if="cell && cell.diff.amount !== 0"
                          class="pi !text-[0.6rem]"
                          :class="cell.diff.amount > 0 ? 'pi-arrow-up' : 'pi-arrow-down'"
                        />
                        {{ cell ? `${cell.saldo.formatted} ${getOrganisationCurrencyCode}` : '-' }}
                      </p>
                    </div>
                  </div>
                  <div
                    class="flex items-center"
                    data-testid="forecast-compare-diff"
                  >
                    <div
                      v-tooltip.top="{ value: `«${shownCompareForecast.scenario.name}» minus aktuelles Szenario`, showDelay: 0 }"
                      class="sticky left-0 z-10 shrink-0 w-28 cursor-default border-b border-l border-zinc-600 dark:border-zinc-400 bg-zinc-300 dark:bg-zinc-700 p-2"
                      :class="stickyDividerClass"
                    >
                      <p class="text-xs">
                        Differenz
                      </p>
                    </div>
                    <div
                      v-for="(cell, i) in compareCells"
                      :key="i"
                      class="w-full border-b border-l last:border-r border-zinc-600 dark:border-zinc-400 p-2 min-w-40"
                      :class="{ 'text-liqui-green': cell && cell.diff.amount > 0, 'text-liqui-red': cell && cell.diff.amount < 0 }"
                    >
                      <p class="text-xs text-center font-bold truncate">
                        {{ cell ? `${cell.diff.amount > 0 ? '+' : ''}${cell.diff.formatted} ${getOrganisationCurrencyCode}` : '-' }}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="compareOptions.length"
        ref="compareBar"
        class="flex flex-wrap items-center gap-2"
      >
        <p class="text-sm whitespace-nowrap">
          Vergleichen mit:
        </p>
        <Select
          v-model="compareScenarioId"
          :options="compareOptions"
          option-label="name"
          option-value="id"
          placeholder="Kein Vergleich"
          show-clear
          size="small"
          class="w-48"
          data-testid="forecast-compare-select"
        />
        <p
          v-if="compareForecast"
          class="text-xs"
        >
          zuletzt berechnet am {{ compareLatestUpdate }}
        </p>
        <Message
          v-if="compareErrorMessage.length"
          severity="error"
          size="small"
          :closable="false"
        >
          {{ compareErrorMessage }}
        </Message>
      </div>

      <div class="bg-zinc-50 dark:bg-zinc-800">
        <ClientOnly>
          <Chart
            type="line"
            :data="chartData"
            :options="chartOptions"
            class="h-80"
          />
          <template #fallback>
            <Skeleton class="!h-80" />
          </template>
        </ClientOnly>
      </div>

      <FullProgressSpinner :show="isLoading" />
    </div>

    <div
      v-if="hasPendingExclusionChanges"
      class="fixed inset-x-0 bottom-4 z-50 flex justify-center px-4"
    >
      <div class="flex w-full max-w-sm flex-col items-center gap-2 rounded-2xl border border-zinc-300 bg-white/95 px-4 py-3 shadow-lg backdrop-blur-sm dark:border-zinc-700 dark:bg-zinc-900/95">
        <p class="text-sm font-medium text-center">
          Geänderte Prognosen speichern?
        </p>
        <div class="flex items-center gap-4">
          <Button
            :disabled="isSavingExclusions"
            severity="secondary"
            icon="pi pi-times"
            rounded
            aria-label="Änderungen verwerfen"
            @click="onCancelForecastExclusionChanges"
          />
          <Button
            :loading="isSavingExclusions"
            icon="pi pi-check"
            rounded
            aria-label="Änderungen speichern"
            @click="onSaveForecastExclusionChanges"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Chart from 'primevue/chart'
import useCharts from '~/composables/useCharts'
import type { ForecastDetailResponse, ForecastDetailRevenueExpenseResponse, ForecastResponse } from '~/models/forecast'
import type { ScenarioForecastResponse } from '~/models/scenario'
import FullProgressSpinner from '~/components/FullProgressSpinner.vue'
import { Config } from '~/config/config'

useHead({
  title: 'Prognose',
})

// The sticky label column only gets its separating shadow once the table is
// actually scrolled sideways, so it stays flat while everything is visible
const forecastScroll = ref<HTMLElement | null>(null)
const isForecastScrolled = ref(false)
const onForecastScroll = () => {
  isForecastScrolled.value = (forecastScroll.value?.scrollLeft ?? 0) > 0
}
// Divider that separates the pinned label column from the months scrolling
// underneath it. Passed down into the nested category rows as a prop.
const stickyDividerClass = computed(() => isForecastScrolled.value
  ? 'border-r shadow-[4px_0_6px_-2px_rgba(0,0,0,0.25)] dark:shadow-[4px_0_6px_-2px_rgba(0,0,0,0.6)]'
  : '')

const utcFormatter = new Intl.DateTimeFormat(Constants.BASE_LOCALE_CODE, { month: 'long', year: 'numeric', timeZone: 'UTC' })
const localFormatter = new Intl.DateTimeFormat(Constants.BASE_LOCALE_CODE, { month: 'long', year: 'numeric' })

const { getOrganisationCurrencyCode, getOrganisationCurrencyLocaleCode } = useAuth()
const { canEdit } = useOrganisations()
const {
  useFetchListForecast,
  listForecasts,
  useFetchListForecastDetails,
  listForecastDetails,
  forecasts,
  forecastDetails,
  calculateForecast,
  forecastExclusionChanges,
  applyForecastExclusionChanges,
  clearForecastExclusionChanges,
} = useForecasts()
const { useFetchListBankAccounts, totalBankSaldoInCHF } = useBankAccounts()
const { forecastPerformance, forecastMonths, forecastShowRevenueDetails, forecastShowExpenseDetails, forecastShowZeroRows, setForecastShowRevenueDetails, setForecastShowExpenseDetails, setForecastPerformance, setForecastMonths, toggleForecastShowZeroRows } = useSettings()
const { setChartData, getChartOptions } = useCharts()
const toast = useToast()

const bankAccountsErrorMessage = ref('')
const forecastErrorMessage = ref('')
const forecastDetailsErrorMessage = ref('')
const forecastCalculateErrorMessage = ref('')

const forecastMonthsComputed = computed(() => {
  return forecastMonths.value + 1
})

await useFetchListForecast(forecastMonthsComputed.value)
  .catch((reason) => {
    forecastErrorMessage.value = reason
  })

await useFetchListBankAccounts()
  .catch((reason) => {
    bankAccountsErrorMessage.value = reason
  })

// Always fetch forecast details to enable VAT scaling with performance slider
await useFetchListForecastDetails(forecastMonthsComputed.value)
  .catch((reason) => {
    forecastDetailsErrorMessage.value = reason
  })

const isLoading = ref(false)
const isSavingExclusions = ref(false)
const chartData = computed(() => setChartData(
  months.value,
  saldos.value.map(s => AmountToFloat(s.amount)),
  compareForecast.value
    ? { label: compareForecast.value.scenario.name, saldos: compareCells.value.map(c => c ? AmountToFloat(c.saldo.amount) : null) }
    : undefined,
))
const chartOptions = computed(() => getChartOptions())
const hasPendingExclusionChanges = computed(() => Object.keys(forecastExclusionChanges.value).length > 0)

const localMonth = computed(() => localFormatter.format(new Date()))
const months = computed(() => {
  return forecasts.value.map(f => utcFormatter.format(Date.parse(f.data.month)))
})
const latestUpdate = computed(() => {
  const forecastWithUpdatedAt = forecasts.value.find(f => f.updatedAt != null)
  if (forecastWithUpdatedAt) {
    return DateStringToFormattedDateTime(forecastWithUpdatedAt.updatedAt, false)
  }
  return '-'
})
const hasNoDataInCurrentMonth = computed(() => {
  return !months.value.includes(localMonth.value)
})

const onCalculateForecast = async () => {
  isLoading.value = true
  forecastCalculateErrorMessage.value = ''
  forecastErrorMessage.value = ''
  forecastDetailsErrorMessage.value = ''
  try {
    await calculateForecast()
    await listForecasts(forecastMonthsComputed.value)
    // Always fetch forecast details to enable VAT scaling with performance slider
    await listForecastDetails(forecastMonthsComputed.value)
  }
  catch (reason) {
    if (typeof reason === 'string' && reason.includes('Prognose Details')) {
      forecastDetailsErrorMessage.value = reason
    }
    else if (typeof reason === 'string' && reason.includes('Berechnen')) {
      forecastCalculateErrorMessage.value = reason
    }
    else if (typeof reason === 'string' && reason.includes('Prognose')) {
      forecastErrorMessage.value = reason
    }
    else {
      forecastCalculateErrorMessage.value = 'Fehler beim Aktualisieren der Prognose'
    }
  }
  finally {
    isLoading.value = false
  }
}

// Details only describe the active scenario, so a comparison hides them without touching the saved setting
// Follows the open state so details and comparison rows animate in the same frame
const isComparing = computed(() => isCompareOpen.value)
const showRevenueDetails = computed(() => forecastShowRevenueDetails.value && !isComparing.value)
const showExpenseDetails = computed(() => forecastShowExpenseDetails.value && !isComparing.value)
const detailToggleTooltip = computed(() => isComparing.value
  ? { value: 'Details sind während eines Vergleichs ausgeblendet', showDelay: 0 }
  : { value: '', disabled: true })

const onToggleRevenueDetails = () => {
  if (isComparing.value) {
    return
  }
  setForecastShowRevenueDetails(!forecastShowRevenueDetails.value)
  // ForecastDetails are already loaded, no need to fetch again
}

const onToggleExpenseDetails = () => {
  if (isComparing.value) {
    return
  }
  setForecastShowExpenseDetails(!forecastShowExpenseDetails.value)
  // ForecastDetails are already loaded, no need to fetch again
}

watch(forecastMonthsComputed, (value) => {
  isLoading.value = true
  listForecasts(value)
    .then(async () => {
      // Always fetch forecast details to enable VAT scaling with performance slider
      await listForecastDetails(value)
        .catch((reason) => {
          forecastDetailsErrorMessage.value = reason
        })
    })
    .catch((reason) => {
      forecastErrorMessage.value = reason
    })
    .finally(() => isLoading.value = false)
})

const onSaveForecastExclusionChanges = async () => {
  if (!hasPendingExclusionChanges.value) {
    return
  }
  isSavingExclusions.value = true
  try {
    await applyForecastExclusionChanges()
    await onCalculateForecast().catch(() => {})
    clearForecastExclusionChanges()
    toast.add({
      summary: 'Gespeichert',
      detail: 'Änderungen an der Prognose wurden übernommen',
      severity: 'success',
      life: Config.TOAST_LIFE_TIME_MEDIUM,
    })
  }
  catch (reason) {
    toast.add({
      summary: 'Fehler',
      detail: reason instanceof Error ? reason.message : typeof reason === 'string' ? reason : 'Änderungen konnten nicht gespeichert werden',
      severity: 'error',
      life: Config.TOAST_LIFE_TIME_MEDIUM,
    })
  }
  finally {
    isSavingExclusions.value = false
  }
}

const onCancelForecastExclusionChanges = () => {
  if (!hasPendingExclusionChanges.value) {
    return
  }
  clearForecastExclusionChanges()
  toast.add({
    summary: 'Verworfen',
    detail: 'Änderungen wurden verworfen',
    severity: 'info',
    life: Config.TOAST_LIFE_TIME_SHORT,
  })
}

const revenueCategories = computed(() => {
  const categories: ForecastDetailRevenueExpenseResponse[] = []

  forecastDetails.value.forEach((data) => {
    (data.revenue || []).forEach((item) => {
      if (!categories.find(c => c.name === item.name)) {
        categories.push(item)
      }
    })
  })

  return categories
    .filter(c => forecastShowZeroRows.value
      || !IsForecastCategoryAlwaysZero(forecastDetails.value, c.name, 'revenue'))
    .sort((a, b) => a.name.localeCompare(b.name))
})
const expenseCategories = computed(() => {
  const categories: ForecastDetailRevenueExpenseResponse[] = []

  forecastDetails.value.forEach((data) => {
    (data.expense || []).forEach((item) => {
      if (!categories.find(c => c.name === item.name)) {
        categories.push(item)
      }
    })
  })

  return categories
    .filter(c => forecastShowZeroRows.value
      || !IsForecastCategoryAlwaysZero(forecastDetails.value, c.name, 'expense'))
    .sort((a, b) => a.name.localeCompare(b.name))
})
// Helper function to find VAT amount recursively in expense tree
const findVATAmount = (expenses: ForecastDetailRevenueExpenseResponse[]): number => {
  for (const expense of expenses) {
    if (expense.name === 'Mwst.') {
      const childrenAmount = expense.children
        ? expense.children.reduce((sum: number, child: ForecastDetailRevenueExpenseResponse) => sum + findVATAmount([child]), 0)
        : 0
      return (expense.amount || 0) + childrenAmount
    }
    if (expense.children && expense.children.length > 0) {
      const childVAT = findVATAmount(expense.children)
      if (childVAT !== 0) {
        return childVAT
      }
    }
  }
  return 0
}

// Monthly rows of one scenario: revenue and VAT scale with the performance slider
const calculateForecastRows = (list: ForecastResponse[], details: ForecastDetailResponse[], startSaldo: number, factor: number) => {
  let saldo = startSaldo
  return list.map((f) => {
    const revenue = f.data.revenue * factor
    let expense = f.data.expense
    // Only scale VAT if forecastDetails are loaded
    if (details.length > 0) {
      const detail = details.find(fd => fd.month === f.data.month)
      const vatAmount = detail?.expense ? findVATAmount(detail.expense) : 0
      expense = f.data.expense - vatAmount + vatAmount * factor
    }
    const cashflow = revenue + expense
    saldo += cashflow
    return { month: f.data.month, revenue, expense, cashflow, saldo }
  })
}
const formatAmount = (amount: number) => ({
  amount,
  formatted: NumberToFormattedCurrency(AmountToFloat(amount), getOrganisationCurrencyLocaleCode.value),
})

const performanceFactor = computed(() => forecastPerformance.value / 100)
const forecastRows = computed(() => calculateForecastRows(forecasts.value, forecastDetails.value, totalBankSaldoInCHF.value, performanceFactor.value))
const revenues = computed(() => forecastRows.value.map(r => formatAmount(r.revenue)))
const expenses = computed(() => forecastRows.value.map(r => formatAmount(r.expense)))
const cashflows = computed(() => forecastRows.value.map(r => formatAmount(r.cashflow)))
const saldos = computed(() => forecastRows.value.map(r => formatAmount(r.saldo)))

// Comparison with another scenario: its stored forecast, never recalculated from here
const { scenarios, getScenarioForecast } = useScenarios()
const { convertAmountToRate } = useGlobalData()
const compareScenarioId = ref<number | null>(null)
const compareForecast = ref<ScenarioForecastResponse | null>(null)
const compareErrorMessage = ref('')
const compareOptions = computed(() => scenarios.value.filter(s => !s.isCurrent))

const loadCompareForecast = async () => {
  compareErrorMessage.value = ''
  if (compareScenarioId.value == null) {
    compareForecast.value = null
    return
  }
  try {
    compareForecast.value = await getScenarioForecast(compareScenarioId.value, forecastMonthsComputed.value)
  }
  catch (reason) {
    compareForecast.value = null
    compareErrorMessage.value = typeof reason === 'string' ? reason : 'Vergleich konnte nicht geladen werden'
  }
}
watch([compareScenarioId, forecastMonthsComputed], loadCompareForecast)
// Drop the selection once the compared scenario is gone or became the active one
watch(compareOptions, (options) => {
  if (compareScenarioId.value != null && !options.some(o => o.id === compareScenarioId.value)) {
    compareScenarioId.value = null
  }
})

// Keeps the last comparison so its rows can collapse smoothly after clearing
const shownCompareForecast = ref<ScenarioForecastResponse | null>(null)
const isCompareOpen = ref(false)

const compareBar = ref<HTMLElement | null>(null)
const findScrollParent = (element: HTMLElement): HTMLElement => {
  for (let node = element.parentElement; node; node = node.parentElement) {
    const overflowY = getComputedStyle(node).overflowY
    if ((overflowY === 'auto' || overflowY === 'scroll') && node.scrollHeight > node.clientHeight) {
      return node
    }
  }
  return document.scrollingElement as HTMLElement
}
// Holds the compare bar still on screen while the rows above it collapse or expand
const pinCompareBar = () => {
  const element = compareBar.value
  if (!element) {
    return
  }
  const scroller = findScrollParent(element)
  const startTop = element.getBoundingClientRect().top
  const until = performance.now() + 450
  const step = () => {
    scroller.scrollTop += element.getBoundingClientRect().top - startTop
    if (performance.now() < until) {
      requestAnimationFrame(step)
    }
  }
  requestAnimationFrame(step)
}

watch(compareForecast, (value, oldValue) => {
  // Runs before the re-render, so the bar position is still the old one
  if ((value == null) !== (oldValue == null)) {
    pinCompareBar()
  }
  if (!value) {
    isCompareOpen.value = false
    return
  }
  const wasShown = shownCompareForecast.value != null
  shownCompareForecast.value = value
  // A fresh wrapper needs one frame at 0fr, else it opens without the transition
  if (!wasShown) {
    requestAnimationFrame(() => requestAnimationFrame(() => isCompareOpen.value = true))
  }
  else {
    isCompareOpen.value = true
  }
})
const onCompareTransitionEnd = (event: TransitionEvent) => {
  if (event.target === event.currentTarget && !isCompareOpen.value) {
    shownCompareForecast.value = null
  }
}

const compareRows = computed(() => {
  const shown = shownCompareForecast.value
  if (!shown) {
    return null
  }
  const startSaldo = (shown.bankAccounts ?? []).reduce((sum, ba) => sum + convertAmountToRate(ba.amount, ba.currency.code), 0)
  return calculateForecastRows(shown.forecasts ?? [], shown.forecastDetails ?? [], startSaldo, performanceFactor.value)
})
// Aligned to the active months, a month missing in the compared scenario stays empty
const compareCells = computed(() => {
  const rows = compareRows.value
  if (!rows) {
    return []
  }
  return forecastRows.value.map((row) => {
    const match = rows.find(c => c.month === row.month)
    return match ? { saldo: formatAmount(match.saldo), diff: formatAmount(match.saldo - row.saldo) } : null
  })
})
const compareLatestUpdate = computed(() => {
  const forecastWithUpdatedAt = compareForecast.value?.forecasts?.find(f => f.updatedAt != null)
  return forecastWithUpdatedAt ? DateStringToFormattedDateTime(forecastWithUpdatedAt.updatedAt, false) : '-'
})
</script>
