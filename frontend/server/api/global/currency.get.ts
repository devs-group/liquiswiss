import {Strapi_ListResponse_Currency, StrapiCurrency} from "~/models/currency";

export default defineEventHandler(async () => {
    const config = useRuntimeConfig()
    const resp = await $fetch<Strapi_ListResponse_Currency>(`/currencies`, {
        baseURL: config.apiHost,
        headers: {
            'Authorization': `Bearer ${config.strapiApiKey}`
        }
    })
    return resp.data as StrapiCurrency[]
})
