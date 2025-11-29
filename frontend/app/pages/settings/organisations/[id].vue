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
            auto-filter-focus
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

    <!-- VAT Settings Section -->
    <div
      v-if="organisation"
      class="p-2 bg-zinc-100 dark:bg-zinc-800"
    >
      <div class="flex justify-between items-center gap-2 mb-4">
        <p class="text-lg font-bold">
          Automatische MwSt.-Abrechnung
        </p>
        <i
          v-tooltip.top="'Summiert die MwSt. von Umsätzen und erstellt automatisch Ausgaben für die MwSt.-Abrechnung'"
          class="pi pi-info-circle"
        />
      </div>
      <form
        class="grid grid-cols-1 sm:grid-cols-2 gap-4"
        @submit.prevent
      >
        <div class="flex flex-col gap-2 col-span-full">
          <div class="flex items-center gap-2">
            <Checkbox
              v-bind="vatEnabledProps"
              id="vat-enabled"
              v-model="vatEnabled"
              binary
            />
            <label
              class="cursor-pointer font-bold"
              for="vat-enabled"
            >MwSt.-Abrechnung aktivieren</label>
          </div>
        </div>

        <div
          v-if="vatEnabled"
          class="flex flex-col gap-2 col-span-full md:col-span-1"
        >
          <div class="flex items-center gap-2">
            <label
              class="text-sm font-bold"
              for="vat-billing-date"
            >Rechnungszeitpunkt *</label>
            <i
              v-tooltip.top="'Datum der MwSt.-Abrechnung (wird zur Berechnung der Periode verwendet)'"
              class="pi pi-info-circle"
            />
          </div>
          <DatePicker
            v-bind="vatBillingDateProps"
            id="vat-billing-date"
            v-model="vatBillingDate"
            :class="{ 'p-invalid': vatErrors['vatBillingDate']?.length }"
            date-format="dd.mm.yy"
            show-button-bar
            show-icon
          />
          <small class="text-liqui-red">{{ vatErrors["vatBillingDate"] }}</small>
        </div>

        <div
          v-if="vatEnabled"
          class="flex flex-col gap-2 col-span-full md:col-span-1"
        >
          <div class="flex items-center gap-2">
            <label
              class="text-sm font-bold"
              for="vat-transaction-date"
            >Transaktionszeitpunkt *</label>
            <i
              v-tooltip.top="'Datum der tatsächlichen Zahlung (erscheint in der Prognose)'"
              class="pi pi-info-circle"
            />
          </div>
          <DatePicker
            v-bind="vatTransactionDateProps"
            id="vat-transaction-date"
            v-model="vatTransactionDate"
            :class="{ 'p-invalid': vatErrors['vatTransactionDate']?.length }"
            date-format="dd.mm.yy"
            show-button-bar
            show-icon
          />
          <small class="text-liqui-red">{{ vatErrors["vatTransactionDate"] }}</small>
        </div>

        <div
          v-if="vatEnabled"
          class="flex flex-col gap-2 col-span-full md:col-span-1"
        >
          <label
            class="text-sm font-bold"
            for="vat-interval"
          >Abrechnungsintervall *</label>
          <Select
            v-bind="vatIntervalProps"
            id="vat-interval"
            v-model="vatInterval"
            :class="{ 'p-invalid': vatErrors['vatInterval']?.length }"
            :options="intervalOptions"
            option-label="label"
            option-value="value"
            placeholder="Bitte wählen"
          />
          <small class="text-liqui-red">{{ vatErrors["vatInterval"] }}</small>
        </div>

        <div class="col-span-full">
          <Message
            v-if="vatSubmitMessage.length"
            severity="success"
            :life="Config.MESSAGE_LIFE_TIME"
            :sticky="false"
            :closable="false"
          >
            {{ vatSubmitMessage }}
          </Message>
          <Message
            v-if="vatSubmitErrorMessage.length"
            severity="error"
            :life="Config.MESSAGE_LIFE_TIME"
            :sticky="false"
            :closable="false"
          >
            {{ vatSubmitErrorMessage }}
          </Message>
        </div>

        <div class="flex justify-end gap-2 col-span-full">
          <Button
            label="MwSt.-Einstellungen speichern"
            type="submit"
            :loading="isVatSubmitting"
            :disabled="!vatMeta.valid || (vatMeta.valid && !vatMeta.dirty) || isVatSubmitting"
            @click="onVatSubmit"
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
import type { VatSettingFormData } from '~/models/vat-setting'
import { Config } from '~/config/config'

const route = useRoute()
const { getOrganisationCurrencyID } = useAuth()
const { useFetchGetOrganisation, updateOrganisation } = useOrganisations()
const { currencies, getCurrencyLabel, showGlobalLoadingSpinner } = useGlobalData()
const { calculateForecast } = useForecasts()
const { getVatSetting, saveVatSetting } = useVatSettings()

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

// VAT Settings Form
const vatSubmitMessage = ref('')
const vatSubmitErrorMessage = ref('')
const isVatSubmitting = ref(false)

// Fetch VAT settings
const existingVatSetting = await getVatSetting()

const intervalOptions = [
  { label: 'Monatlich', value: 'monthly' },
  { label: 'Vierteljährlich', value: 'quarterly' },
  { label: 'Halbjährlich', value: 'biannually' },
  { label: 'Jährlich', value: 'yearly' },
]

const { defineField: defineVatField, errors: vatErrors, handleSubmit: handleVatSubmit, meta: vatMeta, resetForm: resetVatForm } = useForm({
  validationSchema: yup.object({
    vatEnabled: yup.boolean().required(),
    vatBillingDate: yup.date().when('vatEnabled', {
      is: true,
      then: schema => schema.required('Rechnungszeitpunkt wird benötigt'),
      otherwise: schema => schema.nullable(),
    }),
    vatTransactionDate: yup.date().when('vatEnabled', {
      is: true,
      then: schema => schema.required('Transaktionszeitpunkt wird benötigt'),
      otherwise: schema => schema.nullable(),
    }),
    vatInterval: yup.string().when('vatEnabled', {
      is: true,
      then: schema => schema.required('Interval wird benötigt').oneOf(['monthly', 'quarterly', 'biannually', 'yearly']),
      otherwise: schema => schema.nullable(),
    }),
  }),
  initialValues: {
    vatEnabled: existingVatSetting?.enabled ?? false,
    vatBillingDate: existingVatSetting?.billingDate ? new Date(existingVatSetting.billingDate) : new Date(),
    vatTransactionDate: existingVatSetting?.transactionDate ? new Date(existingVatSetting.transactionDate) : new Date(),
    vatInterval: existingVatSetting?.interval ?? 'quarterly',
  } as { vatEnabled: boolean, vatBillingDate: Date, vatTransactionDate: Date, vatInterval: string },
})

const [vatEnabled, vatEnabledProps] = defineVatField('vatEnabled')
const [vatBillingDate, vatBillingDateProps] = defineVatField('vatBillingDate')
const [vatTransactionDate, vatTransactionDateProps] = defineVatField('vatTransactionDate')
const [vatInterval, vatIntervalProps] = defineVatField('vatInterval')

const onVatSubmit = handleVatSubmit((values) => {
  isVatSubmitting.value = true
  vatSubmitMessage.value = ''
  vatSubmitErrorMessage.value = ''

  // Format the dates as YYYY-MM-DD
  const formattedBillingDate = values.vatBillingDate instanceof Date
    ? values.vatBillingDate.toISOString().split('T')[0]
    : values.vatBillingDate

  const formattedTransactionDate = values.vatTransactionDate instanceof Date
    ? values.vatTransactionDate.toISOString().split('T')[0]
    : values.vatTransactionDate

  const payload: VatSettingFormData = {
    enabled: values.vatEnabled,
    billingDate: formattedBillingDate,
    transactionDate: formattedTransactionDate,
    interval: values.vatInterval as 'monthly' | 'quarterly' | 'biannually' | 'yearly',
  }

  saveVatSetting(payload)
    .then(() => {
      resetVatForm({ values })
      vatSubmitMessage.value = 'MwSt.-Einstellungen wurden gespeichert'
      // Trigger forecast recalculation
      showGlobalLoadingSpinner.value = true
      calculateForecast()
        .finally(() => {
          showGlobalLoadingSpinner.value = false
        })
    })
    .catch(() => {
      vatSubmitErrorMessage.value = 'MwSt.-Einstellungen konnten nicht gespeichert werden'
    })
    .finally(() => {
      isVatSubmitting.value = false
    })
})
</script>
