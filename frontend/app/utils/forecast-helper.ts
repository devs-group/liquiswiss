import type { ForecastDetailResponse, ForecastDetailRevenueExpenseResponse } from '~/models/forecast'

/**
 * Sums the amount of a category within one forecast month, including its children.
 * Returns the raw amount, so neither the performance factor nor exclusions are applied.
 */
export const FindForecastAmount = (
  items: ForecastDetailRevenueExpenseResponse[],
  targetName: string,
): number => {
  for (const item of items) {
    if (item.name === targetName) {
      const childrenAmount = item.children
        ? item.children.reduce(
            (sum, child) => sum + FindForecastAmount([child], child.name),
            0,
          )
        : 0
      return (item.amount ?? 0) + childrenAmount
    }

    if (item.children) {
      const childAmount = FindForecastAmount(item.children, targetName)
      if (childAmount !== 0) {
        return childAmount
      }
    }
  }

  return 0
}

/**
 * True when a category has no amount in ANY of the displayed months, so the row
 * would only ever show 0.00. Excluded rows still count as non-zero as long as
 * they carry an amount, otherwise they could no longer be re-included.
 */
export const IsForecastCategoryAlwaysZero = (
  forecastDetails: ForecastDetailResponse[],
  categoryName: string,
  forecastType: 'revenue' | 'expense',
): boolean => {
  return forecastDetails.every(
    detail => FindForecastAmount(detail[forecastType] || [], categoryName) === 0,
  )
}
