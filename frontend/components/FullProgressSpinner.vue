<template>
  <div
      class="flex items-center justify-center bg-white bg-opacity-60 absolute top-0 left-0 right-0 bottom-0 opacity-0 pointer-events-none transition-opacity duration-300"
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
    timeout = setTimeout(() => showInternal.value = value, 1000)
  } else {
    showInternal.value = value
  }
})
</script>