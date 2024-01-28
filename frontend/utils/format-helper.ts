export const DateStringToFormattedDate = (date: string|Date) => {
    const fmt = Intl.DateTimeFormat('de-CH', {day: '2-digit', month: '2-digit', year: 'numeric'})
    return fmt.format(date instanceof Date ? date : new Date(date))
}

export const NumberToFormattedCurrency = (amount: number, locale: string) => {
    const fmt = Intl.NumberFormat(locale)
    return fmt.format(amount)
}
