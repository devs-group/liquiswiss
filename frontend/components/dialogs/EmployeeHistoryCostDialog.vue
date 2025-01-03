<template>
  <form
    class="grid grid-cols-2 gap-2"
    @submit.prevent
  >
    <Message
      severity="secondary"
      class="col-span-full"
    >
      Ausführung der Lohnzahlung am
      {{ employeeHistory.nextExecutionDate ? DateStringToFormattedDate(employeeHistory.nextExecutionDate) : 'Keine weitere' }}
      über
      {{ NumberToFormattedCurrency(AmountToFloat(employeeHistory.salary), employeeHistory.currency.code) }} {{ employeeHistory.currency.code }}
    </Message>

    <div class="col-span-full flex flex-col gap-2">
      <div class="flex items-center gap-2">
        <label
          class="text-sm font-bold"
          for="label-id"
        >Label</label>
        <i
          v-tooltip.top="'Beispiel: BVG, AHV, Quellensteuer, ...'"
          class="pi pi-info-circle"
        />
      </div>
      <Select
        v-bind="labelIDProps"
        id="label-id"
        v-model="labelID"
        empty-message="Keine Labels gefunden"
        show-clear
        :options="employeeHistoryCostsLabels.data"
        option-label="name"
        option-value="id"
        placeholder="Bitte wählen"
        :loading="isLoadingCostLabels"
        :disabled="isLoadingCostLabels"
        :class="{ 'p-invalid': errors['labelID']?.length }"
        type="text"
      >
        <template #option="slotProps">
          <div class="flex items-center w-full justify-between">
            <p>{{ slotProps.option.name }}</p>
            <div
              class="flex gap-2 justify-end"
            >
              <Button
                size="small"
                icon="pi pi-pencil"
                outlined
                rounded
                @click.stop="onEditEmployeeCostLabel(slotProps.option)"
              />
              <Button
                size="small"
                severity="danger"
                icon="pi pi-trash"
                outlined
                rounded
                @click.stop="onDeleteEmployeeCostLabel(slotProps.option)"
              />
            </div>
          </div>
        </template>

        <template #footer>
          <div class="p-1 pt-0">
            <Button
              label="Hinzufügen"
              fluid
              severity="secondary"
              text
              size="small"
              icon="pi pi-plus"
              @click="onCreateEmployeeCostLabel"
            />
          </div>
        </template>
      </Select>
      <div class="flex justify-end gap-2">
        <small class="text-liqui-red">{{ errors["labelID"] }}</small>
        <small class="">Wir empfehlen das Nutzen von Labels für eine bessere Kategorisierung der Kosten</small>
      </div>
    </div>

    <div class="col-span-full md:col-span-1 flex flex-col gap-2">
      <label
        class="text-sm font-bold"
        for="cycle"
      >Frequenz *</label>
      <Select
        v-bind="cycleProps"
        id="cycle"
        v-model="cycle"
        empty-message="Keine Zyklen gefunden"
        :options="CostCycleTypeToOptions()"
        option-label="name"
        option-value="value"
        placeholder="Bitte wählen"
        :class="{ 'p-invalid': errors['cycle']?.length }"
        type="text"
      />
      <div class="flex justify-between gap-2">
        <small class="text-liqui-red">{{ errors["cycle"] }}</small>
      </div>
    </div>

    <div class="col-span-full md:col-span-1 flex flex-col gap-2">
      <div class="flex items-center gap-2">
        <label
          class="text-sm font-bold"
          for="relative-offset"
        >Anzahl der Perioden *</label>
        <i
          v-tooltip.top="'Definiert, wie viele Intervalle der gewählten Frequenz in die Berechnung einbezogen werden'"
          class="pi pi-info-circle"
        />
      </div>
      <InputText
        v-bind="relativeOffsetProps"
        id="relative-offset"
        v-model.number="relativeOffset"
        :class="{ 'p-invalid': errors['relativeOffset']?.length }"
        type="number"
        min="1"
        :disabled="isLoading || isOnce"
      />
      <div class="flex justify-between gap-2">
        <small class="text-liqui-red">{{ errors["relativeOffset"] }}</small>
      </div>
    </div>

    <div class="col-span-full md:col-span-1 flex flex-col gap-2">
      <label
        class="text-sm font-bold"
        for="amount-type"
      >Betragstyp *</label>
      <Select
        v-bind="amountTypeProps"
        id="amount-type"
        v-model="amountType"
        empty-message="Keine Betragstypen gefunden"
        :options="EmployeeCostTypeToOptions()"
        option-label="name"
        option-value="value"
        placeholder="Bitte wählen"
        :class="{ 'p-invalid': errors['amountType']?.length }"
        type="text"
      />
      <div class="flex justify-between gap-2">
        <small class="text-liqui-red">{{ errors["amountType"] }}</small>
      </div>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <label
          class="text-sm font-bold"
          for="amount"
        >Betrag *</label>
        <p
          v-if="!isFixedAmount"
          class="text-indigo-600 text-xs"
        >
          Nur Prozentwerte (von 0 bis 100)
        </p>
      </div>
      <InputText
        v-bind="amountProps"
        id="amount"
        v-model="amount"
        :class="{ 'p-invalid': errors['amount']?.length }"
        type="text"
        :disabled="isLoading"
        @input="onParseAmount"
      />
      <small class="text-liqui-red">{{ errors["amount"] }}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label
          class="text-sm font-bold"
          for="distribution-type"
        >Zu Kosten von *</label>
        <i
          v-tooltip.top="'Arbeitnehmer: Abzug vom Bruttolohn, Arbeitgeber: Zusatz zum Bruttolohn'"
          class="pi pi-info-circle"
        />
      </div>
      <Select
        v-bind="distributionTypeProps"
        id="distribution-type"
        v-model="distributionType"
        empty-message="Keine Option gefunden"
        :options="EmployeeCostDistributionTypeToOptions()"
        option-label="name"
        option-value="value"
        placeholder="Bitte wählen"
        :class="{ 'p-invalid': errors['distributionType']?.length }"
      />
      <small class="text-liqui-red">{{ errors["distributionType"] }}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label
          class="text-sm font-bold"
          for="target-date"
        >Ab spezifischem Zeitpunkt {{ isOnce ? '*' : '' }}</label>
        <i
          v-tooltip.top="(isOnce ? 'Pflichtfeld,' : 'Optional,') + ' ab welchem spezifischen Zeitpunkt die Lohnkosten entstehen'"
          class="pi pi-info-circle"
          :class="{ 'text-red-600': isOnce }"
        />
      </div>
      <DatePicker
        v-model="targetDate"
        v-bind="targetDateProps"
        date-format="dd.mm.yy"
        show-icon
        show-button-bar
        :class="{ 'p-invalid': errors['targetDate']?.length }"
        :disabled="isLoading"
      />
      <small class="text-liqui-red">{{ errors["targetDate"] }}</small>
    </div>

    <Message
      v-if="errorMessage.length"
      severity="error"
      :closable="false"
      class="col-span-full"
    >
      {{ errorMessage }}
    </Message>

    <div class="flex justify-end gap-2 col-span-full">
      <Button
        :disabled="!meta.valid || isLoading"
        :loading="isLoading"
        label="Speichern"
        icon="pi pi-save"
        type="submit"
        @click="onSubmit"
      />
      <Button
        :loading="isLoading"
        label="Abbrechen"
        severity="secondary"
        @click="dialogRef?.close(requiresRefresh)"
      />
    </div>
  </form>
</template>

<script setup lang="ts">
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import type { IEmployeeHistoryCostFormDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'
import type { EmployeeHistoryCostFormData, EmployeeHistoryCostLabelResponse } from '~/models/employee'
import { CycleType, EmployeeCostDistributionType, EmployeeCostType } from '~/config/enums'
import {
  CostCycleTypeToOptions,
  EmployeeCostDistributionTypeToOptions,
  EmployeeCostTypeToOptions,
} from '~/utils/enum-helper'
import { ModalConfig } from '~/config/dialog-props'
import EmployeeHistoryCostLabelDialog from '~/components/dialogs/EmployeeHistoryCostLabelDialog.vue'
import { EmployeeHistoryCostUtils } from '~/utils/models/employee-history-cost-utils'

const dialogRef = inject<IEmployeeHistoryCostFormDialog>('dialogRef')!

const { createEmployeeHistoryCost, updateEmployeeHistoryCost } = useEmployeeHistoryCosts()
const { employeeHistoryCostsLabels, listEmployeeHistoryCostsLabels, deleteEmployeeHistoryCostLabel } = useEmployeeHistoryCostLabels()
const toast = useToast()
const dialog = useDialog()
const confirm = useConfirm()

const employeeHistory = dialogRef.value.data?.employeeHistory
const historyCost = dialogRef.value.data?.employeeCostToEdit
const isClone = dialogRef.value.data?.isClone
const isCreate = isClone || !historyCost?.id
const isLoading = ref(false)
const isLoadingCostLabels = ref(true)
const errorMessage = ref('')
const employeesCostLabelsErrorMessage = ref('')
const requiresRefresh = ref(false)

listEmployeeHistoryCostsLabels(false)
  .catch(() => {
    employeesCostLabelsErrorMessage.value = 'Lohnkosten Labels konnten nicht geladen werden'
  })
  .finally(() => {
    isLoadingCostLabels.value = false
  })

const { defineField, errors, handleSubmit, meta, setFieldValue } = useForm<EmployeeHistoryCostFormData>({
  validationSchema: yup.object({
    labelID: yup.number().nullable().typeError('Ungültiger Wert'),
    cycle: yup.string().required('Zyklus wird benötigt').typeError('Ungültiger Wert'),
    relativeOffset: yup.number().required('Versatz wird benötigt').min(1, 'Relativer Versatz muss mind. 1 sein').typeError('Ungültiger Wert'),
    amountType: yup.string().required('Betragstyp wird benötigt').typeError('Ungültiger Wert'),
    amount: yup.number()
      .test('amountCorrect', 'Betrag ungültig', (value, context) => {
        if (value === undefined) {
          return false
        }
        if (context.parent.amountType === EmployeeCostType.Percentage) {
          return value >= 0 && value <= 100
        }
        return value >= 0
      })
      .min(0, 'Muss mindestens 0 sein').typeError('Ungültiger Wert'),
    distributionType: yup.string().required('Wird benötigt').typeError('Ungültiger Wert'),
    targetDate: yup.date().nullable().typeError('Ungültiger Datumswert')
      .test('onceFulfilled', 'Spezifischer Zeitpunkt wird benötigt bei einmaligen Kosten', (value, context) => {
        if (context.parent.cycle == CycleType.Once) {
          return !!value
        }
        return true
      }),
  }),
  initialValues: {
    id: isClone ? undefined : historyCost?.id ?? undefined,
    labelID: historyCost?.label?.id ?? undefined,
    relativeOffset: historyCost?.relativeOffset ?? 1,
    cycle: historyCost?.cycle ?? employeeHistory?.cycle ?? CycleType.Monthly,
    amountType: historyCost?.amountType ?? EmployeeCostType.Fixed,
    amount: historyCost?.amount
      ? historyCost?.amountType == EmployeeCostType.Fixed
        ? AmountToFloat(historyCost.amount, 2)
        : AmountToFloat(historyCost.amount, 3)
      : 0,
    distributionType: historyCost?.distributionType ?? EmployeeCostDistributionType.Employee,
    targetDate: historyCost?.targetDate ? DateToUTCDate(historyCost.targetDate) : undefined,
  },
})

const [labelID, labelIDProps] = defineField('labelID')
const [cycle, cycleProps] = defineField('cycle')
const [relativeOffset, relativeOffsetProps] = defineField('relativeOffset')
const [amountType, amountTypeProps] = defineField('amountType')
const [amount, amountProps] = defineField('amount')
const [distributionType, distributionTypeProps] = defineField('distributionType')
const [targetDate, targetDateProps] = defineField('targetDate')

const onSubmit = handleSubmit((values) => {
  isLoading.value = true
  errorMessage.value = ''

  if (isCreate) {
    createEmployeeHistoryCost(employeeHistory?.id, values)
      .then(async (historyCost) => {
        dialogRef?.value.close(historyCost.id)
        toast.add({
          summary: 'Erfolg',
          detail: `Lohnnebenkosten "${EmployeeHistoryCostUtils.title(historyCost)}" wurde angelegt`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
      .catch(() => {
        errorMessage.value = 'Lohnnebenkosten konnte nicht angelegt werden'
      })
      .finally(() => {
        isLoading.value = false
      })
  }
  else {
    updateEmployeeHistoryCost(values)
      .then(async (historyCost) => {
        dialogRef?.value.close(historyCost.id)
        toast.add({
          summary: 'Erfolg',
          detail: `Lohnnebenkosten wurde bearbeitet`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
      .catch(() => {
        errorMessage.value = 'Lohnnebenkosten konnte nicht bearbeitet werden'
      })
      .finally(() => {
        isLoading.value = false
      })
  }
})

const onParseAmount = (event: Event) => {
  if (event instanceof InputEvent) {
    parseNumberInput(event, amount as Ref<number>, false)
  }
}

const onCreateEmployeeCostLabel = () => {
  dialog.open(EmployeeHistoryCostLabelDialog, {
    props: {
      header: 'Neues Kostenlabel anlegen',
      ...ModalConfig,
    },
    onClose: (options) => {
      if (options?.data) {
        const isLoadingCostLabels = ref(true)
        listEmployeeHistoryCostsLabels(false)
          .then(() => {
            setFieldValue('labelID', options.data)
          })
          .catch(() => {
            employeesCostLabelsErrorMessage.value = 'Lohnkosten Labels konnten nicht geladen werden'
          })
          .finally(() => {
            isLoadingCostLabels.value = false
          })
      }
    },
  })
}

const onEditEmployeeCostLabel = (employeeCostLabelToEdit: EmployeeHistoryCostLabelResponse) => {
  dialog.open(EmployeeHistoryCostLabelDialog, {
    props: {
      header: 'Kostenlabel bearbeiten',
      ...ModalConfig,
    },
    data: {
      employeeCostLabelToEdit: employeeCostLabelToEdit,
    },
    onClose: (options) => {
      if (options?.data) {
        const isLoadingCostLabels = ref(true)
        listEmployeeHistoryCostsLabels(false)
          .then(() => {
            setFieldValue('labelID', options.data)
          })
          .catch(() => {
            employeesCostLabelsErrorMessage.value = 'Lohnkosten Labels konnten nicht geladen werden'
          })
          .finally(() => {
            isLoadingCostLabels.value = false
          })
      }
    },
  })
}

const onDeleteEmployeeCostLabel = (employeeCostLabelToDelete: EmployeeHistoryCostLabelResponse) => {
  confirm.require({
    header: 'Löschen',
    message: `Kostenlabel "${employeeCostLabelToDelete.name}" vollständig löschen?`,
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      isLoading.value = true
      deleteEmployeeHistoryCostLabel(employeeCostLabelToDelete.id)
        .then(() => {
          // Make sure we refresh also if the cost isn't saved due to the deleted label
          requiresRefresh.value = true
          if (employeeCostLabelToDelete.id === labelID.value) {
            setFieldValue('labelID', null)
          }
          toast.add({
            summary: 'Erfolg',
            detail: `Kostenlabel "${employeeCostLabelToDelete.name}" wurde gelöscht`,
            severity: 'success',
            life: Config.TOAST_LIFE_TIME,
          })
        })
        .catch(() => {
          toast.add({
            summary: 'Fehler',
            detail: `Kostenlabel "${employeeCostLabelToDelete.name}" konnte nicht gelöscht werden`,
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

const isOnce = computed(() => cycle.value as CycleType == CycleType.Once)
const isFixedAmount = computed(() => amountType.value == EmployeeCostType.Fixed)
</script>
