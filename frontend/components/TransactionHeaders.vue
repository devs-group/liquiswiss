<template>
  <div class="grid grid-cols-transactions items-center *:bg-zinc-100 *:dark:bg-zinc-800 *:border *:border-r-0 *:border-b-0 *:border-gray-600 *:p-1 *:text-sm *:font-bold">
    <div @click="onSort('name')" class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300">
      <p class="truncate">Name</p>
      <i :class="getSortIcon('name')"></i>
    </div>
    <div @click="onSort('startDate')" class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300">
      <p class="truncate">Start</p>
      <i :class="getSortIcon('startDate')"></i>
    </div>
    <div @click="onSort('endDate')" class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300">
      <p class="truncate">Ende</p>
      <i :class="getSortIcon('endDate')"></i>
    </div>
    <div class="flex items-center gap-2 cursor-default">
      <p class="truncate">Ausführung</p>
    </div>
    <div @click="onSort('amount')" class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300">
      <p class="truncate">Betrag</p>
      <i :class="getSortIcon('amount')"></i>
    </div>
    <div class="flex items-center gap-2 cursor-default">
      <p class="truncate">MWST</p>
    </div>
    <div @click="onSort('cycle')" class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300">
      <p class="truncate">Häufigkeit</p>
      <i :class="getSortIcon('cycle')"></i>
    </div>
    <div @click="onSort('category')" class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300">
      <p class="truncate">Kategorie</p>
      <i :class="getSortIcon('category')"></i>
    </div>
    <div @click="onSort('employee')" class="flex items-center gap-2 !border-r cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300">
      <p class="truncate">Mitarbeiter</p>
      <i :class="getSortIcon('employee')"></i>
    </div>
  </div>
</template>

<script setup lang="ts">
import type {TransactionSortByType} from "~/utils/types";

const emits = defineEmits(['onSort'])

const {transactionSortBy, transactionSortOrder} = useSettings()

const onSort = (column: TransactionSortByType) => {
  if (column == transactionSortBy.value) {
    transactionSortOrder.value = transactionSortOrder.value == 'ASC' ? 'DESC' : 'ASC'
  } else {
    transactionSortBy.value = column
    transactionSortOrder.value = 'ASC'
  }
  emits('onSort')
}

const getSortIcon = (column: TransactionSortByType) => {
  if (column !== transactionSortBy.value) {
    return 'pi pi-sort';
  }
  return transactionSortOrder.value == 'ASC' ? 'pi pi-sort-up' : 'pi pi-sort-down'
}
</script>
