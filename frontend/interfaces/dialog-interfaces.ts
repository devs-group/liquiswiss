import type {StrapiPerson} from "~/models/person";
import type {StrapiRevenue} from "~/models/revenue";

export interface IPersonFormDialog {
    close: () => {},
    value: {
        close: (person: StrapiPerson|'deleted') => any,
        data: {
            person?: StrapiPerson
        }
    }
}

export interface IRevenueFormDialog {
    close: () => {},
    value: {
        close: (person: StrapiRevenue|'deleted') => any,
        data: {
            revenue?: StrapiRevenue
        }
    }
}
