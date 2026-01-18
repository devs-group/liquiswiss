<template>
  <div class="flex flex-col gap-2">
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
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
          <div class="flex items-center gap-2">
            <p class="text-sm">
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
          <div class="flex items-center gap-2">
            <p class="text-sm">
              Performance ({{ forecastPerformance }}%):
            </p>
            <Slider
              :model-value="forecastPerformance"
              class="w-56"
              @update:model-value="setForecastPerformance"
            />
          </div>
        </div>

        <div class="flex flex-col items-end gap-1">
          <Button
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

        <div class="relative flex flex-col overflow-x-auto pb-2">
          <div class="grid grid-cols-12 items-center">
            <div class="flex items-center col-span-full">
              <div class="border-t border-b border-l border-zinc-600 dark:border-zinc-400 bg-zinc-300 dark:bg-zinc-700 p-2 min-w-28">
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
                class="group flex gap-1 cursor-pointer border-b border-l border-zinc-600 dark:border-zinc-400 bg-zinc-400 dark:bg-zinc-600 p-2 min-w-28"
                @click="onToggleRevenueDetails"
              >
                <p class="text-xs font-bold">
                  Einnahmen
                </p>
                <i
                  class="pi opacity-0 group-hover:opacity-100 transition-opacity"
                  :class="{ 'pi-sort-down': forecastShowRevenueDetails, 'pi-sort-up': !forecastShowRevenueDetails }"
                />
              </div>
              <div
                v-for="(revenue, i) in revenues"
                :key="i"
                class="w-full border-b border-l last:border-r border-zinc-600 dark:border-zinc-400 bg-liqui-green p-2 min-w-40"
              >
                <p class="text-xs text-center">
                  {{ revenue.formatted }} {{ getOrganisationCurrencyCode }}
                </p>
              </div>
            </div>

            <template v-if="forecastShowRevenueDetails">
              <NestedForecastCategory
                v-for="category in revenueCategories"
                :key="category.name"
                :category="category"
                forecast-type="revenue"
                :forecast-details="forecastDetails"
                :currency-code="getOrganisationCurrencyCode"
                :performance-factor="forecastPerformance / 100"
              />
            </template>

            <div class="flex items-center col-span-full">
              <div
                class="group flex gap-1 cursor-pointer border-b border-l border-zinc-600 dark:border-zinc-400 bg-zinc-400 dark:bg-zinc-600 p-2 min-w-28"
                @click="onToggleExpenseDetails"
              >
                <p class="text-xs font-bold">
                  Ausgaben
                </p>
                <i
                  class="pi opacity-0 group-hover:opacity-100 transition-opacity"
                  :class="{ 'pi-sort-down': forecastShowExpenseDetails, 'pi-sort-up': !forecastShowExpenseDetails }"
                />
              </div>
              <div
                v-for="(expense, i) in expenses"
                :key="i"
                class="w-full border-b border-l last:border-r border-zinc-600 dark:border-zinc-400 bg-liqui-red p-2 min-w-40"
              >
                <p class="text-xs text-center">
                  {{ expense.formatted }} {{ getOrganisationCurrencyCode }}
                </p>
              </div>
            </div>

            <template v-if="forecastShowExpenseDetails">
              <NestedForecastCategory
                v-for="category in expenseCategories"
                :key="category.name"
                :category="category"
                forecast-type="expense"
                :forecast-details="forecastDetails"
                :currency-code="getOrganisationCurrencyCode"
                :performance-factor="forecastPerformance / 100"
              />
            </template>

            <div class="flex items-center col-span-full">
              <div class="cursor-default border-b border-l border-zinc-600 dark:border-zinc-400 bg-zinc-300 dark:bg-zinc-700 p-2 min-w-28">
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
              <div class="cursor-default border-b border-l border-zinc-600 dark:border-zinc-400 bg-zinc-300 dark:bg-zinc-700 p-2 min-w-28">
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
          </div>
        </div>
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
import type { ForecastDetailRevenueExpenseResponse } from '~/models/forecast'
import FullProgressSpinner from '~/components/FullProgressSpinner.vue'
import { Config } from '~/config/config'

useHead({
  title: 'Prognose',
})

const utcFormatter = new Intl.DateTimeFormat(Constants.BASE_LOCALE_CODE, { month: 'long', year: 'numeric', timeZone: 'UTC' })
const localFormatter = new Intl.DateTimeFormat(Constants.BASE_LOCALE_CODE, { month: 'long', year: 'numeric' })

const { getOrganisationCurrencyCode, getOrganisationCurrencyLocaleCode } = useAuth()
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
const { forecastPerformance, forecastMonths, forecastShowRevenueDetails, forecastShowExpenseDetails, setForecastShowRevenueDetails, setForecastShowExpenseDetails, setForecastPerformance, setForecastMonths } = useSettings()
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

const onToggleRevenueDetails = () => {
  setForecastShowRevenueDetails(!forecastShowRevenueDetails.value)
  // ForecastDetails are already loaded, no need to fetch again
}

const onToggleExpenseDetails = () => {
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

  return categories.sort((a, b) => a.name.localeCompare(b.name))
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

  return categories.sort((a, b) => a.name.localeCompare(b.name))
})
const revenues = computed(() => forecasts.value.map((f) => {
  const revenue = f.data.revenue * (forecastPerformance.value / 100)
  return {
    amount: revenue,
    formatted: NumberToFormattedCurrency(AmountToFloat(revenue), getOrganisationCurrencyLocaleCode.value),
  }
}))
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

const expenses = computed(() => {
  return forecasts.value.map((f) => {
    let scaledExpense = f.data.expense

    // Only scale VAT if forecastDetails are loaded
    if (forecastDetails.value.length > 0) {
      // Get VAT amount for this month from forecastDetails
      const forecastDetail = forecastDetails.value.find(fd => fd.month === f.data.month)
      let vatAmount = 0

      if (forecastDetail && forecastDetail.expense) {
        vatAmount = findVATAmount(forecastDetail.expense)
      }

      // Calculate scaled expense: non-VAT expenses + scaled VAT
      const totalExpense = f.data.expense
      const nonVATExpense = totalExpense - vatAmount
      const scaledVAT = vatAmount * (forecastPerformance.value / 100)
      scaledExpense = nonVATExpense + scaledVAT
    }

    return {
      amount: scaledExpense,
      formatted: NumberToFormattedCurrency(AmountToFloat(scaledExpense), getOrganisationCurrencyLocaleCode.value),
    }
  })
})
const cashflows = computed(() => {
  return revenues.value.map((r, index) => {
    const e = expenses.value[index]
    const cashflow = r.amount + e.amount
    return {
      amount: cashflow,
      formatted: NumberToFormattedCurrency(AmountToFloat(cashflow), getOrganisationCurrencyLocaleCode.value),
    }
  })
})
const saldos = computed(() => {
  let totalBankSaldo = totalBankSaldoInCHF.value
  return cashflows.value.map((c) => {
    totalBankSaldo += c.amount
    return {
      amount: totalBankSaldo,
      formatted: NumberToFormattedCurrency(AmountToFloat(totalBankSaldo), getOrganisationCurrencyLocaleCode.value),
    }
  })
})
</script>
