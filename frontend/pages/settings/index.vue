<template>
</template>

<script setup lang="ts">
const {settingsTab} = useSettings()
await navigateTo({name: settingsTab.value, replace: true})
</script>
