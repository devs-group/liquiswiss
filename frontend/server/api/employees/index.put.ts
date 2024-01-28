import {type Strapi_PostResponse_Person, StrapiPerson} from "~/models/person";

export default defineEventHandler(async (event) => {
    const body = await readBody<StrapiPerson>(event)

    const config = useRuntimeConfig()
    return await $fetch<Strapi_PostResponse_Person>(`/employees/${body.id}`, {
        baseURL: config.strapiApiUrl,
        method: 'put',
        body: {
            data: body.attributes
        },
        headers: {
            'Authorization': `Bearer ${config.strapiApiKey}`
        }
    })
})
