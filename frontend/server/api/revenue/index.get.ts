import {Strapi_ListResponse_Revenue, StrapiRevenue} from "~/models/revenue";
import {Strapi_RelationResponse_Category} from "~/models/category";
import {Strapi_RelationResponse_Currency} from "~/models/currency";

export default defineEventHandler(async () => {
    const config = useRuntimeConfig()
    const resp = await $fetch<Strapi_ListResponse_Revenue>('/revenues', {
        baseURL: config.strapiApiUrl,
        query: {
            populate: '*',
        },
        headers: {
            'Authorization': `Bearer ${config.strapiApiKey}`
        }
    })

    return resp.data.map((revenue) => {
        return {
            id: revenue.id,
            attributes: {
                ...revenue.attributes,
                // Convert relational object to number
                category: (revenue.attributes.category as Strapi_RelationResponse_Category).data.id,
                currency: (revenue.attributes.currency as Strapi_RelationResponse_Currency).data.id,
            }
        } as StrapiRevenue
    })
})
