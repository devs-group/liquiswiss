import type {Strapi_Meta} from "~/interfaces/strapi-interfaces";

export interface StrapiCategory {
    id?: number;
    attributes: {
        name: string;
        createdAt: string;
        updatedAt: string;
        publishedAt: string;
    }
}

export interface Strapi_RelationResponse_Category extends Strapi_Meta {
    data: StrapiCategory
}

export interface Strapi_ListResponse_Category extends Strapi_Meta {
    data: StrapiCategory[]
}

export interface Strapi_PostResponse_Category extends Strapi_Meta {
    data: StrapiCategory
}
