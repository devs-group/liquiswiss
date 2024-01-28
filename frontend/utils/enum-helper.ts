import {CycleType, RevenueType} from "~/config/enums";

export const CycleTypeToOptions = () => {
    return [
        {
            name: 'Täglich',
            value: CycleType.Daily
        },
        {
            name: 'Wöchentlich',
            value: CycleType.Weekly
        },
        {
            name: 'Monatlich',
            value: CycleType.Monthly
        },
        {
            name: 'Vierteljährlich',
            value: CycleType.Quarterly
        },
        {
            name: 'Halbjährlich',
            value: CycleType.Biannually
        },
        {
            name: 'Jährlich',
            value: CycleType.Yearly
        },
    ]
}

export const RevenueTypeToOptions = () => {
    return [
        {
            name: 'Einmalige Fälligkeit',
            value: RevenueType.Single,
        },
        {
            name: 'Wiederkehrende Zahlung',
            value: RevenueType.Repeating,
        },
    ]
}
