<template>
  <div class="block sm:hidden flex items-center justify-between text-xl text-center w-full font-semibold cursor-default p-2">
    <Logo
      class="!text-left"
      :to-named-route="RouteNames.HOME"
    />
    <Button
      icon="pi pi-bars"
      severity="secondary"
      @click="toggleDrawer"
    />
  </div>
  <Drawer
    v-model:visible="showDrawer"
    :show-close-icon="true"
    header="Menü"
    position="right"
    :pt="{ content: { class: 'flex flex-col' } }"
  >
    <Navigation
      class="flex-1 min-h-0"
      @click="showDrawer = false"
    />
  </Drawer>
</template>

<script setup lang="ts">
import { RouteNames } from '~/config/routes'

const showDrawer = ref(false)

const toggleDrawer = () => {
  showDrawer.value = !showDrawer.value
}
</script>
