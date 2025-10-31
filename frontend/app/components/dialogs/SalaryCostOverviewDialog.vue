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
      class="relative grid grid-cols-1 sm:grid-cols-2 gap-2 col-span-full p-2 bg-zinc-200 dark:bg-zinc-700 rounded-xl"
    >
      <template v-if="costs.length > 0">
        <SalaryCostCard
          v-for="salaryCost of filteredSalaryCosts"
          :key="salaryCost.id"
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
        class="flex flex-wrap col-span-full justify-end gap-2"
      >
        <Button
          v-if="withSeparateCosts"
          icon="pi pi-users"
          severity="help"
          label="Von anderem Mitarbeiter kopieren"
          @click="onCopyFromOtherEmployee"
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
const withSeparateCosts = computed(() => salary.value?.withSeparateCosts ?? false)

const filteredSalaryCosts = computed(() => {
  return costs.value
    .filter(c => !c.label || c.label?.name.toLowerCase().includes(search.value.toLowerCase()))
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

const onCopyFromOtherEmployee = () => {
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
