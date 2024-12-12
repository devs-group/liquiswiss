<template>
  <form @submit.prevent id="transaction-form" class="grid grid-cols-2 gap-2">
    <div class="flex flex-col gap-2 col-span-full">
      <label class="text-sm font-bold" for="name">Name *</label>
      <InputText v-model="name" v-bind="nameProps"
                 :class="{'p-invalid': errors['name']?.length}"
                 id="name" type="text"/>
      <small class="text-liqui-red">{{errors["name"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="name">Kategorie *</label>
      <Select v-model="category" v-bind="categoryProps" empty-message="Keine Kategorien gefunden"
                :options="categories" option-label="name" option-value="id"
                placeholder="Bitte wählen"
                :class="{'p-invalid': errors['category']?.length}"
                id="name" type="text"/>
      <small class="text-liqui-red">{{errors["category"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="name">Mitarbeiter</label>
        <i class="pi pi-info-circle text-liqui-blue" v-tooltip.top="'Optionale Assoziation'"></i>
      </div>
      <Select v-model="employee" v-bind="employeeProps" empty-message="Keine Mitarbeiter gefunden"
                :options="employees.data" option-label="name" option-value="id"
                placeholder="Bitte wählen"
                showClear
                :loading="isLoadingEmployees"
                :disabled="isLoadingEmployees"
                :class="{'p-invalid': errors['employee']?.length}"
                id="name" type="text"/>
      <small class="text-liqui-red">{{errors["employee"] || employeesErrorMessage || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="name">Währung *</label>
      <Select v-model="currency" v-bind="currencyProps" empty-message="Keine Währungen gefunden"
                :options="currencies" option-label="code" option-value="id"
                placeholder="Bitte wählen"
                :class="{'p-invalid': errors['currency']?.length}"
                id="name" type="text"/>
      <small class="text-liqui-red">{{errors["currency"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="name">Betrag *</label>
        <i class="pi pi-info-circle text-liqui-blue" v-tooltip.top="'Negatives Vorzeichen = Ausgabe'"></i>
        <div class="flex-1"></div>
        <small v-if="selectedCurrencyCode && selectedCurrencyCode != Constants.BASE_CURRENCY" class="text-zinc-600 dark:text-zinc-400">{{amountInBaseCurrency}}</small>
      </div>
      <div class="flex item-center gap-2">
        <InputText v-model="amount" v-bind="amountProps"
                   @input="onParseAmount"
                   class="flex-1"
                   :class="{'p-invalid': errors['amount']?.length}"
                   id="name" type="text"/>
        <AmountInvertButton @invert-amount="onInvertAmount" :amount="amount"/>
      </div>
      <small class="text-liqui-red">{{errors["amount"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="vat">Mehrwertsteuer</label>
      <Select v-model="vat" v-bind="vatProps" empty-message="Keine Mehrwertsteuern gefunden"
              :options="vats" option-label="formattedValue" option-value="id"
              placeholder="Wählen (optional)"
              :loading="isLoadingVats"
              :disabled="isLoadingVats"
              show-clear
              :class="{'p-invalid': errors['vat']?.length}"
              id="vat" type="text"
      >
        <template #option="slotProps">
          <div class="flex items-center w-full justify-between">
            <p>{{ slotProps.option.formattedValue }}</p>
            <div v-if="slotProps.option.canEdit" class="flex gap-2 justify-end">
              <Button @click.stop="onEditVat(slotProps.option)" size="small" icon="pi pi-pencil" outlined rounded />
              <Button @click.stop="onDeleteVat(slotProps.option)" size="small" severity="danger" icon="pi pi-trash" outlined rounded />
            </div>
            <i v-else class="pi pi-info-circle text-liqui-blue"
               v-tooltip.top="'Vorgegebene Mehrwertsteuer. Kann nicht bearbeitet bzw. gelöscht werden.'"></i>
          </div>
        </template>

        <template #footer>
          <div class="p-1 pt-0">
            <Button @click="onCreateVat" label="Hinzufügen" fluid severity="secondary" text size="small" icon="pi pi-plus" />
          </div>
        </template>
      </Select>
      <small class="text-liqui-red">{{errors["vat"] || vatsErrorMessage || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col just gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="name">Mehrwertsteuer inklusive?</label>
        <i class="pi pi-info-circle text-liqui-blue" v-tooltip.top="'Anhaken falls Betrag die Mehrwertsteuer bereits enthält'"></i>
      </div>
      <div class="flex items-center flex-1">
        <ToggleSwitch v-model="vatIncluded" v-bind="vatIncludedProps" :disabled="!vat"/>
      </div>
      <small class="text-liqui-red">{{errors["vatIncluded"] || vatsErrorMessage || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="name">Zahlungstyp</label>
      <Select v-model="type" v-bind="typeProps" empty-message="Keine Typen gefunden"
                :options="TransactionTypeToOptions()" option-label="name" option-value="value"
                placeholder="Bitte wählen"
                :class="{'p-invalid': errors['type']?.length}"
                id="name" type="text"/>
      <small class="text-liqui-red">{{errors["type"] || '&nbsp;'}}</small>
    </div>

    <div v-if="isRepeatingTransaction" class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label class="text-sm font-bold" for="name">Zahlungszyklus</label>
      <Select v-model="cycle" v-bind="cycleProps" empty-message="Keine Zyklen gefunden"
                :options="CycleTypeToOptions()" option-label="name" option-value="value"
                placeholder="Bitte wählen"
                :class="{'p-invalid': errors['cycle']?.length}"
                id="name" type="text"/>
      <small class="text-liqui-red">{{errors["cycle"] || '&nbsp;'}}</small>
    </div>
    <span v-else class="hidden md:block"></span>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="vacation-days-per-year">Von *</label>
        <i class="pi pi-info-circle" v-tooltip.top="'Ab wann beginnt diese Transaktion?'"></i>
      </div>
      <DatePicker v-model="startDate" v-bind="startDateProps" date-format="dd.mm.yy" showIcon showButtonBar
                :class="{'p-invalid': errors['startDate']?.length}"/>
      <small class="text-liqui-red">{{errors["startDate"] || '&nbsp;'}}</small>
    </div>

    <div v-if="isRepeatingTransaction" class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="vacation-days-per-year">Bis</label>
        <i class="pi pi-info-circle" v-tooltip.top="'(Optional) Bis wann geht diese Transaktion?'"></i>
      </div>
      <DatePicker v-model="endDate" :min-date="startDate" v-bind="endDateProps" date-format="dd.mm.yy" showIcon showButtonBar
                :class="{'p-invalid': errors['endDate']?.length}"/>
      <small class="text-liqui-red">{{errors["endDate"] || '&nbsp;'}}</small>
    </div>
    <span v-else class="hidden md:block"></span>

    <div v-if="!isClone && !isCreate" class="flex justify-end col-span-full">
      <Button @click="onDeleteTransaction" :loading="isLoading" label="Löschen" icon="pi pi-trash" severity="danger" size="small"/>
    </div>

    <hr class="my-4 col-span-full"/>

    <Message v-if="errorMessage.length" severity="error" :closable="false" class="col-span-full">{{errorMessage}}</Message>

    <div class="flex justify-end gap-2 col-span-full">
      <Button @click="onSubmit" :disabled="!meta.valid || isLoading" :loading="isLoading" label="Speichern" icon="pi pi-save" type="submit"/>
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
import {CycleType, TransactionType} from "~/config/enums";
import {TransactionTypeToOptions, CycleTypeToOptions} from "~/utils/enum-helper";
import {Constants} from "~/utils/constants";
import {NumberToFormattedCurrency} from "~/utils/format-helper";
import {parseNumberInput, scrollToParentBottom} from "~/utils/element-helper";
import {isNumber} from "~/utils/number-helper";
import AmountInvertButton from "~/components/AmountInvertButton.vue";
import {ModalConfig} from "~/config/dialog-props";
import VatDialog from "~/components/dialogs/VatDialog.vue";
import type {VatResponse} from "~/models/vat";

const dialogRef = inject<ITransactionFormDialog>('dialogRef')!;

const {createTransaction, updateTransaction, deleteTransaction} = useTransactions()
const {employees, listEmployees} = useEmployees()
const {vats, listVats, deleteVat} = useVat()
const {categories, currencies, convertAmountToRate} = useGlobalData()
const confirm = useConfirm()
const dialog = useDialog();
const toast = useToast()

// Data
const isLoading = ref(false)
const isLoadingEmployees = ref(true)
const isLoadingVats = ref(true)
const transaction = dialogRef.value.data?.transaction
const isClone = dialogRef.value.data?.isClone
const isCreate = isClone || !transaction?.id
const errorMessage = ref('')
const employeesErrorMessage = ref('')
const vatsErrorMessage = ref('')

listEmployees(false)
    .catch(() => {
      employeesErrorMessage.value = 'Mitarbeiter konnten nicht geladen werden'
    })
    .finally(() => {
      isLoadingEmployees.value = false
    })

listVats()
    .catch(() => {
      vatsErrorMessage.value = 'Mehrwertsteuern konnten nicht geladen werden'
    })
    .finally(() => {
      isLoadingVats.value = false
    })

const { defineField, errors, handleSubmit, meta, setFieldValue } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
    amount: yup.number().required('Betrag wird benötigt').typeError('Ungültiger Betrag'),
    vat: yup.number().nullable().typeError('Ungültige Mehrwertsteuer'),
    vatIncluded: yup.boolean().typeError('Ungültiger Wert'),
    cycle: yup.string().required('Zahlungs-Zyklus wird benötigt'),
    type: yup.string().required('Typ wird benötigt'),
    startDate: yup.date().required('Start wird benötigt').typeError('Bitte Datum eingeben'),
    endDate: yup.date().nullable().typeError('Bitte Datum eingeben'),
    category: yup.number().required('Kategorie wird benötigt').typeError('Ungültige Kategorie'),
    currency: yup.number().required('Währung wird benötigt').typeError('Ungültige Währung'),
    employee: yup.number().nullable().typeError('Ungültiger Mitarbeiter'),
  }),
  initialValues: {
    id: isClone ? undefined : transaction?.id ?? undefined,
    name: transaction?.name ?? '',
    amount: isNumber(transaction?.amount) ? AmountToFloat(transaction!.amount) : '',
    vat: transaction?.vat?.id ?? null,
    vatIncluded: transaction?.vatIncluded ?? false,
    cycle: transaction?.cycle ?? CycleType.Monthly,
    type: transaction?.type ?? TransactionType.Single,
    startDate: transaction?.startDate ? DateToUTCDate(transaction?.startDate) : null,
    endDate: transaction?.endDate ? DateToUTCDate(transaction?.endDate) : undefined,
    category: transaction?.category.id ?? null,
    currency: transaction?.currency.id ?? null,
    employee: transaction?.employee?.id ?? null,
  } as TransactionFormData
});

const [name, nameProps] = defineField('name')
const [amount, amountProps] = defineField('amount')
const [vat, vatProps] = defineField('vat')
const [vatIncluded, vatIncludedProps] = defineField('vatIncluded')
const [cycle, cycleProps] = defineField('cycle')
const [type, typeProps] = defineField('type')
const [startDate, startDateProps] = defineField('startDate')
const [endDate, endDateProps] = defineField('endDate')
const [category, categoryProps] = defineField('category')
const [currency, currencyProps] = defineField('currency')
const [employee, employeeProps] = defineField('employee')

const onSubmit = handleSubmit((values) => {
  isLoading.value = true
  errorMessage.value = ''
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
        .catch(() => {
          errorMessage.value = 'Transaktion konnte nicht angelegt werden'
          nextTick(() => {
            scrollToParentBottom('transaction-form')
          });
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
        .catch(() => {
          errorMessage.value = 'Transaktion konnte nicht bearbeitet werden'
          nextTick(() => {
            scrollToParentBottom('transaction-form')
          });
        })
        .finally(() => {
          isLoading.value = false
        })
  }
})

const onParseAmount = (event: Event) => {
  if (event instanceof InputEvent) {
    parseNumberInput(event, amount, true)
  }
}

const onDeleteTransaction = (event: MouseEvent) => {
  confirm.require({
    header: 'Löschen',
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
              errorMessage.value = 'Transaktion konnte nicht gelöscht werden'
              nextTick(() => {
                scrollToParentBottom('transaction-form')
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

const onCreateVat = () => {
  dialog.open(VatDialog, {
    props: {
      header: 'Neue Mehrwertsteuer anlegen',
      ...ModalConfig,
    },
    onClose: (options) => {
      if (options?.data) {
        setFieldValue('vat', options.data)
      }
    }
  })
}

const onEditVat = (vatToEdit: VatResponse) => {
  dialog.open(VatDialog, {
    props: {
      header: 'Mehrwertsteuer bearbeiten',
      ...ModalConfig,
    },
    data: {
      vatToEdit,
    },
    onClose: (options) => {
      if (options?.data) {
        setFieldValue('vat', options.data)
      }
    }
  })
}

const onDeleteVat = (vatToDelete: VatResponse) => {
  confirm.require({
    header: 'Löschen',
    message: `Mehrwertsteuer "${vatToDelete.formattedValue}" vollständig löschen?`,
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      if (vat) {
        isLoading.value = true
        deleteVat(vatToDelete.id)
            .then(() => {
              if (vatToDelete.id === vat.value) {
                setFieldValue('vat', undefined)
              }
              toast.add({
                summary: 'Erfolg',
                detail: `Mehrwertsteuer "${vatToDelete.formattedValue}" wurde gelöscht`,
                severity: 'success',
                life: Config.TOAST_LIFE_TIME,
              })
            })
            .catch(() => {
              toast.add({
                summary: 'Fehler',
                detail: `Mehrwertsteuer "${vatToDelete.formattedValue}" konnte nicht gelöscht werden`,
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

const onInvertAmount = () => {
  amount.value *= -1
}

const isRepeatingTransaction = computed(() => {
  return type.value === TransactionType.Repeating
})
const selectedCurrencyCode = computed(() => currencies.value.find(c => c.id == currency.value)?.code)
const amountInBaseCurrency = computed(() => {
  let baseAmount = amount.value
  if (selectedCurrencyCode.value) {
    baseAmount = convertAmountToRate(amount.value, selectedCurrencyCode.value)
  }
  return `~ ${NumberToFormattedCurrency(baseAmount, Constants.BASE_LOCALE_CODE)} ${Constants.BASE_CURRENCY}`
})
</script>
