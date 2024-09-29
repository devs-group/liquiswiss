import {ref} from 'vue';
import type {CategoryResponse, ListCategoryResponse} from "~/models/category";
import type {CurrencyResponse, ListCurrencyResponse} from "~/models/currency";

const categories = ref<CategoryResponse[]>([]);
const currencies = ref<CurrencyResponse[]>([]);

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
            console.error('Error fetching global data:', error);
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
            console.error('Error fetching global data:', error);
        }
    }


    return {
        categories,
        fetchCategories,
        currencies,
        fetchCurrencies,
    };
}
