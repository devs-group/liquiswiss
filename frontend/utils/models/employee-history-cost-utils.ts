import type { EmployeeHistoryCostResponse } from '~/models/employee'
import type { CurrencyResponse } from '~/models/currency'

export const EmployeeHistoryCostUtils = {
  title: (historyCost: EmployeeHistoryCostResponse) =>
    historyCost.label?.name ?? Fallbacks.CostLabel,

  isFixed: (historyCost: EmployeeHistoryCostResponse): boolean => historyCost.amountType === 'fixed',

  isOnce: (historyCost: EmployeeHistoryCostResponse): boolean => historyCost.cycle === 'once',

  isInactive: (historyCost: EmployeeHistoryCostResponse) => historyCost.calculatedNextExecutionDate == null,

  amountFormatted: (historyCost: EmployeeHistoryCostResponse, currency: CurrencyResponse) =>
    EmployeeHistoryCostUtils.isFixed(historyCost)
      ? NumberToFormattedCurrency(AmountToFloat(historyCost.amount), currency.localeCode)
      : AmountToFloat(historyCost.amount, 3),

  calculatedAmountFormatted: (historyCost: EmployeeHistoryCostResponse, currency: CurrencyResponse) =>
    NumberToFormattedCurrency(AmountToFloat(historyCost.calculatedAmount), currency.localeCode),

  nextCostFormatted: (historyCost: EmployeeHistoryCostResponse, currency: CurrencyResponse) =>
    NumberToFormattedCurrency(AmountToFloat(historyCost.calculatedNextCost), currency.localeCode),

  amountType: (historyCost: EmployeeHistoryCostResponse) =>
    EmployeeCostTypeToOptions().find(ct => ct.value === historyCost.amountType)?.name ?? '',

  distributionType: (historyCost: EmployeeHistoryCostResponse) =>
    EmployeeCostDistributionTypeToOptions().find(ct => ct.value === historyCost.distributionType)?.name ?? '',

  costCycle: (historyCost: EmployeeHistoryCostResponse) =>
    CostCycleTypeToOptions().find(ct => ct.value === historyCost.cycle)?.name ?? '',

  unit: (historyCost: EmployeeHistoryCostResponse, currency: CurrencyResponse) =>
    EmployeeHistoryCostUtils.isFixed(historyCost) ? currency.code : '%',

  getAmountOffset: (historyCost: EmployeeHistoryCostResponse) => historyCost.relativeOffset == 0
    ? `am selben Tag`
    : `alle ${historyCost.relativeOffset} ${EmployeeHistoryCostUtils.costCycle(historyCost)}`,
}
