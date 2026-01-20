<template>
  <div class="flex flex-col gap-6 w-full max-w-6xl mx-auto">
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
              :class="{ 'p-invalid': errors['name']?.length }"
              type="text"
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
              empty-message="Keine Währungen gefunden"
              :class="{ 'p-invalid': errors['currencyID']?.length }"
              :options="currencies"
              filter
              auto-filter-focus
              empty-filter-message="Keine Resultate gefunden"
              :option-label="getCurrencyLabel"
              option-value="id"
              placeholder="Bitte wählen"
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

          <div class="col-span-full flex justify-end">
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

    <!-- Delete Member Confirmation -->
    <ConfirmDialog />
  </div>
</template>

<script setup lang="ts">
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import type { OrganisationFormData, OrganisationResponse } from '~/models/organisation'
import type { OrganisationMemberResponse } from '~/models/member'
import type { InvitationResponse } from '~/models/invitation'
import { ModalConfig } from '~/config/dialog-props'
import InviteMemberDialog from '~/components/dialogs/InviteMemberDialog.vue'
import { Config } from '~/config/config'

const { user, getOrganisationCurrencyID } = useAuth()
const { updateOrganisation } = useOrganisations()
const { currencies, getCurrencyLabel, showGlobalLoadingSpinner } = useGlobalData()
const { calculateForecast } = useForecasts()
const { members, setMembers, setRefreshMembers, removeMember } = useMembers()
const { invitations, setInvitations, setRefreshInvitations, deleteInvitation } = useInvitations()

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
