import {Strapi_PostResponse_Transaction, StrapiTransaction} from "~/models/transaction";

export default defineEventHandler(async (event) => {
    const body = await readBody<StrapiTransaction>(event)

    const config = useRuntimeConfig()
    return await $fetch<Strapi_PostResponse_Transaction>(`/transactions/${body.id}`, {
        baseURL: config.apiHost,
        method: 'delete',
        headers: {
            'Authorization': `Bearer ${config.strapiApiKey}`
        }
    })
})
