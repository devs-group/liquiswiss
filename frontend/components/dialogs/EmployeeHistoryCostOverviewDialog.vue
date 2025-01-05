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
      <InputText
        v-model="search"
        placeholder="Suchen"
      />
      <Select
        v-model="filterType"
        :options="EmployeeCostOverviewTypeToOptions()"
        option-label="name"
        option-value="value"
      />
    </div>

    <div
      class="grid grid-cols-1 sm:grid-cols-2 gap-2 col-span-full p-2 bg-zinc-200 dark:bg-zinc-700 rounded-xl"
    >
      <template
        v-if="isLoadingCosts"
      >
        <Skeleton class="!h-32" />
        <Skeleton class="!h-32" />
      </template>
      <template v-else>
        <template v-if="costs.length > 0">
          <EmployeeHistoryCostCard
            v-for="cost of filterCosts"
            :key="cost.id"
            :employee-history-cost="cost"
            :employee-history="employeeHistory"
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
          class="flex col-span-full justify-end"
        >
          <Button
            icon="pi pi-plus"
            label="Lohnkosten hinzufügen"
            @click="onCreateCost"
          />
        </div>
      </template>
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
import type { IHistoryCostOverviewDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'
import type { EmployeeHistoryCostResponse } from '~/models/employee'
import EmployeeHistoryCostCard from '~/components/EmployeeHistoryCostCard.vue'
import { ModalConfig } from '~/config/dialog-props'
import EmployeeHistoryCostDialog from '~/components/dialogs/EmployeeHistoryCostDialog.vue'
import { EmployeeHistoryCostUtils } from '~/utils/models/employee-history-cost-utils'
import {
  type EmployeeCostOverviewTypeFilterToStringDefinition,
  EmployeeCostOverviewTypeToOptions,
} from '~/utils/enum-helper'
import { EmployeeCostDistributionType, EmployeeCostOverviewType } from '~/config/enums'

const dialogRef = inject<IHistoryCostOverviewDialog>('dialogRef')!

const { deleteEmployeeHistoryCost, listEmployeeHistoryCosts } = useEmployeeHistoryCosts()
const confirm = useConfirm()
const toast = useToast()
const dialog = useDialog()

// Data
const isLoading = ref(false)
const employeeHistory = ref(dialogRef.value.data!.employeeHistory)
const isLoadingCosts = ref(false)
const requiresRefresh = ref(false)
const costsErrorMessage = ref('')
const costs = ref<EmployeeHistoryCostResponse[]>([])
const search = ref('')
const filterType = ref<EmployeeCostOverviewTypeFilterToStringDefinition>(EmployeeCostOverviewType.All)

const filterCosts = computed(() => {
  return costs.value
    .filter(c => c.label?.name.toLowerCase().includes(search.value.toLowerCase()))
    .filter((c) => {
      switch (filterType.value) {
        case EmployeeCostOverviewType.Employee:
          return c.distributionType == EmployeeCostDistributionType.Employee
        case EmployeeCostOverviewType.Employer:
          return c.distributionType == EmployeeCostDistributionType.Employer
      }
      return c
    })
})

const onListEmployeeHistoryCosts = () => {
  if (employeeHistory.value) {
    costsErrorMessage.value = ''
    isLoadingCosts.value = true
    listEmployeeHistoryCosts(employeeHistory.value.id)
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
  onListEmployeeHistoryCosts()
})

const onCreateCost = () => {
  dialog.open(EmployeeHistoryCostDialog, {
    props: {
      header: 'Neue Lohnkosten anlegen',
      ...ModalConfig,
    },
    data: {
      employeeHistory: employeeHistory.value,
    },
    onClose: (options) => {
      if (options?.data) {
        requiresRefresh.value = true
        onListEmployeeHistoryCosts()
      }
    },
  })
}

const onCloneCost = (costToClone: EmployeeHistoryCostResponse) => {
  dialog.open(EmployeeHistoryCostDialog, {
    props: {
      header: 'Lohnkosten klonen',
      ...ModalConfig,
    },
    data: {
      employeeHistory: employeeHistory.value,
      employeeCostToEdit: costToClone,
      isClone: true,
    },
    onClose: (options) => {
      if (options?.data) {
        requiresRefresh.value = true
        onListEmployeeHistoryCosts()
      }
    },
  })
}

const onEditCost = (costToEdit: EmployeeHistoryCostResponse) => {
  dialog.open(EmployeeHistoryCostDialog, {
    props: {
      header: 'Lohnkosten bearbeiten',
      ...ModalConfig,
    },
    data: {
      employeeHistory: employeeHistory.value,
      employeeCostToEdit: costToEdit,
    },
    onClose: (options) => {
      if (options?.data) {
        requiresRefresh.value = true
        onListEmployeeHistoryCosts()
      }
    },
  })
}

const onDeleteCost = (costToDelete: EmployeeHistoryCostResponse) => {
  confirm.require({
    header: 'Löschen',
    message: `Lohnkosten "${EmployeeHistoryCostUtils.title(costToDelete)}" für Historie vollständig löschen?`,
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      isLoading.value = true
      deleteEmployeeHistoryCost(costToDelete.id)
        .then(() => {
          requiresRefresh.value = true
          onListEmployeeHistoryCosts()
          toast.add({
            summary: 'Erfolg',
            detail: `Lohnkosten "${EmployeeHistoryCostUtils.title(costToDelete)}" für Historie wurde gelöscht`,
            severity: 'success',
            life: Config.TOAST_LIFE_TIME,
          })
        })
        .catch(() => {
          toast.add({
            summary: 'Fehler',
            detail: `Lohnkosten "${EmployeeHistoryCostUtils.title(costToDelete)}" für Historie konnte nicht gelöscht werden`,
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
