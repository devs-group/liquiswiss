<template>
  <form @keyup.enter="onSubmit" class="grid grid-cols-2 gap-2">
    <div class="col-span-2 flex flex-col gap-2">
      <label class="text-sm font-bold" for="name">Name*</label>
      <InputText v-model="name" v-bind="nameProps"
                 :class="{'p-invalid': errors['name']?.length}"
                 id="name" type="text"/>
      <small class="text-red-400">{{errors["name"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="hours-per-month">Max. Stunden pro Monat*</label>
      <InputText v-model.number="hoursPerMonth" v-bind="hoursPerMonthProps"
                 :class="{'p-invalid': errors['hoursPerMonth']?.length}"
                 id="hours-per-month" type="number" min="0"/>
      <small class="text-red-400">{{errors["hoursPerMonth"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="vacation-days-per-year">Urlaubstage pro Jahr*</label>
      <InputText v-model.number="vacationDaysPerYear" v-bind="vacationDaysPerYearProps"
                 :class="{'p-invalid': errors['vacationDaysPerYear']?.length}"
                 id="vacation-days-per-year" type="number" min="0"/>
      <small class="text-red-400">{{errors["vacationDaysPerYear"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="vacation-days-per-year">Eintritt*</label>
        <i class="pi pi-info-circle" v-tooltip="'Wann der Mitarbeiter ins Unternehmen eingetreten ist'"></i>
      </div>
      <Calendar v-model="entryDate" v-bind="entryDateProps" date-format="dd.mm.yy" showIcon showButtonBar
                :class="{'p-invalid': errors['entryDate']?.length}"/>
      <small class="text-red-400">{{errors["entryDate"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="vacation-days-per-year">Austritt</label>
        <i class="pi pi-info-circle" v-tooltip="'Falls der Mitarbeiter aus dem Unternehmen austritt'"></i>
      </div>
      <Calendar v-model="exitDate" v-bind="exitDateProps" date-format="dd.mm.yy" showIcon showButtonBar
                :class="{'p-invalid': errors['exitDate']?.length}"/>
      <small class="text-red-400">{{errors["exitDate"] || '&nbsp;'}}</small>
    </div>

    <div v-if="employee?.id" class="flex justify-end col-span-full">
      <Button @click="onDelete" label="Löschen" severity="danger" size="small"/>
    </div>

    <hr class="my-4 col-span-full"/>

    <div class="flex justify-end gap-2 col-span-full">
      <Button @click="onSubmit" :disabled="!meta.valid" label="Speichern"/>
      <Button @click="dialogRef?.close()" label="Abbrechen" severity="secondary"/>
    </div>
  </form>
</template>

<script setup lang="ts">
import type {IEmployeeFormDialog} from "~/interfaces/dialog-interfaces";
import {useForm} from "vee-validate";
import * as yup from 'yup';
import {Config} from "~/config/config";
import type {EmployeeFormData} from "~/models/employee";
import {DateToUTCDate} from "~/utils/format-helper";

const dialogRef = inject<IEmployeeFormDialog>('dialogRef');

const {deleteEmployee} = useEmployees()
const confirm = useConfirm()
const toast = useToast()

// Data
const employee = dialogRef?.value.data?.employee

const { defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
    hoursPerMonth: yup.number().typeError('Bitte Zahl eingeben').min(0, 'Muss mindestens 0 sein'),
    vacationDaysPerYear: yup.number().typeError('Bitte Zahl eingeben').min(0, 'Muss mindestens 0 sein'),
    entryDate: yup.date().typeError('Bitte Datum eingeben').required('Eintritt wird benötigt'),
    exitDate: yup.date().nullable().typeError('Bitte Datum eingeben'),
  }),
  initialValues: {
    id: employee?.id ?? undefined,
    name: employee?.name ?? '',
    hoursPerMonth: employee?.hoursPerMonth ?? 0,
    vacationDaysPerYear: employee?.vacationDaysPerYear ?? 0,
    entryDate: employee?.entryDate ? DateToUTCDate(employee.entryDate) : null,
    exitDate: employee?.exitDate ? DateToUTCDate(employee.exitDate) : undefined,
  } as EmployeeFormData
});

const [name, nameProps] = defineField('name')
const [hoursPerMonth, hoursPerMonthProps] = defineField('hoursPerMonth')
const [vacationDaysPerYear, vacationDaysPerYearProps] = defineField('vacationDaysPerYear')
const [entryDate, entryDateProps] = defineField('entryDate')
const [exitDate, exitDateProps] = defineField('exitDate')

const onSubmit = handleSubmit((values) => {
  values.entryDate.setMinutes(values.entryDate.getMinutes() - values.entryDate.getTimezoneOffset())
  if (values.exitDate instanceof Date) {
    values.exitDate.setMinutes(values.exitDate.getMinutes() - values.exitDate.getTimezoneOffset())
  }
  dialogRef?.value.close(values)
})

const onDelete = (payload: MouseEvent) => {
  confirm.require({
    target: payload.currentTarget as HTMLElement,
    message: 'Mitarbeiter vollständig löschen?',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      if (employee) {
        deleteEmployee(employee.id)
            .then(async () => {
              dialogRef?.value.close('deleted')
            }).catch(() => {
              toast.add({
                summary: 'Fehler',
                detail: `Mitarbeiter konnte nicht gelöscht werden`,
                severity: 'error',
                life: Config.TOAST_LIFE_TIME,
              })
            })
      }
    },
    reject: () => {
    }
  });
}
</script>
