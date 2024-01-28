import {Strapi_PostResponse_Revenue, StrapiRevenue} from "~/models/revenue";

export default defineEventHandler(async (event) => {
    const body = await readBody<StrapiRevenue>(event)

    const config = useRuntimeConfig()
    return await $fetch<Strapi_PostResponse_Revenue>(`/revenues/${body.id}`, {
        baseURL: config.strapiApiUrl,
        method: 'delete',
        headers: {
            'Authorization': `Bearer ${config.strapiApiKey}`
        }
    })
})
