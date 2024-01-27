import {type Strapi_ListResponse_Person, StrapiPerson} from "~/models/person";

export default defineEventHandler(async () => {
    const config = useRuntimeConfig()
    const resp = await $fetch<Strapi_ListResponse_Person>('/employees', {
        baseURL: config.strapiApiUrl,
        headers: {
            'Authorization': `Bearer ${config.strapiApiKey}`
        }
    })
    return resp.data as StrapiPerson[]
})
