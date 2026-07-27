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

      <!-- Categories Section -->
      <Panel
        :pt="{ root: { class: 'shadow-md' } }"
        data-testid="categories-panel"
      >
        <template #header>
          <div class="flex items-center gap-2">
            <span class="font-bold">Kategorien</span>
            <Tag
              :value="categories.length.toString()"
              severity="secondary"
              rounded
            />
            <i
              v-tooltip.top="'Kategorien für Transaktionen. System-Kategorien sind fix, eigene Kategorien gelten nur für diese Organisation'"
              class="pi pi-info-circle"
            />
          </div>
        </template>
        <template #icons>
          <Button
            v-if="canEdit"
            label="Neue Kategorie"
            icon="pi pi-plus"
            size="small"
            data-testid="create-category-button"
            @click="onOpenCategoryDialog()"
          />
        </template>

        <div class="flex items-center gap-2 mb-4">
          <label
            class="text-sm font-bold"
            for="show-system-categories"
          >System-Kategorien anzeigen:</label>
          <ToggleSwitch
            id="show-system-categories"
            v-model="showSystemCategories"
            class="scale-[0.65] origin-left"
            data-testid="show-system-categories-toggle"
          />
        </div>

        <div
          class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2"
          data-testid="categories-grid"
        >
          <div
            v-for="category in sortedCategories"
            :key="category.id"
            :data-realtime-id="`category:${category.id}`"
            class="flex items-center justify-between gap-2 rounded border border-surface-200 dark:border-surface-700 px-3 py-2"
            data-testid="category-row"
          >
            <div class="flex flex-col gap-1 min-w-0">
              <div class="flex items-center gap-2 min-w-0">
                <span class="truncate">{{ category.name }}</span>
                <Tag
                  v-if="!category.canEdit"
                  value="System"
                  severity="secondary"
                  rounded
                />
              </div>
              <Tag
                v-if="category.inUse"
                v-tooltip.top="'Wird von Transaktionen verwendet'"
                value="In Verwendung"
                severity="info"
                rounded
                class="self-start !text-[10px] !px-2 !py-0.5 leading-none"
              />
            </div>
            <div
              v-if="canEdit && category.canEdit"
              class="flex gap-1 shrink-0"
            >
              <Button
                icon="pi pi-pencil"
                size="small"
                text
                severity="secondary"
                data-testid="edit-category-button"
                @click="onOpenCategoryDialog(category)"
              />
              <Button
                icon="pi pi-trash"
                size="small"
                text
                severity="danger"
                data-testid="delete-category-button"
                @click="onDeleteCategory(category)"
              />
            </div>
          </div>
        </div>
        <p
          v-if="!sortedCategories.length"
          class="text-gray-500"
        >
          Noch keine eigenen Kategorien
        </p>
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
import CategoryDialog from '~/components/dialogs/CategoryDialog.vue'
import CategoryReassignDialog from '~/components/dialogs/CategoryReassignDialog.vue'
import type { CategoryResponse } from '~/models/category'
import { Config } from '~/config/config'
import { DateToApiFormat } from '~/utils/format-helper'

const { user, getOrganisationCurrencyID } = useAuth()
const { updateOrganisation, canEdit } = useOrganisations()
const { currencies, getCurrencyLabel, showGlobalLoadingSpinner } = useGlobalData()
const { calculateForecast } = useForecasts()
const { members, setMembers, setRefreshMembers, removeMember } = useMembers()
const { invitations, setInvitations, setRefreshInvitations, deleteInvitation } = useInvitations()
const { useFetchGetVatSetting, saveVatSetting } = useVatSettings()
const { categories, deleteCategory } = useCategories()

const dialog = useDialog()
const confirm = useConfirm()
const toast = useToast()
const route = useRoute()
const router = useRouter()

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

// Own categories first, system presets after; alphabetical within each group
const showSystemCategories = ref(false)
const sortedCategories = computed(() =>
  categories.value
    .filter(category => showSystemCategories.value || category.canEdit)
    .sort((a, b) => {
      if (a.canEdit !== b.canEdit) return a.canEdit ? -1 : 1
      return a.name.localeCompare(b.name, 'de')
    }),
)

// Reopen dialogs after a page reload (?invite=1, ?category=new|<id>);
// onNuxtReady guarantees the dialog host is mounted
onNuxtReady(() => {
  if (route.query.invite === '1' && canInvite.value) {
    onOpenInviteDialog()
  }
  const categoryQuery = route.query.category
  if (typeof categoryQuery === 'string' && canEdit.value) {
    if (categoryQuery === 'new') {
      onOpenCategoryDialog()
    }
    else {
      const category = categories.value.find(c => c.id === Number(categoryQuery))
      if (category?.canEdit) onOpenCategoryDialog(category)
      else setCategoryQuery(null)
    }
  }
  const reassignQuery = route.query.categoryReassign
  if (typeof reassignQuery === 'string' && canEdit.value) {
    const category = categories.value.find(c => c.id === Number(reassignQuery))
    if (category?.canEdit) onOpenCategoryReassignDialog(category)
    else setCategoryReassignQuery(null)
  }
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
const { onEntityChange } = useRealtimeChanges()
onEntityChange(['organisation', 'member', 'invitation', 'vat_setting'], async (change) => {
  // Own changes are already handled by the submit/dialog flows; refetching
  // here races with their in-flight refresh (Nuxt cancels the older request)
  // and can reset the form to stale data
  if (change.own) return
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

// Keep the open invite dialog in the URL so it survives page reloads
// (?invite=1)
const setInviteQuery = (value: string | null) => {
  const query = { ...route.query }
  if (value === null) delete query.invite
  else query.invite = value
  router.replace({ query })
}

// Keep the open category dialog in the URL (?category=new|<id>)
const setCategoryQuery = (value: string | null) => {
  const query = { ...route.query }
  if (value === null) delete query.category
  else query.category = value
  router.replace({ query })
}

const onOpenCategoryDialog = (category?: CategoryResponse) => {
  setCategoryQuery(category ? String(category.id) : 'new')
  dialog.open(CategoryDialog, {
    props: {
      header: category ? 'Kategorie bearbeiten' : 'Neue Kategorie',
      ...ModalConfig,
      pt: { root: { class: 'dialog-compact' } },
    },
    data: {
      categoryToEdit: category,
    },
    onClose: () => {
      setCategoryQuery(null)
    },
  })
}

const onDeleteCategory = (category: CategoryResponse) => {
  // In-use categories skip the plain confirm: deleting requires reassigning
  // the linked transactions anyway, so go straight to the replace dialog
  if (category.inUse) {
    onOpenCategoryReassignDialog(category)
    return
  }
  confirm.require({
    message: `Möchten Sie die Kategorie "${category.name}" wirklich löschen?`,
    header: 'Kategorie löschen',
    icon: 'pi pi-exclamation-triangle',
    rejectProps: {
      label: 'Abbrechen',
      severity: 'secondary',
    },
    acceptProps: {
      label: 'Löschen',
      severity: 'danger',
    },
    accept: () => {
      deleteCategory(category.id)
        .then(() => {
          toast.add({
            summary: 'Erfolg',
            detail: 'Kategorie wurde gelöscht',
            severity: 'success',
            life: Config.TOAST_LIFE_TIME,
          })
        })
        .catch((err: { status?: number, message: string }) => {
          // Still used by transactions: offer relinking them to another
          // category before deleting
          if (err.status === 409) {
            onOpenCategoryReassignDialog(category)
            return
          }
          toast.add({
            summary: 'Fehler',
            detail: err.message,
            severity: 'error',
            life: Config.TOAST_LIFE_TIME,
          })
        })
    },
  })
}

// Keep the open reassign dialog in the URL (?categoryReassign=<id>)
const setCategoryReassignQuery = (value: string | null) => {
  const query = { ...route.query }
  if (value === null) delete query.categoryReassign
  else query.categoryReassign = value
  router.replace({ query })
}

const onOpenCategoryReassignDialog = (category: CategoryResponse) => {
  setCategoryReassignQuery(String(category.id))
  dialog.open(CategoryReassignDialog, {
    props: {
      header: 'Kategorie ersetzen und löschen',
      ...ModalConfig,
      pt: { root: { class: 'dialog-compact' } },
    },
    data: {
      categoryToDelete: category,
    },
    onClose: () => {
      setCategoryReassignQuery(null)
    },
  })
}

// Member/Invitation handlers
const onOpenInviteDialog = () => {
  setInviteQuery('1')
  dialog.open(InviteMemberDialog, {
    props: {
      header: 'Mitglied einladen',
      ...ModalConfig,
      pt: { root: { class: 'dialog-compact' } },
    },
    data: {
      organisationId: organisation.value?.id,
    },
    onClose: () => {
      setInviteQuery(null)
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
