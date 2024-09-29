import type {StrapiTransaction, TransactionResponse} from "~/models/transaction";
import type {
    EmployeeHistoryResponse,
    EmployeeResponse
} from "~/models/employee";

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
            transactionID: number
            transaction?: TransactionResponse
        }
    }
}
