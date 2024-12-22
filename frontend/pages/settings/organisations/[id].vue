<template>
  <div class="flex flex-col gap-4 w-full">
    <Message
      v-if="organisationError.length"
      severity="error"
    >
      {{ organisationError }}
    </Message>
    <div
      v-else-if="organisation"
      class="p-2 bg-zinc-100 dark:bg-zinc-800"
    >
      <form
        class="grid grid-cols-1 sm:grid-cols-2 gap-2"
        @submit.prevent
      >
        <div class="flex flex-col gap-2 col-span-full">
          <label
            class="text-sm font-bold"
            for="name"
          >Name *</label>
          <InputText
            v-bind="nameProps"
            id="name"
            v-model="name"
            :class="{ 'p-invalid': errors['name']?.length }"
            type="text"
          />
          <small class="text-liqui-red">{{ errors["name"] }}</small>
        </div>

        <div class="col-span-full">
          <Message
            v-if="organisationSubmitMessage.length"
            severity="success"
            :life="Config.MESSAGE_LIFE_TIME"
            :sticky="false"
            :closable="false"
          >
            {{ organisationSubmitMessage }}
          </Message>
          <Message
            v-if="organisationSubmitErrorMessage.length"
            severity="error"
            :life="Config.MESSAGE_LIFE_TIME"
            :sticky="false"
            :closable="false"
          >
            {{ organisationSubmitErrorMessage }}
          </Message>
        </div>

        <div class="flex justify-end gap-2 col-span-full">
          <Button
            label="Speichern"
            type="submit"
            :loading="isSubmitting"
            :disabled="!meta.valid || (meta.valid && !meta.dirty) || isSubmitting"
            @click="onSubmit"
          />
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import type { OrganisationFormData, OrganisationResponse } from '~/models/organisation'
import { Config } from '~/config/config'

const route = useRoute()
const { useFetchGetOrganisation, updateOrganisation } = useOrganisations()

const organisation = ref<OrganisationResponse>()
const organisationError = ref('')
const organisationSubmitMessage = ref('')
const organisationSubmitErrorMessage = ref('')
const isSubmitting = ref(false)

await useFetchGetOrganisation(Number.parseInt(route.params.id as string))
  .then((value) => {
    if (value) {
      organisation.value = value
    }
  })
  .catch(() => {
    organisationError.value = 'Diese Organisation konnte nicht geladen werden'
  })

useHead({
  title: organisation.value?.name ?? '-',
})

const { defineField, errors, handleSubmit, meta, resetForm } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
  }),
  initialValues: {
    id: organisation.value?.id ?? undefined,
    name: organisation.value?.name ?? '',
  } as OrganisationFormData,
})

const [name, nameProps] = defineField('name')

const onSubmit = handleSubmit((values) => {
  if (!organisation.value) {
    return
  }

  isSubmitting.value = true
  organisationSubmitMessage.value = ''
  organisationSubmitErrorMessage.value = ''
  updateOrganisation(organisation.value.id, values)
    .then(() => {
      resetForm({ values })
      organisationSubmitMessage.value = 'Organisation wurde bearbeitet'
    })
    .catch(() => {
      organisationSubmitErrorMessage.value = 'Organisation konnte nicht bearbeitet werden'
    })
    .finally(() => {
      isSubmitting.value = false
    })
})
</script>
