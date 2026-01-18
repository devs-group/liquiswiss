<template>
  <div class="grid grid-cols-employees items-center *:bg-zinc-100 *:dark:bg-zinc-800 *:border *:border-r-0 *:border-b-0 *:border-gray-600 *:p-1 *:text-sm *:font-bold">
    <div
      class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
      @click="onSort('name')"
    >
      <p class="truncate">
        Name
      </p>
      <i :class="getSortIcon('name')" />
    </div>
    <div
      class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
      @click="onSort('hoursPerMonth')"
    >
      <p class="truncate">
        Arbeitsstunden / Monat
      </p>
      <i :class="getSortIcon('hoursPerMonth')" />
    </div>
    <div
      class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
      @click="onSort('salary')"
    >
      <p class="truncate">
        Lohn
      </p>
      <i :class="getSortIcon('salary')" />
    </div>
    <div
      class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
      @click="onSort('vacationDaysPerYear')"
    >
      <p class="truncate">
        Urlaubstage / Jahr
      </p>
      <i :class="getSortIcon('vacationDaysPerYear')" />
    </div>
    <div
      class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
      @click="onSort('fromDate')"
    >
      <p class="truncate">
        Gültig ab/seit
      </p>
      <i :class="getSortIcon('fromDate')" />
    </div>
    <div
      class="flex items-center gap-2 !border-r cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
      @click="onSort('toDate')"
    >
      <p class="truncate">
        Gültig bis
      </p>
      <i :class="getSortIcon('toDate')" />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { EmployeeSortByType } from '~/utils/types'

const emits = defineEmits(['onSort'])

const { employeeSortBy, employeeSortOrder, setEmployeeSort } = useSettings()

const onSort = (column: EmployeeSortByType) => {
  if (column === employeeSortBy.value) {
    setEmployeeSort(column, employeeSortOrder.value === 'ASC' ? 'DESC' : 'ASC')
  }
  else {
    setEmployeeSort(column, 'ASC')
  }
  emits('onSort')
}

const getSortIcon = (column: EmployeeSortByType) => {
  if (column !== employeeSortBy.value) {
    return 'pi pi-sort'
  }
  return employeeSortOrder.value === 'ASC' ? 'pi pi-sort-up' : 'pi pi-sort-down'
}
</script>
