<template>
  <Card>
    <template #header>
      <div class="relative">
        <img class="w-full h-48 object-cover object-center" src="https://100k-faces.glitch.me/random-image" :alt="person.name"/>
        <div class="absolute top-2 right-2 z-10 flex justify-end">
          <Button @click="$emit('onEdit', person)" icon="pi pi-pencil" />
        </div>
      </div>
    </template>
    <template #title>
      <p class="truncate text-base">{{person.name}}</p>
    </template>
    <template #content>
      <div class="flex flex-col gap-2 text-sm">
        <p>
          {{person.hoursPerMonth}} Stunden pro Monat
        </p>
        <p>
          {{person.vacationDaysPerYear}} Urlaubstage im Jahr
        </p>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type {Person} from "~/models/person";

defineProps({
  person: {
    type: Object as PropType<Person>,
    required: true,
  }
})

defineEmits<{
  'onEdit': [person: Person]
}>()
</script>
