<template>
  <form @keyup.enter="onSubmit" class="grid grid-cols-2 gap-2">
    <div class="flex flex-col gap-2 col-span-full">
      <label class="text-sm font-bold" for="name">Name *</label>
      <InputText v-model="name" v-bind="nameProps"
                 :class="{'p-invalid': errors['attributes.name']?.length}"
                 id="name" type="text" autofocus/>
      <small class="text-red-400">{{errors["attributes.name"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full">
      <label class="text-sm font-bold" for="name">Kategorie *</label>
      <Dropdown v-model="category" v-bind="categoryProps" editable empty-message="Keine Kategorien gefunden"
                :options="categories" option-label="attributes.name" option-value="id"
                placeholder="Bitte wählen"
                :class="{'p-invalid': errors['attributes.category']?.length}"
                id="name" type="text" autofocus/>
      <small class="text-red-400">{{errors["attributes.category"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="name">Währung *</label>
      <Dropdown v-model="currency" v-bind="currencyProps" editable empty-message="Keine Währungen gefunden"
                :options="currencies" option-label="attributes.code" option-value="id"
                placeholder="Bitte wählen"
                :class="{'p-invalid': errors['attributes.currency']?.length}"
                id="name" type="text" autofocus/>
      <small class="text-red-400">{{errors["attributes.currency"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="name">Betrag *</label>
      <InputText v-model="amount" v-bind="amountProps"
                 :class="{'p-invalid': errors['attributes.amount']?.length}"
                 id="name" type="text" autofocus/>
      <small class="text-red-400">{{errors["attributes.amount"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="name">Zahlungstyp</label>
      <Dropdown v-model="type" v-bind="typeProps" editable empty-message="Keine Typen gefunden"
                :options="RevenueTypeToOptions()" option-label="name" option-value="value"
                placeholder="Bitte wählen"
                :class="{'p-invalid': errors['attributes.type']?.length}"
                id="name" type="text" autofocus/>
      <small class="text-red-400">{{errors["attributes.type"] || '&nbsp;'}}</small>
    </div>

    <div v-if="isRepeatingRevenue" class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="name">Zahlungszyklus</label>
      <Dropdown v-model="cycle" v-bind="cycleProps" editable empty-message="Keine Zyklen gefunden"
                :options="CycleTypeToOptions()" option-label="name" option-value="value"
                placeholder="Bitte wählen"
                :class="{'p-invalid': errors['attributes.cycle']?.length}"
                id="name" type="text" autofocus/>
      <small class="text-red-400">{{errors["attributes.cycle"] || '&nbsp;'}}</small>
    </div>
    <span v-else class="hidden md:block"></span>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="vacation-days-per-year">Start *</label>
        <i class="pi pi-info-circle" v-tooltip="'Ab wann beginnt die erste Einnahme?'"></i>
      </div>
      <Calendar v-model="start" v-bind="startProps" date-format="dd.mm.yy" showIcon showButtonBar
                :class="{'p-invalid': errors['attributes.start']?.length}"/>
      <small class="text-red-400">{{errors["attributes.start"] || '&nbsp;'}}</small>
    </div>

    <div v-if="isRepeatingRevenue" class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="vacation-days-per-year">Ende</label>
        <i class="pi pi-info-circle" v-tooltip="'(Optional) Wann findet die letzte Einnahme statt?'"></i>
      </div>
      <Calendar v-model="end" v-bind="endProps" date-format="dd.mm.yy" showIcon showButtonBar
                :class="{'p-invalid': errors['attributes.end']?.length}"/>
      <small class="text-red-400">{{errors["attributes.end"] || '&nbsp;'}}</small>
    </div>
    <span v-else class="hidden md:block"></span>

    <div v-if="revenue?.id" class="flex justify-end col-span-full">
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
import type {IRevenueFormDialog} from "~/interfaces/dialog-interfaces";
import {useForm} from "vee-validate";
import * as yup from 'yup';
import {Config} from "~/config/config";
import {type StrapiRevenue} from "~/models/revenue";
import useGlobalData from "~/composables/useGlobalData";
import {CycleType, RevenueType} from "~/config/enums";
import {CycleTypeToOptions, RevenueTypeToOptions} from "~/utils/enum-helper";

const dialogRef = inject<IRevenueFormDialog>('dialogRef');
const confirm = useConfirm()
const toast = useToast()
const {categories, currencies} = useGlobalData()

// Data
const revenue = dialogRef?.value.data?.revenue

const { values, defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    attributes: yup.object({
      name: yup.string().trim().required('Name wird benötigt'),
      category: yup.number().required('Kategorie wird benötigt').typeError('Ungültige Kategorie'),
      currency: yup.number().required('Währung wird benötigt').typeError('Ungültige Währung'),
      type: yup.string().required('Typ wird benötigt'),
      amount: yup.number().required('Betrag wird benötigt').typeError('Ungültiger Betrag'),
      cycle: yup.string().required('Zahlungs-Zyklus wird benötigt'),
      start: yup.date().required('Start wird benötigt').typeError('Bitte Datum eingeben'),
      end: yup.date().nullable().typeError('Bitte Datum eingeben'),
    })
  }),
  initialValues: {
    id: revenue?.id ?? undefined,
    attributes: {
      name: revenue?.attributes.name ?? '',
      category: revenue?.attributes.category ?? undefined,
      currency: revenue?.attributes.currency ?? '',
      type: revenue?.attributes.type ?? RevenueType.Single,
      amount: revenue?.attributes.amount ?? '',
      cycle: revenue?.attributes.cycle ?? CycleType.Monthly,
      start: revenue?.attributes.start ? new Date(revenue?.attributes.start as string) : undefined,
      end: revenue?.attributes.end ? new Date(revenue?.attributes.end as string) : undefined,
    }
  } as StrapiRevenue
});

const [name, nameProps] = defineField('attributes.name')
const [category, categoryProps] = defineField('attributes.category')
const [currency, currencyProps] = defineField('attributes.currency')
const [type, typeProps] = defineField('attributes.type')
const [amount, amountProps] = defineField('attributes.amount')
const [cycle, cycleProps] = defineField('attributes.cycle')
const [start, startProps] = defineField('attributes.start')
const [end, endProps] = defineField('attributes.end')

const onSubmit = handleSubmit((values) => {
  if (values.attributes.start instanceof Date) {
    values.attributes.start.setMinutes(values.attributes.start.getMinutes() - values.attributes.start.getTimezoneOffset())
  }
  if (values.attributes.end instanceof Date) {
    values.attributes.end.setMinutes(values.attributes.end.getMinutes() - values.attributes.end.getTimezoneOffset())
  }
  dialogRef?.value.close(values)
})

const onDelete = (event: PointerEvent) => {
  confirm.require({
    target: event.currentTarget as HTMLElement,
    message: 'Einnahme vollständig löschen?',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      $fetch('/api/revenue', {
        method: 'delete',
        body: revenue,
      }).then(async () => {
        dialogRef?.value.close('deleted')
      }).catch(() => {
        toast.add({
          summary: 'Fehler',
          detail: `Einnahme konnte nicht gelöscht werden`,
          severity: 'error',
          life: Config.TOAST_LIFE_TIME,
        })
      })
    },
    reject: () => {
    }
  });
}

const isRepeatingRevenue = computed(() => {
  return type.value === RevenueType.Repeating
})
</script>
