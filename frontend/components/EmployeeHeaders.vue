<template>
  <div class="grid grid-cols-employees items-center *:bg-gray-100 *:border *:border-r-0 *:border-b-0 *:border-gray-600 *:p-1 *:text-sm *:font-bold">
    <div @click="onSort('name')" class="flex items-center gap-2 cursor-pointer hover:bg-gray-50 transition-colors duration-300">
      <p class="truncate">Name</p>
      <i :class="getSortIcon('name')"></i>
    </div>
    <div @click="onSort('hoursPerMonth')" class="flex items-center gap-2 cursor-pointer hover:bg-gray-50 transition-colors duration-300">
      <p class="truncate">Arbeitsstunden / Monat</p>
      <i :class="getSortIcon('hoursPerMonth')"></i>
    </div>
    <div @click="onSort('salaryPerMonth')" class="flex items-center gap-2 cursor-pointer hover:bg-gray-50 transition-colors duration-300">
      <p class="truncate">Lohn / Monat</p>
      <i :class="getSortIcon('salaryPerMonth')"></i>
    </div>
    <div @click="onSort('vacationDaysPerYear')" class="flex items-center gap-2 cursor-pointer hover:bg-gray-50 transition-colors duration-300">
      <p class="truncate">Urlaubstage / Jahr</p>
      <i :class="getSortIcon('vacationDaysPerYear')"></i>
    </div>
    <div @click="onSort('fromDate')" class="flex items-center gap-2 cursor-pointer hover:bg-gray-50 transition-colors duration-300">
      <p class="truncate">Gültig ab/seit</p>
      <i :class="getSortIcon('fromDate')"></i>
    </div>
    <div @click="onSort('toDate')" class="flex items-center gap-2 !border-r cursor-pointer hover:bg-gray-50 transition-colors duration-300">
      <p class="truncate">Gültig bis</p>
      <i :class="getSortIcon('toDate')"></i>
    </div>
  </div>
</template>

<script setup lang="ts">
import type {EmployeeSortByType} from "~/utils/types";

const emits = defineEmits(['onSort'])

const {employeeSortBy, employeeSortOrder} = useSettings()

const onSort = (column: EmployeeSortByType) => {
  if (column == employeeSortBy.value) {
    employeeSortOrder.value = employeeSortOrder.value == 'ASC' ? 'DESC' : 'ASC'
  } else {
    employeeSortBy.value = column
    employeeSortOrder.value = 'ASC'
  }
  emits('onSort')
}

const getSortIcon = (column: EmployeeSortByType) => {
  if (column !== employeeSortBy.value) {
    return 'pi pi-sort';
  }
  return employeeSortOrder.value == 'ASC' ? 'pi pi-sort-up' : 'pi pi-sort-down'
}
</script>
