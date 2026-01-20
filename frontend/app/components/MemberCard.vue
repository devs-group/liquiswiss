<template>
  <Card>
    <template #title>
      <div class="flex items-center justify-between">
        <p class="truncate text-base">
          {{ member.name || member.email }}
        </p>
        <div
          v-if="canManage"
          class="flex justify-end gap-1"
        >
          <Button
            v-if="member.role !== 'owner'"
            v-tooltip.top="'Bearbeiten'"
            icon="pi pi-cog"
            outlined
            rounded
            @click="$emit('onEdit', member)"
          />
          <Button
            v-if="member.role !== 'owner'"
            v-tooltip.top="'Entfernen'"
            icon="pi pi-trash"
            outlined
            rounded
            severity="danger"
            @click="$emit('onDelete', member)"
          />
        </div>
      </div>
    </template>
    <template #content>
      <div class="flex flex-col gap-2 text-sm">
        <p><strong>E-Mail:</strong> {{ member.email }}</p>
        <p><strong>Rolle:</strong> {{ getRoleLabel(member.role) }}</p>
        <Tag
          v-if="member.role === 'owner'"
          severity="warn"
          value="Eigentümer"
        />
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type { OrganisationMemberResponse } from '~/models/member'

defineProps({
  member: {
    type: Object as PropType<OrganisationMemberResponse>,
    required: true,
  },
  canManage: {
    type: Boolean,
    default: false,
  },
})

defineEmits<{
  onEdit: [member: OrganisationMemberResponse]
  onDelete: [member: OrganisationMemberResponse]
}>()

const getRoleLabel = (role: string) => {
  const labels: Record<string, string> = {
    'owner': 'Eigentümer',
    'admin': 'Admin',
    'editor': 'Editor',
    'read-only': 'Nur Lesen',
  }
  return labels[role] ?? role
}
</script>
