<template>
  <Menubar :model="items" breakpoint="768px">
    <template #item="{ item, props, hasSubmenu }">
      <RouterLink v-if="item.name" v-slot="{ href, navigate, isActive }" :to="{name: item.name}" custom>
        <a v-ripple :href="href" v-bind="props.action" @click="navigate" :class="{'text-green-600': isActive}">
          <span :class="item.icon" />
          <span class="ml-2">{{ item.label }}</span>
        </a>
      </RouterLink>
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
import useAuth from "~/composables/useAuth";
import type {MenuItemCommandEvent} from "primevue/menuitem";

const {logout} = useAuth()
const confirm = useConfirm()

const items = ref([
  { label: 'Prognose', icon: 'pi pi-chart-line', name: Routes.HOME },
  { label: 'Mitarbeitende', icon: 'pi pi-user', name: Routes.EMPLOYEES },
  { label: 'Transaktionen', icon: 'pi pi-money-bill', name: Routes.TRANSACTIONS },
  { label: 'Bankkonten', icon: 'pi pi-building', name: Routes.BANK_ACCOUNTS },
  { label: 'Abmelden', icon: 'pi pi-sign-out', command: async (event: MenuItemCommandEvent) => {
      confirm.require({
        header: 'Abmelden',
        message: 'Möchtest du dich wirklich abmelden?',
        icon: 'pi pi-exclamation-triangle',
        rejectLabel: 'Nein',
        acceptLabel: 'Ja',
        accept: async () => {
          await logout();
          reloadNuxtApp({force: true})
        },
        reject: () => {
        }
      });
    }},
]);
</script>