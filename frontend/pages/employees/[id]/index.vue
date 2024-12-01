<template>
  <div v-if="!employeeLoadErrorMessage.length" class="flex flex-col gap-4">
    <div class="flex justify-between gap-2 col-span-full">
      <Button @click="onGoBack" icon="pi pi-arrow-left" label="Zurück"/>
      <Button @click="onDeleteEmployee" :loading="isSubmitting" label="Löschen" icon="pi pi-trash" severity="danger" size="small"/>
    </div>

    <Message v-if="employeeDeleteErrorMessage.length" severity="error" :life="Config.MESSAGE_LIFE_TIME" :sticky="false" :closable="false" class="col-span-full">{{employeeDeleteErrorMessage}}</Message>

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

      <Message v-if="employeeUpdateMessage.length" severity="success" :life="Config.MESSAGE_LIFE_TIME" :sticky="false" :closable="false" class="col-span-full">{{employeeUpdateMessage}}</Message>
      <Message v-if="employeeUpdateErrorMessage.length" severity="error" :life="Config.MESSAGE_LIFE_TIME" :sticky="false" :closable="false" class="col-span-full">{{employeeUpdateErrorMessage}}</Message>

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

    <Message v-if="historyErrorMessage.length" severity="error" :closable="false" class="col-span-full">{{historyErrorMessage}}</Message>

    <div v-if="employeeHistories?.data.length" class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <EmployeeHistoryCard @on-edit="onUpdateEmployeeHistory" @on-clone="onCloneEmployeeHistory" v-for="employeeHistory in employeeHistories.data" :employee-history="employeeHistory" :is-active="employee?.historyID == employeeHistory.id"/>
    </div>

    <div v-if="employeeHistories?.data.length" class="self-center">
      <Button v-if="!noMoreDataEmployeeHistories" severity="info" label="Mehr anzeigen" @click="onLoadMoreEmployeeHistory" :loading="isLoadingMore"/>
      <p v-else class="text-xs opacity-60">Keine weiteren Historien ...</p>
    </div>
    <p v-else class="text-xs opacity-60">Mitarbeiter hat noch keine Historien. Erstelle die <a @click="onCreateEmployeeHistory" class="underline cursor-pointer font-bold">erste Historie</a>!</p>

  </div>

  <div v-else class="flex flex-col gap-2 items-start bg-red-100 border border-red-200 p-4">
    <span>{{ employeeLoadErrorMessage }}</span>
    <NuxtLink :to="{name: RouteNames.EMPLOYEES, replace: true}">
      <Button label="Zurück zur Übersicht"/>
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
import {ModalConfig} from "~/config/dialog-props";
import {Config} from "~/config/config";
import type {
  EmployeeFormData,
  EmployeeHistoryResponse,
  EmployeeResponse,
  ListEmployeeHistoryResponse
} from "~/models/employee";
import {RouteNames} from "~/config/routes";
import {useForm} from "vee-validate";
import * as yup from "yup";
import EmployeeHistoryDialog from "~/components/dialogs/EmployeeHistoryDialog.vue";
import EmployeeHistoryCard from "~/components/EmployeeHistoryCard.vue";

const {employeeHistories, noMoreDataEmployeeHistories, pageEmployeeHistories, useFetchGetEmployee, getEmployee, updateEmployee, deleteEmployee, useFetchListEmployeeHistory, listEmployeeHistory, limitEmployeeHistories, setEmployeeHistories} = useEmployees()
const dialog = useDialog();
const toast = useToast()
const route = useRoute()
const confirm = useConfirm()

const isSubmitting = ref(false)
const isLoadingMore = ref(false)
const employeeID = Number(route.params.id);
const employee = ref<EmployeeResponse|null>()
const employeeLoadErrorMessage = ref('')
const employeeUpdateMessage = ref('')
const employeeUpdateErrorMessage = ref('')
const employeeDeleteErrorMessage = ref('')
const historyErrorMessage = ref('')

if (route.params.id && !Number.isNaN(employeeID) && Number.isInteger(employeeID)) {
   await useFetchGetEmployee(employeeID)
       .then(value => {
         employee.value = value
       })
      .catch(reason => {
        employeeLoadErrorMessage.value = reason
      })
} else {
  await navigateTo({name: RouteNames.EMPLOYEES})
}

if (employee.value) {
  await useFetchListEmployeeHistory(employeeID)
      .catch(reason => {
        historyErrorMessage.value = reason
      })
}

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

const onGoBack = () => {
  navigateTo({name: RouteNames.EMPLOYEES, replace: true})
}

const onUpdateEmployee = handleSubmit((values) => {
  employeeUpdateErrorMessage.value = ''
  employeeUpdateMessage.value = ''
  isSubmitting.value = true
  updateEmployee(values)
      .then(async (updatedEmployee) => {
        employee.value = updatedEmployee
        resetForm({values: values})
        employeeUpdateMessage.value = 'Mitarbeiter wurde bearbeitet'
      })
      .catch(() => {
        employeeUpdateErrorMessage.value = 'Mitarbeiter konnte nicht bearbeitet werden'
      })
      .finally(() => {
        isSubmitting.value = false
      })
})

const onDeleteEmployee = (payload: MouseEvent) => {
  confirm.require({
    header: 'Löschen',
    message: `Mitarbeiter "${employee.value!.name}" vollständig löschen?`,
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      if (employee.value) {
        employeeDeleteErrorMessage.value = ''
        isSubmitting.value = true
        deleteEmployee(employeeID)
            .then(() => {
              navigateTo({name: RouteNames.EMPLOYEES, replace: true})
              toast.add({
                summary: 'Erfolg',
                detail: `Mitarbeiter "${employee.value!.name}" wurde gelöscht`,
                severity: 'success',
                life: Config.TOAST_LIFE_TIME,
              })
            })
            .catch(() => {
              employeeDeleteErrorMessage.value = 'Mitarbeiter konnte nicht gelöscht werden'
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
    onClose: onClosedHistoryDialog
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
    onClose: onClosedHistoryDialog
  })
}

const onCloneEmployeeHistory = (employeeHistory: EmployeeHistoryResponse) => {
  dialog.open(EmployeeHistoryDialog, {
    data: {
      employeeID,
      employeeHistory,
      isClone: true,
    },
    props: {
      header: 'Historie klonen',
      ...ModalConfig,
    },
    onClose: onClosedHistoryDialog
  })
}

const onClosedHistoryDialog = (options: any) => {
  if (options?.data === true) {
    // Refetch employee to set proper active history
    getEmployee(employeeID)
        .then((value) => {
          employee.value = value
        })
        .catch(reason => {
          employeeLoadErrorMessage.value = reason
        })
  }
}

const onLoadMoreEmployeeHistory = async (event: MouseEvent) => {
  isLoadingMore.value = true
  pageEmployeeHistories.value += 1
  await listEmployeeHistory(employeeID)
  isLoadingMore.value = false
}
</script>