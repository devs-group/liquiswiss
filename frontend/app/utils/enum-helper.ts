import {
  CycleType,
  EmployeeCostDistributionType,
  EmployeeCostOverviewType,
  EmployeeCostType,
  TransactionType,
} from '~/config/enums'

export type TransactionCycleTypeToStringDefinition = CycleType.Monthly | CycleType.Quarterly | CycleType.Biannually | CycleType.Yearly

export const TransactionCycleTypeToOptions = () => {
  return [
    {
      name: 'Monatlich',
      value: CycleType.Monthly,
    },
    {
      name: 'Vierteljährlich',
      value: CycleType.Quarterly,
    },
    {
      name: 'Halbjährlich',
      value: CycleType.Biannually,
    },
    {
      name: 'Jährlich',
      value: CycleType.Yearly,
    },
  ]
}

export type SalaryCycleTypeToStringDefinition = CycleType.Monthly

export const SalaryCycleTypeToOptions = () => {
  return [
    {
      name: 'Monatlich',
      value: CycleType.Monthly,
    },
  ]
}

export type CostCycleTypeToStringDefinition = CycleType.Once | CycleType.Monthly

export const CostCycleTypeToOptions = () => {
  return [
    {
      name: 'Einmalig',
      value: CycleType.Once,
    },
    {
      name: 'Monatlich',
      value: CycleType.Monthly,
    },
    // {
    //   name: 'Vierteljährlich',
    //   value: CycleType.Quarterly,
    // },
    // {
    //   name: 'Halbjährlich',
    //   value: CycleType.Biannually,
    // },
    // {
    //   name: 'Jährlich',
    //   value: CycleType.Yearly,
    // },
  ]
}

export type TransactionTypeToStringDefinition = TransactionType.Single | TransactionType.Repeating

export const TransactionTypeToOptions = () => {
  return [
    {
      name: 'Einmalige Fälligkeit',
      value: TransactionType.Single,
    },
    {
      name: 'Wiederkehrende Zahlung',
      value: TransactionType.Repeating,
    },
  ]
}

export type EmployeeCostTypeToStringDefinition = EmployeeCostType.Fixed | EmployeeCostType.Percentage

export const EmployeeCostTypeToOptions = () => {
  return [
    {
      name: 'Fixer Betrag',
      value: EmployeeCostType.Fixed,
    },
    {
      name: 'Prozentualer Anteil',
      value: EmployeeCostType.Percentage,
    },
  ]
}

export type EmployeeCostDistributionTypeToStringDefinition =
  | EmployeeCostDistributionType.Employee
  | EmployeeCostDistributionType.Employer
  | EmployeeCostDistributionType.Both

export const EmployeeCostDistributionTypeToOptions = () => {
  return [
    {
      name: 'Arbeitnehmer (AN)',
      value: EmployeeCostDistributionType.Employee,
    },
    {
      name: 'Arbeitgeber (AG)',
      value: EmployeeCostDistributionType.Employer,
    },
    {
      name: 'Arbeitgeber & Arbeitnehmer (AG & AN)',
      value: EmployeeCostDistributionType.Both,
    },
  ]
}

export type EmployeeCostOverviewTypeFilterToStringDefinition = EmployeeCostOverviewType.All | EmployeeCostOverviewType.Employee | EmployeeCostOverviewType.Employer

export const EmployeeCostOverviewTypeToOptions = () => {
  return [
    {
      name: 'Alles',
      value: EmployeeCostOverviewType.All,
    },
    {
      name: 'Arbeitnehmer (AN)',
      value: EmployeeCostOverviewType.Employee,
    },
    {
      name: 'Arbeitgeber (AG)',
      value: EmployeeCostOverviewType.Employer,
    },
  ]
}
