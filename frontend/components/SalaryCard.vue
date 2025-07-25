<template>
  <Card>
    <template #title>
      <div class="relative flex items-center justify-between">
        <p class="truncate text-base">
          Von {{ fromDateFormatted }}
        </p>
        <div class="flex gap-2 justify-end">
          <Button
            severity="help"
            icon="pi pi-copy"
            outlined
            rounded
            @click="$emit('onClone', salary)"
          />
          <Button
            icon="pi pi-pencil"
            outlined
            rounded
            @click="$emit('onEdit', salary)"
          />
        </div>
        <p
          v-if="isActive"
          class="absolute -top-9 left-0 whitespace-nowrap text-sm bg-liqui-green p-2 rounded-xl font-bold text-center"
        >
          Aktiver Lohn
        </p>
      </div>
    </template>
    <template #content>
      <div class="flex flex-col gap-2 text-sm">
        <p>{{ salary.hoursPerMonth }} Arbeitsstunden / Monat</p>
        <div class="flex flex-col">
          <p v-if="withSeparateCosts">
            {{ totalSalaryCostFormatted }} {{ salary.currency.code }} Gesamtkosten
          </p>
          <p v-else>
            {{ grossSalaryFormatted }} {{ salary.currency.code }} / {{ cycle }}
          </p>
          <div
            v-if="withSeparateCosts"
            class="flex flex-col text-xs"
          >
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

        <div class="flex items-center gap-2">
          <label
            class="text-sm font-bold"
            for="with-separate-costs"
          >Lohnkosten separat erfassen</label>
          <div class="flex items-center">
            <ToggleSwitch
              id="with-separate-costs"
              v-model="withSeparateCosts"
            />
          </div>
        </div>
        <div
          v-if="withSeparateCosts"
          class="flex items-center gap-2"
        >
          <Button
            v-if="hasCosts"
            v-tooltip.top="'Lohnkosten in anderen Lohn kopieren'"
            icon="pi pi-copy"
            severity="help"
            @click="onCopyAllCosts"
          />
          <Button
            icon="pi pi-pencil"
            @click="onShowCostOverview"
          />
        </div>
      </div>
    </template>
  </Card>
</template>

<script setup lang="ts">
import type { SalaryResponse } from '~/models/employee'
import { Config } from '~/config/config'
import { ModalConfig } from '~/config/dialog-props'
import SalaryCostOverviewDialog from '~/components/dialogs/SalaryCostOverviewDialog.vue'
import SalaryCostCopyDialog from '~/components/dialogs/SalaryCostCopyDialog.vue'
import { SalaryUtils } from '~/utils/models/salary-utils'

const toast = useToast()
const dialog = useDialog()

const { updateSalary, listSalaries } = useSalaries()

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

defineEmits<{
  onEdit: [salary: SalaryResponse]
  onClone: [salary: SalaryResponse]
}>()

const withSeparateCosts = ref(props.salary.withSeparateCosts)

watch(withSeparateCosts, (value) => {
  updateSalary(props.salary.employeeID, {
    id: props.salary.id,
    cycle: props.salary.cycle,
    withSeparateCosts: value,
  })
    .then(() => {
      toast.add({
        summary: 'Erfolg',
        detail: `Änderung gespeichert`,
        severity: 'info',
        life: Config.TOAST_LIFE_TIME_SHORT,
      })
    })
})

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
const hasCosts = computed(
  () => SalaryUtils.hasCosts(props.salary),
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
    onClose: (options) => {
      if (options?.data) {
        listSalaries(props.salary.employeeID)
      }
    },
  })
}

const onCopyAllCosts = () => {
  dialog.open(SalaryCostCopyDialog, {
    props: {
      header: `Lohnkosten kopieren`,
      ...ModalConfig,
    },
    data: {
      salary: props.salary,
    },
    onClose: (options) => {
      console.log(options?.data)
      if (options?.data) {
        listSalaries(props.salary.employeeID)
      }
    },
  })
}
</script>
