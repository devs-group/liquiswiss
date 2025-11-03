<template>
  <Card :class="{ 'opacity-60': isDisabled }">
    <template #title>
      <div class="relative flex items-center justify-between">
        <p class="truncate text-base">
          Von {{ fromDateFormatted }}
        </p>
        <div class="flex gap-2 items-center justify-end">
          <Button
            v-if="!isTermination"
            v-tooltip.top="'Lohnkosten kopieren'"
            severity="help"
            icon="pi pi-copy"
            outlined
            rounded
            @click="$emit('onClone', salary)"
          />
          <Button
            v-if="!isTermination"
            icon="pi pi-pencil"
            rounded
            @click="$emit('onEdit', salary)"
          />
          <Button
            v-if="isTermination"
            icon="pi pi-trash"
            severity="danger"
            outlined
            rounded
            @click="onDeleteSalary"
          />
        </div>
        <p
          v-if="isActive && !isTermination"
          class="absolute -top-9 left-0 whitespace-nowrap text-sm bg-liqui-green p-2 rounded-xl font-bold text-center"
        >
          Aktiver Lohn
        </p>
        <p
          v-else-if="!isTermination && isDisabled"
          class="absolute -top-9 left-0 whitespace-nowrap text-sm bg-zinc-400 p-2 rounded-xl font-bold text-center text-zinc-900"
        >
          Deaktiviert
        </p>
      </div>
    </template>
    <template #content>
      <div
        v-if="!isTermination"
        class="flex flex-col gap-2 text-sm"
      >
        <Message
          v-if="isDisabled"
          severity="warn"
          size="small"
          :closable="false"
        >
          Dieser Lohn ist deaktiviert und wird nicht berechnet.
        </Message>
        <p>{{ salary.hoursPerMonth }} Arbeitsstunden / Monat</p>
        <div class="flex flex-col">
          <p>
            {{ totalSalaryCostFormatted }} {{ salary.currency.code }} Gesamtkosten
          </p>
          <div class="flex flex-col text-xs">
            <p>Brutto: {{ grossSalaryFormatted }} {{ salary.currency.code }} / {{ cycle }}</p>
            <p>Netto: {{ netSalaryFormatted }} {{ salary.currency.code }} / {{ cycle }}</p>
          </div>
        </div>
        <p>{{ salary.vacationDaysPerYear }} Urlaubstage im Jahr</p>
        <p
          v-if="salary.toDate"
          class="text-orange-500"
        >
          Bis {{ toDateFormatted }}
        </p>
        <p
          v-else
          class="text-orange-500"
        >
          Dauerhaft
        </p>

        <div class="flex items-center justify-end">
          <Button
            size="small"
            label="Lohnkosten"
            icon="pi pi-pencil"
            :disabled="isDisabled"
            @click="onShowCostOverview"
          />
        </div>
        <div class="flex items-center gap-2">
          <p>Status:</p>
          <ToggleSwitch
            id="salary-card-active"
            class="scale-[0.65] origin-left"
            :model-value="!isDisabled"
            :disabled="isUpdatingDisabled"
            @update:model-value="onToggleDisabled"
          />
        </div>
      </div>
      <div
        v-else
        class="flex flex-col gap-2 text-sm"
      >
        <Message
          severity="warn"
          size="small"
        >
          <p
            v-if="salary.toDate"
          >
            Austritt bis {{ toDateFormatted }}
          </p>
          <p v-else>
            Dauerhafter Austritt
          </p>
        </Message>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type { SalaryResponse } from '~/models/employee'
import { Config } from '~/config/config'
import { ModalConfig } from '~/config/dialog-props'
import SalaryCostOverviewDialog from '~/components/dialogs/SalaryCostOverviewDialog.vue'
import { SalaryUtils } from '~/utils/models/salary-utils'

const toast = useToast()
const dialog = useDialog()
const confirm = useConfirm()

const isDeletingSalary = ref(false)

const { updateSalary, listSalaries, deleteSalary } = useSalaries()

const props = defineProps({
  salary: {
    type: Object as PropType<SalaryResponse>,
    required: true,
  },
  isActive: {
    type: Boolean,
    required: true,
  },
})

const emits = defineEmits<{
  onEdit: [salary: SalaryResponse]
  onClone: [salary: SalaryResponse]
  onDeleted: []
}>()

const isTermination = ref(props.salary.isTermination)
const isDisabled = ref(props.salary.isDisabled)
const isUpdatingDisabled = ref(false)

watch(() => props.salary.isDisabled, (value) => {
  isDisabled.value = value
})

const onToggleDisabled = (isActive: boolean) => {
  const previous = isDisabled.value
  const nextDisabled = !isActive
  if (previous === nextDisabled) {
    return
  }
  isDisabled.value = nextDisabled
  isUpdatingDisabled.value = true
  updateSalary(props.salary.employeeID, {
    id: props.salary.id,
    cycle: props.salary.cycle,
    isDisabled: nextDisabled,
  })
    .then(() => {
      toast.add({
        summary: 'Erfolg',
        detail: nextDisabled ? 'Lohn deaktiviert' : 'Lohn aktiviert',
        severity: 'info',
        life: Config.TOAST_LIFE_TIME_SHORT,
      })
    })
    .catch(() => {
      isDisabled.value = previous
      toast.add({
        summary: 'Fehler',
        detail: 'Status konnte nicht geändert werden',
        severity: 'error',
        life: Config.TOAST_LIFE_TIME_SHORT,
      })
    })
    .finally(() => {
      isUpdatingDisabled.value = false
    })
}

const grossSalaryFormatted = computed(
  () => SalaryUtils.grossSalaryFormatted(props.salary),
)
const netSalaryFormatted = computed(
  () => SalaryUtils.netSalaryFormatted(props.salary),
)
const totalSalaryCostFormatted = computed(
  () => SalaryUtils.totalSalaryCostFormatted(props.salary),
)
const fromDateFormatted = computed(
  () => SalaryUtils.fromDateFormatted(props.salary),
)
const toDateFormatted = computed(
  () => SalaryUtils.toDateFormatted(props.salary),
)
const cycle = computed(
  () => SalaryUtils.cycle(props.salary),
)
const onShowCostOverview = () => {
  dialog.open(SalaryCostOverviewDialog, {
    props: {
      header: `Lohnkostenübersicht`,
      ...ModalConfig,
    },
    data: {
      salary: props.salary,
    },
    onClose: () => {
      listSalaries(props.salary.employeeID)
    },
  })
}

const onDeleteSalary = () => {
  confirm.require({
    header: 'Löschen',
    message: `Austritt vom "${fromDateFormatted.value}" vollständig löschen?`,
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      if (props.salary) {
        isDeletingSalary.value = false
        deleteSalary(props.salary.employeeID, props.salary.id)
          .then(() => {
            toast.add({
              summary: 'Erfolg',
              detail: `Austritt wurde gelöscht`,
              severity: 'success',
              life: Config.TOAST_LIFE_TIME,
            })
            emits('onDeleted')
            listSalaries(props.salary.employeeID)
          })
          .catch(() => {
            toast.add({
              summary: 'Fehler',
              detail: `Austritt konnte nicht gelöscht werden`,
              severity: 'error',
              life: Config.TOAST_LIFE_TIME_SHORT,
            })
            nextTick(() => {
              scrollToParentBottom('salary-form')
            })
          })
          .finally(() => {
            isDeletingSalary.value = false
          })
      }
    },
    reject: () => {
    },
  })
}
</script>
