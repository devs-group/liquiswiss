export const TransactionSortByOptions = ['name', 'startDate', 'endDate', 'amount', 'cycle', 'category', 'employee', 'nextExecutionDate'] as const
export type TransactionSortByType = typeof TransactionSortByOptions[number]

export const EmployeeSortByOptions = ['name', 'hoursPerMonth', 'salaryPerMonth', 'vacationDaysPerYear', 'fromDate', 'toDate'] as const
export type EmployeeSortByType = typeof EmployeeSortByOptions[number]

export const SortOrderOptions = ['ASC', 'DESC'] as const
export type SortOrderType = typeof SortOrderOptions[number]