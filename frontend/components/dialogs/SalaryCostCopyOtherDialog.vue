<template>
  <div class="flex flex-col gap-2">
    <Message
      v-if="employeesErrorMessage.length"
      severity="error"
      size="small"
      class="col-span-full"
      closable
    >
      {{ employeesErrorMessage }}
    </Message>

    <div class="flex flex-col gap-2">
      <p>Kopiere von Mitarbeiter:</p>
      <Select
        v-model="selectedSourceEmployeeID"
        :loading="isLoadingEmployees"
        :options="employeeOptions"
        option-label="name"
        option-value="id"
        placeholder="Bitte wählen"
      />
    </div>

    <div
      v-if="selectedSourceEmployeeID"
      class="flex flex-col gap-2"
    >
      <p>Kopiere von Lohn:</p>
      <Select
        v-model="selectedSourceSalaryID"
        :loading="isLoadingSourceSalaries"
        :options="sourceSalaryOptions"
        :option-label="(data) => SalaryUtils.title(data)"
        option-value="id"
        placeholder="Bitte wählen"
        :disabled="!sourceSalaryOptions.length"
      />
    </div>

    <Message
      v-if="salaryCostsErrorMessage.length"
      severity="error"
      size="small"
      class="col-span-full"
      closable
    >
      {{ salaryCostsErrorMessage }}
    </Message>

    <div
      v-if="isLoadingSalaryCosts"
      class="flex justify-center items-center"
    >
      <ProgressSpinner />
    </div>
    <div
      v-else-if="salaryCosts.length"
      class="flex flex-col overflow-x-auto pb-2"
    >
      <SalaryCostCopyHeaders v-model="selectAll" />
      <SalaryCostCopyRow
        v-for="cost of salaryCosts"
        :key="cost.id"
        :salary-cost="cost"
        :currency="salary.currency"
        :select-all="selectAll"
        @on-selection="onSalaryCostSelection"
      />
    </div>
    <div
      v-else-if="selectedSourceSalaryID"
      class="text-sm text-surface-500"
    >
      Keine Lohnkosten verfügbar
    </div>

    <hr class="my-2 col-span-full">

    <Message
      severity="warn"
      size="small"
      class="col-span-full"
    >
      Beim Kopieren werden vorhandene Lohnkosten dieses Lohns vollständig ersetzt.
    </Message>

    <div class="flex justify-end gap-2 col-span-full">
      <Button
        label="Kopieren"
        :disabled="!canCopy || isCopying"
        @click="onCopy"
      />
      <Button
        :disabled="isCopying"
        label="Schliessen"
        severity="contrast"
        @click="dialogRef.close()"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type {
  EmployeeResponse,
  ListEmployeeResponse,
  ListSalaryCostResponse,
  ListSalaryResponse,
  SalaryCostResponse,
  SalaryResponse,
} from '~/models/employee'
import type { ISalaryCostCopyDialog } from '~/interfaces/dialog-interfaces'
import { SalaryUtils } from '~/utils/models/salary-utils'
import { Config } from '~/config/config'
import SalaryCostCopyHeaders from '~/components/SalaryCostCopyHeaders.vue'

const dialogRef = inject<ISalaryCostCopyDialog>('dialogRef')!

const toast = useToast()
const confirm = useConfirm()
const { copySalaryCost } = useSalaryCosts()

const salary = ref(dialogRef.value.data!.salary)

const employees = ref<EmployeeResponse[]>([])
const isLoadingEmployees = ref(false)
const employeesErrorMessage = ref('')
const selectedSourceEmployeeID = ref<number | null>(null)

const sourceSalaries = ref<SalaryResponse[]>([])
const isLoadingSourceSalaries = ref(false)
const selectedSourceSalaryID = ref<number | null>(null)

const salaryCosts = ref<SalaryCostResponse[]>([])
const selectedSalaryCosts = ref<SalaryCostResponse[]>([])
const selectAll = ref(true)
const isLoadingSalaryCosts = ref(false)
const salaryCostsErrorMessage = ref('')
const isCopying = ref(false)

const employeeOptions = computed(() => employees.value.filter(emp => emp.id !== salary.value.employeeID))

const sourceSalaryOptions = computed(() => sourceSalaries.value)
const canCopy = computed(() => selectedSalaryCosts.value.length > 0 && !!selectedSourceSalaryID.value)

const fetchEmployees = async () => {
  isLoadingEmployees.value = true
  employeesErrorMessage.value = ''
  try {
    const data = await $fetch<ListEmployeeResponse>('/api/employees', {
      method: 'GET',
      query: {
        page: 1,
        limit: 200,
      },
    })
    employees.value = data.data?.filter(emp => emp.id !== salary.value.employeeID) ?? []
  }
  catch {
    employeesErrorMessage.value = 'Es gab einen Fehler beim Laden der Mitarbeiter'
  }
  finally {
    isLoadingEmployees.value = false
  }
}

const fetchSalaries = async (employeeID: number) => {
  const data = await $fetch<ListSalaryResponse>(`/api/employees/${employeeID}/salary`, {
    method: 'GET',
    query: {
      page: 1,
      limit: 100,
    },
  })
  return data.data ?? []
}

const fetchSalaryCosts = async (salaryID: number) => {
  const data = await $fetch<ListSalaryCostResponse>(`/api/employees/salary/${salaryID}/costs`, {
    method: 'GET',
    query: {
      page: 1,
      limit: 100,
    },
  })
  return data.data ?? []
}

const loadSourceSalaries = async (employeeID: number) => {
  selectedSourceSalaryID.value = null
  salaryCosts.value = []
  selectedSalaryCosts.value = []
  salaryCostsErrorMessage.value = ''
  selectAll.value = true
  isLoadingSourceSalaries.value = true
  try {
    sourceSalaries.value = await fetchSalaries(employeeID)
  }
  catch {
    sourceSalaries.value = []
    salaryCostsErrorMessage.value = 'Es gab einen Fehler beim Laden der Lohnkosten'
  }
  finally {
    isLoadingSourceSalaries.value = false
  }
}

const loadSalaryCosts = async (salaryID: number) => {
  isLoadingSalaryCosts.value = true
  salaryCostsErrorMessage.value = ''
  try {
    salaryCosts.value = await fetchSalaryCosts(salaryID)
    if (selectAll.value) {
      selectedSalaryCosts.value = [...salaryCosts.value]
    }
  }
  catch {
    salaryCosts.value = []
    selectedSalaryCosts.value = []
    salaryCostsErrorMessage.value = 'Es gab einen Fehler beim Laden der Lohnkosten'
  }
  finally {
    isLoadingSalaryCosts.value = false
  }
}

watch(selectAll, (value) => {
  if (value) {
    selectedSalaryCosts.value = [...salaryCosts.value]
  }
  else {
    selectedSalaryCosts.value = []
  }
})

watch(salaryCosts, (newCosts) => {
  if (selectAll.value) {
    selectedSalaryCosts.value = [...newCosts]
  }
})

watch(selectedSourceEmployeeID, async (employeeID) => {
  if (!employeeID) {
    sourceSalaries.value = []
    salaryCosts.value = []
    selectedSalaryCosts.value = []
    return
  }
  await loadSourceSalaries(employeeID)
})

watch(selectedSourceSalaryID, async (salaryID) => {
  if (!salaryID) {
    salaryCosts.value = []
    selectedSalaryCosts.value = []
    return
  }
  await loadSalaryCosts(salaryID)
})

onMounted(async () => {
  await fetchEmployees()
})

const performCopy = async () => {
  if (!selectedSourceSalaryID.value) {
    return
  }
  isCopying.value = true
  const idsToCopy = selectedSalaryCosts.value.map(cost => cost.id)
  copySalaryCost(salary.value.id, {
    ids: idsToCopy,
    sourceSalaryID: selectedSourceSalaryID.value,
  })
    .then(() => {
      dialogRef.value.close(true)
      toast.add({
        summary: 'Erfolg',
        detail: 'Lohnkosten wurden erfolgreich kopiert',
        severity: 'success',
        life: Config.TOAST_LIFE_TIME,
      })
    })
    .catch(() => {
      toast.add({
        summary: 'Fehler',
        detail: 'Konnte Lohnkosten nicht kopieren',
        severity: 'error',
        life: Config.TOAST_LIFE_TIME,
      })
    })
    .finally(() => {
      isCopying.value = false
    })
}

const onCopy = () => {
  if (!selectedSourceSalaryID.value || selectedSalaryCosts.value.length === 0) {
    return
  }
  confirm.require({
    header: 'Lohnkosten ersetzen',
    message: 'Bestehende Lohnkosten dieses Lohns werden vollständig ersetzt. Fortfahren?',
    icon: 'pi pi-exclamation-triangle',
    rejectLabel: 'Nein',
    acceptLabel: 'Ja',
    accept: performCopy,
  })
}

const onSalaryCostSelection = (salaryCost: SalaryCostResponse, isSelected: boolean) => {
  if (isSelected) {
    const exists = selectedSalaryCosts.value.find(cost => cost.id === salaryCost.id)
    if (!exists) {
      selectedSalaryCosts.value.push(salaryCost)
    }
    if (selectedSalaryCosts.value.length === salaryCosts.value.length) {
      selectAll.value = true
    }
  }
  else {
    selectedSalaryCosts.value = selectedSalaryCosts.value.filter(cost => cost.id !== salaryCost.id)
    selectAll.value = false
  }
}
</script>
