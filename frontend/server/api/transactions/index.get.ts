import {Strapi_ListResponse_Transaction, StrapiTransaction} from "~/models/transaction";
import {Strapi_RelationResponse_Category} from "~/models/category";
import {Strapi_RelationResponse_Currency} from "~/models/currency";

export default defineEventHandler(async () => {
    const config = useRuntimeConfig()
    const resp = await $fetch<Strapi_ListResponse_Transaction>('/transactions', {
        baseURL: config.apiHost,
        query: {
            populate: '*',
        },
        headers: {
            'Authorization': `Bearer ${config.strapiApiKey}`
        }
    })

    return resp.data.map((transaction) => {
        return {
            id: transaction.id,
            attributes: {
                ...transaction.attributes,
                // Convert relational object to number
                category: (transaction.attributes.category as Strapi_RelationResponse_Category).data.id,
                currency: (transaction.attributes.currency as Strapi_RelationResponse_Currency).data.id,
            }
        } as StrapiTransaction
    })
})
