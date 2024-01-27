import {Person} from "~/models/person";

export default defineEventHandler(async (event) => {
    const body = await readBody<Person>(event)
    const isNew = body.id === undefined
    // TODO: Write User to DB as new or update
    return body
})
