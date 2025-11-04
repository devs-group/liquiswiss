<template>
  <div class="flex flex-col gap-4">
    <div class="flex flex-col sm:flex-row gap-2 justify-between items-center">
      <div class="flex items-center gap-2 w-full sm:w-auto">
        <InputText
          v-model="search"
          placeholder="Suchen"
        />
      </div>
      <Button
        class="self-end"
        label="Szenario hinzufügen"
        icon="pi pi-plus"
        @click="onCreateScenario"
      />
    </div>

    <Message
      v-if="scenariosErrorMessage.length"
      severity="error"
      :closable="false"
      class="col-span-full"
    >
      {{ scenariosErrorMessage }}
    </Message>

    <TreeTable
      v-else
      :value="treeScenarios"
      :loading="isLoading"
      :expanded-keys="expandedKeys"
      striped-rows
      table-style="min-width: 50rem"
      empty-message="Es gibt noch keine Szenarien"
    >
      <Column
        field="name"
        header="Name"
        expander
      >
        <template #body="slotProps">
          <div class="flex items-center gap-2">
            <span>{{ slotProps.node.data.name }}</span>
            <Tag
              v-if="slotProps.node.data.id === user?.currentScenarioID"
              value="Aktiv"
              severity="success"
              rounded
            />
          </div>
        </template>
      </Column>
      <Column
        field="createdAt"
        header="Erstellt am"
      >
        <template #body="slotProps">
          {{ formatDate(slotProps.node.data.createdAt) }}
        </template>
      </Column>
      <Column header="Aktionen">
        <template #body="slotProps">
          <div class="flex gap-2">
            <Button
              icon="pi pi-pencil"
              text
              rounded
              severity="info"
              @click="onEditScenario(slotProps.node.data)"
            />
            <Button
              v-if="!slotProps.node.data.isDefault"
              icon="pi pi-trash"
              text
              rounded
              severity="danger"
              @click="onDeleteScenarioConfirm(slotProps.node.data)"
            />
          </div>
        </template>
      </Column>
    </TreeTable>
  </div>
</template>

<script setup lang="ts">
import { ModalConfig } from '~/config/dialog-props'
import ScenarioDialog from '~/components/dialogs/ScenarioDialog.vue'
import type { ScenarioListItem } from '~/models/scenario'
import { Config } from '~/config/config'

useHead({
  title: 'Szenarien',
})

const dialog = useDialog()
const confirm = useConfirm()
const toast = useToast()
const { scenarios, useFetchListScenarios, deleteScenario } = useScenarios()
const { user } = useAuth()

const isLoading = ref(false)
const search = ref('')
const scenariosErrorMessage = ref('')

const buildTree = (parentId: number | null = null): any[] => {
  const children = scenarios.value
    .filter(s => s.parentScenarioId === parentId)
    .map(scenario => ({
      key: scenario.id.toString(),
      data: scenario,
      children: buildTree(scenario.id),
    }))

  // If search is active, only return nodes that match or have matching children
  if (search.value) {
    return children.filter(node =>
      node.data.name.toLowerCase().includes(search.value.toLowerCase()) ||
      node.children.length > 0
    )
  }

  return children
}

const treeScenarios = computed(() => buildTree(null))

// Expand all nodes by default
const expandedKeys = computed(() => {
  const keys: Record<string, boolean> = {}
  const addKeys = (nodes: any[]) => {
    nodes.forEach(node => {
      keys[node.key] = true
      if (node.children && node.children.length > 0) {
        addKeys(node.children)
      }
    })
  }
  addKeys(treeScenarios.value)
  return keys
})

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleDateString('de-DE', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

await useFetchListScenarios()
  .catch((reason) => {
    scenariosErrorMessage.value = reason
  })

const onCreateScenario = () => {
  dialog.open(ScenarioDialog, {
    props: {
      header: 'Neues Szenario anlegen',
      ...ModalConfig,
    },
  })
}

const onEditScenario = (scenario: ScenarioListItem) => {
  dialog.open(ScenarioDialog, {
    data: {
      scenario: scenario,
    },
    props: {
      header: 'Szenario bearbeiten',
      ...ModalConfig,
    },
  })
}

const onDeleteScenarioConfirm = (scenario: ScenarioListItem) => {
  confirm.require({
    header: 'Löschen',
    message: `Szenario "${scenario.name}" vollständig löschen?`,
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      isLoading.value = true
      deleteScenario(scenario.id)
        .then(() => {
          toast.add({
            summary: 'Erfolg',
            detail: `Szenario "${scenario.name}" wurde gelöscht`,
            severity: 'success',
            life: Config.TOAST_LIFE_TIME,
          })
        })
        .catch(() => {
          scenariosErrorMessage.value = 'Szenario konnte nicht gelöscht werden'
        })
        .finally(() => {
          isLoading.value = false
        })
    },
    reject: () => {
    },
  })
}
</script>
