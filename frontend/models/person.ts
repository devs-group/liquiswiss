import type {Strapi_Meta} from "~/interfaces/strapi-interfaces";

export interface StrapiPerson {
    id?: number;
    attributes: {
        name: string;
        hoursPerMonth: number;
        vacationDaysPerYear: number;
        entry: Date|string;
        exit: Date|string;
        createdAt: string;
        updatedAt: string;
        publishedAt: string;
    }
}

export interface Strapi_ListResponse_Person extends Strapi_Meta {
    data: StrapiPerson[]
}

export interface Strapi_PostResponse_Person extends Strapi_Meta {
    data: StrapiPerson
}
