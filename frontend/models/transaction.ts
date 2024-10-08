import type {CycleType, TransactionType} from "~/config/enums";
import type {CategoryResponse, Strapi_RelationResponse_Category} from "~/models/category";
import type {CurrencyResponse, Strapi_RelationResponse_Currency} from "~/models/currency";
import type {PaginationResponse} from "~/models/pagination";

export interface StrapiTransaction {
    id?: number;
    attributes: {
        name: string;
        category?: number|Strapi_RelationResponse_Category;
        currency?: number|Strapi_RelationResponse_Currency;
        type?: TransactionType;
        amount: number,
        cycle?: CycleType,
        start: Date|string,
        end?: Date|string,
        createdAt: string;
        updatedAt: string;
        publishedAt: string;
    }
}

export interface TransactionResponse {
    id: number;
    name: string;
    amount: number;
    cycle: CycleTypeToStringDefinition | null;
    type: TransactionTypeToStringDefinition;
    startDate: string;
    endDate: string | null;
    category: CategoryResponse;
    currency: CurrencyResponse;
    employee: TransactionEmployeeResponse | null;
}

export interface TransactionEmployeeResponse {
    id: number;
    name: string;
}

export interface ListTransactionResponse {
    data: TransactionResponse[];
    pagination: PaginationResponse;
}

export interface TransactionFormData {
    id: number;
    name: string;
    amount: number;
    cycle?: CycleTypeToStringDefinition;
    type: TransactionTypeToStringDefinition;
    startDate: Date;
    endDate?: Date;
    category: number;
    currency: number;
    employee: number;
}