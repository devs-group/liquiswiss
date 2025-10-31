export const DateDiffInMilliseconds = (d1: Date, d2: Date) => {
  return d1.getTime() - d2.getTime()
}

export const DateDiffInDays = (d1: Date, d2: Date) => {
  const diffInMs = DateDiffInMilliseconds(d1, d2)
  return diffInMs / (1000 * 60 * 60 * 24)
}

export const DateAddDays = (date: Date, days: number): Date => {
  const result = new Date(date)
  result.setDate(result.getDate() + days)
  return result
}
