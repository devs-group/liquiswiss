import type {Strapi_Meta} from "~/interfaces/strapi-interfaces";
import type {CycleType, TransactionType} from "~/config/enums";
import type {Strapi_RelationResponse_Category} from "~/models/category";
import type {Strapi_RelationResponse_Currency} from "~/models/currency";

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
