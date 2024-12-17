<template>
  <div class="text-xl text-center w-full font-manrope font-semibold cursor-default">
    <NuxtLink v-if="toNamedRoute" :to="{name: toNamedRoute}">LIQUI<span class="text-liqui-red">SWISS</span></NuxtLink>
    <p v-else>LIQUI<span class="text-liqui-red">SWISS</span></p>
  </div>
</template>

<script setup lang="ts">
defineProps({
  toNamedRoute: {
    type: String,
    required: false,
  }
})
</script>