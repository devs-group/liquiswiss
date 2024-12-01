<template>
  <div class="block sm:hidden flex items-center justify-between text-xl text-center w-full font-semibold cursor-default p-2">
    <Logo class="!text-left" :to-named-route="RouteNames.HOME"/>
    <Button @click="toggleDrawer" icon="pi pi-bars" severity="secondary"/>
  </div>
  <Drawer v-model:visible="showDrawer" :show-close-icon="true" header="Menü" position="right">
    <Navigation @click="showDrawer = false"/>
  </Drawer>
</template>

<script setup lang="ts">
import {RouteNames} from "~/config/routes";

const showDrawer = ref(false)

const toggleDrawer = () => {
  showDrawer.value = !showDrawer.value
}
</script>