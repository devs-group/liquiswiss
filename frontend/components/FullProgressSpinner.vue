<template>
  <div
      class="flex items-center justify-center bg-white dark:bg-black !bg-opacity-50 backdrop-blur-sm fixed top-0 left-0 right-0 bottom-0 opacity-0 z-50 pointer-events-none transition-opacity duration-300"
      :class="{'opacity-100 pointer-events-auto': showInternal}">
    <ProgressSpinner/>
  </div>
</template>

<script setup lang="ts">
const props = defineProps({
  show: Boolean,
})
const showInternal = ref(false)
let timeout: NodeJS.Timeout | number = 0

watch(() => props.show, value => {
  if (timeout) {
    clearTimeout(timeout)
  }
  if (value) {
    timeout = setTimeout(() => showInternal.value = value, 300)
  } else {
    showInternal.value = value
  }
})
</script>