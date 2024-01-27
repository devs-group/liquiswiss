import type {Person} from "~/models/person";

export interface IPersonFormDialog {
    close: () => {},
    value: {
        close: (person: Person) => any,
        data: {
            person?: Person
        }
    }
}
