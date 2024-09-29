import type {Strapi_Meta} from "~/interfaces/strapi-interfaces";
import type {CycleType, TransactionType} from "~/config/enums";
import type {CategoryResponse, Strapi_RelationResponse_Category} from "~/models/category";
import type {CurrencyResponse, Strapi_RelationResponse_Currency} from "~/models/currency";
import type {PaginationResponse} from "~/models/pagination";
import type {CycleTypeToStringDefinition, TransactionTypeToStringDefinition} from "~/utils/enum-helper";

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

export interface Strapi_ListResponse_Transaction extends Strapi_Meta {
    data: StrapiTransaction[]
}

export interface Strapi_PostResponse_Transaction extends Strapi_Meta {
    data: StrapiTransaction
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
}