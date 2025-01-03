<template>
  <Button
    :icon="getDisplayIcon"
    @click="callback"
  />
</template>

<script setup lang="ts">
import type { DisplayType } from '~/utils/types'

defineProps<{
  callback: (payload: MouseEvent) => void
}>()

const model = defineModel<DisplayType>({ default: 'list' })

const getDisplayIcon = computed(() => model.value == 'list' ? `pi pi-microsoft` : `pi pi-list`)
</script>
