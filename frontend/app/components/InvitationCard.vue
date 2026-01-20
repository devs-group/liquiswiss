<template>
  <Card class="border-dashed border-2">
    <template #title>
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <i class="pi pi-clock text-yellow-500" />
          <p class="truncate text-base">
            {{ invitation.email }}
          </p>
        </div>
        <div class="flex justify-end gap-1">
          <Button
            v-tooltip.top="'Erneut senden'"
            icon="pi pi-refresh"
            outlined
            rounded
            :loading="isResending"
            @click="onResend"
          />
          <Button
            v-tooltip.top="'Widerrufen'"
            icon="pi pi-times"
            outlined
            rounded
            severity="danger"
            @click="$emit('onDelete', invitation)"
          />
        </div>
      </div>
    </template>
    <template #content>
      <div class="flex flex-col gap-2 text-sm">
        <p><strong>Rolle:</strong> {{ getRoleLabel(invitation.role) }}</p>
        <p><strong>Eingeladen von:</strong> {{ invitation.invitedByName }}</p>
        <p>
          <strong>Läuft ab:</strong>
          <span :class="{ 'text-red-500': isExpired }">
            {{ formatDate(invitation.expiresAt) }}
          </span>
        </p>
        <Tag
          v-if="isExpired"
          severity="danger"
          value="Abgelaufen"
        />
        <Tag
          v-else
          severity="info"
          value="Ausstehend"
        />
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type { InvitationResponse } from '~/models/invitation'
import { Config } from '~/config/config'

const props = defineProps({
  invitation: {
    type: Object as PropType<InvitationResponse>,
    required: true,
  },
  organisationId: {
    type: Number,
    required: true,
  },
})

defineEmits<{
  onDelete: [invitation: InvitationResponse]
}>()

const { resendInvitation } = useInvitations()
const toast = useToast()

const isResending = ref(false)

const isExpired = computed(() => {
  return new Date(props.invitation.expiresAt) < new Date()
})

const getRoleLabel = (role: string) => {
  const labels: Record<string, string> = {
    'admin': 'Admin',
    'editor': 'Editor',
    'read-only': 'Nur Lesen',
  }
  return labels[role] ?? role
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('de-CH', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const onResend = async () => {
  isResending.value = true
  await resendInvitation(props.organisationId, props.invitation.id)
    .then(() => {
      toast.add({
        summary: 'Erfolg',
        detail: 'Einladung wurde erneut gesendet',
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
    .finally(() => {
      isResending.value = false
    })
}
</script>
