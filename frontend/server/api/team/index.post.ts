import {type Strapi_ListResponse_Person, StrapiPerson} from "~/models/person";

export default defineEventHandler(async (event) => {
    const body = await readBody<StrapiPerson>(event)
    const isNew = body.id === undefined

    const config = useRuntimeConfig()
    const resp = await $fetch<Strapi_ListResponse_Person>('/employees', {
        baseURL: config.strapiApiUrl,
        method: 'post',
        body: body,
        headers: {
            'Authorization': `Bearer ${config.strapiApiKey}`
        }
    })

    return resp
})
