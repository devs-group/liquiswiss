<template>
  <div class="flex flex-col gap-6 w-full">
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div>
        <h2 class="text-xl font-bold">
          Ihre Organisationen
        </h2>
        <p class="text-sm text-gray-500">
          Wechseln Sie zwischen Ihren Organisationen oder erstellen Sie eine neue
        </p>
      </div>
      <Button
        label="Organisation hinzufügen"
        icon="pi pi-plus"
        @click="onCreateOrganisation"
      />
    </div>

    <div
      v-if="myPendingInvitations.length"
      class="flex flex-col gap-3"
    >
      <Message severity="warn">
        Sie haben {{ myPendingInvitations.length }} ausstehende
        {{ myPendingInvitations.length === 1 ? 'Einladung' : 'Einladungen' }} zu einer Organisation.
      </Message>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card
          v-for="invitation in myPendingInvitations"
          :key="invitation.id"
          class="border-dashed border-2"
          :pt="{ body: { class: 'p-4' }, content: { class: 'p-0' } }"
        >
          <template #content>
            <div class="flex flex-col gap-3">
              <div>
                <p class="font-semibold">
                  {{ invitation.organisationName }}
                </p>
                <p class="text-xs text-gray-500">
                  Eingeladen von {{ invitation.invitedByName }}
                </p>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <Tag
                  :severity="getInvitationRoleSeverity(invitation.role)"
                  :value="getRoleLabel(invitation.role)"
                />
                <Tag
                  severity="info"
                  icon="pi pi-clock"
                  value="Ausstehend"
                />
              </div>
              <NuxtLink
                :to="{ name: RouteNames.AUTH_INVITATION, query: { token: invitation.token } }"
                class="w-full"
              >
                <Button
                  label="Annehmen"
                  icon="pi pi-check"
                  class="w-full"
                  size="small"
                />
              </NuxtLink>
            </div>
          </template>
        </Card>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <Card
        v-for="org in organisations"
        :key="org.id"
        :class="[
          'transition-all',
          org.id === currentOrganisationID
            ? 'ring-2 ring-liqui-green'
            : 'hover:shadow-lg',
        ]"
        :pt="{ root: { class: '!bg-green-900/5 dark:!bg-green-900/10' }, body: { class: 'p-4' }, content: { class: 'p-0' } }"
        data-testid="organisation-card"
      >
        <template #content>
          <div class="flex items-center justify-between gap-4">
            <div class="flex items-center gap-3 min-w-0">
              <div
                class="w-10 h-10 rounded-full flex items-center justify-center text-white font-bold text-lg shrink-0"
                :class="org.id === currentOrganisationID ? 'bg-liqui-green' : 'bg-gray-400'"
              >
                {{ org.name.charAt(0).toUpperCase() }}
              </div>
              <div class="min-w-0">
                <p
                  class="font-semibold truncate"
                  :class="{ 'text-liqui-green': org.id === currentOrganisationID }"
                >
                  {{ org.name }}
                </p>
                <p class="text-xs text-gray-500">
                  {{ getRoleLabel(org.role) }}
                </p>
              </div>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <Tag
                v-if="org.id === currentOrganisationID"
                value="Aktiv"
                severity="success"
              />
              <Button
                v-else
                label="Wechseln"
                size="small"
                severity="contrast"
                outlined
                data-testid="switch-organisation-button"
                @click="onSwitchOrganisation(org)"
              />
            </div>
          </div>
        </template>
      </Card>
    </div>

    <Message
      v-if="!organisations.length"
      severity="info"
    >
      Sie haben noch keine Organisationen. Erstellen Sie Ihre erste Organisation.
    </Message>
  </div>
</template>

<script setup lang="ts">
import type { OrganisationResponse } from '~/models/organisation'
import { RouteNames } from '~/config/routes'
import { ModalConfig } from '~/config/dialog-props'
import OrganisationDialog from '~/components/dialogs/OrganisationDialog.vue'
import { Config } from '~/config/config'

useHead({
  title: 'Organisationen',
})

const dialog = useDialog()
const confirm = useConfirm()
const toast = useToast()
const { settingsTab } = useSettings()
const { organisations } = useOrganisations()
const { user, updateCurrentOrganisation } = useAuth()
const { showGlobalLoadingSpinner } = useGlobalData()
const { skipOrganisationSwitchQuestion } = useSettings()
const { myPendingInvitations, loadMyPendingInvitations } = useInvitations()

const currentOrganisationID = computed(() => user.value?.currentOrganisationID)

const getRoleLabel = (role: string): string => {
  const labels: Record<string, string> = {
    'owner': 'Eigentümer',
    'admin': 'Administrator',
    'editor': 'Bearbeiter',
    'read-only': 'Nur Lesen',
  }
  return labels[role] ?? role
}

const getInvitationRoleSeverity = (role: string) => {
  const severities: Record<string, string> = {
    'admin': 'info',
    'editor': 'success',
    'read-only': 'secondary',
  }
  return severities[role] ?? 'secondary'
}

const onCreateOrganisation = () => {
  dialog.open(OrganisationDialog, {
    props: {
      header: 'Neue Organisation anlegen',
      ...ModalConfig,
    },
  })
}

const onSwitchOrganisation = (org: OrganisationResponse) => {
  if (skipOrganisationSwitchQuestion.value) {
    doSwitchOrganisation(org.id)
  }
  else {
    confirm.require({
      header: 'Organisation wechseln',
      message: `Möchten Sie zur Organisation "${org.name}" wechseln?`,
      icon: 'pi pi-question-circle',
      rejectLabel: 'Nein',
      acceptLabel: 'Ja',
      accept: () => doSwitchOrganisation(org.id),
    })
  }
}

const doSwitchOrganisation = (organisationId: number) => {
  showGlobalLoadingSpinner.value = true
  updateCurrentOrganisation({ organisationId })
    .then(() => {
      reloadNuxtApp({ force: true })
    })
    .catch(() => {
      showGlobalLoadingSpinner.value = false
      toast.add({
        summary: 'Fehler',
        detail: 'Die Organisation konnte nicht gewechselt werden',
        severity: 'error',
        life: Config.TOAST_LIFE_TIME,
      })
    })
}

onMounted(() => {
  settingsTab.value = RouteNames.SETTINGS_ORGANISATIONS
  loadMyPendingInvitations()
})
</script>
