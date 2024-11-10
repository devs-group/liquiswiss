<template>
  <div class="flex flex-col gap-4">
    <div class="grid grid-cols-1 sm:grid-cols-3">
      <div class="flex items-center gap-2">
        <p class="text-sm">Zeitraum:</p>
        <Dropdown v-model="forecastMonths"
                  :options="monthChoices" option-label="label" option-value="value"
                  empty-message="Keine Auswahl gefunden"
        />
      </div>
      <div class="flex items-center gap-2">
        <p class="text-sm">Performance ({{forecastPerformance}}%): </p>
        <Slider v-model="forecastPerformance" class="w-56" />
      </div>
    </div>

    <div class="relative flex flex-col overflow-x-auto p-4">
      <div class="grid grid-cols-12 items-center">
        <div class="flex items-center col-span-full">
          <div class="border-t border-b border-l border-gray-600 bg-gray-300 p-2 min-w-28">
            <p class="text-xs">&nbsp;</p>
          </div>
          <div v-for="month in months" class="border-t border-b border-l last:border-r border-gray-600 bg-gray-300 p-2 min-w-40">
            <p class="text-xs text-center font-bold">{{ month }}</p>
          </div>
        </div>

        <div class="flex items-center col-span-full">
          <div @click="onToggleRevenueDetails" class="flex gap-1 cursor-pointer border-b border-l border-gray-600 bg-gray-300 p-2 min-w-28">
            <p class="text-xs">Einnahmen</p>
            <i class="pi" :class="{'pi-sort-down': forecastShowRevenueDetails, 'pi-sort-up': !forecastShowRevenueDetails}"></i>
          </div>
          <div v-for="revenue in revenues" class="border-b border-l last:border-r border-gray-600 bg-green-100 p-2 min-w-40">
            <p class="text-xs text-center">
              {{revenue.formatted}} CHF
            </p>
          </div>
        </div>

        <template v-if="forecastShowRevenueDetails">
          <div v-for="category of revenueCategories" class="flex items-center col-span-full">
            <div class="flex gap-1 cursor-pointer border-b border-l border-gray-600 bg-gray-300 p-1 min-w-28">
              <p class="w-full text-xs text-right">{{category}}</p>
            </div>
            <div v-for="data in forecastDetails" class="border-b border-l last:border-r border-gray-600 bg-gray-100 p-1 min-w-40">
              <p class="text-xs text-center">
                {{getCategoryAmount(data, category, 'revenue')}} CHF
              </p>
            </div>
          </div>
        </template>

        <div class="flex items-center col-span-full">
          <div @click="onToggleExpenseDetails" class="flex gap-1 cursor-pointer border-b border-l border-gray-600 bg-gray-300 p-2 min-w-28">
            <p class="text-xs">Ausgaben</p>
            <i class="pi" :class="{'pi-sort-down': forecastShowExpenseDetails, 'pi-sort-up': !forecastShowExpenseDetails}"></i>
          </div>
          <div v-for="expense in expenses" class="border-b border-l last:border-r border-gray-600 bg-red-100 p-2 min-w-40">
            <p class="text-xs text-center">
              {{expense.formatted}} CHF
            </p>
          </div>
        </div>

        <template v-if="forecastShowExpenseDetails">
          <div v-for="category of expenseCategories" class="flex items-center col-span-full">
            <div class="flex gap-1 cursor-pointer border-b border-l border-gray-600 bg-gray-300 p-1 min-w-28">
              <p class="w-full text-xs text-right">{{category}}</p>
            </div>
            <div v-for="data in forecastDetails" class="border-b border-l last:border-r border-gray-600 bg-gray-100 p-1 min-w-40">
              <p class="text-xs text-center">
                {{getCategoryAmount(data, category, 'expense')}} CHF
              </p>
            </div>
          </div>
        </template>

        <div class="flex items-center col-span-full">
          <div class="border-b border-l border-gray-600 bg-gray-300 p-2 min-w-28">
            <p class="text-xs">Cashflow</p>
          </div>
          <div v-for="cashflow in cashflows" class="border-b border-l last:border-r border-gray-600 p-2 min-w-40"
               :class="{'text-red-600': cashflow.amount < 0, 'text-green-600': cashflow.amount > 0}">
            <p class="text-xs text-center">
              {{cashflow.formatted}} CHF
            </p>
          </div>
        </div>

        <div class="flex items-center col-span-full">
          <div class="border-b border-b border-l border-gray-600 bg-gray-300 p-2 min-w-28">
            <p class="text-xs">Saldo</p>
          </div>
          <div v-for="saldo in saldos" class="border-b border-l last:border-r border-gray-600 p-2 min-w-40"
               :class="{'bg-red-100': saldo.amount < 0, 'bg-green-100': saldo.amount > 0}">
            <p class="text-xs text-center font-bold truncate">
              {{saldo.formatted}} CHF
            </p>
          </div>
        </div>
      </div>

      <FullProgressSpinner :show="isLoading"/>
    </div>
  </div>

  <div class="bg-gray-50">
    <Chart type="line" :data="chartData" :options="chartOptions" class="h-80"/>
  </div>
</template>

<script setup lang="ts">
import Chart from "primevue/chart";
import useCharts from "~/composables/useCharts";
import {Constants} from "~/utils/constants";
import type {ForecastDetailResponse} from "~/models/forecast";
import FullProgressSpinner from "~/components/FullProgressSpinner.vue";

const formatter = new Intl.DateTimeFormat(Constants.BASE_LOCALE_CODE, { month: 'long', year: '2-digit' })
const monthChoices = [
  {
    label: '6 Monate',
    value: 6
  },
  {
    label: '12 Monate',
    value: 13
  },
  {
    label: '24 Monate',
    value: 25
  },
  {
    label: '36 Monate',
    value: 37
  }
]

const {listForecasts, listForecastDetails, forecasts, forecastDetails} = useForecasts()
const {listBankAccounts, totalBankSaldoInCHF} = useBankAccounts()
const {forecastPerformance, forecastMonths, forecastShowRevenueDetails, forecastShowExpenseDetails} = useSettings()
const {setChartData, setChartOptions} = useCharts()

await listForecasts(forecastMonths.value)
await listBankAccounts()

if (forecastShowRevenueDetails.value || forecastShowExpenseDetails.value) {
  await listForecastDetails(forecastMonths.value)
}

const isLoading = ref(false)
const chartData = computed(() => setChartData(
    months.value,
    saldos.value.map(s => AmountToFloat(s.amount)),
))
const chartOptions = setChartOptions()

const now = new Date()
const months = computed(() => {
  return Array.from({length: forecastMonths.value}, (_, i) => {
    const nextMonth = new Date(now.getFullYear(), now.getMonth() + i)
    return formatter.format(nextMonth)
  })
})

const onToggleRevenueDetails = () => {
  forecastShowRevenueDetails.value = !forecastShowRevenueDetails.value
  if (forecastShowRevenueDetails.value) {
    isLoading.value = true
    listForecastDetails(forecastMonths.value)
        .finally(() => isLoading.value = false)
  }
}

const onToggleExpenseDetails = async () => {
  forecastShowExpenseDetails.value = !forecastShowExpenseDetails.value
  if (forecastShowExpenseDetails.value) {
    isLoading.value = true
    await listForecastDetails(forecastMonths.value)
        .finally(() => isLoading.value = false)
  }
}

const getCategoryAmount = (data: ForecastDetailResponse, category: string, type: 'revenue'|'expense') => {
  const amount = data[type].find(e => e.name == category)?.amount ?? 0
  return NumberToFormattedCurrency(AmountToFloat(amount), Constants.BASE_LOCALE_CODE)
}

watch(forecastMonths, (value) => {
  isLoading.value = true
  listForecasts(value)
      .then(async () => {
        if (forecastShowRevenueDetails.value || forecastShowExpenseDetails.value) {
          await listForecastDetails(value)
              .finally(() => isLoading.value = false)
        }
      })
      .finally(() => isLoading.value = false)
})

const revenueCategories = computed(() => {
  return Array.from(new Set(
      forecastDetails.value.flatMap(data => [
        ...data.revenue.map(r => r.name),
      ])
  )).sort((a, b) => a.localeCompare(b))
})
const expenseCategories = computed(() => {
  return Array.from(new Set(
      forecastDetails.value.flatMap(data => [
        ...data.expense.map(e => e.name),
      ])
  )).sort((a, b) => a.localeCompare(b))
})
const revenues = computed(() => forecasts.value.map(f => {
  const revenue = f.revenue * (forecastPerformance.value / 100)
  return {
    amount: revenue,
    formatted: NumberToFormattedCurrency(AmountToFloat(revenue), Constants.BASE_LOCALE_CODE),
  }
}))
const expenses = computed(() => forecasts.value.map(f => ({
  amount: f.expense,
  formatted: NumberToFormattedCurrency(AmountToFloat(f.expense), Constants.BASE_LOCALE_CODE),
})))
const cashflows = computed(() => {
  return revenues.value.map((r, index) => {
    const e = expenses.value[index]
    const cashflow = r.amount + e.amount
    return {
      amount: cashflow,
      formatted: NumberToFormattedCurrency(AmountToFloat(cashflow), Constants.BASE_LOCALE_CODE),
    }
  })
})
const saldos = computed(() => {
  let totalBankSaldo = totalBankSaldoInCHF.value
  return cashflows.value.map(c => {
    totalBankSaldo += c.amount
    return {
      amount: totalBankSaldo,
      formatted: NumberToFormattedCurrency(AmountToFloat(totalBankSaldo), Constants.BASE_LOCALE_CODE),
    }
  })
})
</script>