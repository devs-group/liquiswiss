<template>
  <div class="flex flex-col gap-4">
    <Button @click="addEmployee" class="self-end" label="Mitarbeiter hinzufügen" icon="pi pi-user"/>

    <div v-if="employees?.data.length" class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <EmployeeCard @on-edit="onEdit" v-for="employee in employees.data" :employee="employee"/>
    </div>
    <p v-else>Es gibt noch keine Mitarbeiter</p>

    <div v-if="employees?.data.length" class="self-center">
      <Button v-if="!noMoreData" severity="info" label="Mehr anzeigen" @click="onLoadMore" :loading="isLoadingMore"/>
      <p v-else class="text-xs opacity-60">Keine weiteren Mitarbeiter ...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ModalConfig} from "~/config/dialog-props";
import {Config} from "~/config/config";
import EmployeeDialog from "~/components/dialogs/EmployeeDialog.vue";
import type {EmployeeFormData, EmployeeResponse} from "~/models/employee";

const {employees, noMoreData, page, getEmployees, getEmployeesPagination, createEmployee, updateEmployee} = useEmployees()
const dialog = useDialog();
const toast = useToast()
const isLoadingMore = ref(false)

await getEmployees(false)

const onLoadMore = async (event: MouseEvent) => {
  isLoadingMore.value = true
  page.value += 1
  await getEmployees(false)
  isLoadingMore.value = false
}
const onEdit = (employee: EmployeeResponse) => {
  dialog.open(EmployeeDialog, {
    data: {
      employee,
    },
    props: {
      header: 'Mitarbeiter bearbeiten',
      ...ModalConfig,
    },
    onClose: async (opt) => {
      if (!opt?.data) {
        // TODO: Handle error?
        return
      }
      const payload = opt.data as EmployeeFormData|'deleted'

      if (payload == 'deleted') {
        await getEmployeesPagination()
        toast.add({
          summary: 'Erfolg',
          detail: `Mitarbeiter "${employee.name}" wurde gelöscht`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
        return
      }

      updateEmployee(payload)
          .then(async () => {
            toast.add({
              summary: 'Erfolg',
              detail: `Mitarbeiter "${employee.name}" wurde aktualisiert`,
              severity: 'success',
              life: Config.TOAST_LIFE_TIME,
            })
          })
    }
  })
}

const addEmployee = () => {
  dialog.open(EmployeeDialog, {
    props: {
      header: 'Neuen Mitarbeiter anlegen',
      ...ModalConfig,
    },
    onClose: (opt) => {
      if (!opt?.data) {
        // TODO: Handle error?
        return
      }
      const payload = opt.data as EmployeeFormData
      createEmployee(payload)
          .then(async () => {
            toast.add({
              summary: 'Erfolg',
              detail: `Mitarbeiter "${payload.name}" wurde angelegt`,
              severity: 'success',
              life: Config.TOAST_LIFE_TIME,
            })
          })
    }
  })
}
</script>