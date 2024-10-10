import {CreateSettingsCookie} from "~/utils/cookie-helper";
import {ref} from "vue";

const transactionDisplay = ref<'grid'|'list'>('grid')
const forecastPerformance = ref(100)
const forecastMonths = ref(12)

export default function useSettings() {
    const transactionDisplayCookie = CreateSettingsCookie('transaction-display')
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

    watch(forecastMonths, (value) => {
        forecastMonthsCookie.value = value.toString()
    })

    return {
        toggleDisplayType,
        transactionDisplay,
        forecastPerformance,
        forecastMonths,
    };
}
