<template>
  <div class="flex flex-col gap-4">
    <Menubar
      :model="items"
      class="justify-end sm:justify-start"
    >
      <template #item="{ item, props }">
        <router-link
          v-if="item.routeName"
          v-slot="{ href, navigate, isActive }"
          :to="{ name: item.routeName }"
          custom
        >
          <a
            v-ripple
            :href="href"
            v-bind="props.action"
            @click="navigate"
          >
            <span :class="item.icon" />
            <span
              class="ml-2"
              :class="{ 'text-liqui-green': isActive }"
            >{{ item.label }}</span>
          </a>
        </router-link>
        <a
          v-else
          v-ripple
          :href="item.url"
          :target="item.target"
          v-bind="props.action"
        >
          <span :class="item.icon" />
          <span class="ml-2">{{ item.label }}</span>
        </a>
      </template>
    </Menubar>
    <NuxtPage />
  </div>
</template>

<script setup lang="ts">
import type { MenuItem } from 'primevue/menuitem'
import { RouteNames } from '~/config/routes'

const items = ref<MenuItem[]>([
  { label: 'Profil', icon: 'pi pi-user', routeName: RouteNames.SETTINGS_PROFILE },
  { label: 'Organisationen', icon: 'pi pi-building', routeName: RouteNames.SETTINGS_ORGANISATIONS },
  { label: 'App', icon: 'pi pi-mobile', routeName: RouteNames.SETTINGS_APP },
])

definePageMeta({
  redirect: () => {
    const { settingsTab } = useSettings()
    return { name: settingsTab.value }
  },
})
</script>
