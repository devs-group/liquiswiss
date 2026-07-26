<template>
  <div
    class="grid grid-cols-2 gap-2"
  >
    <Message
      v-if="costsErrorMessage.length"
      severity="error"
      size="small"
      class="col-span-full"
      closable
    >
      {{ costsErrorMessage }}
    </Message>

    <div class="flex items-center gap-2">
      <SearchInput v-model="search" />
      <Select
        v-model="filterType"
        :options="EmployeeCostOverviewTypeToOptions()"
        option-label="name"
        option-value="value"
      />
    </div>

    <div
      class="relative grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-2 col-span-full p-2 bg-zinc-200 dark:bg-zinc-700 rounded-xl"
    >
      <template v-if="costs.length > 0">
        <SalaryCostCard
          v-for="salaryCost of filteredSalaryCosts"
          :key="salaryCost.id"
          :data-realtime-id="`salary_cost:${salaryCost.id}`"
          :salary-cost="salaryCost"
          :salary="salary"
          @on-clone="onCloneCost"
          @on-edit="onEditCost"
          @on-delete="onDeleteCost"
        />
      </template>
      <p
        v-else
        class="text-sm col-span-full"
      >
        Noch keine Lohnkosten vorhanden
      </p>
      <div
        v-if="canEdit"
        class="flex flex-wrap col-span-full justify-end gap-2"
      >
        <Button
          icon="pi pi-users"
          severity="help"
          label="Von Mitarbeiter kopieren"
          @click="onCopyFromEmployee"
        />
        <Button
          icon="pi pi-plus"
          label="Lohnkosten hinzufügen"
          @click="onCreateCost"
        />
      </div>

      <FullProgressSpinner :show="isLoadingCosts" />
    </div>

    <hr class="my-4 col-span-full">

    <div class="flex justify-end gap-2 col-span-full">
      <Button
        :disabled="isLoading"
        label="Schliessen"
        severity="contrast"
        @click="dialogRef.close(requiresRefresh)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ISalaryCostOverviewDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'
import type { SalaryCostResponse } from '~/models/employee'
import SalaryCostCard from '~/components/SalaryCostCard.vue'
import { ModalConfig } from '~/config/dialog-props'
import SalaryCostDialog from '~/components/dialogs/SalaryCostDialog.vue'
import SalaryCostCopyOtherDialog from '~/components/dialogs/SalaryCostCopyOtherDialog.vue'
import { SalaryCostUtils } from '~/utils/models/salary-cost-utils'
import {
  type EmployeeCostOverviewTypeFilterToStringDefinition,
  EmployeeCostOverviewTypeToOptions,
} from '~/utils/enum-helper'
import { EmployeeCostDistributionType, EmployeeCostOverviewType } from '~/config/enums'

const dialogRef = inject<ISalaryCostOverviewDialog>('dialogRef')!

const { deleteSalaryCost, listSalaryCosts } = useSalaryCosts()
const { canEdit } = useOrganisations()
const confirm = useConfirm()
const toast = useToast()
const dialog = useDialog()

// Data
const isLoading = ref(false)
const salary = ref(dialogRef.value.data!.salary)
const isLoadingCosts = ref(false)
const requiresRefresh = ref(false)
const costsErrorMessage = ref('')
const costs = ref<SalaryCostResponse[]>([])
const search = ref('')
const filterType = ref<EmployeeCostOverviewTypeFilterToStringDefinition>(EmployeeCostOverviewType.All)
const filteredSalaryCosts = computed(() => {
  return costs.value
    .filter(c => !c.label || c.label?.name.toLowerCase().includes(search.value.toLowerCase()))
    .filter((c) => {
      switch (filterType.value) {
        case EmployeeCostOverviewType.Employee:
          return [
            EmployeeCostDistributionType.Employee,
            EmployeeCostDistributionType.Both,
          ].includes(c.distributionType)
        case EmployeeCostOverviewType.Employer:
          return [
            EmployeeCostDistributionType.Employer,
            EmployeeCostDistributionType.Both,
          ].includes(c.distributionType)
      }
      return c
    })
})

// Real-time: refresh the cost list when costs of this salary change via
// MCP or another member; highlighting is handled by the SSE plugin
const sseLastChange = useState<{ entity: string, action: string, id?: number, parentId?: number, ts: number } | null>('sse-last-change', () => null)
watch(sseLastChange, (change) => {
  if (!change || !salary.value) return
  if (change.entity !== 'salary_cost') return
  if (change.parentId && change.parentId !== salary.value.id) return
  onListSalaryCosts()
  toast.add({
    severity: 'info',
    summary: 'Aktualisiert',
    detail: 'Die Lohnkosten wurden soeben aktualisiert',
    life: 3000,
  })
})

const onListSalaryCosts = () => {
  if (salary.value) {
    costsErrorMessage.value = ''
    isLoadingCosts.value = true
    listSalaryCosts(salary.value.id)
      .then((resp) => {
        costs.value = resp.data
      })
      .catch(() => {
        costsErrorMessage.value = 'Es gab einen Fehler beim Laden der Lohnkosten'
      })
      .finally(() => {
        isLoadingCosts.value = false
      })
  }
}

onMounted(() => {
  onListSalaryCosts()
})

const onCreateCost = () => {
  dialog.open(SalaryCostDialog, {
    props: {
      header: 'Neue Lohnkosten anlegen',
      ...ModalConfig,
    },
    data: {
      salary: salary.value,
    },
    onClose: () => {
      requiresRefresh.value = true
      onListSalaryCosts()
    },
  })
}

const onCopyFromEmployee = () => {
  dialog.open(SalaryCostCopyOtherDialog, {
    props: {
      header: 'Lohnkosten kopieren',
      ...ModalConfig,
    },
    data: {
      salary: salary.value,
    },
    onClose: (copied?: boolean) => {
      if (copied) {
        requiresRefresh.value = true
        onListSalaryCosts()
      }
    },
  })
}

const onCloneCost = (costToClone: SalaryCostResponse) => {
  dialog.open(SalaryCostDialog, {
    props: {
      header: 'Lohnkosten klonen',
      ...ModalConfig,
    },
    data: {
      salary: salary.value,
      salaryCostToEdit: costToClone,
      isClone: true,
    },
    onClose: () => {
      requiresRefresh.value = true
      onListSalaryCosts()
    },
  })
}

const onEditCost = (costToEdit: SalaryCostResponse) => {
  dialog.open(SalaryCostDialog, {
    props: {
      header: 'Lohnkosten bearbeiten',
      ...ModalConfig,
    },
    data: {
      salary: salary.value,
      salaryCostToEdit: costToEdit,
    },
    onClose: () => {
      requiresRefresh.value = true
      onListSalaryCosts()
    },
  })
}

const onDeleteCost = (costToDelete: SalaryCostResponse) => {
  confirm.require({
    header: 'Löschen',
    message: `Lohnkosten "${SalaryCostUtils.title(costToDelete)}" für Lohn vollständig löschen?`,
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      isLoading.value = true
      deleteSalaryCost(costToDelete.id)
        .then(() => {
          requiresRefresh.value = true
          onListSalaryCosts()
          toast.add({
            summary: 'Erfolg',
            detail: `Lohnkosten "${SalaryCostUtils.title(costToDelete)}" für Lohn wurde gelöscht`,
            severity: 'success',
            life: Config.TOAST_LIFE_TIME,
          })
        })
        .catch(() => {
          toast.add({
            summary: 'Fehler',
            detail: `Lohnkosten "${SalaryCostUtils.title(costToDelete)}" für Lohn konnte nicht gelöscht werden`,
            severity: 'error',
            life: Config.TOAST_LIFE_TIME,
          })
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
