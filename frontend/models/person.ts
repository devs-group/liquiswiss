import type {Strapi_ListResponse} from "~/interfaces/strapi-interfaces";

export interface StrapiPerson {
    id?: number;
    attributes: {
        name: string;
        hoursPerMonth: number;
        vacationDaysPerYear: number;
        createdAt: string;
        updatedAt: string;
        publishedAt: string;
    }
}

export interface Strapi_ListResponse_Person extends Strapi_ListResponse {
    data: StrapiPerson[]
}
