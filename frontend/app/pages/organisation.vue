<template>
  <div
    class="flex flex-col gap-6 w-full max-w-6xl mx-auto"
  >
    <Message
      v-if="organisationError.length"
      severity="error"
    >
      {{ organisationError }}
    </Message>

    <template v-else-if="organisation">
      <!-- Page Header -->
      <div class="flex items-center gap-3">
        <i class="pi pi-building text-2xl text-liqui-green" />
        <h1 class="text-2xl font-bold">
          {{ organisation.name }}
        </h1>
      </div>

      <!-- General Settings Section -->
      <Panel
        header="Allgemein"
        :pt="{ root: { class: 'shadow-md' } }"
        data-testid="general-settings-panel"
      >
        <template #icons>
          <i class="pi pi-cog text-lg" />
        </template>
        <form
          class="grid grid-cols-1 md:grid-cols-2 gap-4"
          @submit.prevent
        >
          <div class="flex flex-col gap-2">
            <label
              class="text-sm font-bold"
              for="name"
            >Name *</label>
            <InputText
              v-bind="nameProps"
              id="name"
              v-model="name"
              data-realtime-field="org-name"
              :class="[{ 'p-invalid': errors['name']?.length }, !canEditOrganisation ? '!opacity-100 !cursor-not-allowed' : '']"
              type="text"
              :disabled="!canEditOrganisation"
              data-testid="organisation-name-input"
            />
            <small class="text-liqui-red">{{ errors["name"] }}</small>
          </div>

          <div class="flex flex-col gap-2">
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
              data-realtime-field="org-currency"
              empty-message="Keine Währungen gefunden"
              :class="[{ 'p-invalid': errors['currencyID']?.length }, !canEditOrganisation ? '!opacity-100 !cursor-not-allowed [&_*]:!pointer-events-auto [&_*]:!cursor-not-allowed' : '']"
              :options="currencies"
              filter
              auto-filter-focus
              empty-filter-message="Keine Resultate gefunden"
              :option-label="getCurrencyLabel"
              option-value="id"
              placeholder="Bitte wählen"
              :disabled="!canEditOrganisation"
              data-testid="organisation-currency-select"
            />
            <small class="text-liqui-red">{{ errors["currencyID"] }}</small>
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

          <div
            v-if="canEditOrganisation"
            class="col-span-full flex justify-end"
          >
            <Button
              label="Speichern"
              type="submit"
              :loading="isSubmitting"
              :disabled="!meta.valid || (meta.valid && !meta.dirty) || isSubmitting"
              data-testid="organisation-save-button"
              @click="onSubmit"
            />
          </div>
        </form>
      </Panel>

      <!-- VAT Automation Section -->
      <Panel
        :pt="{ root: { class: 'shadow-md' } }"
        data-testid="vat-automation-panel"
      >
        <template #header>
          <div class="flex items-center gap-2">
            <span class="font-bold">Automatische MwSt.-Abrechnung</span>
            <i
              v-tooltip.top="'Summiert die MwSt. von Umsätzen und erstellt automatisch Ausgaben für die MwSt.-Abrechnung'"
              class="pi pi-info-circle"
            />
          </div>
        </template>
        <template #icons>
          <i class="pi pi-sync text-lg" />
        </template>
        <form
          class="grid grid-cols-1 sm:grid-cols-2 gap-4"
          @submit.prevent
        >
          <div class="flex flex-col gap-2 col-span-full">
            <div class="flex items-center gap-2">
              <p class="font-bold">
                MwSt.-Abrechnung aktivieren:
              </p>
              <ToggleSwitch
                v-bind="vatEnabledProps"
                id="vat-enabled"
                data-realtime-field="vat-enabled"
                class="scale-[0.65] origin-left"
                :class="!canEdit ? '!opacity-100 !cursor-not-allowed [&_*]:!pointer-events-auto [&_*]:!cursor-not-allowed' : ''"
                :model-value="vatEnabled"
                :disabled="!canEdit"
                @update:model-value="vatEnabled = $event"
              />
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
              data-realtime-field="vat-billing-date"
              :class="[{ 'p-invalid': vatErrors['vatBillingDate']?.length }, !canEdit ? '!opacity-100 !cursor-not-allowed [&_*]:!pointer-events-auto [&_*]:!cursor-not-allowed' : '']"
              date-format="dd.mm.yy"
              show-button-bar
              show-icon
              :disabled="!canEdit"
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
                for="vat-transaction-month-offset"
              >Transaktionszeitpunkt *</label>
              <i
                v-tooltip.top="'Zeitpunkt der tatsächlichen Zahlung relativ zum Rechnungszeitpunkt. Dies definiert die Anzeige in der Prognose'"
                class="pi pi-info-circle"
              />
            </div>
            <Select
              v-bind="vatTransactionMonthOffsetProps"
              id="vat-transaction-month-offset"
              v-model="vatTransactionMonthOffset"
              data-realtime-field="vat-offset"
              :class="[{ 'p-invalid': vatErrors['vatTransactionMonthOffset']?.length }, !canEdit ? '!opacity-100 !cursor-not-allowed [&_*]:!pointer-events-auto [&_*]:!cursor-not-allowed' : '']"
              :options="transactionMonthOffsetOptions"
              option-label="label"
              option-value="value"
              placeholder="Bitte wählen"
              :disabled="!canEdit"
            />
            <small class="text-liqui-red">{{ vatErrors["vatTransactionMonthOffset"] }}</small>
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
              data-realtime-field="vat-interval"
              :class="[{ 'p-invalid': vatErrors['vatInterval']?.length }, !canEdit ? '!opacity-100 !cursor-not-allowed [&_*]:!pointer-events-auto [&_*]:!cursor-not-allowed' : '']"
              :options="intervalOptions"
              option-label="label"
              option-value="value"
              placeholder="Bitte wählen"
              :disabled="!canEdit"
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

          <div
            v-if="canEdit"
            class="flex justify-end gap-2 col-span-full"
          >
            <Button
              label="MwSt.-Einstellungen speichern"
              type="submit"
              :loading="isVatSubmitting"
              :disabled="!vatMeta.valid || (vatMeta.valid && !vatMeta.dirty) || isVatSubmitting"
              @click="onVatSubmit"
            />
          </div>
        </form>
      </Panel>

      <!-- Members Section -->
      <Panel
        :pt="{ root: { class: 'shadow-md' } }"
        data-testid="members-panel"
      >
        <template #header>
          <div class="flex items-center gap-2">
            <span class="font-bold">Mitglieder</span>
            <Tag
              :value="members.length.toString()"
              severity="secondary"
              rounded
            />
          </div>
        </template>
        <template #icons>
          <Button
            v-if="canInvite"
            label="Einladen"
            icon="pi pi-user-plus"
            size="small"
            data-testid="invite-member-button"
            @click="onOpenInviteDialog"
          />
        </template>

        <!-- Pending Invitations -->
        <div
          v-if="invitations.length && canInvite"
          class="mb-6"
        >
          <div class="flex items-center gap-2 mb-3">
            <h3 class="text-md font-semibold text-gray-500">
              Ausstehende Einladungen
            </h3>
            <Tag
              :value="invitations.length.toString()"
              severity="warn"
              rounded
              data-testid="invitations-count"
            />
          </div>
          <div
            class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
            data-testid="invitations-grid"
          >
            <InvitationCard
              v-for="invitation in invitations"
              :key="invitation.id"
              :data-realtime-id="`invitation:${invitation.id}`"
              :invitation="invitation"
              :organisation-id="organisation.id"
              data-testid="invitation-card"
              @on-delete="onDeleteInvitation"
            />
          </div>
          <Divider />
        </div>

        <!-- Members List -->
        <div
          v-if="members.length"
          class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
          data-testid="members-grid"
        >
          <MemberCard
            v-for="member in members"
            :key="member.userId"
            :data-realtime-id="`member:${member.userId}`"
            :member="member"
            :can-manage="canManageMembers"
            data-testid="member-card"
            @on-edit="onEditMember"
            @on-delete="onDeleteMember"
          />
        </div>
        <p
          v-else
          class="text-gray-500"
        >
          Noch keine Mitglieder
        </p>
      </Panel>
    </template>
  </div>
</template>

<script setup lang="ts">
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import type { OrganisationFormData, OrganisationResponse } from '~/models/organisation'
import type { OrganisationMemberResponse } from '~/models/member'
import type { InvitationResponse } from '~/models/invitation'
import type { VatSettingFormData } from '~/models/vat-setting'
import { ModalConfig } from '~/config/dialog-props'
import InviteMemberDialog from '~/components/dialogs/InviteMemberDialog.vue'
import { Config } from '~/config/config'
import { DateToApiFormat } from '~/utils/format-helper'

const { user, getOrganisationCurrencyID } = useAuth()
const { updateOrganisation, canEdit } = useOrganisations()
const { currencies, getCurrencyLabel, showGlobalLoadingSpinner } = useGlobalData()
const { calculateForecast } = useForecasts()
const { members, setMembers, setRefreshMembers, removeMember } = useMembers()
const { invitations, setInvitations, setRefreshInvitations, deleteInvitation } = useInvitations()
const { useFetchGetVatSetting, saveVatSetting } = useVatSettings()

const dialog = useDialog()
const confirm = useConfirm()
const toast = useToast()

const organisation = ref<OrganisationResponse>()
const organisationError = ref('')
const organisationSubmitMessage = ref('')
const organisationSubmitErrorMessage = ref('')
const isSubmitting = ref(false)

const organisationId = computed(() => user.value?.currentOrganisationID)

// Fetch organisation data
const { data: orgData, error: orgError, refresh: refreshOrganisation } = await useFetch<OrganisationResponse>(
  () => `/api/organisations/${organisationId.value}`,
  { method: 'GET' },
)

if (orgError.value) {
  organisationError.value = 'Diese Organisation konnte nicht geladen werden'
}
else {
  organisation.value = orgData.value ?? undefined
}

// Get current user's role
const currentUserRole = computed(() => organisation.value?.role ?? '')
const canInvite = computed(() => ['owner', 'admin'].includes(currentUserRole.value))
const canManageMembers = computed(() => currentUserRole.value === 'owner')
const canEditOrganisation = computed(() => ['owner', 'admin'].includes(currentUserRole.value))

// Fetch members and invitations
const { data: membersData, error: membersError, refresh: refreshMembersData } = await useFetch<OrganisationMemberResponse[]>(
  () => `/api/organisations/${organisationId.value}/members`,
  { method: 'GET' },
)

const { data: invitationsData, refresh: refreshInvitationsData } = await useFetch<InvitationResponse[]>(
  () => `/api/organisations/${organisationId.value}/invitations`,
  {
    method: 'GET',
    immediate: canInvite.value,
  },
)

if (membersError.value) {
  organisationError.value = 'Fehler beim Laden der Mitglieder'
}

// Sync data to composable state
watchEffect(() => {
  setMembers(membersData.value ?? [])
})

watchEffect(() => {
  setInvitations(invitationsData.value ?? [])
})

// Register refresh callbacks
onMounted(() => {
  setRefreshMembers(async () => {
    await refreshMembersData()
  })
  setRefreshInvitations(async () => {
    await refreshInvitationsData()
  })
})

useHead({
  title: organisation.value?.name ?? 'Organisation',
})

// Form setup
const { defineField, errors, handleSubmit, meta, resetForm, isFieldDirty } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
    currencyID: yup.number().required('Währung wird benötigt').typeError('Bitte gültige Währung eingeben'),
  }),
  initialValues: {
    id: organisation.value?.id,
    name: organisation.value?.name ?? '',
    currencyID: getOrganisationCurrencyID.value,
  } as OrganisationFormData,
})

const [name, nameProps] = defineField('name')
const [currencyID, currencyIDProps] = defineField('currencyID')

// VAT Settings Form
const vatSubmitMessage = ref('')
const vatSubmitErrorMessage = ref('')
const isVatSubmitting = ref(false)

const existingVatSetting = await useFetchGetVatSetting()

const intervalOptions = [
  { label: 'Monatlich', value: 'monthly' },
  { label: 'Vierteljährlich', value: 'quarterly' },
  { label: 'Halbjährlich', value: 'biannually' },
  { label: 'Jährlich', value: 'yearly' },
]

const transactionMonthOffsetOptions = [
  { label: 'Gleich wie Rechnungszeitpunkt', value: 0 },
  { label: '1 Monat später', value: 1 },
  { label: '2 Monate später', value: 2 },
  { label: '3 Monate später', value: 3 },
  { label: '4 Monate später', value: 4 },
  { label: '5 Monate später', value: 5 },
  { label: '6 Monate später', value: 6 },
  { label: '7 Monate später', value: 7 },
  { label: '8 Monate später', value: 8 },
  { label: '9 Monate später', value: 9 },
  { label: '10 Monate später', value: 10 },
  { label: '11 Monate später', value: 11 },
  { label: '12 Monate später', value: 12 },
]

const { defineField: defineVatField, errors: vatErrors, handleSubmit: handleVatSubmit, meta: vatMeta, resetForm: resetVatForm } = useForm({
  validationSchema: yup.object({
    vatEnabled: yup.boolean().required(),
    vatBillingDate: yup.date().when('vatEnabled', {
      is: true,
      then: schema => schema.required('Rechnungszeitpunkt wird benötigt'),
      otherwise: schema => schema.nullable(),
    }),
    vatTransactionMonthOffset: yup.number().when('vatEnabled', {
      is: true,
      then: schema => schema.required('Transaktionszeitpunkt wird benötigt').min(0).max(12),
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
    vatBillingDate: existingVatSetting?.billingDate ? new Date(existingVatSetting.billingDate) : null,
    vatTransactionMonthOffset: existingVatSetting?.transactionMonthOffset ?? 0,
    vatInterval: existingVatSetting?.interval ?? 'quarterly',
  } as { vatEnabled: boolean, vatBillingDate: Date | null, vatTransactionMonthOffset: number, vatInterval: string },
})

const [vatEnabled, vatEnabledProps] = defineVatField('vatEnabled')
const [vatBillingDate, vatBillingDateProps] = defineVatField('vatBillingDate')
const [vatTransactionMonthOffset, vatTransactionMonthOffsetProps] = defineVatField('vatTransactionMonthOffset')
const [vatInterval, vatIntervalProps] = defineVatField('vatInterval')

// Real-time: refresh local page data when it changes via MCP or another
// member; highlighting is handled by the SSE plugin
const { getVatSetting } = useVatSettings()
const sseLastChange = useState<{ entity: string, action: string, id?: number, ts: number } | null>('sse-last-change', () => null)
watch(sseLastChange, async (change) => {
  if (!change) return
  switch (change.entity) {
    case 'organisation': {
      const previousName = organisation.value?.name
      const previousCurrencyID = organisation.value?.currency?.id
      await refreshOrganisation()
      organisation.value = orgData.value ?? undefined
      resetForm({
        values: {
          id: organisation.value?.id,
          name: organisation.value?.name ?? '',
          currencyID: organisation.value?.currency?.id ?? getOrganisationCurrencyID.value,
        } as OrganisationFormData,
      })
      await nextTick()
      // Only the inputs whose value actually changed blink
      if (organisation.value?.name !== previousName) {
        flashRealtimeSelector('[data-realtime-field="org-name"]')
      }
      if (organisation.value?.currency?.id !== previousCurrencyID) {
        flashRealtimeSelector('[data-realtime-field="org-currency"]')
      }
      break
    }
    case 'member':
      await refreshMembersData()
      break
    case 'invitation':
      if (canInvite.value) await refreshInvitationsData()
      break
    case 'vat_setting': {
      const previous = {
        enabled: vatEnabled.value,
        billingDate: vatBillingDate.value ? DateToApiFormat(vatBillingDate.value as Date) : null,
        offset: vatTransactionMonthOffset.value,
        interval: vatInterval.value,
      }
      const setting = await getVatSetting()
      resetVatForm({
        values: {
          vatEnabled: setting?.enabled ?? false,
          vatBillingDate: setting?.billingDate ? new Date(setting.billingDate) : null,
          vatTransactionMonthOffset: setting?.transactionMonthOffset ?? 0,
          vatInterval: setting?.interval ?? 'quarterly',
        },
      })
      await nextTick()
      if ((setting?.enabled ?? false) !== previous.enabled) {
        flashRealtimeSelector('[data-realtime-field="vat-enabled"]')
      }
      const newBillingDate = setting?.billingDate ? DateToApiFormat(new Date(setting.billingDate)) : null
      if (newBillingDate !== previous.billingDate) {
        // Target the inner input: the DatePicker root spans the full column
        // while the visible input is narrower
        flashRealtimeSelector('[data-realtime-field="vat-billing-date"] input')
      }
      if ((setting?.transactionMonthOffset ?? 0) !== previous.offset) {
        flashRealtimeSelector('[data-realtime-field="vat-offset"]')
      }
      if ((setting?.interval ?? 'quarterly') !== previous.interval) {
        flashRealtimeSelector('[data-realtime-field="vat-interval"]')
      }
      break
    }
  }
})

const onVatSubmit = handleVatSubmit((values) => {
  if (!values.vatBillingDate) {
    return
  }

  isVatSubmitting.value = true
  vatSubmitMessage.value = ''
  vatSubmitErrorMessage.value = ''

  values.vatBillingDate.setMinutes(values.vatBillingDate.getMinutes() - values.vatBillingDate.getTimezoneOffset())

  const payload: VatSettingFormData = {
    enabled: values.vatEnabled,
    billingDate: DateToApiFormat(values.vatBillingDate),
    transactionMonthOffset: values.vatTransactionMonthOffset,
    interval: values.vatInterval as 'monthly' | 'quarterly' | 'biannually' | 'yearly',
  }

  saveVatSetting(payload)
    .then(() => {
      resetVatForm({ values })
      vatSubmitMessage.value = 'MwSt.-Einstellungen wurden gespeichert'
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

const onSubmit = handleSubmit((values) => {
  if (!organisation.value) {
    return
  }

  const requiresReload = isFieldDirty('currencyID')
  isSubmitting.value = true
  organisationSubmitMessage.value = ''
  organisationSubmitErrorMessage.value = ''
  updateOrganisation(organisation.value.id, values)
    .then(async () => {
      resetForm({ values })
      organisationSubmitMessage.value = 'Organisation wurde bearbeitet'
      await refreshOrganisation()
      if (orgData.value) {
        organisation.value = orgData.value
      }
      if (requiresReload) {
        showGlobalLoadingSpinner.value = true
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

// Member/Invitation handlers
const onOpenInviteDialog = () => {
  dialog.open(InviteMemberDialog, {
    props: {
      header: 'Mitglied einladen',
      ...ModalConfig,
    },
    data: {
      organisationId: organisation.value?.id,
    },
  })
}

const onEditMember = (_member: OrganisationMemberResponse) => {
  toast.add({
    summary: 'Info',
    detail: 'Bearbeitungsfunktion wird in einer späteren Version hinzugefügt',
    severity: 'info',
    life: Config.TOAST_LIFE_TIME,
  })
}

const onDeleteMember = (member: OrganisationMemberResponse) => {
  confirm.require({
    message: `Möchten Sie "${member.name || member.email}" wirklich aus der Organisation entfernen?`,
    header: 'Mitglied entfernen',
    icon: 'pi pi-exclamation-triangle',
    rejectProps: {
      label: 'Abbrechen',
      severity: 'secondary',
    },
    acceptProps: {
      label: 'Entfernen',
      severity: 'danger',
    },
    accept: () => {
      removeMember(organisation.value!.id, member.userId)
        .then(() => {
          toast.add({
            summary: 'Erfolg',
            detail: 'Mitglied wurde entfernt',
            severity: 'success',
            life: Config.TOAST_LIFE_TIME,
          })
        })
        .catch((err) => {
          toast.add({
            summary: 'Fehler',
            detail: err,
            severity: 'error',
            life: Config.TOAST_LIFE_TIME,
          })
        })
    },
  })
}

const onDeleteInvitation = (invitation: InvitationResponse) => {
  confirm.require({
    message: `Möchten Sie die Einladung an "${invitation.email}" wirklich widerrufen?`,
    header: 'Einladung widerrufen',
    icon: 'pi pi-exclamation-triangle',
    rejectProps: {
      label: 'Abbrechen',
      severity: 'secondary',
    },
    acceptProps: {
      label: 'Widerrufen',
      severity: 'danger',
    },
    accept: () => {
      deleteInvitation(organisation.value!.id, invitation.id)
        .then(() => {
          toast.add({
            summary: 'Erfolg',
            detail: 'Einladung wurde widerrufen',
            severity: 'success',
            life: Config.TOAST_LIFE_TIME,
          })
        })
        .catch((err) => {
          toast.add({
            summary: 'Fehler',
            detail: err,
            severity: 'error',
            life: Config.TOAST_LIFE_TIME,
          })
        })
    },
  })
}
</script>
