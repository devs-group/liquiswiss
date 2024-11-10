import {ref} from 'vue';
import type {ForecastDetailResponse, ForecastResponse} from "~/models/forecast";

const forecasts = ref<ForecastResponse[]>([]);
const forecastDetails = ref<ForecastDetailResponse[]>([]);

export default function useForecasts() {
    const listForecasts = async (months: number)  => {
        const {data, status} = await useFetch<ForecastResponse[]>('/api/forecasts', {
            method: 'GET',
            query: {
                limit: months,
            }
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Laden der Prognose')
        } else {
            if (data.value) {
                forecasts.value = data.value
            } else {
                forecasts.value = []
            }
        }
        return Promise.resolve()
    }

    const listForecastDetails = async (months: number)  => {
        const {data, status} = await useFetch<ForecastDetailResponse[]>('/api/forecast-details', {
            method: 'GET',
            query: {
                limit: months,
            }
        });

        if (status.value === 'error') {
            return Promise.reject('Fehler beim Laden der Prognose Details')
        } else {
            if (data.value) {
                forecastDetails.value = data.value
            } else {
                forecastDetails.value = []
            }
        }
        return Promise.resolve()
    }


    return {
        listForecasts,
        listForecastDetails,
        forecasts,
        forecastDetails,
    };
}
