<template>
  <div class="text-xl text-center w-full font-semibold cursor-default">
    <RouterLink v-if="toNamedRoute" :to="{name: toNamedRoute}">LIQUI<span class="text-liqui-red">SWISS</span></RouterLink>
    <p v-else>LIQUI<span class="text-liqui-red">SWISS</span></p>
  </div>
</template>

<script setup lang="ts">
import {RouteNames} from "~/config/routes.js";

defineProps({
  toNamedRoute: {
    type: String,
    required: false,
  }
})
</script>