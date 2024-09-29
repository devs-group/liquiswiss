<template>
  <form @submit.prevent class="grid grid-cols-2 gap-2">
    <div class="flex flex-col gap-2 col-span-full">
      <label class="text-sm font-bold" for="name">Name *</label>
      <InputText v-model="name" v-bind="nameProps"
                 :class="{'p-invalid': errors['name']?.length}"
                 id="name" type="text" autofocus/>
      <small class="text-red-400">{{errors["name"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full">
      <label class="text-sm font-bold" for="name">Kategorie *</label>
      <Dropdown v-model="category" v-bind="categoryProps" editable empty-message="Keine Kategorien gefunden"
                :options="categories" option-label="name" option-value="id"
                placeholder="Bitte wählen"
                :class="{'p-invalid': errors['category']?.length}"
                id="name" type="text" autofocus/>
      <small class="text-red-400">{{errors["category"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="name">Währung *</label>
      <Dropdown v-model="currency" v-bind="currencyProps" editable empty-message="Keine Währungen gefunden"
                :options="currencies" option-label="code" option-value="id"
                placeholder="Bitte wählen"
                :class="{'p-invalid': errors['currency']?.length}"
                id="name" type="text" autofocus/>
      <small class="text-red-400">{{errors["currency"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="name">Betrag *</label>
        <i class="pi pi-info-circle text-red-600" v-tooltip="'Negatives Vorzeichen = Ausgabe'"></i>
      </div>
      <InputText v-model="amount" v-bind="amountProps"
                 :class="{'p-invalid': errors['amount']?.length}"
                 id="name" type="text" autofocus/>
      <small class="text-red-400">{{errors["amount"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="name">Zahlungstyp</label>
      <Dropdown v-model="type" v-bind="typeProps" editable empty-message="Keine Typen gefunden"
                :options="TransactionTypeToOptions()" option-label="name" option-value="value"
                placeholder="Bitte wählen"
                :class="{'p-invalid': errors['type']?.length}"
                id="name" type="text" autofocus/>
      <small class="text-red-400">{{errors["type"] || '&nbsp;'}}</small>
    </div>

    <div v-if="isRepeatingTransaction" class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="name">Zahlungszyklus</label>
      <Dropdown v-model="cycle" v-bind="cycleProps" editable empty-message="Keine Zyklen gefunden"
                :options="CycleTypeToOptions()" option-label="name" option-value="value"
                placeholder="Bitte wählen"
                :class="{'p-invalid': errors['cycle']?.length}"
                id="name" type="text" autofocus/>
      <small class="text-red-400">{{errors["cycle"] || '&nbsp;'}}</small>
    </div>
    <span v-else class="hidden md:block"></span>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="vacation-days-per-year">Von *</label>
        <i class="pi pi-info-circle" v-tooltip="'Ab wann beginnt diese Transaktion?'"></i>
      </div>
      <Calendar v-model="startDate" v-bind="startDateProps" date-format="dd.mm.yy" showIcon showButtonBar
                :class="{'p-invalid': errors['startDate']?.length}"/>
      <small class="text-red-400">{{errors["startDate"] || '&nbsp;'}}</small>
    </div>

    <div v-if="isRepeatingTransaction" class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="vacation-days-per-year">Bis</label>
        <i class="pi pi-info-circle" v-tooltip="'(Optional) Bis wann geht diese Transaktion?'"></i>
      </div>
      <Calendar v-model="endDate" v-bind="endDateProps" date-format="dd.mm.yy" showIcon showButtonBar
                :class="{'p-invalid': errors['endDate']?.length}"/>
      <small class="text-red-400">{{errors["endDate"] || '&nbsp;'}}</small>
    </div>
    <span v-else class="hidden md:block"></span>

    <div v-if="transaction?.id" class="flex justify-end col-span-full">
      <Button @click="onDeleteTransaction" :loading="isLoading" label="Löschen" severity="danger" size="small"/>
    </div>

    <hr class="my-4 col-span-full"/>

    <div class="flex justify-end gap-2 col-span-full">
      <Button @click="onSubmit" :disabled="!meta.valid || isLoading" :loading="isLoading" label="Speichern" type="submit"/>
      <Button @click="dialogRef?.close()" :loading="isLoading" label="Abbrechen" severity="secondary"/>
    </div>
  </form>
</template>

<script setup lang="ts">
import type {ITransactionFormDialog} from "~/interfaces/dialog-interfaces";
import {useForm} from "vee-validate";
import * as yup from 'yup';
import {Config} from "~/config/config";
import {type TransactionFormData} from "~/models/transaction";
import useGlobalData from "~/composables/useGlobalData";
import {CycleType, TransactionType} from "~/config/enums";
import {CycleTypeToOptions, TransactionTypeToOptions} from "~/utils/enum-helper";
import {DateToUTCDate} from "~/utils/format-helper";
import {AmountToInteger} from "~/utils/number-helper";

const dialogRef = inject<ITransactionFormDialog>('dialogRef')!;

const {createTransaction, updateTransaction, deleteTransaction} = useTransactions()
const {categories, currencies} = useGlobalData()
const confirm = useConfirm()
const toast = useToast()

// Data
const isLoading = ref(false)
const transaction = dialogRef.value.data?.transaction
const isCreate = !transaction?.id

const { defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
    amount: yup.number().required('Betrag wird benötigt').typeError('Ungültiger Betrag')
        .test('Not 0', 'Muss grösser oder kleiner 0 sein', (value) => {
          return AmountToInteger(value) !== 0;
        }),
    cycle: yup.string().required('Zahlungs-Zyklus wird benötigt'),
    type: yup.string().required('Typ wird benötigt'),
    startDate: yup.date().required('Start wird benötigt').typeError('Bitte Datum eingeben'),
    endDate: yup.date().nullable().typeError('Bitte Datum eingeben'),
    category: yup.number().required('Kategorie wird benötigt').typeError('Ungültige Kategorie'),
    currency: yup.number().required('Währung wird benötigt').typeError('Ungültige Währung'),
  }),
  initialValues: {
    id: transaction?.id ?? undefined,
    name: transaction?.name ?? '',
    amount: transaction?.amount ? AmountToFloat(transaction.amount) : '',
    cycle: transaction?.cycle ?? CycleType.Monthly,
    type: transaction?.type ?? TransactionType.Single,
    startDate: transaction?.startDate ? DateToUTCDate(transaction?.startDate) : null,
    endDate: transaction?.endDate ? DateToUTCDate(transaction?.endDate) : undefined,
    category: transaction?.category.id ?? null,
    currency: transaction?.currency.id ?? null,
  } as TransactionFormData
});

const [name, nameProps] = defineField('name')
const [amount, amountProps] = defineField('amount')
const [cycle, cycleProps] = defineField('cycle')
const [type, typeProps] = defineField('type')
const [category, categoryProps] = defineField('category')
const [currency, currencyProps] = defineField('currency')
const [startDate, startDateProps] = defineField('startDate')
const [endDate, endDateProps] = defineField('endDate')

const onSubmit = handleSubmit((values) => {
  isLoading.value = true
  values.startDate.setMinutes(values.startDate.getMinutes() - values.startDate.getTimezoneOffset())
  if (values.type == TransactionType.Single) {
    values.endDate = undefined
  }
  else if (values.endDate instanceof Date) {
    values.endDate.setMinutes(values.endDate.getMinutes() - values.endDate.getTimezoneOffset())
  }

  if (isCreate) {
    createTransaction(values)
        .then(() => {
          dialogRef.value.close()
          toast.add({
            summary: 'Erfolg',
            detail: `Transaktion "${values.name}" wurde angelegt`,
            severity: 'success',
            life: Config.TOAST_LIFE_TIME,
          })
        })
        .finally(() => {
          isLoading.value = false
        })
  } else {
    updateTransaction(values)
        .then(() => {
          dialogRef.value.close()
          toast.add({
            summary: 'Erfolg',
            detail: `Transaktion "${values.name}" wurde bearbeitet`,
            severity: 'success',
            life: Config.TOAST_LIFE_TIME,
          })
        })
        .finally(() => {
          isLoading.value = false
        })
  }
})

const onDeleteTransaction = (event: MouseEvent) => {
  confirm.require({
    target: event.currentTarget as HTMLElement,
    message: 'Transaktion vollständig löschen?',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      if (transaction) {
        isLoading.value = true
        deleteTransaction(transaction.id)
            .then(() => {
              toast.add({
                summary: 'Erfolg',
                detail: `Transaktion "${transaction.name}" wurde gelöscht`,
                severity: 'success',
                life: Config.TOAST_LIFE_TIME,
              })
              dialogRef.value.close()
            })
            .catch(() => {
              toast.add({
                summary: 'Fehler',
                detail: `Transaktion konnte nicht gelöscht werden`,
                severity: 'error',
                life: Config.TOAST_LIFE_TIME,
              })
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

const isRepeatingTransaction = computed(() => {
  return type.value === TransactionType.Repeating
})
</script>
