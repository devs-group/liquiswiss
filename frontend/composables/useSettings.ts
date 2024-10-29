import {CreateSettingsCookie} from "~/utils/cookie-helper";
import {ref} from "vue";
import type {SortOrderType, TransactionSortByType} from "~/utils/types";
import {SortOrderOptions, TransactionSortByOptions} from "~/utils/types";

const transactionDisplay = ref<'grid'|'list'>('grid')
const transactionSortBy = ref<TransactionSortByType>('name')
const transactionSortOrder = ref<SortOrderType>('ASC')
const forecastPerformance = ref(100)
const forecastMonths = ref(12)

export default function useSettings() {
    const transactionDisplayCookie = CreateSettingsCookie('transaction-display')
    const transactionSortByCookie = CreateSettingsCookie('transaction-sort-by')
    const transactionSortOrderCookie = CreateSettingsCookie('transaction-sort-order')
    const forecastPerformanceCookie = CreateSettingsCookie('forecast-performance')
    const forecastMonthsCookie = CreateSettingsCookie('forecast-months')

    if (transactionDisplayCookie.value !== undefined) {
        const val = transactionDisplayCookie.value
        if (val == 'list' || val == 'grid') {
            transactionDisplay.value = val
        }
    } else {
        transactionDisplay.value = 'grid'
    }

    if (transactionSortByCookie.value !== undefined) {
        const val = transactionSortByCookie.value as typeof transactionSortBy.value
        if (TransactionSortByOptions.includes(val)) {
            transactionSortBy.value = val
        }
    } else {
        transactionSortBy.value = 'name'
    }

    if (transactionSortOrderCookie.value !== undefined) {
        const val = transactionSortOrderCookie.value as typeof transactionSortOrder.value
        if (SortOrderOptions.includes(val)) {
            transactionSortOrder.value = val
        }
    } else {
        transactionSortOrder.value = 'ASC'
    }

    if (forecastPerformanceCookie.value !== undefined) {
        const val = forecastPerformanceCookie.value
        if (val !== null && Number.isInteger(Number.parseInt(val))) {
            forecastPerformance.value = Number.parseInt(val)
        }
    } else {
        forecastPerformance.value = 100
    }

    if (forecastMonthsCookie.value !== undefined) {
        const val = forecastMonthsCookie.value
        if (val !== null && Number.isInteger(Number.parseInt(val))) {
            forecastMonths.value = Number.parseInt(val)
        }
    } else {
        forecastMonths.value = 12
    }

    const toggleDisplayType = () => {
        switch (transactionDisplay.value) {
            case 'grid':
                transactionDisplay.value = 'list'
                break
            case 'list':
                transactionDisplay.value = 'grid'
        }
        transactionDisplayCookie.value = transactionDisplay.value
    };

    // Watchers
    watch(forecastPerformance, (value) => {
        forecastPerformanceCookie.value = value.toString()
    })

    watch(transactionSortBy, (value) => {
        transactionSortByCookie.value = value.toString()
    })

    watch(transactionSortOrder, (value) => {
        transactionSortOrderCookie.value = value.toString()
    })

    watch(forecastMonths, (value) => {
        forecastMonthsCookie.value = value.toString()
    })

    return {
        toggleDisplayType,
        transactionDisplay,
        transactionSortBy,
        transactionSortOrder,
        forecastPerformance,
        forecastMonths,
    };
}
