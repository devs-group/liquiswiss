<template>
  <div class="flex flex-col gap-4">
    <Button @click="addEmployee" class="self-end" label="Mitarbeiter hinzufügen" icon="pi pi-user"/>

    <div v-if="people.length" class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <PersonCard @on-edit="onEdit" v-for="person in people" :person="person"/>
    </div>
    <p v-else>Es gibt noch keine Mitarbeiter</p>
  </div>
</template>

<script setup lang="ts">
import type {StrapiPerson} from "~/models/person";
import {ModalConfig} from "~/config/dialog-props";
import PersonDialog from "~/components/dialogs/PersonDialog.vue";
import {Config} from "~/config/config";

const dialog = useDialog();
const toast = useToast()

const {data: people} = await useFetch('/api/team')

const onEdit = (person: StrapiPerson) => {
  dialog.open(PersonDialog, {
    data: {
      person: person,
    },
    props: {
      header: 'Mitarbeiter bearbeiten',
      ...ModalConfig,
    },
    onClose: (opt) => {
      if (!opt?.data) {
        // TODO: Handle error?
        return
      }
      const payload = opt.data as StrapiPerson
      $fetch('/api/team', {
        method: 'post',
        body: payload,
      }).then(() => {
        toast.add({
          summary: 'Erfolg',
          detail: `Mitarbeiter wurde aktualisiert`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
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
      if (!opt?.data) {
        // TODO: Handle error?
        return
      }
      const payload = opt.data as StrapiPerson
      $fetch('/api/team', {
        method: 'post',
        body: payload,
      }).then((resp) => {
        toast.add({
          summary: 'Erfolg',
          detail: `Mitarbeiter "${resp.attributes.name}" wurde angelegt`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
    }
  })
}
</script>