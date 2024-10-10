import {ref} from 'vue';
import type {ForecastResponse} from "~/models/forecast";

const forecasts = ref<ForecastResponse[]>([]);

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


    return {
        listForecasts,
        forecasts,
    };
}
