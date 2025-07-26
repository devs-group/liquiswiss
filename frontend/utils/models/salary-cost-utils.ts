import type { SalaryCostResponse } from '~/models/employee'
import type { CurrencyResponse } from '~/models/currency'

export const SalaryCostUtils = {
  title: (salaryCost: SalaryCostResponse) =>
    salaryCost.label?.name ?? Fallbacks.CostLabel,

  isFixed: (salaryCost: SalaryCostResponse): boolean => salaryCost.amountType === 'fixed',

  isOnce: (salaryCost: SalaryCostResponse): boolean => salaryCost.cycle === 'once',

  isInactive: (salaryCost: SalaryCostResponse) => salaryCost.calculatedNextExecutionDate == null,

  amountFormatted: (salaryCost: SalaryCostResponse, currency: CurrencyResponse) =>
    SalaryCostUtils.isFixed(salaryCost)
      ? NumberToFormattedCurrency(AmountToFloat(salaryCost.amount), currency.localeCode)
      : AmountToFloat(salaryCost.amount, 3),

  calculatedAmountFormatted: (salaryCost: SalaryCostResponse, currency: CurrencyResponse) =>
    NumberToFormattedCurrency(AmountToFloat(salaryCost.calculatedAmount), currency.localeCode),

  nextCostFormatted: (salaryCost: SalaryCostResponse, currency: CurrencyResponse) =>
    NumberToFormattedCurrency(AmountToFloat(salaryCost.calculatedNextCost), currency.localeCode),

  amountType: (salaryCost: SalaryCostResponse) =>
    EmployeeCostTypeToOptions().find(ct => ct.value === salaryCost.amountType)?.name ?? '',

  distributionType: (salaryCost: SalaryCostResponse) =>
    EmployeeCostDistributionTypeToOptions().find(ct => ct.value === salaryCost.distributionType)?.name ?? '',

  costCycle: (salaryCost: SalaryCostResponse) =>
    CostCycleTypeToOptions().find(ct => ct.value === salaryCost.cycle)?.name ?? '',

  unit: (salaryCost: SalaryCostResponse, currency: CurrencyResponse) =>
    SalaryCostUtils.isFixed(salaryCost) ? currency.code : '%',

  getAmountOffset: (salaryCost: SalaryCostResponse) => salaryCost.relativeOffset == 0
    ? `am selben Tag`
    : `alle ${salaryCost.relativeOffset} ${SalaryCostUtils.costCycle(salaryCost)}`,
}
