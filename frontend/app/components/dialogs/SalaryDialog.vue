<template>
  <form
    id="salary-form"
    class="grid grid-cols-2 gap-2"
    @submit.prevent
  >
    <template v-if="!isTermination">
      <div class="flex flex-col gap-2 col-span-full md:col-span-1">
        <label
          class="text-sm font-bold"
          for="hours-per-month"
        >Arbeitsstunden pro Monat *</label>
        <InputText
          v-bind="hoursPerMonthProps"
          id="hours-per-month"
          v-model.number="hoursPerMonth"
          :class="{ 'p-invalid': errors['hoursPerMonth']?.length }"
          type="number"
          min="0"
          :disabled="isLoading"
        />
        <small class="text-liqui-red">{{ errors["hoursPerMonth"] || '&nbsp;' }}</small>
      </div>

      <div class="flex flex-col gap-2 col-span-full md:col-span-1">
        <label
          class="text-sm font-bold"
          for="vacation-days-per-year"
        >Urlaubstage pro Jahr *</label>
        <InputText
          v-bind="vacationDaysPerYearProps"
          id="vacation-days-per-year"
          v-model.number="vacationDaysPerYear"
          :class="{ 'p-invalid': errors['vacationDaysPerYear']?.length }"
          type="number"
          min="0"
          :disabled="isLoading"
        />
        <small class="text-liqui-red">{{ errors["vacationDaysPerYear"] || '&nbsp;' }}</small>
      </div>

      <div class="flex flex-col gap-2 col-span-full md:col-span-1">
        <label
          class="text-sm font-bold"
          for="salary-currencyID"
        >Währung des Lohns *</label>
        <Select
          v-bind="currencyIDProps"
          id="currency-id"
          v-model="currencyID"
          empty-message="Keine Währungen gefunden"
          filter
          auto-filter-focus
          empty-filter-message="Keine Resultate gefunden"
          :disabled="isLoading"
          :class="{ 'p-invalid': errors['currencyID']?.length }"
          :options="currencies"
          :option-label="getCurrencyLabel"
          option-value="id"
          placeholder="Bitte wählen"
        />
        <small class="text-liqui-red">{{ errors["currencyID"] || '&nbsp;' }}</small>
      </div>

      <div class="flex flex-col gap-2 col-span-full md:col-span-1">
        <div class="flex items-center gap-2">
          <label
            class="text-sm font-bold"
            for="salary-per-month"
          >Bruttolohn *</label>
          <i
            v-tooltip.top="'Empfehlung: Bruttolohn angeben und Lohnkosten separat erfassen'"
            class="pi pi-info-circle"
          />
        </div>
        <InputNumber
          v-bind="amountProps"
          id="amount"
          v-model="amount"
          :class="{ 'p-invalid': errors['amount']?.length }"
          mode="currency"
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
        <small class="text-liqui-red">{{ errors["amount"] || '&nbsp;' }}</small>
      </div>

      <div class="flex flex-col gap-2 col-span-full md:col-span-1">
        <div class="flex items-center gap-2">
          <label
            class="text-sm font-bold"
            for="vacation-days-per-year"
          >Von *</label>
          <i
            v-tooltip.top="'Das &quot;Bis&quot; Datum wird automatisch berechnet'"
            class="pi pi-info-circle"
          />
        </div>
        <DatePicker
          v-model="fromDate"
          v-bind="fromDateProps"
          :disabled-dates="getDisabledDates"
          date-format="dd.mm.yy"
          show-icon
          show-button-bar
          :class="{ 'p-invalid': errors['fromDate']?.length }"
          :disabled="isLoading"
        />
        <small class="text-liqui-red">{{ errors["fromDate"] || '&nbsp;' }}</small>
      </div>

      <div class="flex flex-col gap-2 col-span-full md:col-span-1">
        <label
          class="text-sm font-bold"
          for="name"
        >Zahlungszyklus *</label>
        <Select
          v-bind="cycleProps"
          id="cycle"
          v-model="cycle"
          empty-message="Keine Zyklen gefunden"
          :options="SalaryCycleTypeToOptions()"
          option-label="name"
          option-value="value"
          placeholder="Bitte wählen"
          :class="{ 'p-invalid': errors['cycle']?.length }"
          type="text"
        />
        <small class="text-liqui-red">{{ errors["cycle"] || '&nbsp;' }}</small>
      </div>

      <Message
        v-if="isClone && SalaryUtils.hasSeparateCostsDefined(salary!)"
        severity="info"
        size="small"
        class="col-span-full"
      >
        Die Lohnkosten werden ebenfalls kopiert
      </Message>
      <Message
        v-else-if="isCreate"
        severity="info"
        size="small"
        class="col-span-full"
      >
        Lohnkosten können separat, direkt nach dem Erstellen hinzugefügt werden
      </Message>

      <Message
        v-if="!isCreate && salary?.withSeparateCosts"
        severity="info"
        size="small"
        class="col-span-full"
      >
        Hinweis: Für diesen Lohn werden Lohnkosten separat geführt
      </Message>
    </template>
    <template v-else>
      <div class="flex flex-col gap-2 col-span-full md:col-span-1">
        <div class="flex items-center gap-2">
          <label
            class="text-sm font-bold"
            for="vacation-days-per-year"
          >Von *</label>
          <i
            v-tooltip.top="'Empfehlung: Letzter Arbeitstag bzw. NACH letzter Lohnzahlung'"
            class="pi pi-info-circle"
          />
        </div>
        <DatePicker
          v-model="fromDate"
          v-bind="fromDateProps"
          :disabled-dates="getDisabledDates"
          date-format="dd.mm.yy"
          show-icon
          show-button-bar
          :class="{ 'p-invalid': errors['fromDate']?.length }"
          :disabled="isLoading || terminationErrorMessage.length > 0"
        />
        <small
          v-if="getEarliestRecommendedTerminationDate"
          class="text-liqui-blue"
        >
          Vorschlag für nächsten Austritt: {{ DateStringToFormattedDate(getEarliestRecommendedTerminationDate) }}
        </small>
        <small class="text-liqui-red">{{ errors["fromDate"] || '&nbsp;' }}</small>
      </div>

      <Message
        v-if="terminationErrorMessage.length"
        severity="error"
        size="small"
        :closable="false"
        class="col-span-full"
      >
        {{ terminationErrorMessage }}
      </Message>

      <Message
        severity="warn"
        size="small"
        class="col-span-full"
      >
        Austritt = Auflösung des Arbeitsverhältnisses
      </Message>
    </template>

    <hr class="my-4 col-span-full">

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
        :label="isClone ? 'Klonen' : 'Speichern'"
        icon="pi pi-save"
        type="submit"
        @click="onSubmit"
      />
      <Button
        :disabled="isLoading"
        label="Abbrechen"
        severity="contrast"
        @click="dialogRef.close()"
      />
      <div
        v-if="!isCreate && !isTermination"
        class="flex justify-end col-span-full"
      >
        <Button
          :disabled="isLoading"
          severity="danger"
          size="small"
          icon="pi pi-trash"
          @click="onDeleteSalary"
        />
      </div>
    </div>
  </form>
</template>

<script setup lang="ts">
import { nextTick, computed } from 'vue'
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import type { ISalaryFormDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'
import type { SalaryPUTFormData } from '~/models/employee'
import { CycleType } from '~/config/enums'
import { SalaryUtils } from '~/utils/models/salary-utils'
import { SalaryCycleTypeToOptions } from '~/utils/enum-helper'
import { selectAllOnFocus } from '~/utils/element-helper'
import { DateToApiFormat, DateToUTCDate } from '~/utils/format-helper'
import { DateFirstDayOfNextMonth } from '~/utils/date-helper'

const dialogRef = inject<ISalaryFormDialog>('dialogRef')!

const { getOrganisationCurrencyID } = useAuth()
const { createSalary, updateSalary, deleteSalary, listSalaries, salaries } = useSalaries()
const { copySalaryCost } = useSalaryCosts()
const { currencies, getCurrencyLabel } = useGlobalData()
const confirm = useConfirm()
const toast = useToast()

// Data
const isLoading = ref(false)
const employeeID = dialogRef.value.data!.employeeID
const salary = ref(dialogRef.value.data!.salary)
const isTermination = ref(dialogRef.value.data?.isTermination)
const isClone = ref(dialogRef.value.data?.isClone)
const isCreate = computed(() => !isTermination.value && (isClone.value || !salary.value?.id))
const errorMessage = ref('')
const terminationErrorMessage = ref('')

if (isTermination.value) {
  // Only for termination to be able to set the earliest allowed termination date
  isLoading.value = true
  await listSalaries(employeeID)
    .catch(() => {
      terminationErrorMessage.value = 'Fehler beim Laden der Löhne'
    })
    .finally(() => {
      isLoading.value = false
    })
}
else if (isClone.value || isCreate.value) {
  try {
    await listSalaries(employeeID)
  }
  catch {
    // swallowing error here - user can still attempt manual input
  }
}

const selectedCurrencyCode = computed(() => currencies.value.find(c => c.id == currencyID.value)?.code)
const selectedLocalCode = computed(() => currencies.value.find(c => c.id == currencyID.value)?.localeCode)
const activeSalaries = computed(() => salaries.value.data.filter(s => !s.isTermination))
const currentEditableSalaryId = computed(() => {
  if (isClone.value || isTermination.value) {
    return null
  }
  return salary.value?.id ?? null
})

const getEarliestRecommendedTerminationDate = computed(() => {
  if (activeSalaries.value.length > 0) {
    const latestSalary = activeSalaries.value[0]
    if (latestSalary.nextExecutionDate) {
      return DateFirstDayOfNextMonth(DateToUTCDate(latestSalary.nextExecutionDate))
    }
  }
  return undefined
})
const getDisabledDates = computed(() => {
  return activeSalaries.value
    .filter(sl => currentEditableSalaryId.value === null || sl.id !== currentEditableSalaryId.value)
    .map(sl => DateToUTCDate(sl.fromDate))
})

const disabledFromDateKeys = computed(() => {
  return new Set(
    activeSalaries.value
      .map(sl => DateToApiFormat(DateToUTCDate(sl.fromDate))),
  )
})

const cloneDefaultFromDate = computed(() => {
  if (!isClone.value || !salary.value?.fromDate) {
    return salary.value?.fromDate ? DateToUTCDate(salary.value.fromDate) : null
  }
  const sourceDay = DateToUTCDate(salary.value.fromDate).getDate()
  const used = new Set(disabledFromDateKeys.value)
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  let year = today.getFullYear()
  let monthIndex = today.getMonth() + 1

  for (let i = 0; i < 120; i += 1) {
    if (monthIndex > 11) {
      monthIndex = 0
      year += 1
    }
    const daysInMonth = new Date(year, monthIndex + 1, 0).getDate()
    const targetDay = Math.min(sourceDay, daysInMonth)
    const candidate = new Date(year, monthIndex, targetDay)
    candidate.setHours(0, 0, 0, 0)
    const candidateKey = DateToApiFormat(candidate)
    if (!used.has(candidateKey)) {
      return candidate
    }
    monthIndex += 1
  }
  return today
})

const { defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    hoursPerMonth: yup.number().typeError('Bitte Zahl eingeben').min(0, 'Muss mindestens 0 sein').max(480, 'Kann maximal 480 sein'),
    amount: yup.number().typeError('Bitte Gehalt eingeben').min(0, 'Muss mindestens 0 sein'),
    currencyID: yup.number().required('Währung wird benötigt').typeError('Bitte gültige Währung eingeben'),
    vacationDaysPerYear: yup.number().typeError('Bitte Zahl eingeben').min(0, 'Muss mindestens 0 sein').max(365, 'Kann maximal 365 sein'),
    fromDate: yup.date().typeError('Bitte Datum eingeben').required('Von wird benötigt'),
    cycle: yup.string().required('Zahlungs-Zyklus wird benötigt'),
  }),
  initialValues: {
    id: isClone.value ? undefined : salary.value?.id ?? undefined,
    hoursPerMonth: salary.value?.hoursPerMonth ?? 0,
    amount: isNumber(salary.value?.amount) ? AmountToFloat(salary.value!.amount) : 0,
    currencyID: salary.value?.currency.id ?? getOrganisationCurrencyID.value,
    vacationDaysPerYear: salary.value?.vacationDaysPerYear ?? 0,
    fromDate: isClone.value
      ? cloneDefaultFromDate.value
      : salary.value?.fromDate ? DateToUTCDate(salary.value.fromDate) : isTermination.value ? getEarliestRecommendedTerminationDate.value : null,
    cycle: salary.value?.cycle ?? CycleType.Monthly,
  } as SalaryPUTFormData,
})

const [hoursPerMonth, hoursPerMonthProps] = defineField('hoursPerMonth')
const [amount, amountProps] = defineField('amount')
const [currencyID, currencyIDProps] = defineField('currencyID')
const [vacationDaysPerYear, vacationDaysPerYearProps] = defineField('vacationDaysPerYear')
const [fromDate, fromDateProps] = defineField('fromDate')
const [cycle, cycleProps] = defineField('cycle')

const onParseAmount = (event: Event) => {
  if (event instanceof ClipboardEvent) {
    const pastedText = event.clipboardData?.getData('text') ?? ''
    const parsedAmount = parseCurrency(pastedText, false)
    amount.value = parsedAmount.length > 0 ? parseFloat(parsedAmount) : 0
  }
}

const cloneSalaryCostsIfNeeded = async (newSalaryID: number) => {
  if (!isClone.value || !salary.value?.id || !salary.value.hasSeparateCostsDefined) {
    return
  }
  try {
    await copySalaryCost(newSalaryID, {
      ids: [],
      sourceSalaryID: salary.value.id,
    })
    await listSalaries(employeeID)
  }
  catch {
    toast.add({
      summary: 'Hinweis',
      detail: 'Lohnkosten konnten nicht übernommen werden',
      severity: 'warn',
      life: Config.TOAST_LIFE_TIME_SHORT,
    })
  }
}

const onSubmit = handleSubmit(async (values) => {
  isLoading.value = true
  errorMessage.value = ''
  values.fromDate.setMinutes(values.fromDate.getMinutes() - values.fromDate.getTimezoneOffset())

  if (isCreate.value) {
    try {
      const createdSalary = await createSalary(employeeID, values)
      await cloneSalaryCostsIfNeeded(createdSalary.id)
      dialogRef.value.close(true)
      toast.add({
        summary: 'Erfolg',
        detail: `Lohn wurde angelegt`,
        severity: 'success',
        life: Config.TOAST_LIFE_TIME,
      })
    }
    catch {
      errorMessage.value = `Lohn konnte nicht angelegt werden`
      nextTick(() => {
        scrollToParentBottom('salary-form')
      })
    }
    finally {
      isLoading.value = false
    }
    return
  }

  if (isTermination.value) {
    try {
      await createSalary(employeeID, {
        ...values,
        isTermination: true,
      })
      dialogRef.value.close(true)
      toast.add({
        summary: 'Erfolg',
        detail: `Austritt wurde angelegt`,
        severity: 'success',
        life: Config.TOAST_LIFE_TIME,
      })
    }
    catch {
      errorMessage.value = `Austritt konnte nicht angelegt werden`
      nextTick(() => {
        scrollToParentBottom('salary-form')
      })
    }
    finally {
      isLoading.value = false
    }
    return
  }

  try {
    await updateSalary(employeeID, values)
    dialogRef.value.close(true)
    toast.add({
      summary: 'Erfolg',
      detail: `Lohn wurde bearbeitet`,
      severity: 'success',
      life: Config.TOAST_LIFE_TIME,
    })
  }
  catch {
    errorMessage.value = `Lohn konnte nicht bearbeitet werden`
    nextTick(() => {
      scrollToParentBottom('salary-form')
    })
  }
  finally {
    isLoading.value = false
  }
})

const onDeleteSalary = () => {
  confirm.require({
    header: 'Löschen',
    message: 'Lohn vollständig löschen?',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      if (salary.value) {
        isLoading.value = true
        deleteSalary(employeeID, salary.value.id)
          .then(() => {
            toast.add({
              summary: 'Erfolg',
              detail: `Lohn wurde gelöscht`,
              severity: 'success',
              life: Config.TOAST_LIFE_TIME,
            })
            dialogRef.value.close(true)
            listSalaries(employeeID)
          })
          .catch(() => {
            errorMessage.value = `Lohn konnte nicht gelöscht werden`
            nextTick(() => {
              scrollToParentBottom('salary-form')
            })
          })
          .finally(() => {
            isLoading.value = false
          })
      }
    },
    reject: () => {
    },
  })
}
</script>
