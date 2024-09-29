<template>
  <div v-if="!!employee" class="flex flex-col gap-4">
    <div class="flex justify-end gap-2 col-span-full">
      <Button @click="onDeleteEmployee" :loading="isSubmitting" label="Mitarbeiter Löschen" icon="pi pi-trash" severity="danger" size="small"/>
    </div>

    <div class="flex justify-between items-center gap-2">
      <hr class="h-0.5 bg-black flex-1"/>
      <p class="text-xl">Allgemeine Informationen</p>
      <hr class="h-0.5 bg-black flex-1"/>
    </div>
    <form @submit.prevent class="grid grid-cols-2 gap-2">
      <div class="flex flex-col gap-2 col-span-full md:col-span-1">
        <label class="text-sm font-bold" for="name">Name*</label>
        <InputText v-model="name" v-bind="nameProps"
                   :class="{'p-invalid': errors['name']?.length}"
                   id="name" type="text"/>
        <small class="text-red-400">{{errors["name"] || '&nbsp;'}}</small>
      </div>

      <div class="flex justify-end gap-2 col-span-full">
        <Button @click="onUpdateEmployee" severity="info" :disabled="!meta.valid || (meta.valid && !meta.dirty) || isSubmitting" :loading="isSubmitting" label="Speichern" icon="pi pi-save" type="submit"/>
      </div>
    </form>

    <div class="flex justify-between items-center gap-2">
      <hr class="h-0.5 bg-black flex-1"/>
      <p class="text-xl">Historie</p>
      <hr class="h-0.5 bg-black flex-1"/>
    </div>
    <Button @click="onCreateEmployeeHistory" class="self-end" :loading="isSubmitting" label="Historie hinzufügen" icon="pi pi-history"/>
    <div v-if="employeeHistories?.data.length" class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <EmployeeHistoryCard @on-edit="onUpdateEmployeeHistory" v-for="employeeHistory in employeeHistories.data" :employee-history="employeeHistory"/>
    </div>

    <div v-if="employeeHistories?.data.length" class="self-center">
      <Button v-if="!noMoreDataEmployeeHistories" severity="info" label="Mehr anzeigen" @click="onLoadMoreEmployeeHistory" :loading="isLoadingMore"/>
      <p v-else class="text-xs opacity-60">Keine weiteren Historien ...</p>
    </div>
  </div>

  <div v-else class="flex flex-col gap-2 items-start bg-red-100 border border-red-200 p-4">
    <span>Mitarbeiter mit der ID "{{route.params.id}}" wurde nicht gefunden</span>
    <NuxtLink :to="{name: Routes.EMPLOYEES, replace: true}">
      <Button label="Zurück zur Übersicht"/>
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
import {ModalConfig} from "~/config/dialog-props";
import {Config} from "~/config/config";
import type {EmployeeFormData, EmployeeHistoryResponse, EmployeeResponse} from "~/models/employee";
import {Routes} from "~/config/routes";
import {useForm} from "vee-validate";
import * as yup from "yup";
import EmployeeHistoryDialog from "~/components/dialogs/EmployeeHistoryDialog.vue";
import EmployeeHistoryCard from "~/components/EmployeeHistoryCard.vue";

const {employeeHistories, noMoreDataEmployeeHistories, pageEmployeeHistories, getEmployee, updateEmployee, deleteEmployee, getEmployeeHistory} = useEmployees()
const dialog = useDialog();
const toast = useToast()
const route = useRoute()
const confirm = useConfirm()

const isSubmitting = ref(false)
const isLoadingMore = ref(false)
const employeeID = Number(route.params.id);
const employee = ref<EmployeeResponse|null>()

if (route.params.id && !Number.isNaN(employeeID) && Number.isInteger(employeeID)) {
  employee.value = await getEmployee(employeeID)
}

await getEmployeeHistory(employeeID)

const { defineField, errors, handleSubmit, meta, resetForm } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
  }),
  initialValues: {
    id: employee.value?.id ?? undefined,
    name: employee.value?.name ?? '',
  } as EmployeeFormData
});

const [name, nameProps] = defineField('name')

const onUpdateEmployee = handleSubmit((values) => {
  isSubmitting.value = true
  updateEmployee(values)
      .then(async () => {
        employee.value = await getEmployee(employeeID)
        resetForm({values: values})
        toast.add({
          summary: 'Erfolg',
          detail: `Mitarbeiter wurde bearbeitet`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
      .finally(() => {
        isSubmitting.value = false
      })
})

const onDeleteEmployee = (payload: MouseEvent) => {
  confirm.require({
    target: payload.currentTarget as HTMLElement,
    message: 'Mitarbeiter vollständig löschen?',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      if (employee.value) {
        isSubmitting.value = true
        deleteEmployee(employeeID)
            .then(() => {
              navigateTo({name: Routes.EMPLOYEES, replace: true})
              toast.add({
                summary: 'Erfolg',
                detail: `Mitarbeiter "${employee.value!.name}" wurde gelöscht`,
                severity: 'success',
                life: Config.TOAST_LIFE_TIME,
              })
            })
            .catch(() => {
              toast.add({
                summary: 'Fehler',
                detail: `Mitarbeiter konnte nicht gelöscht werden`,
                severity: 'error',
                life: Config.TOAST_LIFE_TIME,
              })
            })
            .finally(() => {
              isSubmitting.value = false
            })
      }
    },
    reject: () => {
    }
  });
}

const onCreateEmployeeHistory = () => {
  dialog.open(EmployeeHistoryDialog, {
    data: {
      employeeID,
    },
    props: {
      header: 'Neue Historie anlegen',
      ...ModalConfig,
    },
  })
}

const onUpdateEmployeeHistory = (employeeHistory: EmployeeHistoryResponse) => {
  dialog.open(EmployeeHistoryDialog, {
    data: {
      employeeID,
      employeeHistory,
    },
    props: {
      header: 'Historie bearbeiten',
      ...ModalConfig,
    },
  })
}
const onLoadMoreEmployeeHistory = async (event: MouseEvent) => {
  isLoadingMore.value = true
  pageEmployeeHistories.value += 1
  await getEmployeeHistory(employeeID)
  isLoadingMore.value = false
}
</script>