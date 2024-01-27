import {Person} from "~/models/person";

export default defineEventHandler(async () => {
    // TODO: Fetch from database
    return [
        {
            id: 1,
            name: "John Doe",
            hoursPerMonth: 160,
            vacationDaysPerYear: 25,
        },
        {
            id: 2,
            name: "Jane Doe",
            hoursPerMonth: 160,
            vacationDaysPerYear: 25,
        },
        {
            id: 2,
            name: "Michelle Doe",
            hoursPerMonth: 160,
            vacationDaysPerYear: 25,
        },
        {
            id: 3,
            name: "Michael Doe",
            hoursPerMonth: 120,
            vacationDaysPerYear: 25,
        }
    ] as Person[]
})
