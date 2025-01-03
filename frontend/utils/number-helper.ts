export const AmountToInteger = (amount: number, precision: number = 2): number => {
  const factor = Math.pow(10, precision)
  return Math.round(amount * factor)
}

export const AmountToFloat = (amount: number, precision: number = 2): number => {
  const factor = Math.pow(10, precision)
  return amount / factor
}

export const isNumber = (value: unknown): boolean => {
  return typeof value === 'number' && !isNaN(value)
}

export const parseCurrency = (input: string | number, allowNegative: boolean) => {
  if (input === undefined) {
    input = ''
  }
  if (typeof input === 'number') {
    input = input.toString(10)
  }

  // Replace all commas with dots (unifying decimal separator)
  let unifiedInput = input.replace(/,/g, '.')

  const isNegative = allowNegative && unifiedInput.startsWith('-')

  // Remove all invalid characters except numbers and dots
  unifiedInput = unifiedInput.replace(/[^0-9.]/g, '')

  // Ensure only one decimal separator is allowed
  const parts = unifiedInput.split('.')
  if (parts.length > 2) {
    // Keep the last detected decimal part
    const decimals = parts.pop()
    // Reassemble with a single dot
    unifiedInput = parts.join('') + '.' + decimals
  }

  if (isNegative && unifiedInput.length > 0) {
    unifiedInput = '-' + unifiedInput
  }

  return unifiedInput
}
