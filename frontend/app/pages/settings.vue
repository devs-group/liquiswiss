<template>
  <div class="flex flex-col gap-4">
    <Menubar
      :model="items"
      breakpoint="640px"
      class="justify-between sm:justify-start"
    >
      <template #start>
        <div class="sm:hidden flex items-center gap-2 px-2 text-base font-medium">
          <span :class="activeItem?.icon" />
          <span>{{ activeItem?.label ?? '' }}</span>
        </div>
      </template>
      <template #button="{ toggleCallback }">
        <button
          type="button"
          class="p-menubar-button"
          aria-label="Menü"
          @click="toggleCallback"
        >
          <i class="pi pi-chevron-down" />
        </button>
      </template>
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
            <Badge
              v-if="item.badge"
              :value="item.badge"
              severity="warn"
              class="ml-2"
            />
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

const { myPendingInvitations } = useInvitations()

const route = useRoute()

const items = computed<MenuItem[]>(() => [
  { label: 'Profil', icon: 'pi pi-user', routeName: RouteNames.SETTINGS_PROFILE },
  {
    label: 'Organisationen',
    icon: 'pi pi-building',
    routeName: RouteNames.SETTINGS_ORGANISATIONS,
    badge: myPendingInvitations.value.length > 0 ? myPendingInvitations.value.length.toString() : undefined,
  },
  { label: 'Automatisierung', icon: 'pi pi-sync', routeName: RouteNames.SETTINGS_AUTOMATION },
  { label: 'App', icon: 'pi pi-mobile', routeName: RouteNames.SETTINGS_APP },
])

const activeItem = computed(() => items.value.find(i => i.routeName === route.name))

definePageMeta({
  redirect: () => {
    const { settingsTab } = useSettings()
    return { name: settingsTab.value }
  },
})
</script>
