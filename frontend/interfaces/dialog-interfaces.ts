import type {StrapiPerson} from "~/models/person";

export interface IPersonFormDialog {
    close: () => {},
    value: {
        close: (person: StrapiPerson|'deleted') => any,
        data: {
            person?: StrapiPerson
        }
    }
}
