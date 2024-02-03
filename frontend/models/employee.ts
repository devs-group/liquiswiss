import type {PaginationResponse} from "~/models/pagination";

export interface EmployeeFormData {
    id: number;
    name: string;
    hoursPerMonth: number;
    vacationDaysPerYear: number;
    entryDate: Date;
    exitDate?: Date;
}

export interface EmployeeResponse {
    id: number;
    name: string;
    hoursPerMonth: number;
    vacationDaysPerYear: number;
    entryDate: string;
    exitDate?: string;
}

export interface ListEmployeeResponse {
    data: EmployeeResponse[];
    pagination: PaginationResponse;
}