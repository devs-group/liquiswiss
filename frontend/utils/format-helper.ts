export const DateStringToFormattedDate = (date: string | Date, asUtc: boolean = true) => {
  const fmt = Intl.DateTimeFormat(Constants.BASE_LOCALE_CODE, { day: '2-digit', month: '2-digit', year: 'numeric', timeZone: asUtc ? 'UTC' : undefined })
  return fmt.format(date instanceof Date ? date : new Date(date))
}

export const DateStringToFormattedWordDate = (date: string | Date, asUtc: boolean = true) => {
  const fmt = Intl.DateTimeFormat(Constants.BASE_LOCALE_CODE, { month: 'long', year: 'numeric', timeZone: asUtc ? 'UTC' : undefined })
  return fmt.format(date instanceof Date ? date : new Date(date))
}

export const DateStringToFormattedDateTime = (date: string | Date, asUtc: boolean = true) => {
  const fmt = Intl.DateTimeFormat(Constants.BASE_LOCALE_CODE, {
    day: '2-digit', month: '2-digit', year: 'numeric',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
    timeZone: asUtc ? 'UTC' : undefined,
  })
  return fmt.format(date instanceof Date ? date : new Date(date))
}

export const DateToUTCDate = (date: string | Date) => {
  const dateToFormat = date instanceof Date ? date : new Date(date)
  return new Date(dateToFormat.toLocaleDateString('en-US', { timeZone: 'UTC' }))
}

export const DateToEuropeZurichDate = (date: string | Date) => {
  const dateToFormat = date instanceof Date ? date : new Date(date)
  return new Date(dateToFormat.toLocaleDateString('en-US', { timeZone: 'Europe/Zurich' }))
}

export const DateToApiFormat = (date: string | Date) => {
  const dateToFormat = date instanceof Date ? date : new Date(date)

  const year = dateToFormat.getFullYear()
  const month = (dateToFormat.getMonth() + 1).toString().padStart(2, '0')
  const day = dateToFormat.getDate().toString().padStart(2, '0')

  return `${year}-${month}-${day}`
}

export const NumberToFormattedCurrency = (amount: number, locale: string) => {
  const fmt = Intl.NumberFormat(locale, { maximumFractionDigits: 2, minimumFractionDigits: 2 })
  return fmt.format(amount)
}
