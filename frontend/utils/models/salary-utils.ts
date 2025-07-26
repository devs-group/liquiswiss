import type { SalaryResponse } from '~/models/employee'

export const SalaryUtils = {
  title: (salary: SalaryResponse) =>
    `Von ${SalaryUtils.fromDateFormatted(salary)}, ${SalaryUtils.grossSalaryFormatted(salary)} ${salary.currency.code} / ${SalaryUtils.cycle(salary)}`,
  nextExecutionDate: (salary: SalaryResponse) =>
    salary.nextExecutionDate ? DateStringToFormattedDate(salary.nextExecutionDate) : '-',
  grossSalaryFormatted: (salary: SalaryResponse) =>
    NumberToFormattedCurrency(AmountToFloat(salary.amount), salary.currency.localeCode),
  netSalaryFormatted: (salary: SalaryResponse) =>
    NumberToFormattedCurrency(AmountToFloat(salary.amount - salary.employeeDeductions), salary.currency!.localeCode),
  totalSalaryCostFormatted: (salary: SalaryResponse) =>
    NumberToFormattedCurrency(AmountToFloat(salary.amount + salary.employerCosts), salary.currency!.localeCode),
  fromDateFormatted: (salary: SalaryResponse) =>
    DateStringToFormattedDate(salary.fromDate),
  toDateFormatted: (salary: SalaryResponse) =>
    salary.toDate ? DateStringToFormattedDate(salary.toDate) : '',
  cycle: (salary: SalaryResponse) =>
    SalaryCycleTypeToOptions().find(ct => ct.value === salary.cycle)?.name ?? '',
  hasCosts: (salary: SalaryResponse) =>
    (salary.employeeDeductions + salary.employerCosts) > 0,
}
