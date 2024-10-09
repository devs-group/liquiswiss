import type {StrapiTransaction, TransactionResponse} from "~/models/transaction";
import type {
    EmployeeHistoryResponse,
    EmployeeResponse
} from "~/models/employee";
import type {BankAccountResponse} from "~/models/bank-account";

export interface IEmployeeFormDialog {
    close: () => {},
    value: {
        close: () => any,
        data: {
            employee?: EmployeeResponse
        }
    }
}

export interface IHistoryFormDialog {
    close: () => {},
    value: {
        close: () => any,
        data: {
            employeeID: number
            employeeHistory?: EmployeeHistoryResponse
        }
    }
}

export interface ITransactionFormDialog {
    close: () => {},
    value: {
        close: () => any,
        data: {
            isClone: boolean
            transaction?: TransactionResponse
        }
    }
}
export interface IBankAccountFormDialog {
    close: () => {},
    value: {
        close: () => any,
        data: {
            bankAccountID: number
            bankAccount?: BankAccountResponse
        }
    }
}
