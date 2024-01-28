<template>
  <form @keyup.enter="onSubmit" class="grid grid-cols-2 gap-2">
    <div class="col-span-2 flex flex-col gap-2">
      <label class="text-sm font-bold" for="name">Name*</label>
      <InputText v-model="name" v-bind="nameProps"
                 :class="{'p-invalid': errors['attributes.name']?.length}"
                 id="name" type="text" autofocus/>
      <small class="text-red-400">{{errors["attributes.name"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="hours-per-month">Max. Stunden pro Monat*</label>
      <InputText v-model.number="hoursPerMonth" v-bind="hoursPerMonthProps"
                 :class="{'p-invalid': errors['attributes.hoursPerMonth']?.length}"
                 id="hours-per-month" type="number" min="0"/>
      <small class="text-red-400">{{errors["attributes.hoursPerMonth"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="vacation-days-per-year">Urlaubstage pro Jahr*</label>
      <InputText v-model.number="vacationDaysPerYear" v-bind="vacationDaysPerYearProps"
                 :class="{'p-invalid': errors['attributes.vacationDaysPerYear']?.length}"
                 id="vacation-days-per-year" type="number" min="0"/>
      <small class="text-red-400">{{errors["attributes.vacationDaysPerYear"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="vacation-days-per-year">Eintritt</label>
        <i class="pi pi-info-circle" v-tooltip="'Wann der Mitarbeiter ins Unternehmen eingetreten ist'"></i>
      </div>
      <Calendar v-model="entry" v-bind="entryProps" date-format="dd.mm.yy" showIcon showButtonBar
                :class="{'p-invalid': errors['attributes.entry']?.length}"/>
      <small class="text-red-400">{{errors["attributes.entry"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="vacation-days-per-year">Austritt</label>
        <i class="pi pi-info-circle" v-tooltip="'Falls der Mitarbeiter aus dem Unternehmen austritt'"></i>
      </div>
      <Calendar v-model="exit" v-bind="exitProps" date-format="dd.mm.yy" showIcon showButtonBar
                :class="{'p-invalid': errors['attributes.exit']?.length}"/>
      <small class="text-red-400">{{errors["attributes.exit"] || '&nbsp;'}}</small>
    </div>

    <div v-if="person?.id" class="flex justify-end col-span-full">
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
import type {IPersonFormDialog} from "~/interfaces/dialog-interfaces";
import type {StrapiPerson} from "~/models/person";
import {useForm} from "vee-validate";
import * as yup from 'yup';
import {Config} from "~/config/config";

const dialogRef = inject<IPersonFormDialog>('dialogRef');
const confirm = useConfirm()
const toast = useToast()

// Data
const person = dialogRef?.value.data?.person

const { values, defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    attributes: yup.object({
      name: yup.string().trim().required('Name wird benötigt'),
      hoursPerMonth: yup.number().typeError('Bitte Zahl eingeben').min(0, 'Muss mindestens 0 sein'),
      vacationDaysPerYear: yup.number().typeError('Bitte Zahl eingeben').min(0, 'Muss mindestens 0 sein'),
      entry: yup.date().nullable().typeError('Bitte Datum eingeben'),
      exit: yup.date().nullable().typeError('Bitte Datum eingeben'),
    })
  }),
  initialValues: {
    id: person?.id ?? undefined,
    attributes: {
      name: person?.attributes.name ?? '',
      hoursPerMonth: person?.attributes.hoursPerMonth ?? 0,
      vacationDaysPerYear: person?.attributes.vacationDaysPerYear ?? 0,
      entry: person?.attributes.entry ? new Date(person?.attributes.entry as string) : undefined,
      exit: person?.attributes.exit ? new Date(person?.attributes.exit as string) : undefined,
    }
  } as StrapiPerson
});

const [name, nameProps] = defineField('attributes.name')
const [hoursPerMonth, hoursPerMonthProps] = defineField('attributes.hoursPerMonth')
const [vacationDaysPerYear, vacationDaysPerYearProps] = defineField('attributes.vacationDaysPerYear')
const [entry, entryProps] = defineField('attributes.entry')
const [exit, exitProps] = defineField('attributes.exit')

const onSubmit = handleSubmit((values) => {
  if (values.attributes.entry instanceof Date) {
    values.attributes.entry.setMinutes(values.attributes.entry.getMinutes() - values.attributes.entry.getTimezoneOffset())
  }
  if (values.attributes.exit instanceof Date) {
    values.attributes.exit.setMinutes(values.attributes.exit.getMinutes() - values.attributes.exit.getTimezoneOffset())
  }
  dialogRef?.value.close(values)
})
const onDelete = (event: PointerEvent) => {
  confirm.require({
    target: event.currentTarget as HTMLElement,
    message: 'Mitarbeiter vollständig löschen?',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      $fetch('/api/employees', {
        method: 'delete',
        body: person,
      }).then(async () => {
        dialogRef?.value.close('deleted')
      }).catch(() => {
        toast.add({
          summary: 'Fehler',
          detail: `Mitarbeiter konnte nicht gelöscht werden`,
          severity: 'error',
          life: Config.TOAST_LIFE_TIME,
        })
      })
    },
    reject: () => {
    }
  });
}
</script>
