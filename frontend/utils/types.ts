import { RouteNames } from '~/config/routes'

export const TransactionSortByOptions = ['name', 'startDate', 'endDate', 'amount', 'cycle', 'category', 'employee', 'nextExecutionDate'] as const
export type TransactionSortByType = typeof TransactionSortByOptions[number]

export const EmployeeSortByOptions = ['name', 'hoursPerMonth', 'salaryAmount', 'vacationDaysPerYear', 'fromDate', 'toDate'] as const
export type EmployeeSortByType = typeof EmployeeSortByOptions[number]

export const BankAccountSortByOptions = ['name', 'amount'] as const
export type BankAccountSortByType = typeof BankAccountSortByOptions[number]

export const OrganisationRoleOptions = ['owner', 'admin', 'editor', 'read-only'] as const
export type OrganisationRoleType = typeof OrganisationRoleOptions[number]

export const SortOrderOptions = ['ASC', 'DESC'] as const
export type SortOrderType = typeof SortOrderOptions[number]

export const SettingsTabOptions = [RouteNames.SETTINGS_PROFILE, RouteNames.SETTINGS_ORGANISATIONS, RouteNames.SETTINGS_APP] as const
export type SettingsTabType = typeof SettingsTabOptions[number]

export const DarkModeOptions = ['system', 'dark', 'light'] as const
export type DarkModeType = typeof DarkModeOptions[number]

export const DisplayTypeOptions = ['grid', 'list'] as const
export type DisplayType = typeof DisplayTypeOptions[number]

export const GroupingTypeOptions = ['ungrouped', 'grouped'] as const
export type GroupingType = typeof GroupingTypeOptions[number]
