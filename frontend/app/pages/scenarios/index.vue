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
        icon="pi pi-calculator"
        @click="onCreateScenario"
      />
    </div>

    <Message
      v-if="scenarioErrorMessage.length"
      severity="error"
      :closable="false"
      class="col-span-full"
    >
      {{ scenarioErrorMessage }}
    </Message>

    <TreeTable
      v-else-if="filteredTreeNodes.length"
      :value="filteredTreeNodes"
      :expanded-keys="expandedKeys"
      class="w-full"
    >
      <Column
        field="name"
        header="Name"
        expander
      >
        <template #body="{ node }">
          <div class="flex items-center gap-2">
            <span
              v-if="node.data.level > 0 && node.children.length === 0"
              class="w-0.5 bg-gray-400 self-stretch"
            />
            <div class="flex items-center gap-2">
              <p>{{ node.data.name }}</p>
              <Badge
                v-if="node.data.isDefault"
                size="small"
                severity="warn"
              >
                Standard
              </Badge>
              <Badge
                v-if="isActiveSzenario(node.data)"
                size="small"
              >
                Aktiv
              </Badge>
            </div>
          </div>
        </template>
      </Column>
      <Column header="Aktionen">
        <template #body="{ node }">
          <span
            class="pi pi-pencil cursor-pointer text-primary"
            @click="onEditScenario(node.data)"
          />
        </template>
      </Column>
    </TreeTable>

    <Message
      v-else
      severity="warn"
      size="small"
    >
      Es gibt noch keine Szenarien
    </Message>
  </div>
</template>

<script setup lang="ts">
import { ModalConfig } from '~/config/dialog-props'
import ScenarioDialog from '~/components/dialogs/ScenarioDialog.vue'
import type { ScenarioResponse, ScenarioTreeNode } from '~/models/scenario'
import useScenarios from '~/composables/useScenarios'

useHead({
  title: 'Szenarien',
})

const dialog = useDialog()
const { scenarioTreeNodes, useFetchListScenarios } = useScenarios()
const { user } = useAuth()

const search = ref('')
const scenarioErrorMessage = ref('')

const filterTreeNodes = (nodes: ScenarioTreeNode[], query: string): ScenarioTreeNode[] => {
  if (!query) return nodes
  const lowerQuery = query.toLowerCase()

  return nodes.reduce<ScenarioTreeNode[]>((acc, node) => {
    const filteredChildren = filterTreeNodes(node.children, query)
    const nameMatches = node.data.name.toLowerCase().includes(lowerQuery)

    if (nameMatches || filteredChildren.length > 0) {
      acc.push({
        ...node,
        children: nameMatches ? node.children : filteredChildren,
      })
    }
    return acc
  }, [])
}

const filteredTreeNodes = computed(() => filterTreeNodes(scenarioTreeNodes.value, search.value))

const collectAllKeys = (nodes: ScenarioTreeNode[]): Record<string, boolean> => {
  const keys: Record<string, boolean> = {}
  for (const node of nodes) {
    keys[node.key] = true
    Object.assign(keys, collectAllKeys(node.children))
  }
  return keys
}

const expandedKeys = computed(() => collectAllKeys(filteredTreeNodes.value))

await useFetchListScenarios()
  .catch((reason) => {
    scenarioErrorMessage.value = reason
  })

const onCreateScenario = () => {
  dialog.open(ScenarioDialog, {
    props: {
      header: 'Neues Szenario anlegen',
      ...ModalConfig,
    },
  })
}

const onEditScenario = (scenario: ScenarioResponse) => {
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

const isActiveSzenario = (scenario: ScenarioResponse) => {
  return scenario.id === user.value?.currentScenarioID
}
</script>
