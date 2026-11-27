<template>
  <!-- Grid rows 0fr to 1fr animates the height without knowing it -->
  <div
    class="col-span-full grid transition-[grid-template-rows] duration-300 ease-in-out"
    :class="open ? 'grid-rows-[1fr]' : 'grid-rows-[0fr]'"
  >
    <!-- clip, not hidden: hidden would become the scroll box and break the sticky labels -->
    <div class="min-h-0 overflow-y-clip">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ open: boolean }>()
</script>
