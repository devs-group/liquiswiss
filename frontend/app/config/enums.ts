export enum TransactionType {
  Single = 'single',
  Repeating = 'repeating',
}

export enum CycleType {
  Once = 'once',
  Monthly = 'monthly',
  Quarterly = 'quarterly',
  Biannually = 'biannually',
  Yearly = 'yearly',
}

export enum EmployeeCostType {
  Fixed = 'fixed',
  Percentage = 'percentage',
}

export enum EmployeeCostDistributionType {
  Employee = 'employee',
  Employer = 'employer',
  Both = 'both',
}

export enum EmployeeCostOverviewType {
  All = 'all',
  Employee = 'employee',
  Employer = 'employer',
}
