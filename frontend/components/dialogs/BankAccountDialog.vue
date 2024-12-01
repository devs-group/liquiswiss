<template>
  <form @submit.prevent id="bank-account-form" class="grid grid-cols-2 gap-2">
    <div class="flex flex-col gap-2 col-span-full">
      <label class="text-sm font-bold" for="name">Name *</label>
      <InputText v-model="name" v-bind="nameProps"
                 :class="{'p-invalid': errors['name']?.length}"
                 id="name" type="text"/>
      <small class="text-liqui-red">{{errors["name"] || '&nbsp;'}}</small>
    </div>

    <div class="flex flex-col gap-2 col-span-full md:col-span-1">
      <div class="flex items-center gap-2">
        <label class="text-sm font-bold" for="name">Kontostand *</label>
        <i class="pi pi-info-circle text-liqui-blue" v-tooltip="'Negatives Vorzeichen möglich'"></i>
      </div>
      <InputText v-model="amount" v-bind="amountProps"
                 @input="onParseAmount"
                 :class="{'p-invalid': errors['amount']?.length}"
                 id="name" type="text"/>
      <small class="text-liqui-red">{{errors["amount"] || '&nbsp;'}}</small>
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

    <div v-if="!isCreate" class="flex justify-end col-span-full">
      <Button @click="onDeleteBankAccount" :loading="isLoading" label="Löschen" icon="pi pi-trash" severity="danger" size="small"/>
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
import type {IBankAccountFormDialog} from "~/interfaces/dialog-interfaces";
import {useForm} from "vee-validate";
import * as yup from 'yup';
import {Config} from "~/config/config";
import type {BankAccountFormData} from "~/models/bank-account";
import {parseNumberInput} from "~/utils/element-helper";

const dialogRef = inject<IBankAccountFormDialog>('dialogRef')!;

const {createBankAccount, updateBankAccount, deleteBankAccount} = useBankAccounts()
const {currencies} = useGlobalData()
const confirm = useConfirm()
const toast = useToast()

// Data
const isLoading = ref(false)
const bankAccount = dialogRef.value.data?.bankAccount
const isCreate = !bankAccount?.id
const errorMessage = ref('')

const { defineField, errors, handleSubmit, meta } = useForm({
  validationSchema: yup.object({
    name: yup.string().trim().required('Name wird benötigt'),
    amount: yup.number().required('Betrag wird benötigt').typeError('Ungültiger Betrag')
        .test('Not 0', 'Muss grösser oder kleiner 0 sein', (value) => {
          return AmountToInteger(value) !== 0;
        }),
    currency: yup.number().required('Währung wird benötigt').typeError('Ungültige Währung'),
  }),
  initialValues: {
    id: bankAccount?.id ?? undefined,
    name: bankAccount?.name ?? '',
    amount: bankAccount?.amount ? AmountToFloat(bankAccount.amount) : '',
    currency: bankAccount?.currency.id ?? null,
  } as BankAccountFormData
});

const [name, nameProps] = defineField('name')
const [amount, amountProps] = defineField('amount')
const [currency, currencyProps] = defineField('currency')

const onParseAmount = (event: Event) => {
  if (event instanceof InputEvent) {
    parseNumberInput(event, amount)
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
          });
        })
        .finally(() => {
          isLoading.value = false
        })
  } else {
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
          });
        })
        .finally(() => {
          isLoading.value = false
        })
  }
})

const onDeleteBankAccount = (event: MouseEvent) => {
  confirm.require({
    header: 'Löschen',
    message: 'Bankkonto vollständig löschen?',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: () => {
      if (bankAccount) {
        isLoading.value = true
        deleteBankAccount(bankAccount.id)
            .then(() => {
              toast.add({
                summary: 'Erfolg',
                detail: `Bankkonto "${bankAccount.name}" wurde gelöscht`,
                severity: 'success',
                life: Config.TOAST_LIFE_TIME,
              })
              dialogRef.value.close()
            })
            .catch(() => {
              errorMessage.value = 'Bankkonto konnte nicht gelöscht werden'
              nextTick(() => {
                scrollToParentBottom('bank-account-form')
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
