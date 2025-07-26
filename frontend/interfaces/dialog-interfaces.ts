import type { TransactionResponse } from '~/models/transaction'
import type { EmployeeResponse, SalaryCostLabelResponse, SalaryCostResponse, SalaryResponse } from '~/models/employee'
import type { BankAccountResponse } from '~/models/bank-account'
import type { VatResponse } from '~/models/vat'
import type { OrganisationResponse } from '~/models/organisation'

export interface IEmployeeFormDialog {
  close: () => object
  value: {
    close: () => void
    data: {
      employee?: EmployeeResponse
    }
  }
}

export interface ISalaryFormDialog {
  close: () => object
  value: {
    close: (data: boolean) => boolean
    data: {
      isClone: boolean
      employeeID: number
      salary?: SalaryResponse
    }
  }
}

export interface ISalaryCostOverviewDialog {
  close: (requiresRefresh: boolean) => unknown
  value: {
    close: (data: boolean) => boolean
    data: {
      salary: SalaryResponse
    }
  }
}

export interface ISalaryCostCopyDialog {
  close: () => object
  value: {
    close: (data: boolean) => boolean
    data: {
      salaryAmount: SalaryResponse
    }
  }
}

export interface ITransactionFormDialog {
  close: () => object
  value: {
    close: () => void
    data: {
      isClone: boolean
      transaction?: TransactionResponse
    }
  }
}
export interface IBankAccountFormDialog {
  close: () => object
  value: {
    close: () => void
    data: {
      isClone: boolean
      bankAccount?: BankAccountResponse
    }
  }
}

export interface IVatFormDialog {
  close: () => object
  value: {
    close: (vatId?: number) => number | undefined
    data: {
      vatToEdit?: VatResponse
    }
  }
}

export interface IOrganisationFormDialog {
  close: () => object
  value: {
    close: () => void
    data: {
      organisation?: OrganisationResponse
    }
  }
}

export interface ISalaryCostFormDialog {
  close: (requiresRefresh: boolean) => unknown
  value: {
    close: (salaryCostId?: number) => number | undefined
    data: {
      isClone: boolean
      salary: SalaryResponse
      salaryCostToEdit?: SalaryCostResponse
    }
  }
}

export interface ISalaryCostLabelFormDialog {
  close: () => object
  value: {
    close: (salaryCostID?: number) => number | undefined
    data: {
      employeeCostLabelToEdit?: SalaryCostLabelResponse
    }
  }
}
