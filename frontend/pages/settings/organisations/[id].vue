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
        <div class="flex flex-col gap-2 col-span-full md:col-span-1">
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

        <div class="flex flex-col gap-2 col-span-full md:col-span-1">
          <div class="flex items-center gap-2">
            <label
              class="text-sm font-bold"
              for="base-currency"
            >Hauptwährung *</label>
            <i
              v-tooltip.top="'Legt die Anzeige für die Prognose und den Umwandlungskurs fest. Währungen von bereits bestehenden Daten werden nicht geändert'"
              class="pi pi-info-circle"
            />
          </div>
          <Select
            v-bind="currencyIDProps"
            id="base-currency"
            v-model="currencyID"
            empty-message="Keine Währungen gefunden"
            :class="{ 'p-invalid': errors['currencyID']?.length }"
            :options="currencies"
            filter
            empty-filter-message="Keine Resultate gefunden"
            :option-label="getCurrencyLabel"
            option-value="id"
            placeholder="Bitte wählen"
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
const { getOrganisationCurrencyID } = useAuth()
const { useFetchGetOrganisation, updateOrganisation } = useOrganisations()
const { currencies, getCurrencyLabel, showGlobalLoadingSpinner } = useGlobalData()
const { calculateForecast } = useForecasts()

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

const { defineField, errors, handleSubmit, meta, resetForm, isFieldDirty } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
    currencyID: yup.number().required('Währung wird benötigt').typeError('Bitte gültige Währung eingeben'),
  }),
  initialValues: {
    id: organisation.value?.id ?? undefined,
    name: organisation.value?.name ?? '',
    currencyID: getOrganisationCurrencyID.value,
  } as OrganisationFormData,
})

const [name, nameProps] = defineField('name')
const [currencyID, currencyIDProps] = defineField('currencyID')

const onSubmit = handleSubmit((values) => {
  if (!organisation.value) {
    return
  }

  const requiresReload = isFieldDirty('currencyID')
  isSubmitting.value = true
  organisationSubmitMessage.value = ''
  organisationSubmitErrorMessage.value = ''
  updateOrganisation(organisation.value.id, values)
    .then(() => {
      resetForm({ values })
      organisationSubmitMessage.value = 'Organisation wurde bearbeitet'
      if (requiresReload) {
        showGlobalLoadingSpinner.value = true
        // Trigger forecast and reload the whole app if the base currency has changed
        calculateForecast()
          .finally(() => {
            reloadNuxtApp({ force: true })
          })
      }
    })
    .catch(() => {
      organisationSubmitErrorMessage.value = 'Organisation konnte nicht bearbeitet werden'
    })
    .finally(() => {
      isSubmitting.value = false
    })
})
</script>
