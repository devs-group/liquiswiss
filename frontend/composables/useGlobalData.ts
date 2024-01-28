import { ref } from 'vue';
import type {StrapiCategory} from "~/models/category";
import type {StrapiCurrency} from "~/models/currency";

const categories = ref<StrapiCategory[]>([]);
const currencies = ref<StrapiCurrency[]>([]);

export default function useGlobalData() {
    const fetchCategories = async () => {
        try {
            const {data} = await useFetch('/api/global/category');
            categories.value = data.value ?? [];
        } catch (error) {
            console.error('Error fetching global data:', error);
        }
    }

    const fetchCurrencies = async () => {
        try {
            const {data} = await useFetch('/api/global/currency');
            currencies.value = data.value ?? [];
        } catch (error) {
            console.error('Error fetching global data:', error);
        }
    }

    const getCategoryNameFromId = (id: number) => {
        return categories.value.find((category) => category.id === id)?.attributes.name ?? ''
    }

    const getCurrencyCodeFromId = (id: number) => {
        return currencies.value.find((currency) => currency.id === id)?.attributes.code ?? ''
    }

    const getCurrencyLocaleCodeFromId = (id: number) => {
        return currencies.value.find((currency) => currency.id === id)?.attributes.localeCode ?? ''
    }

    return {
        categories,
        fetchCategories,
        getCategoryNameFromId,
        currencies,
        fetchCurrencies,
        getCurrencyCodeFromId,
        getCurrencyLocaleCodeFromId,
    };
}
