import type {Strapi_Meta} from "~/interfaces/strapi-interfaces";

export interface StrapiCurrency {
    id?: number;
    attributes: {
        code: string;
        description: string;
        localeCode: string;
        createdAt: string;
        updatedAt: string;
        publishedAt: string;
    }
}

export interface Strapi_RelationResponse_Currency extends Strapi_Meta {
    data: StrapiCurrency
}

export interface Strapi_ListResponse_Currency extends Strapi_Meta {
    data: StrapiCurrency[]
}

export interface Strapi_PostResponse_Currency extends Strapi_Meta {
    data: StrapiCurrency
}
