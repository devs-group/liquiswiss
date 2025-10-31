<template>
  <div class="grid grid-cols-bank-accounts items-center *:bg-zinc-100 *:dark:bg-zinc-800 *:border *:border-r-0 *:border-b-0 *:border-gray-600 *:p-1 *:text-sm *:font-bold">
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
      class="flex items-center gap-2 !border-r cursor-pointer hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
      @click="onSort('amount')"
    >
      <p class="truncate">
        Kontostand
      </p>
      <i :class="getSortIcon('amount')" />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { BankAccountSortByType } from '~/utils/types'

const emits = defineEmits(['onSort'])

const { bankAccountSortBy, bankAccountSortOrder } = useSettings()

const onSort = (column: BankAccountSortByType) => {
  if (column == bankAccountSortBy.value) {
    bankAccountSortOrder.value = bankAccountSortOrder.value == 'ASC' ? 'DESC' : 'ASC'
  }
  else {
    bankAccountSortBy.value = column
    bankAccountSortOrder.value = 'ASC'
  }
  emits('onSort')
}

const getSortIcon = (column: BankAccountSortByType) => {
  if (column !== bankAccountSortBy.value) {
    return 'pi pi-sort'
  }
  return bankAccountSortOrder.value == 'ASC' ? 'pi pi-sort-up' : 'pi pi-sort-down'
}
</script>
