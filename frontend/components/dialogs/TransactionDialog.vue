<template>
  <form
    id="transaction-form"
    class="grid grid-cols-2 gap-2"
    @submit.prevent
  >
    <div class="flex flex-col gap-2 col-span-full">
      <label
        class="text-sm font-bold"
        for="name"
      >Name *</label>
      <InputText
        v-bind="nameProps"
        id="name"
        v-model="name"
        :class="{ 'p-invalid': errors['name']?.length }"
        type="text"
      />
      <small class="text-liqui-red">{{ errors["name"] || '&nbsp;' }}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label
        class="text-sm font-bold"
        for="name"
      >Kategorie *</label>
      <Select
        v-bind="categoryProps"
        id="name"
        v-model="category"
        empty-message="Keine Kategorien gefunden"
        :options="categories"
        option-label="name"
        option-value="id"
        placeholder="Bitte wählen"
        :class="{ 'p-invalid': errors['category']?.length }"
        type="text"
      />
      <small class="text-liqui-red">{{ errors["category"] || '&nbsp;' }}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label
          class="text-sm font-bold"
          for="name"
        >Mitarbeiter</label>
        <i
          v-tooltip.top="'Optionale Assoziation'"
          class="pi pi-info-circle text-liqui-blue"
        />
      </div>
      <Select
        v-bind="employeeProps"
        id="name"
        v-model="employee"
        empty-message="Keine Mitarbeiter gefunden"
        :options="employees.data"
        option-label="name"
        option-value="id"
        placeholder="Bitte wählen"
        show-clear
        :loading="isLoadingEmployees"
        :disabled="isLoadingEmployees"
        :class="{ 'p-invalid': errors['employee']?.length }"
        type="text"
      />
      <small class="text-liqui-red">{{ errors["employee"] || employeesErrorMessage || '&nbsp;' }}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label
        class="text-sm font-bold"
        for="name"
      >Währung *</label>
      <Select
        v-bind="currencyProps"
        id="name"
        v-model="currency"
        empty-message="Keine Währungen gefunden"
        filter
        auto-filter-focus
        empty-filter-message="Keine Resultate gefunden"
        :options="currencies"
        :option-label="getCurrencyLabel"
        option-value="id"
        placeholder="Bitte wählen"
        :class="{ 'p-invalid': errors['currency']?.length }"
        type="text"
      />
      <small class="text-liqui-red">{{ errors["currency"] || '&nbsp;' }}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label
          class="text-sm font-bold"
          for="name"
        >Betrag *</label>
        <i
          v-tooltip.top="'Negatives Vorzeichen = Ausgabe'"
          class="pi pi-info-circle text-liqui-blue"
        />
        <div class="flex-1" />
        <small
          v-if="selectedCurrencyCode && selectedCurrencyCode != getOrganisationCurrencyCode"
          class="text-zinc-600 dark:text-zinc-400"
        >{{ amountInBaseCurrency }}</small>
      </div>
      <div class="flex item-center gap-2">
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
          @paste="onParseAmount"
          @input="event => amount = event.value"
          @focus="selectAllOnFocus"
        />
        <AmountInvertButton
          :amount="amount"
          @invert-amount="onInvertAmount"
        />
      </div>
      <small class="text-liqui-red">{{ errors["amount"] || '&nbsp;' }}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label
        class="text-sm font-bold"
        for="vat"
      >Mehrwertsteuer</label>
      <Select
        v-bind="vatProps"
        id="vat"
        v-model="vat"
        empty-message="Keine Mehrwertsteuern gefunden"
        :options="vats"
        option-label="formattedValue"
        option-value="id"
        placeholder="Wählen (optional)"
        :loading="isLoadingVats"
        :disabled="isLoadingVats"
        show-clear
        :class="{ 'p-invalid': errors['vat']?.length }"
        type="text"
      >
        <template #option="slotProps">
          <div class="flex items-center w-full justify-between">
            <p>{{ slotProps.option.formattedValue }}</p>
            <div
              v-if="slotProps.option.canEdit"
              class="flex gap-2 justify-end"
            >
              <Button
                size="small"
                icon="pi pi-pencil"
                outlined
                rounded
                @click.stop="onEditVat(slotProps.option)"
              />
              <Button
                size="small"
                severity="danger"
                icon="pi pi-trash"
                outlined
                rounded
                @click.stop="onDeleteVat(slotProps.option)"
              />
            </div>
            <i
              v-else
              v-tooltip.top="'Vorgegebene Mehrwertsteuer. Kann nicht bearbeitet bzw. gelöscht werden.'"
              class="pi pi-info-circle text-liqui-blue"
            />
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
              @click="onCreateVat"
            />
          </div>
        </template>
      </Select>
      <small class="text-liqui-red">{{ errors["vat"] || vatsErrorMessage || '&nbsp;' }}</small>
    </div>

    <div class="flex flex-col just gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label
          class="text-sm font-bold"
          for="name"
        >Mehrwertsteuer inklusive?</label>
        <i
          v-tooltip.top="'Anhaken falls Betrag die Mehrwertsteuer bereits enthält'"
          class="pi pi-info-circle text-liqui-blue"
        />
      </div>
      <div class="flex items-center flex-1">
        <ToggleSwitch
          v-model="vatIncluded"
          v-bind="vatIncludedProps"
          :disabled="!vat"
        />
      </div>
      <small class="text-liqui-red">{{ errors["vatIncluded"] || vatsErrorMessage || '&nbsp;' }}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <label
        class="text-sm font-bold"
        for="name"
      >Zahlungstyp</label>
      <Select
        v-bind="typeProps"
        id="name"
        v-model="type"
        empty-message="Keine Typen gefunden"
        :options="TransactionTypeToOptions()"
        option-label="name"
        option-value="value"
        placeholder="Bitte wählen"
        :class="{ 'p-invalid': errors['type']?.length }"
        type="text"
      />
      <small class="text-liqui-red">{{ errors["type"] || '&nbsp;' }}</small>
    </div>

    <div
      v-if="isRepeatingTransaction"
      class="flex flex-col gap-2 col-span-full md:col-span-1"
    >
      <label
        class="text-sm font-bold"
        for="name"
      >Zahlungszyklus</label>
      <Select
        v-bind="cycleProps"
        id="name"
        v-model="cycle"
        empty-message="Keine Zyklen gefunden"
        :options="TransactionCycleTypeToOptions()"
        option-label="name"
        option-value="value"
        placeholder="Bitte wählen"
        :class="{ 'p-invalid': errors['cycle']?.length }"
        type="text"
      />
      <small class="text-liqui-red">{{ errors["cycle"] || '&nbsp;' }}</small>
    </div>
    <span
      v-else
      class="hidden md:block"
    />

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label
          class="text-sm font-bold"
          for="vacation-days-per-year"
        >Von *</label>
        <i
          v-tooltip.top="'Ab wann beginnt diese Transaktion?'"
          class="pi pi-info-circle"
        />
      </div>
      <DatePicker
        v-model="startDate"
        v-bind="startDateProps"
        date-format="dd.mm.yy"
        show-icon
        show-button-bar
        :class="{ 'p-invalid': errors['startDate']?.length }"
      />
      <small class="text-liqui-red">{{ errors["startDate"] || '&nbsp;' }}</small>
    </div>

    <div
      v-if="isRepeatingTransaction"
      class="flex flex-col gap-2 col-span-full md:col-span-1"
    >
      <div class="flex items-center gap-2">
        <label
          class="text-sm font-bold"
          for="vacation-days-per-year"
        >Bis</label>
        <i
          v-tooltip.top="'(Optional) Bis wann geht diese Transaktion?'"
          class="pi pi-info-circle"
        />
      </div>
      <DatePicker
        v-model="endDate"
        :min-date="startDate"
        v-bind="endDateProps"
        date-format="dd.mm.yy"
        show-icon
        show-button-bar
        :class="{ 'p-invalid': errors['endDate']?.length }"
      />
      <Message
        v-if="endDateErrorMessage.length"
        severity="warn"
        size="small"
        :sticky="false"
        :closable="true"
      >
        {{ endDateErrorMessage }}
      </Message>
      <small class="text-liqui-red">{{ errors["endDate"] || '&nbsp;' }}</small>
    </div>
    <span
      v-else
      class="hidden md:block"
    />

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
        :loading="isLoading"
        label="Abbrechen"
        severity="contrast"
        @click="dialogRef?.close()"
      />
      <Button
        v-if="!isCreate"
        :disabled="isLoading"
        severity="danger"
        size="small"
        icon="pi pi-trash"
        @click="onDeleteTransaction"
      />
    </div>
  </form>
</template>

<script setup lang="ts">
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import type { ITransactionFormDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'
import type { TransactionFormData } from '~/models/transaction'
import { CycleType, TransactionType } from '~/config/enums'
import AmountInvertButton from '~/components/AmountInvertButton.vue'
import { ModalConfig } from '~/config/dialog-props'
import VatDialog from '~/components/dialogs/VatDialog.vue'
import type { VatResponse } from '~/models/vat'
import { selectAllOnFocus } from '~/utils/element-helper'

const dialogRef = inject<ITransactionFormDialog>('dialogRef')!

const { getOrganisationCurrencyID, getOrganisationCurrencyCode, getOrganisationCurrencyLocaleCode } = useAuth()
const { createTransaction, updateTransaction, deleteTransaction } = useTransactions()
const { employees, listEmployees } = useEmployees()
const { vats, listVats, deleteVat } = useVat()
const { categories, currencies, getCurrencyLabel, convertAmountToRate } = useGlobalData()
const confirm = useConfirm()
const dialog = useDialog()
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
const endDateErrorMessage = ref('')

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
    currency: transaction?.currency.id ?? getOrganisationCurrencyID.value,
    employee: transaction?.employee?.id ?? null,
  } as TransactionFormData,
})

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

watch(startDate, (value) => {
  if (endDate.value && value > endDate.value) {
    endDate.value = undefined
    endDateErrorMessage.value = `Das "Bis" Datum wurde entfernt, da das "Von" Datum nach diesem liegt`
  }
})

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
        })
      })
      .finally(() => {
        isLoading.value = false
      })
  }
  else {
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
        })
      })
      .finally(() => {
        isLoading.value = false
      })
  }
})

const onParseAmount = (event: Event) => {
  if (event instanceof ClipboardEvent) {
    const pastedText = event.clipboardData?.getData('text') ?? ''
    const parsedAmount = parseCurrency(pastedText, true)
    amount.value = parsedAmount.length > 0 ? parseFloat(parsedAmount) : 0
  }
}

const onDeleteTransaction = () => {
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

const onCreateVat = () => {
  dialog.open(VatDialog, {
    props: {
      header: 'Neue Mehrwertsteuer anlegen',
      ...ModalConfig,
    },
    onClose: () => {
      if (options?.data) {
        setFieldValue('vat', options.data)
      }
    },
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
    onClose: () => {
      if (options?.data) {
        setFieldValue('vat', options.data)
      }
    },
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
    },
  })
}

const onInvertAmount = () => {
  amount.value *= -1
}

const isRepeatingTransaction = computed(() => {
  return type.value === TransactionType.Repeating
})
const selectedCurrencyCode = computed(() => currencies.value.find(c => c.id == currency.value)?.code)
const selectedLocalCode = computed(() => currencies.value.find(c => c.id == currency.value)?.localeCode)
const amountInBaseCurrency = computed(() => {
  let baseAmount = amount.value
  if (selectedCurrencyCode.value) {
    baseAmount = convertAmountToRate(amount.value, selectedCurrencyCode.value)
  }
  return `~ ${NumberToFormattedCurrency(baseAmount, getOrganisationCurrencyLocaleCode.value)} ${getOrganisationCurrencyCode.value}`
})
</script>
