import {ref} from 'vue';
import type {CategoryResponse, ListCategoryResponse} from "~/models/category";
import type {CurrencyResponse, ListCurrencyResponse} from "~/models/currency";
import type {FiatRateResponse} from "~/models/fiat-rate";

const categories = ref<CategoryResponse[]>([]);
const currencies = ref<CurrencyResponse[]>([]);
const fiatRates = ref<FiatRateResponse[]>([]);

export default function useGlobalData() {
    const fetchCategories = async () => {
        try {
            const {data} = await useFetch<ListCategoryResponse>('/api/categories', {
                method: 'GET',
                query: {
                    page: 1,
                    limit: 100,
                }
            })
            categories.value = data.value?.data ?? [];
        } catch (error) {
            console.error('Error fetching categories:', error);
        }
    }

    const fetchCurrencies = async () => {
        try {
            const {data} = await useFetch<ListCurrencyResponse>('/api/currencies', {
                method: 'GET',
                query: {
                    page: 1,
                    limit: 100,
                }
            })
            currencies.value = data.value?.data ?? [];
        } catch (error) {
            console.error('Error fetching currencies:', error);
        }
    }

    const fetchFiatRates = async () => {
        try {
            const {data} = await useFetch<FiatRateResponse[]>('/api/fiat-rates', {
                method: 'GET',
            })
            fiatRates.value = data.value ?? [];
        } catch (error) {
            console.error('Error fetching fiat rates:', error);
        }
    }

    const convertAmountToRate = (amount: number, currency: string) => {
        const fiatRate = fiatRates.value.find(fr => fr.target == currency)
        if (fiatRate) {
            return amount / fiatRate.rate
        }
        return amount
    }

    return {
        categories,
        fetchCategories,
        currencies,
        fetchCurrencies,
        fiatRates,
        fetchFiatRates,
        convertAmountToRate,
    };
}
