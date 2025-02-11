<template>
  <div class="flex flex-col gap-4">
    <div class="flex flex-col sm:flex-row gap-2 justify-between items-center">
      <div class="flex items-center gap-2 w-full sm:w-auto">
        <InputText
          v-model="search"
          placeholder="Suchen"
        />
        <Button
          :icon="getDisplayIcon"
          @click="toggleEmployeeDisplayType"
        />
      </div>
      <Button
        class="self-end"
        label="Mitarbeiter hinzufügen"
        icon="pi pi-user"
        @click="onCreateEmployee"
      />
    </div>

    <Message
      v-if="employeesErrorMessage.length"
      severity="error"
      :closable="false"
      class="col-span-full"
    >
      {{ employeesErrorMessage }}
    </Message>
    <template v-else-if="filterEmployees.length">
      <div
        v-if="employeeDisplay == 'grid'"
        class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4"
      >
        <EmployeeCard
          v-for="employee in filterEmployees"
          :key="employee.id"
          :employee="employee"
          @on-edit="onEditEmployee"
        />
      </div>
      <div
        v-else
        class="flex flex-col overflow-x-auto pb-2"
      >
        <EmployeeHeaders @on-sort="onSort" />
        <EmployeeRow
          v-for="employee in filterEmployees"
          :key="employee.id"
          :employee="employee"
          @on-edit="onEditEmployee"
        />
      </div>
    </template>
    <p v-else>
      Es gibt noch keine Mitarbeiter
    </p>

    <div
      v-if="employees?.data.length"
      class="self-center"
    >
      <Button
        v-if="!noMoreDataEmployees"
        severity="info"
        label="Mehr anzeigen"
        :loading="isLoadingMore"
        @click="onLoadMoreEmployees"
      />
      <p
        v-else
        class="text-xs opacity-60"
      >
        Keine weiteren Mitarbeitenden ...
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ModalConfig } from '~/config/dialog-props'
import EmployeeDialog from '~/components/dialogs/EmployeeDialog.vue'
import type { EmployeeResponse } from '~/models/employee'
import { RouteNames } from '~/config/routes'

useHead({
  title: 'Mitarbeitende',
})

const { employees, noMoreDataEmployees, pageEmployees, useFetchListEmployees, listEmployees } = useEmployees()
const { toggleEmployeeDisplayType, employeeDisplay } = useSettings()
const dialog = useDialog()

const isLoading = ref(false)
const isLoadingMore = ref(false)
const employeesErrorMessage = ref('')
const search = ref('')

// Init
await useFetchListEmployees()
  .catch((reason) => {
    employeesErrorMessage.value = reason
  })

// Computed
const getDisplayIcon = computed(() => employeeDisplay.value == 'list' ? 'pi pi-microsoft' : 'pi pi-list')
const filterEmployees = computed(() => {
  return employees.value.data
    .filter(t => t.name.toLowerCase().includes(search.value.toLowerCase()))
})

// Functions
const onSort = () => {
  isLoading.value = false
  nextTick(() => {
    isLoading.value = true
    listEmployees(false)
      .then(() => {
        isLoading.value = false
      })
      .catch((err) => {
        if (err !== 'aborted') {
          isLoading.value = false
          employeesErrorMessage.value = err
        }
      })
  })
}

const onCreateEmployee = () => {
  dialog.open(EmployeeDialog, {
    props: {
      header: 'Neuen Mitarbeiter anlegen',
      ...ModalConfig,
    },
  })
}

const onEditEmployee = (employee: EmployeeResponse) => {
  navigateTo({ name: RouteNames.EMPLOYEES_EDIT, params: { id: employee.id } })
}

const onLoadMoreEmployees = async () => {
  isLoadingMore.value = true
  pageEmployees.value += 1
  listEmployees(false)
    .catch((err) => {
      if (err !== 'aborted') {
        employeesErrorMessage.value = err
      }
    })
    .finally(() => isLoadingMore.value = false)
}
</script>
