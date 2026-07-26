<template>
  <form
    id="bank-account-form"
    class="grid grid-cols-2 gap-2"
    @submit.prevent
  >
    <Message
      v-if="externalChange"
      severity="warn"
      class="col-span-full"
      :closable="false"
    >
      <div class="flex items-center justify-between gap-4 w-full">
        <span v-if="externalChange === 'deleted'">
          Dieses Bankkonto wurde soeben extern gelöscht.
        </span>
        <span v-else>
          Dieses Bankkonto wurde soeben extern geändert. Beim Speichern werden die externen Änderungen überschrieben.
        </span>
        <Button
          v-if="externalChange === 'updated'"
          label="Neu laden"
          icon="pi pi-refresh"
          size="small"
          severity="warn"
          @click="onReloadExternalChange"
        />
      </div>
    </Message>
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
        >Kontostand *</label>
        <i
          v-tooltip.top="'Negatives Vorzeichen möglich'"
          class="pi pi-info-circle text-liqui-blue"
        />
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

    <hr class="my-4 col-span-full">

    <Message
      v-if="errorMessage.length"
      severity="error"
      :closable="false"
      class="col-span-full"
    >
      {{ errorMessage }}
    </Message>

    <div class="flex items-center justify-end gap-2 col-span-full">
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
        outlined
        rounded
        icon="pi pi-trash"
        @click="onDeleteBankAccount"
      />
    </div>
  </form>
</template>

<script setup lang="ts">
import { useForm } from 'vee-validate'
import * as yup from 'yup'
import type { IBankAccountFormDialog } from '~/interfaces/dialog-interfaces'
import { Config } from '~/config/config'
import type { BankAccountFormData } from '~/models/bank-account'
import AmountInvertButton from '~/components/AmountInvertButton.vue'
import { selectAllOnFocus } from '~/utils/element-helper'

const dialogRef = inject<IBankAccountFormDialog>('dialogRef')!

const { getOrganisationCurrencyID } = useAuth()
const { getBankAccount, createBankAccount, updateBankAccount, deleteBankAccount } = useBankAccounts()
const { currencies, getCurrencyLabel } = useGlobalData()
const confirm = useConfirm()
const toast = useToast()

// Data
const isLoading = ref(false)
const bankAccount = ref(dialogRef.value.data?.bankAccount)
const isClone = dialogRef.value.data?.isClone
const isCreate = isClone || !bankAccount.value?.id
const errorMessage = ref('')

const { defineField, errors, handleSubmit, meta, resetForm } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
    amount: yup.number().required('Betrag wird benötigt').typeError('Ungültiger Betrag'),
    currency: yup.number().required('Währung wird benötigt').typeError('Ungültige Währung'),
  }),
  initialValues: {
    id: isClone ? undefined : bankAccount.value?.id ?? undefined,
    name: bankAccount.value?.name ?? '',
    amount: isNumber(bankAccount.value?.amount) ? AmountToFloat(bankAccount.value!.amount) : '',
    currency: bankAccount.value?.currency.id ?? getOrganisationCurrencyID.value,
  } as BankAccountFormData,
})

// Real-time: warn when the bank account being edited was changed or deleted
// externally (form values stay untouched, saving would overwrite)
const { useExternalChangeBanner } = useRealtimeChanges()
const externalChange = useExternalChangeBanner('bank_account', () => isCreate ? undefined : bankAccount.value?.id)

const onReloadExternalChange = async () => {
  if (!bankAccount.value?.id) return
  try {
    const fresh = await getBankAccount(bankAccount.value.id)
    bankAccount.value = fresh
    resetForm({
      values: {
        id: fresh.id,
        name: fresh.name ?? '',
        amount: isNumber(fresh.amount) ? AmountToFloat(fresh.amount) : '',
        currency: fresh.currency.id ?? getOrganisationCurrencyID.value,
      } as BankAccountFormData,
    })
    externalChange.value = null
  }
  catch {
    // Keep the banner; reloading failed
  }
}

const [name, nameProps] = defineField('name')
const [amount, amountProps] = defineField('amount')
const [currency, currencyProps] = defineField('currency')

const onParseAmount = (event: Event) => {
  if (event instanceof ClipboardEvent) {
    const pastedText = event.clipboardData?.getData('text') ?? ''
    const parsedAmount = parseCurrency(pastedText, true)
    amount.value = parsedAmount.length > 0 ? parseFloat(parsedAmount) : 0
  }
}

const onSubmit = handleSubmit((values) => {
  isLoading.value = true
  errorMessage.value = ''

  if (isCreate) {
    createBankAccount(values)
      .then(() => {
        dialogRef.value.close()
        toast.add({
          summary: 'Erfolg',
          detail: `Bankkonto "${values.name}" wurde angelegt`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
      .catch(() => {
        errorMessage.value = 'Bankkonto konnte nicht angelegt werden'
        nextTick(() => {
          scrollToParentBottom('bank-account-form')
        })
      })
      .finally(() => {
        isLoading.value = false
      })
  }
  else {
    updateBankAccount(values)
      .then(() => {
        dialogRef.value.close()
        toast.add({
          summary: 'Erfolg',
          detail: `Bankkonto "${values.name}" wurde bearbeitet`,
          severity: 'success',
          life: Config.TOAST_LIFE_TIME,
        })
      })
      .catch(() => {
        errorMessage.value = 'Bankkonto konnte nicht bearbeitet werden'
        nextTick(() => {
          scrollToParentBottom('bank-account-form')
        })
      })
      .finally(() => {
        isLoading.value = false
      })
  }
})

const onDeleteBankAccount = () => {
  confirm.require({
    header: 'Löschen',
    message: 'Bankkonto vollständig löschen?',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      if (bankAccount.value) {
        isLoading.value = true
        deleteBankAccount(bankAccount.value.id)
          .then(() => {
            toast.add({
              summary: 'Erfolg',
              detail: `Bankkonto "${bankAccount.value!.name}" wurde gelöscht`,
              severity: 'success',
              life: Config.TOAST_LIFE_TIME,
            })
            dialogRef.value.close()
          })
          .catch(() => {
            errorMessage.value = 'Bankkonto konnte nicht gelöscht werden'
            nextTick(() => {
              scrollToParentBottom('bank-account-form')
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

const selectedCurrencyCode = computed(() => currencies.value.find(c => c.id == currency.value)?.code)
const selectedLocalCode = computed(() => currencies.value.find(c => c.id == currency.value)?.localeCode)
</script>
