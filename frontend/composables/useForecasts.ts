import type {ForecastDetailResponse, ForecastResponse} from "~/models/forecast";

export default function useForecasts() {
    const forecasts = useState<ForecastResponse[]>('forecasts', () => []);
    const forecastDetails = useState<ForecastDetailResponse[]>('forecastDetails', () => []);

    const useFetchListForecast = async (months: number) => {
        const {data, error} = await useFetch<ForecastResponse[]>('/api/forecasts', {
            method: 'GET',
            query: {
                limit: months,
            }
        });
        if (error.value) {
            return Promise.reject('Prognose konnten nicht geladen werden')
        }
        setForecasts(data.value, false)
    }

    const listForecasts = async (months: number)  => {
        try {
            const data = await $fetch<ForecastResponse[]>('/api/forecasts', {
                method: 'GET',
                query: {
                    limit: months,
                }
            });
            setForecasts(data, false)
        } catch (err) {
            return Promise.reject('Fehler beim Laden der Prognose')
        }
    }

    const useFetchListForecastDetails = async (months: number)  => {
        const {data, error} = await useFetch<ForecastDetailResponse[]>('/api/forecast-details', {
            method: 'GET',
            query: {
                limit: months,
            }
        });
        if (error.value) {
            return Promise.reject('Fehler beim Laden der Prognose Details')
        }
        setForecastDetails(data.value, false)
    }

    const listForecastDetails = async (months: number)  => {
        try {
            const data = await $fetch<ForecastDetailResponse[]>('/api/forecast-details', {
                method: 'GET',
                query: {
                    limit: months,
                }
            });
            setForecastDetails(data, false)
        } catch (err) {
            return Promise.reject('Fehler beim Laden der Prognose Details')
        }
    }

    const setForecasts = (data: ForecastResponse[]|null, append: boolean) => {
        if (data) {
            if (append) {
                forecasts.value = forecasts.value.concat(data ?? [])
            } else {
                forecasts.value = data
            }
        } else {
            forecasts.value = []
        }
    }

    const setForecastDetails = (data: ForecastDetailResponse[]|null, append: boolean) => {
        if (data) {
            if (append) {
                forecastDetails.value = forecastDetails.value.concat(data ?? [])
            } else {
                forecastDetails.value = data
            }
        } else {
            forecastDetails.value = []
        }
    }

    return {
        forecasts,
        forecastDetails,
        useFetchListForecast,
        listForecasts,
        useFetchListForecastDetails,
        listForecastDetails,
        setForecasts,
    };
}
