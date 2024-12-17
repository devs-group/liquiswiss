<template>
</template>

<script setup lang="ts">
import {RouteNames} from "~/config/routes";
const {organisations} = useOrganisations()
await navigateTo({name: RouteNames.SETTINGS_ORGANISATION_EDIT, params: {id: organisations.value[0].id}, replace: true})
</script>
