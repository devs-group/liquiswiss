import {
  CycleType,
  EmployeeCostDistributionType,
  EmployeeCostOverviewType,
  EmployeeCostType,
  TransactionType,
} from '~/config/enums'

export type CycleTypeToStringDefinition = CycleType.Daily | CycleType.Weekly | CycleType.Monthly | CycleType.Quarterly | CycleType.Biannually | CycleType.Yearly

export const CycleTypeToOptions = () => {
  return [
    // TODO: Due to complexity disabled for now
    // {
    //   name: 'Täglich',
    //   value: CycleType.Daily,
    // },
    // {
    //   name: 'Wöchentlich',
    //   value: CycleType.Weekly,
    // },
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

export type CostCycleTypeToStringDefinition = CycleType.Once | CycleTypeToStringDefinition

export const CostCycleTypeToOptions = () => {
  return [
    {
      name: 'Einmalig',
      value: CycleType.Once,
    },
    {
      name: 'Täglich',
      value: CycleType.Daily,
    },
    {
      name: 'Wöchenlich',
      value: CycleType.Weekly,
    },
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

export type EmployeeCostDistributionTypeToStringDefinition = EmployeeCostDistributionType.Employee | EmployeeCostDistributionType.Employer

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
