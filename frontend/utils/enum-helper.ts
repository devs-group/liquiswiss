import { CycleType, TransactionType } from '~/config/enums'

export type CycleTypeToStringDefinition = CycleType.Daily | CycleType.Weekly | CycleType.Monthly | CycleType.Quarterly | CycleType.Biannually | CycleType.Yearly

export const CycleTypeToOptions = () => {
  return [
    {
      name: 'Täglich',
      value: CycleType.Daily,
    },
    {
      name: 'Wöchentlich',
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
