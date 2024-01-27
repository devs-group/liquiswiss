import type {StrapiPerson} from "~/models/person";

export interface IPersonFormDialog {
    close: () => {},
    value: {
        close: (person: StrapiPerson) => any,
        data: {
            person?: StrapiPerson
        }
    }
}
