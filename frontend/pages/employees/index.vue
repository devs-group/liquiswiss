<template>
  <div class="flex flex-col gap-4">
    <Button @click="onCreateEmployee" class="self-end" label="Mitarbeiter hinzufügen" icon="pi pi-user"/>

    <Message v-if="employeesErrorMessage.length" severity="error" :closable="false" class="col-span-full">{{employeesErrorMessage}}</Message>
    <div v-else-if="employees?.data.length" class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <EmployeeCard @on-edit="onEditEmployee" v-for="employee in employees.data" :employee="employee"/>
    </div>
    <p v-else>Es gibt noch keine Mitarbeiter</p>

    <div v-if="employees?.data.length" class="self-center">
      <Button v-if="!noMoreDataEmployees" severity="info" label="Mehr anzeigen" @click="onLoadMoreEmployees" :loading="isLoadingMore"/>
      <p v-else class="text-xs opacity-60">Keine weiteren Mitarbeiter ...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ModalConfig} from "~/config/dialog-props";
import EmployeeDialog from "~/components/dialogs/EmployeeDialog.vue";
import type {EmployeeResponse} from "~/models/employee";
import {Routes} from "~/config/routes";

const {employees, noMoreDataEmployees, pageEmployees, getEmployees} = useEmployees()
const dialog = useDialog();

const isLoadingMore = ref(false)
const employeesErrorMessage = ref('')

await getEmployees(false)
    .catch(() => {
      employeesErrorMessage.value = 'Mitarbeiter konnten nicht geladen werden'
    })

const onCreateEmployee = () => {
  dialog.open(EmployeeDialog, {
    props: {
      header: 'Neuen Mitarbeiter anlegen',
      ...ModalConfig,
    },
  })
}

const onEditEmployee = (employee: EmployeeResponse) => {
  navigateTo({name: Routes.EMPLOYEE_EDIT, params: {id: employee.id}})
}

const onLoadMoreEmployees = async (event: MouseEvent) => {
  isLoadingMore.value = true
  pageEmployees.value += 1
  await getEmployees(false)
  isLoadingMore.value = false
}
</script>