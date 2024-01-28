<template>
  <Menubar :model="items" breakpoint="768px">
    <template #item="{ item, props, hasSubmenu }">
      <router-link v-if="item.route" v-slot="{ href, navigate }" :to="item.route" custom>
        <a v-ripple :href="href" v-bind="props.action" @click="navigate">
          <span :class="item.icon" />
          <span class="ml-2">{{ item.label }}</span>
        </a>
      </router-link>
      <a v-else v-ripple :href="item.url" :target="item.target" v-bind="props.action">
        <span :class="item.icon" />
        <span class="ml-2">{{ item.label }}</span>
        <span v-if="hasSubmenu" class="pi pi-fw pi-angle-down ml-2" />
      </a>
    </template>
  </Menubar>

</template>

<script setup lang="ts">
import {Routes} from "~/config/routes";

const items = ref([
  { label: 'Übersicht', icon: 'pi pi-home', route: Routes.HOME },
  { label: 'Mitarbeitende', icon: 'pi pi-user', route: Routes.EMPLOYEES },
  { label: 'Einnahmen', icon: 'pi pi-angle-double-up', route: Routes.REVENUE },
  { label: 'Abmelden', icon: 'pi pi-sign-out', command: () => {alert("logout")} },
]);
</script>