import type { EmployeeHistoryResponse } from '~/models/employee'

export const EmployeeHistoryUtils = {
  title: (history: EmployeeHistoryResponse) =>
    `Von ${EmployeeHistoryUtils.fromDateFormatted(history)}, ${EmployeeHistoryUtils.grossSalaryFormatted(history)} ${history.currency.code} / ${EmployeeHistoryUtils.cycle(history)}`,
  nextExecutionDate: (history: EmployeeHistoryResponse) =>
    history.nextExecutionDate ? DateStringToFormattedDate(history.nextExecutionDate) : '-',
  grossSalaryFormatted: (history: EmployeeHistoryResponse) =>
    NumberToFormattedCurrency(AmountToFloat(history.salary), history.currency.localeCode),
  netSalaryFormatted: (history: EmployeeHistoryResponse) =>
    NumberToFormattedCurrency(AmountToFloat(history.salary - history.employeeDeductions), history.currency!.localeCode),
  totalSalaryCostFormatted: (history: EmployeeHistoryResponse) =>
    NumberToFormattedCurrency(AmountToFloat(history.salary + history.employerCosts), history.currency!.localeCode),
  fromDateFormatted: (history: EmployeeHistoryResponse) =>
    DateStringToFormattedDate(history.fromDate),
  toDateFormatted: (history: EmployeeHistoryResponse) =>
    history.toDate ? DateStringToFormattedDate(history.toDate) : '',
  cycle: (history: EmployeeHistoryResponse) =>
    CycleTypeToOptions().find(ct => ct.value === history.cycle)?.name ?? '',
  hasCosts: (history: EmployeeHistoryResponse) =>
    (history.employeeDeductions + history.employerCosts) > 0,
}
