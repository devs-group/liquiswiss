<template>
  <div class="block sm:hidden flex items-center justify-between text-xl text-center w-full font-semibold cursor-default p-2">
    <p>LIQUI<span class="text-red-400">SWISS</span></p>
    <Button @click="toggleDrawer" icon="pi pi-bars" severity="secondary"/>
  </div>
  <Drawer v-model:visible="showDrawer" :show-close-icon="false">
    <Navigation @click="showDrawer = false"/>
  </Drawer>
</template>

<script setup lang="ts">
const showDrawer = ref(false)

const toggleDrawer = () => {
  showDrawer.value = !showDrawer.value
}
</script>