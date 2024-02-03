import {CycleType, TransactionType} from "~/config/enums";

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

export const TransactionTypeToOptions = () => {
    return [
        {
            name: 'Einmalige Fälligkeit',
            value: TransactionType.Single,
        },
        {
            name: 'Wiederkehrende Zahlung',
            value: TransactionType.Repeating,
        },
    ]
}
