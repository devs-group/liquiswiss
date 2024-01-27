<template>
  <div class="flex flex-col gap-4 p-4">
    <MainMenu/>
    <div class="w-full h-full">
      <NuxtPage/>
    </div>
  </div>
  <DynamicDialog/>
  <Toast position="bottom-right"/>
</template>

<script setup lang="ts">
useHead({
  title: 'LiquiSwiss'
})
</script>