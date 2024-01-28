<template>
  <Card>
    <template #title>
      <div class="flex items-center justify-between">
        <p class="truncate text-base">{{person.attributes.name}}</p>
        <div class="flex justify-end">
          <Button @click="$emit('onEdit', person)" icon="pi pi-pencil" outlined rounded />
        </div>
      </div>
    </template>
    <template #content>
      <div class="flex flex-col gap-2 text-sm">
        <p>
          {{person.attributes.hoursPerMonth}} max. Stunden pro Monat
        </p>
        <p>
          {{person.attributes.vacationDaysPerYear}} Urlaubstage im Jahr
        </p>
        <p v-if="person.attributes.entry">
          Zugehörig seit {{DateStringToFormattedDate(person.attributes.entry)}}
        </p>
        <p v-if="person.attributes.exit">
          Austritt am {{DateStringToFormattedDate(person.attributes.exit)}}
        </p>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type {StrapiPerson} from "~/models/person";
import {DateStringToFormattedDate} from "~/utils/format-helper";

defineProps({
  person: {
    type: Object as PropType<StrapiPerson>,
    required: true,
  }
})

defineEmits<{
  'onEdit': [person: StrapiPerson]
}>()
</script>
