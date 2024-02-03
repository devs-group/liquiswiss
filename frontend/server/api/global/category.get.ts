import type {StrapiCategory, Strapi_ListResponse_Category} from "~/models/category";

export default defineEventHandler(async () => {
    const config = useRuntimeConfig()
    const locale = 'de'
    const resp = await $fetch<Strapi_ListResponse_Category>(`/categories?locale=${locale}`, {
        baseURL: config.apiHost,
        headers: {
            'Authorization': `Bearer ${config.strapiApiKey}`
        }
    })
    return resp.data as StrapiCategory[]
})
