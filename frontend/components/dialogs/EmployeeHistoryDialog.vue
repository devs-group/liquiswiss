<template>
  <form
    id="employee-history-form"
    class="grid grid-cols-2 gap-2"
    @submit.prevent
  >
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
      <InputText
        v-bind="salaryProps"
        id="salary"
        v-model="salary"
        :class="{ 'p-invalid': errors['salary']?.length }"
        type="text"
        :disabled="isLoading"
        @input="onParseAmount"
      />
      <small class="text-liqui-red">{{ errors["salary"] || '&nbsp;' }}</small>
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
        :options="CycleTypeToOptions()"
        option-label="name"
        option-value="value"
        placeholder="Bitte wählen"
        :class="{ 'p-invalid': errors['cycle']?.length }"
        type="text"
      />
      <small class="text-liqui-red">{{ errors["cycle"] || '&nbsp;' }}</small>
    </div>

    <Message
      v-if="isClone && EmployeeHistoryUtils.hasCosts(employeeHistory!)"
      severity="warn"
      size="small"
      class="col-span-full"
    >
      Achtung: Lohnkosten werden aktuell nicht geklont, diese können nachträglich separat übertragen werden
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
      v-if="!isCreate && employeeHistory?.withSeparateCosts"
      severity="info"
      size="small"
      class="col-span-full"
    >
      Hinweis: Für diese Historie werden Lohnkosten separat geführt
    </Message>

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
        v-if="!isCreate"
        class="flex justify-end col-span-full"
      >
        <Button
          :disabled="isLoading"
          severity="danger"
          size="small"
          icon="pi pi-trash"
          @click="onDeleteEmployeeHistory"
        />
      </div>
    </div>
  </form>
</template>

<script setup lang="ts">
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import type { IHistoryFormDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'
import type { EmployeeHistoryPUTFormData } from '~/models/employee'
import { CycleType } from '~/config/enums'
import { EmployeeHistoryUtils } from '~/utils/models/employee-history-utils'

const dialogRef = inject<IHistoryFormDialog>('dialogRef')!

const { getOrganisationCurrencyID } = useAuth()
const { createEmployeeHistory, updateEmployeeHistory, deleteEmployeeHistory } = useEmployeeHistories()
const { currencies, getCurrencyLabel } = useGlobalData()
const confirm = useConfirm()
const toast = useToast()

// Data
const isLoading = ref(false)
const employeeID = dialogRef.value.data!.employeeID
const employeeHistory = ref(dialogRef.value.data!.employeeHistory)
const isClone = ref(dialogRef.value.data?.isClone)
const isCreate = computed(() => isClone.value || !employeeHistory.value?.id)
const errorMessage = ref('')

const { defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    hoursPerMonth: yup.number().typeError('Bitte Zahl eingeben').min(0, 'Muss mindestens 0 sein').max(480, 'Kann maximal 480 sein'),
    salary: yup.number().typeError('Bitte Gehalt eingeben').min(0, 'Muss mindestens 0 sein'),
    currencyID: yup.number().required('Währung wird benötigt').typeError('Bitte gültige Währung eingeben'),
    vacationDaysPerYear: yup.number().typeError('Bitte Zahl eingeben').min(0, 'Muss mindestens 0 sein').max(365, 'Kann maximal 365 sein'),
    fromDate: yup.date().typeError('Bitte Datum eingeben').required('Von wird benötigt'),
    cycle: yup.string().required('Zahlungs-Zyklus wird benötigt'),
  }),
  initialValues: {
    id: isClone.value ? undefined : employeeHistory.value?.id ?? undefined,
    hoursPerMonth: employeeHistory.value?.hoursPerMonth ?? 0,
    salary: isNumber(employeeHistory.value?.salary) ? AmountToFloat(employeeHistory.value!.salary) : 0,
    currencyID: employeeHistory.value?.currency.id ?? getOrganisationCurrencyID.value,
    vacationDaysPerYear: employeeHistory.value?.vacationDaysPerYear ?? 0,
    fromDate: employeeHistory.value?.fromDate ? DateToUTCDate(employeeHistory.value.fromDate) : null,
    cycle: employeeHistory.value?.cycle ?? CycleType.Monthly,
  } as EmployeeHistoryPUTFormData,
})

const [hoursPerMonth, hoursPerMonthProps] = defineField('hoursPerMonth')
const [salary, salaryProps] = defineField('salary')
const [currencyID, currencyIDProps] = defineField('currencyID')
const [vacationDaysPerYear, vacationDaysPerYearProps] = defineField('vacationDaysPerYear')
const [fromDate, fromDateProps] = defineField('fromDate')
const [cycle, cycleProps] = defineField('cycle')

const onParseAmount = (event: Event) => {
  if (event instanceof InputEvent) {
    parseNumberInput(event, salary, false)
  }
}

const onSubmit = handleSubmit((values) => {
  isLoading.value = true
  errorMessage.value = ''
  values.fromDate.setMinutes(values.fromDate.getMinutes() - values.fromDate.getTimezoneOffset())

  if (isCreate.value) {
    createEmployeeHistory(employeeID, values)
      .then(() => {
        dialogRef.value.close(true)
        toast.add({
          summary: 'Erfolg',
          detail: `Historie wurde angelegt`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
      .catch(() => {
        errorMessage.value = `Historie konnte nicht angelegt werden`
        nextTick(() => {
          scrollToParentBottom('employee-history-form')
        })
      })
      .finally(() => {
        isLoading.value = false
      })
  }
  else {
    updateEmployeeHistory(employeeID, values)
      .then(() => {
        dialogRef.value.close(true)
        toast.add({
          summary: 'Erfolg',
          detail: `Historie wurde bearbeitet`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
      .catch(() => {
        errorMessage.value = `Historie konnte nicht bearbeitet werden`
        nextTick(() => {
          scrollToParentBottom('employee-history-form')
        })
      })
      .finally(() => {
        isLoading.value = false
      })
  }
})

const onDeleteEmployeeHistory = () => {
  confirm.require({
    header: 'Löschen',
    message: 'Historie vollständig löschen?',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      if (employeeHistory.value) {
        isLoading.value = true
        deleteEmployeeHistory(employeeID, employeeHistory.value.id)
          .then(() => {
            toast.add({
              summary: 'Erfolg',
              detail: `Historie wurde gelöscht`,
              severity: 'success',
              life: Config.TOAST_LIFE_TIME,
            })
            dialogRef.value.close(true)
          })
          .catch(() => {
            errorMessage.value = `Historie konnte nicht gelöscht werden`
            nextTick(() => {
              scrollToParentBottom('employee-history-form')
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
