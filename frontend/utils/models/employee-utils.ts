import type { EmployeeResponse } from '~/models/employee'

export const EmployeeUtils = {
  salaryFormatted: (employee: EmployeeResponse) =>
    employee.salaryAmount ? NumberToFormattedCurrency(AmountToFloat(employee.salaryAmount ?? 0), employee.currency!.localeCode) : '-',

  fromDate: (employee: EmployeeResponse) =>
    employee.fromDate ? DateStringToFormattedDate(employee.fromDate) : '-',

  toDate: (employee: EmployeeResponse) =>
    employee.toDate ? DateStringToFormattedDate(employee.toDate) : '-',

  cycle: (employee: EmployeeResponse) =>
    CycleTypeToOptions().find(ct => ct.value === employee.cycle)?.name ?? '',

  hasSalaryAmount: (employee: EmployeeResponse) =>
    employee.salaryAmount !== null,
}
