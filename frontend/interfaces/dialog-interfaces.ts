import type {StrapiTransaction} from "~/models/transaction";
import type {EmployeeFormData, EmployeeResponse} from "~/models/employee";

export interface IEmployeeFormDialog {
    close: () => {},
    value: {
        close: (employee: EmployeeFormData|'deleted') => any,
        data: {
            employee?: EmployeeResponse
        }
    }
}

export interface ITransactionFormDialog {
    close: () => {},
    value: {
        close: (person: StrapiTransaction|'deleted') => any,
        data: {
            transaction?: StrapiTransaction
        }
    }
}
