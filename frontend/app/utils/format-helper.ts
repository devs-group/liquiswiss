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
  return new Date(Date.UTC(
    dateToFormat.getUTCFullYear(),
    dateToFormat.getUTCMonth(),
    dateToFormat.getUTCDate(),
  ))
}

export const DateToEuropeZurichDate = (date: string | Date) => {
  const dateToFormat = date instanceof Date ? date : new Date(date)
  const formatter = new Intl.DateTimeFormat('en-GB', {
    timeZone: 'Europe/Zurich',
    year: 'numeric',
    month: 'numeric',
    day: 'numeric',
    hour: 'numeric',
    minute: 'numeric',
    second: 'numeric',
    hourCycle: 'h23',
  })
  const parts = formatter.formatToParts(dateToFormat)
  const getPart = (type: string, fallback: number) => {
    const part = parts.find(p => p.type === type)
    return part ? parseInt(part.value, 10) : fallback
  }

  const year = getPart('year', dateToFormat.getUTCFullYear())
  const month = getPart('month', dateToFormat.getUTCMonth() + 1) - 1
  const day = getPart('day', dateToFormat.getUTCDate())
  const hour = getPart('hour', dateToFormat.getUTCHours())
  const minute = getPart('minute', dateToFormat.getUTCMinutes())
  const second = getPart('second', dateToFormat.getUTCSeconds())

  return new Date(Date.UTC(year, month, day, hour, minute, second))
}

export const DateToApiFormat = (date: string | Date) => {
  const dateToFormat = date instanceof Date ? date : new Date(date)

  const year = dateToFormat.getUTCFullYear()
  const month = (dateToFormat.getUTCMonth() + 1).toString().padStart(2, '0')
  const day = dateToFormat.getUTCDate().toString().padStart(2, '0')

  return `${year}-${month}-${day}`
}

export const NumberToFormattedCurrency = (amount: number, locale: string) => {
  const fmt = Intl.NumberFormat(locale, { maximumFractionDigits: 2, minimumFractionDigits: 2 })
  return fmt.format(amount)
}

export const NormalizeUrl = (url: string): string => {
  if (!url) return ''
  const trimmed = url.trim()
  if (trimmed.startsWith('http://') || trimmed.startsWith('https://')) {
    return trimmed
  }
  return `https://${trimmed}`
}
