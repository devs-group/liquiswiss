import {ref} from 'vue';
import type {ForecastResponse} from "~/models/forecast";

const limitForecasts = ref(12)
const forecasts = ref<ForecastResponse[]>([]);

export default function useForecasts() {
    const listForecasts = async ()  => {
        const {data, status} = await useFetch<ForecastResponse[]>('/api/forecasts', {
            method: 'GET',
            query: {
                limit: limitForecasts.value,
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
        limitForecasts,
        forecasts,
    };
}
