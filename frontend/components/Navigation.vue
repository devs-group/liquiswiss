<template>
  <Menu class="sm:!rounded-none sm:!border-t-0 sm:!border-b-0" :model="items">
    <template #start>
      <Logo class="hidden sm:block p-4"/>
    </template>

    <template #item="{ item, props }">
      <router-link v-if="item.routeName" v-slot="{ href, navigate, isActive }" :to="{name: item.routeName}" custom>
        <a v-ripple :href="href" v-bind="props.action" @click="navigate">
          <span :class="item.icon" />
          <span class="ml-2" :class="{'text-liqui-green': isActive}">{{ item.label }}</span>
        </a>
      </router-link>
      <a v-else v-ripple :href="item.url" :target="item.target" v-bind="props.action">
        <span :class="item.icon" />
        <span class="ml-2">{{ item.label }}</span>
      </a>
    </template>
  </Menu>
</template>

<script setup lang="ts">
import {RouteNames} from "~/config/routes";
import useAuth from "~/composables/useAuth";
import type {MenuItem, MenuItemCommandEvent} from "primevue/menuitem";

const {logout} = useAuth()
const confirm = useConfirm()

const items = ref<MenuItem[]>([
  { label: 'Prognose', icon: 'pi pi-chart-line', routeName: RouteNames.HOME },
  { label: 'Mitarbeitende', icon: 'pi pi-users', routeName: RouteNames.EMPLOYEES },
  { label: 'Transaktionen', icon: 'pi pi-money-bill', routeName: RouteNames.TRANSACTIONS },
  { label: 'Bankkonten', icon: 'pi pi-building', routeName: RouteNames.BANK_ACCOUNTS },
  { label: 'Konto', icon: 'pi pi-user', routeName: RouteNames.ACCOUNT },
  { label: 'Abmelden', icon: 'pi pi-sign-out', command: async (event: MenuItemCommandEvent) => {
      confirm.require({
        header: 'Abmelden',
        message: 'Möchtest du dich wirklich abmelden?',
        icon: 'pi pi-exclamation-triangle',
        rejectLabel: 'Nein',
        acceptLabel: 'Ja',
        accept: async () => {
          await logout();
          reloadNuxtApp({force: true}) // , path: RoutePaths.AUTH
        },
        reject: () => {
        }
      });
    }},
]);
</script>