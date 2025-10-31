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
      {{ salary.nextExecutionDate ? DateStringToFormattedDate(salary.nextExecutionDate) : 'Keine weitere' }}
      über
      {{ NumberToFormattedCurrency(AmountToFloat(salary.amount), salary.currency.localeCode) }} {{
        salary.currency.code
      }}
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
        :options="salaryCostsLabels.data"
        option-label="name"
        option-value="id"
        filter
        auto-filter-focus
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
      <InputNumber
        v-bind="amountProps"
        id="amount"
        v-model="amount"
        :class="{ 'p-invalid': errors['amount']?.length }"
        :mode="isFixedAmount ? 'currency' : 'decimal'"
        :suffix="isFixedAmount ? '' : ' %'"
        :allow-empty="false"
        :currency="selectedCurrencyCode"
        currency-display="code"
        :locale="selectedLocalCode"
        fluid
        :max-fraction-digits="2"
        :disabled="isLoading"
        @paste="onParseAmount"
        @input="event => amount = event.value"
        @focus="selectAllOnFocus"
      />
      <small class="text-liqui-red">{{ errors["amount"] }}</small>
    </div>

    <div
      v-if="showBaseCostSelect"
      class="flex flex-col gap-2 col-span-full md:col-span-1"
    >
      <div class="flex items-center gap-2">
        <label
          class="text-sm font-bold"
          for="base-salary-cost-ids"
        >Berechnungsgrundlage</label>
        <i
          v-tooltip.top="'Optional: Auswahl bestehender Lohnkosten als Basis für den Prozentwert, leer lassen um auf dem Lohn zu basieren'"
          class="pi pi-info-circle"
        />
      </div>
      <MultiSelect
        v-bind="baseSalaryCostIDsProps"
        id="base-salary-cost-ids"
        v-model="baseSalaryCostIDs"
        empty-message="Keine Lohnkosten verfügbar"
        :options="baseCostSelectOptions"
        option-label="name"
        option-value="value"
        placeholder="Bitte wählen"
        display="chip"
        :loading="isLoadingBaseCosts"
        :disabled="isLoadingBaseCosts || isLoading"
        :class="{ 'p-invalid': errors['baseSalaryCostIDs']?.length }"
      />
      <div class="flex justify-between gap-2">
        <small class="text-liqui-red">{{ errors['baseSalaryCostIDs'] }}</small>
        <small
          v-if="baseCostsErrorMessage.length"
          class="text-liqui-red"
        >{{ baseCostsErrorMessage }}</small>
      </div>
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
import type { ISalaryCostFormDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'
import type { SalaryCostFormData, SalaryCostLabelResponse, SalaryCostResponse } from '~/models/employee'
import { CycleType, EmployeeCostDistributionType, EmployeeCostType } from '~/config/enums'
import {
  CostCycleTypeToOptions,
  EmployeeCostDistributionTypeToOptions,
  EmployeeCostTypeToOptions,
} from '~/utils/enum-helper'
import { ModalConfig } from '~/config/dialog-props'
import SalaryCostLabelDialog from '~/components/dialogs/SalaryCostLabelDialog.vue'
import { SalaryCostUtils } from '~/utils/models/salary-cost-utils'
import { selectAllOnFocus } from '~/utils/element-helper'

const dialogRef = inject<ISalaryCostFormDialog>('dialogRef')!

const { createSalaryCost, updateSalaryCost, listSalaryCosts } = useSalaryCosts()
const { salaryCostsLabels, listSalaryCostsLabels, deleteSalaryCostLabel } = useSalaryCostLabels()
const toast = useToast()
const dialog = useDialog()
const confirm = useConfirm()

const salary = dialogRef.value.data?.salary
const salaryCost = dialogRef.value.data?.salaryCostToEdit
const isClone = dialogRef.value.data?.isClone
const isCreate = isClone || !salaryCost?.id
const isLoading = ref(false)
const isLoadingCostLabels = ref(true)
const errorMessage = ref('')
const employeesCostLabelsErrorMessage = ref('')
const requiresRefresh = ref(false)
const isLoadingBaseCosts = ref(false)
const baseCostsErrorMessage = ref('')
const baseCostOptions = ref<SalaryCostResponse[]>([])

listSalaryCostsLabels(false)
  .catch(() => {
    employeesCostLabelsErrorMessage.value = 'Lohnkosten Labels konnten nicht geladen werden'
  })
  .finally(() => {
    isLoadingCostLabels.value = false
  })

const { defineField, errors, handleSubmit, meta, setFieldValue } = useForm<SalaryCostFormData>({
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
    baseSalaryCostIDs: yup.array()
      .of(yup.number().typeError('Ungültiger Wert').min(1, 'Ungültiger Wert'))
      .ensure()
      .test('noDuplicates', 'Doppelte Auswahl nicht erlaubt', (value) => {
        if (!value) {
          return true
        }
        return new Set(value).size === value.length
      })
      .test('notSelf', 'Lohnkosten können nicht auf sich selbst basieren', (value) => {
        if (!value || value.length === 0) {
          return true
        }
        if (!salaryCost?.id || isClone) {
          return true
        }
        return !value.includes(salaryCost.id)
      }),
  }),
  initialValues: {
    id: isClone ? undefined : salaryCost?.id ?? undefined,
    labelID: salaryCost?.label?.id ?? undefined,
    relativeOffset: salaryCost?.relativeOffset ?? 1,
    cycle: salaryCost?.cycle ?? salary?.cycle ?? CycleType.Monthly,
    amountType: salaryCost?.amountType ?? EmployeeCostType.Fixed,
    amount: salaryCost?.amount
      ? salaryCost?.amountType == EmployeeCostType.Fixed
        ? AmountToFloat(salaryCost.amount, 2)
        : AmountToFloat(salaryCost.amount, 3)
      : 0,
    distributionType: salaryCost?.distributionType ?? EmployeeCostDistributionType.Employee,
    targetDate: salaryCost?.targetDate ? DateToUTCDate(salaryCost.targetDate) : undefined,
    baseSalaryCostIDs: salaryCost?.baseSalaryCostIDs ?? [],
  },
})

const [labelID, labelIDProps] = defineField('labelID')
const [cycle, cycleProps] = defineField('cycle')
const [relativeOffset, relativeOffsetProps] = defineField('relativeOffset')
const [amountType, amountTypeProps] = defineField('amountType')
const [amount, amountProps] = defineField('amount')
const [distributionType, distributionTypeProps] = defineField('distributionType')
const [targetDate, targetDateProps] = defineField('targetDate')
const [baseSalaryCostIDs, baseSalaryCostIDsProps] = defineField('baseSalaryCostIDs')

watch(amountType, (value) => {
  if (value !== EmployeeCostType.Percentage) {
    setFieldValue('baseSalaryCostIDs', [])
  }
})

const showBaseCostSelect = computed(() => amountType.value === EmployeeCostType.Percentage)

const availableBaseCostOptions = computed(() => {
  return baseCostOptions.value.filter((cost) => {
    if (!salaryCost?.id || isClone) {
      return true
    }
    return cost.id !== salaryCost.id
  })
})

const baseCostDistributionLabel = (cost: SalaryCostResponse) =>
  cost.distributionType === EmployeeCostDistributionType.Employer ? 'Arbeitgeber' : 'Arbeitnehmer'

const baseCostSelectOptions = computed(() => {
  return availableBaseCostOptions.value.map(cost => ({
    value: cost.id,
    name: `${SalaryCostUtils.title(cost)} (${baseCostDistributionLabel(cost)})`,
  }))
})

onMounted(() => {
  if (salary?.id) {
    isLoadingBaseCosts.value = true
    listSalaryCosts(salary.id)
      .then((response) => {
        baseCostOptions.value = response.data
      })
      .catch(() => {
        baseCostsErrorMessage.value = 'Lohnkosten konnten nicht geladen werden'
      })
      .finally(() => {
        isLoadingBaseCosts.value = false
      })
  }
})

const onSubmit = handleSubmit((values) => {
  isLoading.value = true
  errorMessage.value = ''

  if (isCreate) {
    createSalaryCost(salary?.id, values)
      .then(async (salaryCost) => {
        dialogRef?.value.close(salaryCost.id)
        toast.add({
          summary: 'Erfolg',
          detail: `Lohnnebenkosten "${SalaryCostUtils.title(salaryCost)}" wurde angelegt`,
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
    updateSalaryCost(values)
      .then(async (salaryCost) => {
        dialogRef?.value.close(salaryCost.id)
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
  if (event instanceof ClipboardEvent) {
    const pastedText = event.clipboardData?.getData('text') ?? ''
    const parsedAmount = parseCurrency(pastedText, false)
    amount.value = parsedAmount.length > 0 ? parseFloat(parsedAmount) : 0
  }
}

const onCreateEmployeeCostLabel = () => {
  dialog.open(SalaryCostLabelDialog, {
    props: {
      header: 'Neues Kostenlabel anlegen',
      ...ModalConfig,
    },
    onClose: (options) => {
      const isLoadingCostLabels = ref(true)
      listSalaryCostsLabels(false)
        .then(() => {
          if (options?.data) {
            setFieldValue('labelID', options.data)
          }
        })
        .catch(() => {
          employeesCostLabelsErrorMessage.value = 'Lohnkosten Labels konnten nicht geladen werden'
        })
        .finally(() => {
          isLoadingCostLabels.value = false
        })
    },
  })
}

const onEditEmployeeCostLabel = (employeeCostLabelToEdit: SalaryCostLabelResponse) => {
  dialog.open(SalaryCostLabelDialog, {
    props: {
      header: 'Kostenlabel bearbeiten',
      ...ModalConfig,
    },
    data: {
      employeeCostLabelToEdit: employeeCostLabelToEdit,
    },
    onClose: (options) => {
      const isLoadingCostLabels = ref(true)
      listSalaryCostsLabels(false)
        .then(() => {
          if (options?.data) {
            setFieldValue('labelID', options.data)
          }
        })
        .catch(() => {
          employeesCostLabelsErrorMessage.value = 'Lohnkosten Labels konnten nicht geladen werden'
        })
        .finally(() => {
          isLoadingCostLabels.value = false
        })
    },
  })
}

const onDeleteEmployeeCostLabel = (employeeCostLabelToDelete: SalaryCostLabelResponse) => {
  confirm.require({
    header: 'Löschen',
    message: `Kostenlabel "${employeeCostLabelToDelete.name}" vollständig löschen?`,
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      isLoading.value = true
      deleteSalaryCostLabel(employeeCostLabelToDelete.id)
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
const selectedCurrencyCode = computed(() => salary.currency.code)
const selectedLocalCode = computed(() => salary.currency.localeCode)
</script>
