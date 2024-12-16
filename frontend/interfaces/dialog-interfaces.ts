import type {StrapiTransaction, TransactionResponse} from "~/models/transaction";
import type {
    EmployeeHistoryResponse,
    EmployeeResponse
} from "~/models/employee";
import type {BankAccountResponse} from "~/models/bank-account";
import type {VatResponse} from "~/models/vat";
import type {OrganisationResponse} from "~/models/organisation";

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
        close: (data: boolean) => any,
        data: {
            isClone: boolean
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

export interface IVatFormDialog {
    close: () => {},
    value: {
        close: (vatId?: number) => any,
        data: {
            vat?: VatResponse
        }
    }
}

export interface IOrganisationFormDialog {
    close: () => {},
    value: {
        close: () => any,
        data: {
            organisation?: OrganisationResponse
        }
    }
}