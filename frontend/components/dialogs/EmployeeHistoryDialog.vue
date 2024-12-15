<template>
  <form @submit.prevent id="employee-history-form" class="grid grid-cols-2 gap-2">
    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="hours-per-month">Arbeitsstunden pro Monat*</label>
      <InputText v-model.number="hoursPerMonth" v-bind="hoursPerMonthProps"
                 :class="{'p-invalid': errors['hoursPerMonth']?.length}"
                 id="hours-per-month" type="number" min="0"
                 :disabled="isLoading"/>
      <small class="text-liqui-red">{{errors["hoursPerMonth"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="vacation-days-per-year">Urlaubstage pro Jahr*</label>
      <InputText v-model.number="vacationDaysPerYear" v-bind="vacationDaysPerYearProps"
                 :class="{'p-invalid': errors['vacationDaysPerYear']?.length}"
                 id="vacation-days-per-year" type="number" min="0"
                 :disabled="isLoading"/>
      <small class="text-liqui-red">{{errors["vacationDaysPerYear"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="salary-per-month">Lohnkosten pro Monat*</label>
        <i class="pi pi-info-circle" v-tooltip.top="'Bruttolohn + Arbeitgeberkosten'"></i>
      </div>
      <InputText v-model="salaryPerMonth" v-bind="salaryPerMonthProps"
                 @input="onParseAmount"
                 :class="{'p-invalid': errors['salaryPerMonth']?.length}"
                 id="salary-per-month" type="text"
                 :disabled="isLoading"/>
      <small class="text-liqui-red">{{errors["salaryPerMonth"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="salary-currencyID">Währung des Lohns*</label>
      <Select v-model="currencyID" v-bind="currencyIDProps"
                empty-message="Keine Währungen gefunden"
                :disabled="isLoading"
                :class="{'p-invalid': errors['currencyID']?.length}"
                :options="currencies" option-label="code" option-value="id"
                placeholder="Bitte wählen"
                id="salary-currency-id"/>
      <small class="text-liqui-red">{{errors["currencyID"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="vacation-days-per-year">Von*</label>
        <i class="pi pi-info-circle" v-tooltip.top="'Ab wann gelten diese Daten?'"></i>
      </div>
      <DatePicker v-model="fromDate" v-bind="fromDateProps" :min-date="getDisabledPreviousDate" :max-date="getDisabledNextDate" date-format="dd.mm.yy" showIcon showButtonBar
                :class="{'p-invalid': errors['fromDate']?.length}" :disabled="isLoading"/>
      <small class="text-liqui-red">{{errors["fromDate"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="vacation-days-per-year">Bis</label>
        <i class="pi pi-info-circle" v-tooltip.top="'Bis wann gelten diese Daten?'"></i>
      </div>
      <Message severity="secondary">Wird automatisch berechnet...</Message>
    </div>

    <div v-if="!isClone && !isCreate" class="flex justify-end col-span-full">
      <Button @click="onDeleteEmployeeHistory" :disabled="isLoading" label="Löschen" severity="danger" size="small" icon="pi pi-trash"/>
    </div>

    <hr class="my-4 col-span-full"/>

    <Message v-if="errorMessage.length" severity="error" :closable="false" class="col-span-full">{{errorMessage}}</Message>

    <div class="flex justify-end gap-2 col-span-full">
      <Button @click="onSubmit" severity="info" :disabled="!meta.valid || isLoading" :loading="isLoading" label="Speichern" icon="pi pi-save" type="submit"/>
      <Button @click="dialogRef.close()" :disabled="isLoading" label="Abbrechen" severity="secondary"/>
    </div>
  </form>
</template>

<script setup lang="ts">
import type {IHistoryFormDialog} from "~/interfaces/dialog-interfaces";
import {useForm} from "vee-validate";
import * as yup from 'yup';
import {Config} from "~/config/config";
import type {EmployeeHistoryFormData} from "~/models/employee";
import {DateToUTCDate} from "~/utils/format-helper";
import {parseNumberInput, scrollToParentBottom} from "~/utils/element-helper";
import {isNumber} from "~/utils/number-helper";

const dialogRef = inject<IHistoryFormDialog>('dialogRef')!;

const {employeeHistories, createEmployeeHistory, updateEmployeeHistory, deleteEmployeeHistory} = useEmployees()
const {currencies} = useGlobalData()
const confirm = useConfirm()
const toast = useToast()

// Data
const isLoading = ref(false)
const employeeID = dialogRef.value.data!.employeeID
const employeeHistory = dialogRef.value.data!.employeeHistory
const isClone = dialogRef.value.data?.isClone
const isCreate = isClone || !employeeHistory?.id
const errorMessage = ref('')

const { defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    hoursPerMonth: yup.number().typeError('Bitte Zahl eingeben').min(0, 'Muss mindestens 0 sein'),
    salaryPerMonth: yup.number().typeError('Bitte Gehalt eingeben').min(0, 'Muss mindestens 0 sein'),
    currencyID: yup.number().required('Währung wird benötigt').typeError('Bitte gültige Währung eingeben'),
    vacationDaysPerYear: yup.number().typeError('Bitte Zahl eingeben').min(0, 'Muss mindestens 0 sein'),
    fromDate: yup.date().typeError('Bitte Datum eingeben').required('Von wird benötigt'),
  }),
  initialValues: {
    id: isClone ? undefined : employeeHistory?.id ?? undefined,
    hoursPerMonth: employeeHistory?.hoursPerMonth ?? 0,
    salaryPerMonth: isNumber(employeeHistory?.salaryPerMonth) ? AmountToFloat(employeeHistory!.salaryPerMonth) : 0,
    currencyID: employeeHistory?.currency.id ?? null,
    vacationDaysPerYear: employeeHistory?.vacationDaysPerYear ?? 0,
    fromDate: employeeHistory?.fromDate ? DateToUTCDate(employeeHistory.fromDate) : null,
  } as EmployeeHistoryFormData
});

const [hoursPerMonth, hoursPerMonthProps] = defineField('hoursPerMonth')
const [salaryPerMonth, salaryPerMonthProps] = defineField('salaryPerMonth')
const [currencyID, currencyIDProps] = defineField('currencyID')
const [vacationDaysPerYear, vacationDaysPerYearProps] = defineField('vacationDaysPerYear')
const [fromDate, fromDateProps] = defineField('fromDate')

const getDisabledPreviousDate = computed(() => {
  const fromDate = employeeHistory?.fromDate ? DateToUTCDate(employeeHistory?.fromDate) : DateToUTCDate(new Date())
  const otherFutureHistories = employeeHistories.value.data
      .filter(h => h.id !== employeeHistory?.id && h.toDate && DateToUTCDate(h.toDate) <= fromDate)
      .map(h => DateToUTCDate(h.toDate!))
      .sort((a, b) => b.getTime() - a.getTime())
  if (otherFutureHistories.length > 0) {
    const previousLockedDate = otherFutureHistories[0]
    previousLockedDate.setDate(previousLockedDate.getDate() + 1)
    return previousLockedDate
  }
})

const getDisabledNextDate = computed(() => {
  const fromDate = employeeHistory?.fromDate ? DateToUTCDate(employeeHistory?.fromDate) : DateToUTCDate(new Date())
  const otherFutureHistories = employeeHistories.value.data
      .filter(h => h.id !== employeeHistory?.id && DateToUTCDate(h.fromDate) >= fromDate)
      .map(h => DateToUTCDate(h.fromDate))
      .sort((a, b) => a.getTime() - b.getTime())
  if (otherFutureHistories.length > 0) {
    const nextLockedDate = otherFutureHistories[0]
    nextLockedDate.setDate(nextLockedDate.getDate() - 1)
    return nextLockedDate
  }
})

const onParseAmount = (event: Event) => {
  if (event instanceof InputEvent) {
    parseNumberInput(event, salaryPerMonth, false)
  }
}

const onSubmit = handleSubmit((values) => {
  isLoading.value = true
  errorMessage.value = ''
  values.fromDate.setMinutes(values.fromDate.getMinutes() - values.fromDate.getTimezoneOffset())

  if (isCreate) {
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
          });
        })
        .finally(() => {
          isLoading.value = false
        })
  } else {
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
          });
        })
        .finally(() => {
          isLoading.value = false
        })
  }
})

const onDeleteEmployeeHistory = (event: MouseEvent) => {
  confirm.require({
    header: 'Löschen',
    message: 'Historie vollständig löschen?',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      if (employeeHistory) {
        isLoading.value = true
        deleteEmployeeHistory(employeeID, employeeHistory.id)
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
              });
            })
            .finally(() => {
              isLoading.value = false
            })
      }
    },
    reject: () => {
    }
  });
}
</script>
