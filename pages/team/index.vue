<template>
  <div class="flex flex-col gap-4">
    <Button @click="addEmployee" class="self-end" label="Mitarbeiter hinzufügen" icon="pi pi-user"/>

    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <PersonCard @on-edit="onEdit" v-for="person in people" :person="person"/>
    </div>
  </div>
</template>

<script setup lang="ts">
import type {Person} from "~/models/person";
import {ModalConfig} from "~/config/dialog-props";
import PersonDialog from "~/components/dialogs/PersonDialog.vue";

const dialog = useDialog();

const people = ref<Person[]>([
  {
    id: 1,
    name: "Ralph Segi",
    hoursPerMonth: 160,
    vacationDaysPerYear: 25,
  },
  {
    id: 2,
    name: "Robert Pupel",
    hoursPerMonth: 160,
    vacationDaysPerYear: 25,
  },
  {
    id: 3,
    name: "Matthias Hillert-Wernicke",
    hoursPerMonth: 120,
    vacationDaysPerYear: 25,
  }
]);

const onEdit = (person: Person) => {
  dialog.open(PersonDialog, {
    data: {
      person: person,
    },
    props: {
      header: 'Mitarbeiter bearbeiten',
      ...ModalConfig,
    },
    onClose: (opt) => {
      if (!opt) {
        // TODO: Handle error?
        return
      }
      const person = opt.data as Person
      // TODO: Update Person
      console.log(person)
    }
  })
}

const addEmployee = () => {
  dialog.open(PersonDialog, {
    props: {
      header: 'Neuen Mitarbeiter anlegen',
      ...ModalConfig,
    },
    onClose: (opt) => {
      if (!opt) {
        // TODO: Handle error?
        return
      }
      const person = opt.data as Person
      // TODO: Create Person
      console.log(person)
    }
  })
}
</script>