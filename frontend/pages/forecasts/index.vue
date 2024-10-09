<template>
  <div class="flex flex-col gap-4">
    <div class="grid grid-cols-1 sm:grid-cols-3">
      <div class="flex items-center gap-2">
        <p class="text-sm">Zeitraum:</p>
        <Dropdown v-model="limitForecasts"
                  :options="monthChoices" option-label="label" option-value="value"
                  empty-message="Keine Auswahl gefunden"
        />
      </div>
      <div class="flex items-center gap-2">
        <p class="text-sm">Performance ({{forecastPerformance}}%): </p>
        <Slider v-model="forecastPerformance" class="w-56" />
      </div>
    </div>

    <div class="flex flex-col overflow-x-auto p-4">
      <div class="grid grid-cols-12 items-center">
        <div class="flex items-center col-span-full">
          <div class="border-t border-b border-l border-gray-600 bg-gray-300 p-2 min-w-24">
            <p class="text-xs">&nbsp;</p>
          </div>
          <div v-for="month in months" class="border-t border-b border-l last:border-r border-gray-600 bg-gray-300 p-2 min-w-40">
            <p class="text-xs text-center font-bold">{{ month }}</p>
          </div>
        </div>

        <div class="flex items-center col-span-full">
          <div class="border-b border-l border-gray-600 bg-gray-300 p-2 min-w-24">
            <p class="text-xs">Einnahmen</p>
          </div>
          <div v-for="revenue in revenues" class="border-b border-l last:border-r border-gray-600 bg-green-100 p-2 min-w-40">
            <p class="text-xs text-center">
              {{revenue.formatted}} CHF
            </p>
          </div>
        </div>

        <div class="flex items-center col-span-full">
          <div class="border-b border-l border-gray-600 bg-gray-300 p-2 min-w-24">
            <p class="text-xs">Ausgaben</p>
          </div>
          <div v-for="expense in expenses" class="border-b border-l last:border-r border-gray-600 bg-red-100 p-2 min-w-40">
            <p class="text-xs text-center">
              {{expense.formatted}} CHF
            </p>
          </div>
        </div>

        <div class="flex items-center col-span-full">
          <div class="border-b border-l border-gray-600 bg-gray-300 p-2 min-w-24">
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
          <div class="border-b border-b border-l border-gray-600 bg-gray-300 p-2 min-w-24">
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
    </div>
  </div>
</template>

<script setup lang="ts">
const formatter = new Intl.DateTimeFormat('de-CH', { month: 'long', year: '2-digit' })
const monthChoices = [
  {
    label: '6 Monate',
    value: 6
  },
  {
    label: '12 Monate',
    value: 12
  },
  {
    label: '24 Monate',
    value: 24
  },
  {
    label: '36 Monate',
    value: 36
  }
]

const {listForecasts, limitForecasts, forecastPerformance, forecasts} = useForecasts()
const {listBankAccounts, bankAccounts} = useBankAccounts()

await listForecasts()
await listBankAccounts()

const now = new Date()
const months = computed(() => {
  return Array.from({length: limitForecasts.value}, (_, i) => {
    const nextMonth = new Date(now.getFullYear(), now.getMonth() + i)
    return formatter.format(nextMonth)
  })
})

watch(limitForecasts, () => {
  listForecasts()
})

const revenues = computed(() => forecasts.value.map(f => {
  const revenue = f.revenue * (forecastPerformance.value / 100)
  return {
    amount: revenue,
    formatted: NumberToFormattedCurrency(AmountToFloat(revenue), 'de-CH'),
  }
}))
const expenses = computed(() => forecasts.value.map(f => ({
  amount: f.expense,
  formatted: NumberToFormattedCurrency(AmountToFloat(f.expense), 'de-CH'),
})))
const cashflows = computed(() => {
  return revenues.value.map((r, index) => {
    const e = expenses.value[index]
    const cashflow = r.amount + e.amount
    return {
      amount: cashflow,
      formatted: NumberToFormattedCurrency(AmountToFloat(cashflow), 'de-CH'),
    }
  })
  // forecasts.value.map(f => ({
  //   amount: f.cashflow,
  //   formatted: NumberToFormattedCurrency(AmountToFloat(f.cashflow), 'de-CH'),
  // }))
})
const saldos = computed(() => {
  let totalBankSaldo = bankAccounts.value.reduce((previousValue, currentValue) => previousValue + currentValue.amount, 0)
  return cashflows.value.map(c => {
    totalBankSaldo += c.amount
    return {
      amount: totalBankSaldo,
      formatted: NumberToFormattedCurrency(AmountToFloat(totalBankSaldo), 'de-CH'),
    }
  })
})
</script>