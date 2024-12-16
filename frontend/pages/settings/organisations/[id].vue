<template>
  <div class="flex flex-col gap-4 w-full">
    <Message severity="error" v-if="organisationError.length">{{organisationError}}</Message>
    <div v-else-if="organisation" class="p-2 bg-zinc-100 dark:bg-zinc-800">
      <form @submit.prevent class="grid grid-cols-1 sm:grid-cols-2 gap-2">
        <div class="flex flex-col gap-2 col-span-full">
          <label class="text-sm font-bold" for="name">Name *</label>
          <InputText v-model="name" v-bind="nameProps"
                     :class="{'p-invalid': errors['name']?.length}"
                     id="name" type="text"/>
          <small class="text-liqui-red">{{errors["name"]}}</small>
        </div>

        <div class="col-span-full">
          <Message v-if="organisationSubmitError.length" severity="error" :life="Config.MESSAGE_LIFE_TIME" :sticky="false" :closable="false">
            {{organisationSubmitError}}
          </Message>
        </div>

        <div class="flex justify-end gap-2 col-span-full">
          <Button @click="onSubmit" label="Speichern" type="submit" :loading="isSubmitting" :disabled="!meta.valid || isSubmitting"/>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import type {OrganisationFormData, OrganisationResponse} from "~/models/organisation";
import {useForm} from "vee-validate";
import * as yup from "yup";
import {Config} from "~/config/config";

const route = useRoute()
const {useFetchGetOrganisation, updateOrganisation} = useOrganisations()

const organisation = ref<OrganisationResponse>()
const organisationError = ref('')
const organisationSubmitError = ref('')
const isSubmitting = ref(false)

await useFetchGetOrganisation(Number.parseInt(route.params.id as string))
    .then(value => {
      if (value) {
        organisation.value = value
      }
    })
    .catch(() => {
      organisationError.value = "Diese Organisation konnte nicht geladen werden"
    })

useHead({
  title: organisation.value?.name ?? '-',
})

const { defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
  }),
  initialValues: {
    id: organisation.value?.id ?? undefined,
    name: organisation.value?.name ?? '',
  } as OrganisationFormData
});

const [name, nameProps] = defineField('name')

const onSubmit = handleSubmit((values) => {
  if (!organisation.value) {
    return
  }

  isSubmitting.value = true
  organisationSubmitError.value = ''
  updateOrganisation(organisation.value.id, values)
      .catch(() => {
        organisationSubmitError.value = 'Konnte Organisation nicht aktualisieren'
      })
      .finally(() => {
        isSubmitting.value = false
      })
})
</script>
