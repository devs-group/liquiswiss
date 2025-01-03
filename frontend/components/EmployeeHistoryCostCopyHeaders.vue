<template>
  <div class="grid grid-cols-employee-history-costs-copy items-center *:bg-zinc-100 *:dark:bg-zinc-800 *:border *:border-r-0 *:border-b-0 *:border-gray-600 *:p-1 *:text-sm *:font-bold">
    <div class="flex justify-center">
      <Checkbox
        v-model="model"
        binary
      />
    </div>
    <div
      class="flex items-center gap-2 cursor-default hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
    >
      <p class="truncate">
        Label
      </p>
    </div>
    <div
      class="flex items-center gap-2 cursor-default hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
    >
      <p class="truncate">
        Betragstyp
      </p>
    </div>
    <div
      class="flex items-center gap-2 cursor-default hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
    >
      <p class="truncate">
        Zyklus
      </p>
    </div>
    <div
      class="flex items-center gap-2 !border-r cursor-default hover:bg-zinc-50 hover:dark:bg-zinc-700 transition-colors duration-300"
    >
      <p class="truncate">
        Zu Kosten von
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
const model = defineModel<boolean>()
</script>
