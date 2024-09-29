export const AmountToInteger = (amount: number) => {
    return Math.round(amount * 100)
}

export const AmountToFloat = (amount: number) => {
    return Math.round(amount / 100 * 100) / 100
}