export const TransactionSortByOptions = ['name', 'startDate', 'endDate', 'amount', 'cycle', 'category', 'employee'] as const
export type TransactionSortByType = typeof TransactionSortByOptions[number]

export const SortOrderOptions = ['ASC', 'DESC'] as const
export type SortOrderType = typeof SortOrderOptions[number]