import type {PaginationResponse} from "~/models/pagination";
import type {CurrencyResponse} from "~/models/currency";

export interface EmployeeFormData {
    id: number;
    name: string;
}

export interface EmployeeResponse {
    id: number;
    name: string;
    hoursPerMonth: number | null;
    salaryPerMonth: number | null;
    salaryCurrency: CurrencyResponse | null;
    vacationDaysPerYear?: number | null;
    fromDate?: string | null;
    toDate?: string | null;
}

export interface ListEmployeeResponse {
    data: EmployeeResponse[];
    pagination: PaginationResponse;
}

export interface EmployeeHistoryFormData {
    id: number;
    hoursPerMonth: number;
    salaryPerMonth: number;
    salaryCurrency: number;
    vacationDaysPerYear: number;
    fromDate: Date;
    toDate?: Date;
}

export interface EmployeeHistoryResponse {
    id: number;
    employeeID: string;
    hoursPerMonth: number;
    salaryPerMonth: number;
    salaryCurrency: CurrencyResponse;
    vacationDaysPerYear: number;
    fromDate: string;
    toDate: string | null;
}

export interface ListEmployeeHistoryResponse {
    data: EmployeeHistoryResponse[];
    pagination: PaginationResponse;
}