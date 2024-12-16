<template>
  <Message severity="info">Bitte wählen Sie links eine Organisation aus</Message>
</template>

<script setup lang="ts">
import {RouteNames} from "~/config/routes";

const {organisations} = useOrganisations()

await navigateTo({name: RouteNames.ORGANISATION_EDIT, params: {id: organisations.value[0].id}, replace: true})
</script>
