import type {Strapi_Meta} from "~/interfaces/strapi-interfaces";
import type {CycleType, RevenueType} from "~/config/enums";
import type {Strapi_RelationResponse_Category} from "~/models/category";
import type {Strapi_RelationResponse_Currency} from "~/models/currency";

export interface StrapiRevenue {
    id?: number;
    attributes: {
        name: string;
        category?: number|Strapi_RelationResponse_Category;
        currency?: number|Strapi_RelationResponse_Currency;
        type?: RevenueType;
        amount: number,
        cycle?: CycleType,
        start: Date|string,
        end: Date|string,
        createdAt: string;
        updatedAt: string;
        publishedAt: string;
    }
}

export interface Strapi_ListResponse_Revenue extends Strapi_Meta {
    data: StrapiRevenue[]
}

export interface Strapi_PostResponse_Revenue extends Strapi_Meta {
    data: StrapiRevenue
}
