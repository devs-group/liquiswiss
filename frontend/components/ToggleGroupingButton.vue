<template>
  <Button
    v-tooltip.top="getTooltip"
    :icon="getDisplayIcon"
    @click="callback"
  />
</template>

<script setup lang="ts">
import type { GroupingType } from '~/utils/types'

defineProps<{
  callback: (payload: MouseEvent) => void
}>()

const model = defineModel<GroupingType>({ default: 'ungrouped' })

const getDisplayIcon = computed(() => model.value == 'ungrouped' ? `pi pi-link` : `pi pi-th-large`)
const getTooltip = computed(() => model.value == 'ungrouped' ? `Mit Gruppierung anzeigen` : `Ohne Gruppierung anzeigen`)
</script>
