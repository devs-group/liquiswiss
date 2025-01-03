import type { TransactionResponse } from '~/models/transaction'
import type {
  EmployeeHistoryCostLabelResponse,
  EmployeeHistoryCostResponse,
  EmployeeHistoryResponse,
  EmployeeResponse,
} from '~/models/employee'
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

export interface IHistoryFormDialog {
  close: () => object
  value: {
    close: (data: boolean) => boolean
    data: {
      isClone: boolean
      employeeID: number
      employeeHistory?: EmployeeHistoryResponse
    }
  }
}

export interface IHistoryCostOverviewDialog {
  close: (requiresRefresh: boolean) => unknown
  value: {
    close: (data: boolean) => boolean
    data: {
      employeeHistory: EmployeeHistoryResponse
    }
  }
}

export interface IHistoryCostCopyDialog {
  close: () => object
  value: {
    close: (data: boolean) => boolean
    data: {
      employeeHistory: EmployeeHistoryResponse
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

export interface IEmployeeHistoryCostFormDialog {
  close: (requiresRefresh: boolean) => unknown
  value: {
    close: (historyCostId?: number) => number | undefined
    data: {
      isClone: boolean
      employeeHistory: EmployeeHistoryResponse
      employeeCostToEdit?: EmployeeHistoryCostResponse
    }
  }
}

export interface IEmployeeHistoryCostLabelFormDialog {
  close: () => object
  value: {
    close: (historyCostId?: number) => number | undefined
    data: {
      employeeCostLabelToEdit?: EmployeeHistoryCostLabelResponse
    }
  }
}
