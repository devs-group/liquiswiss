<template>
  <div
    v-if="!employeeLoadErrorMessage.length"
    class="flex flex-col gap-4"
  >
    <div class="flex justify-between gap-2 col-span-full">
      <Button
        icon="pi pi-arrow-left"
        label="Zurück"
        @click="onGoBack"
      />
      <Button
        :loading="isSubmitting"
        label="Löschen"
        icon="pi pi-trash"
        severity="danger"
        size="small"
        @click="onDeleteEmployee"
      />
    </div>

    <Message
      v-if="employeeDeleteErrorMessage.length"
      severity="error"
      :life="Config.MESSAGE_LIFE_TIME"
      :sticky="false"
      :closable="false"
      class="col-span-full"
    >
      {{ employeeDeleteErrorMessage }}
    </Message>

    <div class="flex justify-between items-center gap-2">
      <hr class="h-0.5 bg-black flex-1">
      <p class="text-xl">
        Allgemeine Informationen
      </p>
      <hr class="h-0.5 bg-black flex-1">
    </div>
    <form
      class="grid grid-cols-2 gap-2"
      @submit.prevent
    >
      <div class="flex flex-col gap-2 col-span-full md:col-span-1">
        <label
          class="text-sm font-bold"
          for="name"
        >Name*</label>
        <InputText
          v-bind="nameProps"
          id="name"
          v-model="name"
          :class="{ 'p-invalid': errors['name']?.length }"
          type="text"
        />
        <small class="text-liqui-red">{{ errors["name"] || "&nbsp;" }}</small>
      </div>

      <Message
        v-if="employeeUpdateMessage.length"
        severity="success"
        :life="Config.MESSAGE_LIFE_TIME"
        :sticky="false"
        :closable="false"
        class="col-span-full"
      >
        {{ employeeUpdateMessage }}
      </Message>
      <Message
        v-if="employeeUpdateErrorMessage.length"
        severity="error"
        :life="Config.MESSAGE_LIFE_TIME"
        :sticky="false"
        :closable="false"
        class="col-span-full"
      >
        {{ employeeUpdateErrorMessage }}
      </Message>

      <div class="flex justify-end gap-2 col-span-full">
        <Button
          :disabled="!meta.valid || (meta.valid && !meta.dirty) || isSubmitting"
          :loading="isSubmitting"
          label="Speichern"
          icon="pi pi-save"
          type="submit"
          @click="onUpdateEmployee"
        />
      </div>
    </form>

    <div class="flex justify-between items-center gap-2">
      <hr class="h-0.5 bg-black flex-1">
      <p class="text-xl">
        Lohnverlauf
      </p>
      <hr class="h-0.5 bg-black flex-1">
    </div>
    <Button
      class="self-end"
      :loading="isSubmitting"
      label="Lohn hinzufügen"
      icon="pi pi-history"
      @click="onCreateSalary"
    />

    <Message
      v-if="salaryErrorMessage.length"
      severity="error"
      :closable="false"
      class="col-span-full"
    >
      {{ salaryErrorMessage }}
    </Message>

    <div
      v-if="salaries?.data.length"
      class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4 gap-4"
    >
      <SalaryCard
        v-for="salary in salaries.data"
        :key="salary.id"
        :salary="salary"
        :is-active="employee?.salaryID == salary.id"
        @on-edit="onUpdateSalary"
        @on-clone="onCloneSalary"
      />
    </div>

    <div
      v-if="salaries?.data.length"
      class="self-center"
    >
      <Button
        v-if="!noMoreDataSalaries"
        severity="info"
        label="Mehr anzeigen"
        :loading="isLoadingMore"
        @click="onLoadMoreSalaries"
      />
      <p
        v-else
        class="text-xs opacity-60"
      >
        Keine weiteren Löhne ...
      </p>
    </div>
    <p
      v-else
      class="text-xs opacity-60"
    >
      Mitarbeiter hat noch keinen Lohn. Erstelle den
      <a
        class="underline cursor-pointer font-bold"
        @click="onCreateSalary"
      >ersten Lohn</a>!
    </p>
  </div>

  <div
    v-else
    class="flex flex-col gap-2 items-start bg-liqui-red border border-red-200 p-4"
  >
    <span>{{ employeeLoadErrorMessage }}</span>
    <NuxtLink :to="{ name: RouteNames.EMPLOYEES, replace: true }">
      <Button label="Zurück zur Übersicht" />
    </NuxtLink>
  </div>
</template>

<script setup lang="ts">
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import type { DynamicDialogCloseOptions } from 'primevue/dynamicdialogoptions'
import { ModalConfig } from '~/config/dialog-props'
import { Config } from '~/config/config'
import type { EmployeeFormData, EmployeeResponse, SalaryResponse } from '~/models/employee'
import { RouteNames } from '~/config/routes'
import SalaryDialog from '~/components/dialogs/SalaryDialog.vue'
import SalaryCard from '~/components/SalaryCard.vue'

const {
  useFetchGetEmployee,
  getEmployee,
  updateEmployee,
  deleteEmployee,
} = useEmployees()
const {
  salaries,
  noMoreDataSalaries,
  pageSalaries,
  useFetchListSalaries,
  listSalaries,
} = useSalaries()
const dialog = useDialog()
const toast = useToast()
const route = useRoute()
const confirm = useConfirm()

const isSubmitting = ref(false)
const isLoadingMore = ref(false)
const employeeID = Number(route.params.id)
const employee = ref<EmployeeResponse | null>()
const employeeLoadErrorMessage = ref('')
const employeeUpdateMessage = ref('')
const employeeUpdateErrorMessage = ref('')
const employeeDeleteErrorMessage = ref('')
const salaryErrorMessage = ref('')

if (
  route.params.id
  && !Number.isNaN(employeeID)
  && Number.isInteger(employeeID)
) {
  await useFetchGetEmployee(employeeID)
    .then((value) => {
      employee.value = value
    })
    .catch((reason) => {
      employeeLoadErrorMessage.value = reason
    })
}
else {
  await navigateTo({ name: RouteNames.EMPLOYEES })
}

if (employee.value) {
  await useFetchListSalaries(employeeID).catch((reason) => {
    salaryErrorMessage.value = reason
  })
}

useHead({
  title: `${employee.value?.name ?? ''} Mitarbeitende`.trim(),
})

const { defineField, errors, handleSubmit, meta, resetForm } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
  }),
  initialValues: {
    id: employee.value?.id ?? undefined,
    name: employee.value?.name ?? '',
  } as EmployeeFormData,
})

const [name, nameProps] = defineField('name')

const onGoBack = () => {
  navigateTo({ name: RouteNames.EMPLOYEES, replace: true })
}

const onUpdateEmployee = handleSubmit((values) => {
  employeeUpdateErrorMessage.value = ''
  employeeUpdateMessage.value = ''
  isSubmitting.value = true
  updateEmployee(values)
    .then(async (updatedEmployee) => {
      employee.value = updatedEmployee
      resetForm({ values: values })
      employeeUpdateMessage.value = 'Mitarbeiter wurde bearbeitet'
    })
    .catch(() => {
      employeeUpdateErrorMessage.value
        = 'Mitarbeiter konnte nicht bearbeitet werden'
    })
    .finally(() => {
      isSubmitting.value = false
    })
})

const onDeleteEmployee = () => {
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
            navigateTo({ name: RouteNames.EMPLOYEES, replace: true })
            toast.add({
              summary: 'Erfolg',
              detail: `Mitarbeiter "${employee.value!.name}" wurde gelöscht`,
              severity: 'success',
              life: Config.TOAST_LIFE_TIME,
            })
          })
          .catch(() => {
            employeeDeleteErrorMessage.value
              = 'Mitarbeiter konnte nicht gelöscht werden'
          })
          .finally(() => {
            isSubmitting.value = false
          })
      }
    },
    reject: () => {},
  })
}

const onCreateSalary = () => {
  dialog.open(SalaryDialog, {
    data: {
      employeeID,
    },
    props: {
      header: 'Neuen Lohn anlegen',
      ...ModalConfig,
    },
    onClose: onClosedSalaryDialog,
  })
}

const onUpdateSalary = (salary: SalaryResponse) => {
  dialog.open(SalaryDialog, {
    data: {
      employeeID,
      salary,
    },
    props: {
      header: 'Lohn bearbeiten',
      ...ModalConfig,
    },
    onClose: onClosedSalaryDialog,
  })
}

const onCloneSalary = (salary: SalaryResponse) => {
  dialog.open(SalaryDialog, {
    data: {
      employeeID,
      salary,
      isClone: true,
    },
    props: {
      header: 'Lohn klonen',
      ...ModalConfig,
    },
    onClose: onClosedSalaryDialog,
  })
}

const onClosedSalaryDialog = (options: DynamicDialogCloseOptions) => {
  if (options?.data === true) {
    // Refetch employee to set proper active salary
    getEmployee(employeeID)
      .then((value) => {
        employee.value = value
      })
      .catch((reason) => {
        employeeLoadErrorMessage.value = reason
      })
  }
}

const onLoadMoreSalaries = async () => {
  isLoadingMore.value = true
  pageSalaries.value += 1
  await listSalaries(employeeID)
  isLoadingMore.value = false
}
</script>
