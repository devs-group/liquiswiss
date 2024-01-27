<template>
  <form @keyup.enter="onSubmit" class="flex flex-col gap-2">
    <span class="flex flex-col gap-2">
      <label class="text-sm font-bold" for="name">Name</label>
      <InputText v-model.trim="name" v-bind="nameProps"
                 :class="{'p-invalid': errors.name?.length}"
                 id="name" type="text" autofocus/>
      <small class="text-red-400">{{errors.name || '&nbsp;'}}</small>
    </span>

    <span class="flex flex-col gap-2">
      <label class="text-sm font-bold" for="hours-per-month">Stunden pro Monat</label>
      <InputText v-model="hoursPerMonth" v-bind="hoursPerMonthProps"
                 :class="{'p-invalid': errors.hoursPerMonth?.length}"
                 id="hours-per-month" type="number" min="0"/>
      <small class="text-red-400">{{errors.hoursPerMonth || '&nbsp;'}}</small>
    </span>

    <span class="flex flex-col gap-2">
      <label class="text-sm font-bold" for="vacation-days-per-year">Urlaubstage pro Jahr</label>
      <InputText v-model="vacationDaysPerYear" v-bind="vacationDaysPerYearProps"
                 :class="{'p-invalid': errors.vacationDaysPerYear?.length}"
                 id="vacation-days-per-year" type="number" min="0"/>
      <small class="text-red-400">{{errors.vacationDaysPerYear || '&nbsp;'}}</small>
    </span>

    <hr class="my-4"/>

    <div class="flex justify-end gap-2">
      <Button @click="onSubmit" :disabled="!meta.valid" label="Speichern"/>
      <Button @click="dialogRef?.close()" label="Abbrechen" severity="secondary"/>
    </div>
  </form>
</template>

<script setup lang="ts">
import type {IPersonFormDialog} from "~/interfaces/DialogInterfaces";
import type {Person} from "~/models/person";
import {useForm} from "vee-validate";
import * as yup from 'yup';

const dialogRef = inject<IPersonFormDialog>('dialogRef');

// Data
const person = dialogRef?.value.data?.person

const { values, defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    name: yup.string().required('Name wird benötigt'),
    hoursPerMonth: yup.number().typeError('Bitte Zahl eingeben').min(0, 'Muss mindestens 0 sein'),
    vacationDaysPerYear: yup.number().typeError('Bitte Zahl eingeben').min(0, 'Muss mindestens 0 sein'),
  }),
  initialValues: {
    id: person?.id ?? undefined,
    name: person?.name ?? '',
    hoursPerMonth: person?.hoursPerMonth ?? 0,
    vacationDaysPerYear: person?.vacationDaysPerYear ?? 0,
  } as Person
});

const [name, nameProps] = defineField('name')
const [hoursPerMonth, hoursPerMonthProps] = defineField('hoursPerMonth')
const [vacationDaysPerYear, vacationDaysPerYearProps] = defineField('vacationDaysPerYear')

const onSubmit = handleSubmit((values) => {
  dialogRef?.value.close(values)
})
</script>
