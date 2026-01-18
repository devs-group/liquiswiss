<template>
  <div class="grid grid-cols-transactions items-center *:bg-zinc-100 *:dark:bg-zinc-800 *:border *:border-r-0 *:border-b-0 *:border-gray-600 *:p-1 *:text-sm *:font-bold">
    <div
      class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
      @click="onSort('name')"
    >
      <p class="truncate">
        Name
      </p>
      <i :class="getSortIcon('name')" />
    </div>
    <div class="flex items-center justify-center cursor-default">
      <p class="truncate">
        Status
      </p>
    </div>
    <div
      class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
      @click="onSort('startDate')"
    >
      <p class="truncate">
        Start
      </p>
      <i :class="getSortIcon('startDate')" />
    </div>
    <div
      class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
      @click="onSort('endDate')"
    >
      <p class="truncate">
        Ende
      </p>
      <i :class="getSortIcon('endDate')" />
    </div>
    <div class="flex items-center gap-2 cursor-default">
      <p class="truncate">
        Ausführung
      </p>
    </div>
    <div
      class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
      @click="onSort('amount')"
    >
      <p class="truncate">
        Betrag
      </p>
      <i :class="getSortIcon('amount')" />
    </div>
    <div class="flex items-center gap-2 cursor-default">
      <p class="truncate">
        MWST
      </p>
    </div>
    <div
      class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
      @click="onSort('cycle')"
    >
      <p class="truncate">
        Häufigkeit
      </p>
      <i :class="getSortIcon('cycle')" />
    </div>
    <div
      class="flex items-center gap-2 cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
      @click="onSort('category')"
    >
      <p class="truncate">
        Kategorie
      </p>
      <i :class="getSortIcon('category')" />
    </div>
    <div
      class="flex items-center gap-2 !border-r cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
      @click="onSort('employee')"
    >
      <p class="truncate">
        Mitarbeiter
      </p>
      <i :class="getSortIcon('employee')" />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TransactionSortByType } from '~/utils/types'

const emits = defineEmits(['onSort'])

const { transactionSortBy, transactionSortOrder, setTransactionSort } = useSettings()

const onSort = (column: TransactionSortByType) => {
  if (column === transactionSortBy.value) {
    setTransactionSort(column, transactionSortOrder.value === 'ASC' ? 'DESC' : 'ASC')
  }
  else {
    setTransactionSort(column, 'ASC')
  }
  emits('onSort')
}

const getSortIcon = (column: TransactionSortByType) => {
  if (column !== transactionSortBy.value) {
    return 'pi pi-sort'
  }
  return transactionSortOrder.value === 'ASC' ? 'pi pi-sort-up' : 'pi pi-sort-down'
}
</script>
