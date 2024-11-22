import {CreateSettingsCookie} from "~/utils/cookie-helper";
import type {SortOrderType, TransactionSortByType} from "~/utils/types";
import {SortOrderOptions, TransactionSortByOptions} from "~/utils/types";
import {SymbolKind} from "vscode-languageserver-types";

export default function useSettings() {
    const forecastShowRevenueDetails = useState('forecastShowRevenueDetails', () => false)
    const forecastShowExpenseDetails = useState('forecastShowExpenseDetails', () => false)
    const transactionDisplay = useState<'grid'|'list'>('transactionDisplay', () => 'grid')
    const transactionSortBy = useState<TransactionSortByType>('transactionSortBy', () => 'name')
    const transactionSortOrder = useState<SortOrderType>('transactionSortOrder', () => 'ASC')
    const forecastPerformance = useState('forecastPerformance', () => 100)
    // 13 to include from current to the same month
    const forecastMonths = useState('forecastMonths', () => 13)

    const forecastShowRevenueDetailsCookie = CreateSettingsCookie('forecast-revenue-details')
    const forecastShowExpenseDetailsCookie = CreateSettingsCookie('forecast-expense-details')
    const transactionDisplayCookie = CreateSettingsCookie('transaction-display')
    const transactionSortByCookie = CreateSettingsCookie('transaction-sort-by')
    const transactionSortOrderCookie = CreateSettingsCookie('transaction-sort-order')
    const forecastPerformanceCookie = CreateSettingsCookie('forecast-performance')
    const forecastMonthsCookie = CreateSettingsCookie('forecast-months')

    if (forecastShowRevenueDetailsCookie.value !== undefined) {
        const val = forecastShowRevenueDetailsCookie.value
        if (val == 'true') {
            forecastShowRevenueDetails.value = true
        }
        else if (val == 'false') {
            forecastShowRevenueDetails.value = false
        } else if (typeof val === 'boolean') {
            // Can be boolean for some reason
            forecastShowRevenueDetails.value = val
        }
    } else {
        forecastShowRevenueDetails.value = false
    }

    if (forecastShowExpenseDetailsCookie.value !== undefined) {
        const val = forecastShowExpenseDetailsCookie.value
        if (val == 'true') {
            forecastShowExpenseDetails.value = true
        }
        else if (val == 'false') {
            forecastShowExpenseDetails.value = false
        } else if (typeof val === 'boolean') {
            // Can be boolean for some reason
            forecastShowExpenseDetails.value = val
        }
    } else {
        forecastShowExpenseDetails.value = false
    }

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
        // 13 to include from current to the same month
        forecastMonths.value = 13
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
    watch(forecastShowRevenueDetails, (value) => {
        forecastShowRevenueDetailsCookie.value = value.toString()
    })

    watch(forecastShowExpenseDetails, (value) => {
        forecastShowExpenseDetailsCookie.value = value.toString()
    })

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
        forecastShowRevenueDetails,
        forecastShowExpenseDetails,
        transactionDisplay,
        transactionSortBy,
        transactionSortOrder,
        forecastPerformance,
        forecastMonths,
    };
}
