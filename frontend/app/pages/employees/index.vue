<template>
  <div class="flex flex-col gap-4 relative">
    <div class="flex flex-col sm:flex-row gap-2 justify-between items-center">
      <div class="flex items-center gap-2 w-full sm:w-auto">
        <SearchInput
          :model-value="searchEmployees"
          @update:model-value="onSearch"
          @clear="onClearSearch"
        />
        <Button
          :icon="getDisplayIcon"
          @click="toggleEmployeeDisplayType"
        />
        <Button
          v-tooltip.bottom="employeeHideTerminated ? 'Ausgetretene Mitarbeiter anzeigen' : 'Ausgetretene Mitarbeiter ausblenden'"
          :icon="employeeHideTerminated ? 'pi pi-eye-slash' : 'pi pi-eye'"
          :severity="employeeHideTerminated ? 'secondary' : 'contrast'"
          @click="onToggleHideTerminated"
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
    <template v-else-if="employees.data.length">
      <div
        v-if="employeeDisplay == 'grid'"
        class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4"
      >
        <EmployeeCard
          v-for="employee in employees.data"
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
          v-for="employee in employees.data"
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

    <FullProgressSpinner :show="isLoading" />
  </div>
</template>

<script setup lang="ts">
import { ModalConfig } from '~/config/dialog-props'
import EmployeeDialog from '~/components/dialogs/EmployeeDialog.vue'
import type { EmployeeResponse } from '~/models/employee'
import { RouteNames } from '~/config/routes'
import FullProgressSpinner from '~/components/FullProgressSpinner.vue'

useHead({
  title: 'Mitarbeitende',
})

const { employees, noMoreDataEmployees, pageEmployees, searchEmployees, useFetchListEmployees, listEmployees } = useEmployees()
const { toggleEmployeeDisplayType, toggleEmployeeHideTerminated, employeeDisplay, employeeHideTerminated } = useSettings()
const dialog = useDialog()

const isLoading = ref(false)
const isLoadingMore = ref(false)
const employeesErrorMessage = ref('')

// Computed
const getDisplayIcon = computed(() => employeeDisplay.value == 'list' ? 'pi pi-microsoft' : 'pi pi-list')

// Debounced search
let searchTimeout: ReturnType<typeof setTimeout> | null = null
const onSearch = (value: string) => {
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
  const trimmedValue = value.trim()
  if (trimmedValue === searchEmployees.value) {
    return
  }
  searchTimeout = setTimeout(() => {
    searchEmployees.value = trimmedValue
    pageEmployees.value = 1
    isLoading.value = true
    listEmployees(false)
      .catch((err) => {
        if (err !== 'aborted') {
          employeesErrorMessage.value = err
        }
      })
      .finally(() => {
        isLoading.value = false
      })
  }, 300)
}

const onClearSearch = () => {
  if (searchEmployees.value === '') return
  searchEmployees.value = ''
  pageEmployees.value = 1
  isLoading.value = true
  listEmployees(false)
    .catch((err) => {
      if (err !== 'aborted') {
        employeesErrorMessage.value = err
      }
    })
    .finally(() => {
      isLoading.value = false
    })
}

// Init
await useFetchListEmployees()
  .catch((reason) => {
    employeesErrorMessage.value = reason
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

const onToggleHideTerminated = () => {
  toggleEmployeeHideTerminated()
  pageEmployees.value = 1
  isLoading.value = true
  listEmployees(false)
    .catch((err) => {
      if (err !== 'aborted') {
        employeesErrorMessage.value = err
      }
    })
    .finally(() => {
      isLoading.value = false
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
