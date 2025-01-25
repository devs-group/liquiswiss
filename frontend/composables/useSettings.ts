import {
  BankAccountSortByOptions,
  type BankAccountSortByType,
  type DisplayType,
  DisplayTypeOptions,
  GroupingTypeOptions,
  SettingsTabOptions,
  type SettingsTabType,
  SortOrderOptions,
  type SortOrderType,
  TransactionSortByOptions,
  type TransactionSortByType,
} from '~/utils/types'
import { RouteNames } from '~/config/routes'

export default function useSettings() {
  const forecastShowRevenueDetails = useState('forecastShowRevenueDetails', () => false)
  const forecastShowExpenseDetails = useState('forecastShowExpenseDetails', () => false)
  const forecastShowChildDetails = useState('forecastShowChildDetails', () => false)
  const forecastPerformance = useState('forecastPerformance', () => 100)
  // 13 to include from current to the same month
  const forecastMonths = useState('forecastMonths', () => 13)
  const forecastShowRevenueDetailsCookie = CreateSettingsCookie('forecast-revenue-details')
  const forecastShowExpenseDetailsCookie = CreateSettingsCookie('forecast-expense-details')
  const forecastShowChildDetailsCookie = CreateSettingsCookie('forecast-child-details')
  const forecastPerformanceCookie = CreateSettingsCookie('forecast-performance')
  const forecastMonthsCookie = CreateSettingsCookie('forecast-months')

  const employeeDisplay = useState<DisplayType>('employeeDisplay', () => 'grid')
  const employeeSortBy = useState<EmployeeSortByType>('employeeSortBy', () => 'name')
  const employeeSortOrder = useState<SortOrderType>('employeeSortOrder', () => 'ASC')
  const employeeDisplayCookie = CreateSettingsCookie('employee-display')
  const employeeSortByCookie = CreateSettingsCookie('employee-sort-by')
  const employeeSortOrderCookie = CreateSettingsCookie('employee-sort-order')

  const transactionDisplay = useState<DisplayType>('transactionDisplay', () => 'grid')
  const transactionSortBy = useState<TransactionSortByType>('transactionSortBy', () => 'name')
  const transactionSortOrder = useState<SortOrderType>('transactionSortOrder', () => 'ASC')
  const transactionDisplayCookie = CreateSettingsCookie('transaction-display')
  const transactionSortByCookie = CreateSettingsCookie('transaction-sort-by')
  const transactionSortOrderCookie = CreateSettingsCookie('transaction-sort-order')

  const bankAccountDisplay = useState<DisplayType>('bankAccountDisplay', () => 'grid')
  const bankAccountSortBy = useState<BankAccountSortByType>('bankAccountSortBy', () => 'name')
  const bankAccountSortOrder = useState<SortOrderType>('bankAccountSortOrder', () => 'ASC')
  const bankAccountDisplayCookie = CreateSettingsCookie('bank-account-display')
  const bankAccountSortByCookie = CreateSettingsCookie('bank-account-sort-by')
  const bankAccountSortOrderCookie = CreateSettingsCookie('bank-account-sort-order')

  const costOverviewDisplay = useState<DisplayType>('costOverviewDisplay', () => 'grid')
  const costOverviewDisplayCookie = CreateSettingsCookie('cost-overview-display')
  const costOverviewGrouping = useState<GroupingType>('costOverviewGrouping', () => 'ungrouped')
  const costOverviewGroupingCookie = CreateSettingsCookie('cost-overview-grouping')

  const settingsTab = useState<SettingsTabType>('settingsTab', () => RouteNames.SETTINGS_PROFILE)
  const settingsTabCookie = CreateSettingsCookie('settings-tab')

  const skipOrganisationSwitchQuestion = useState<boolean>('skipOrganisationSwitchQuestion', () => false)
  const skipOrganisationSwitchQuestionCookie = CreateSettingsCookie('skip-organisation-switch-question')

  if (forecastShowRevenueDetailsCookie.value !== undefined) {
    const val = forecastShowRevenueDetailsCookie.value
    if (val == 'true') {
      forecastShowRevenueDetails.value = true
    }
    else if (val == 'false') {
      forecastShowRevenueDetails.value = false
    }
    else if (typeof val === 'boolean') {
      // Can be boolean for some reason
      forecastShowRevenueDetails.value = val
    }
  }
  else {
    forecastShowRevenueDetails.value = false
  }

  if (forecastShowExpenseDetailsCookie.value !== undefined) {
    const val = forecastShowExpenseDetailsCookie.value
    if (val == 'true') {
      forecastShowExpenseDetails.value = true
    }
    else if (val == 'false') {
      forecastShowExpenseDetails.value = false
    }
    else if (typeof val === 'boolean') {
      // Can be boolean for some reason
      forecastShowExpenseDetails.value = val
    }
  }
  else {
    forecastShowExpenseDetails.value = false
  }

  if (forecastShowChildDetailsCookie.value !== undefined) {
    const val = forecastShowChildDetailsCookie.value
    if (val == 'true') {
      forecastShowChildDetails.value = true
    }
    else if (val == 'false') {
      forecastShowChildDetails.value = false
    }
    else if (typeof val === 'boolean') {
      // Can be boolean for some reason
      forecastShowChildDetails.value = val
    }
  }
  else {
    forecastShowChildDetails.value = true
  }

  if (employeeDisplayCookie.value !== undefined) {
    const val = employeeDisplayCookie.value as DisplayType
    if (val !== null && DisplayTypeOptions.includes(val)) {
      employeeDisplay.value = val
    }
  }
  else {
    employeeDisplay.value = 'grid'
  }

  if (employeeSortByCookie.value !== undefined) {
    const val = employeeSortByCookie.value as typeof employeeSortBy.value
    if (EmployeeSortByOptions.includes(val)) {
      employeeSortBy.value = val
    }
  }
  else {
    employeeSortBy.value = 'name'
  }

  if (employeeSortOrderCookie.value !== undefined) {
    const val = employeeSortOrderCookie.value as typeof employeeSortOrder.value
    if (SortOrderOptions.includes(val)) {
      employeeSortOrder.value = val
    }
  }
  else {
    employeeSortOrder.value = 'ASC'
  }

  if (transactionDisplayCookie.value !== undefined) {
    const val = transactionDisplayCookie.value as DisplayType
    if (val !== null && DisplayTypeOptions.includes(val)) {
      transactionDisplay.value = val
    }
  }
  else {
    transactionDisplay.value = 'grid'
  }

  if (transactionSortByCookie.value !== undefined) {
    const val = transactionSortByCookie.value as typeof transactionSortBy.value
    if (TransactionSortByOptions.includes(val)) {
      transactionSortBy.value = val
    }
  }
  else {
    transactionSortBy.value = 'name'
  }

  if (transactionSortOrderCookie.value !== undefined) {
    const val = transactionSortOrderCookie.value as typeof transactionSortOrder.value
    if (SortOrderOptions.includes(val)) {
      transactionSortOrder.value = val
    }
  }
  else {
    transactionSortOrder.value = 'ASC'
  }

  if (forecastPerformanceCookie.value !== undefined) {
    const val = forecastPerformanceCookie.value
    if (val !== null && Number.isInteger(Number.parseInt(val))) {
      forecastPerformance.value = Number.parseInt(val)
    }
  }
  else {
    forecastPerformance.value = 100
  }

  if (forecastMonthsCookie.value !== undefined) {
    const val = forecastMonthsCookie.value
    if (val !== null && Number.isInteger(Number.parseInt(val))) {
      forecastMonths.value = Number.parseInt(val)
    }
  }
  else {
    // 13 to include from current to the same month
    forecastMonths.value = 13
  }

  if (bankAccountDisplayCookie.value !== undefined) {
    const val = bankAccountDisplayCookie.value as DisplayType
    if (val !== null && DisplayTypeOptions.includes(val)) {
      bankAccountDisplay.value = val
    }
  }
  else {
    bankAccountDisplay.value = 'grid'
  }

  if (bankAccountSortByCookie.value !== undefined) {
    const val = bankAccountSortByCookie.value as typeof bankAccountSortBy.value
    if (BankAccountSortByOptions.includes(val)) {
      bankAccountSortBy.value = val
    }
  }
  else {
    bankAccountSortBy.value = 'name'
  }

  if (bankAccountSortOrderCookie.value !== undefined) {
    const val = bankAccountSortOrderCookie.value as typeof bankAccountSortOrder.value
    if (SortOrderOptions.includes(val)) {
      bankAccountSortOrder.value = val
    }
  }
  else {
    bankAccountSortOrder.value = 'ASC'
  }

  if (costOverviewDisplayCookie.value !== undefined) {
    const val = costOverviewDisplayCookie.value as DisplayType
    if (val !== null && DisplayTypeOptions.includes(val)) {
      costOverviewDisplay.value = val
    }
  }
  else {
    costOverviewDisplay.value = 'grid'
  }

  if (costOverviewGroupingCookie.value !== undefined) {
    const val = costOverviewGroupingCookie.value as GroupingType
    if (val !== null && GroupingTypeOptions.includes(val)) {
      costOverviewGrouping.value = val
    }
  }
  else {
    costOverviewGrouping.value = 'ungrouped'
  }

  if (settingsTabCookie.value !== undefined) {
    const val = settingsTabCookie.value
    if (val !== null && SettingsTabOptions.includes(val)) {
      settingsTab.value = val
    }
  }
  else {
    settingsTab.value = RouteNames.SETTINGS_PROFILE
  }

  if (skipOrganisationSwitchQuestionCookie.value !== undefined) {
    const val = skipOrganisationSwitchQuestionCookie.value
    if (val !== null) {
      if (typeof val === 'boolean') {
        skipOrganisationSwitchQuestion.value = val
      }
      else {
        skipOrganisationSwitchQuestion.value = val === 'true'
      }
    }
  }
  else {
    skipOrganisationSwitchQuestion.value = false
  }

  const toggleTransactionDisplayType = () => {
    switch (transactionDisplay.value) {
      case 'grid':
        transactionDisplay.value = 'list'
        break
      case 'list':
        transactionDisplay.value = 'grid'
    }
    transactionDisplayCookie.value = transactionDisplay.value
  }

  const toggleEmployeeDisplayType = () => {
    switch (employeeDisplay.value) {
      case 'grid':
        employeeDisplay.value = 'list'
        break
      case 'list':
        employeeDisplay.value = 'grid'
    }
    employeeDisplayCookie.value = employeeDisplay.value
  }

  const toggleBankAccountDisplayType = () => {
    switch (bankAccountDisplay.value) {
      case 'grid':
        bankAccountDisplay.value = 'list'
        break
      case 'list':
        bankAccountDisplay.value = 'grid'
    }
    bankAccountDisplayCookie.value = bankAccountDisplay.value
  }

  const toggleCostOverviewDisplayType = () => {
    switch (costOverviewDisplay.value) {
      case 'grid':
        costOverviewDisplay.value = 'list'
        break
      case 'list':
        costOverviewDisplay.value = 'grid'
    }
    costOverviewDisplayCookie.value = costOverviewDisplay.value
  }

  const toggleCostOverviewGroupingType = () => {
    switch (costOverviewGrouping.value) {
      case 'ungrouped':
        costOverviewGrouping.value = 'grouped'
        break
      case 'grouped':
        costOverviewGrouping.value = 'ungrouped'
    }
    costOverviewGroupingCookie.value = costOverviewGrouping.value
  }

  // Watchers
  watch(forecastShowRevenueDetails, (value) => {
    forecastShowRevenueDetailsCookie.value = value.toString()
  })

  watch(forecastShowExpenseDetails, (value) => {
    forecastShowExpenseDetailsCookie.value = value.toString()
  })

  watch(forecastShowChildDetails, (value) => {
    forecastShowChildDetailsCookie.value = value.toString()
  })

  watch(forecastPerformance, (value) => {
    forecastPerformanceCookie.value = value.toString()
  })

  watch(employeeSortBy, (value) => {
    employeeSortByCookie.value = value.toString()
  })

  watch(employeeSortOrder, (value) => {
    employeeSortOrderCookie.value = value.toString()
  })

  watch(transactionSortBy, (value) => {
    transactionSortByCookie.value = value.toString()
  })

  watch(transactionSortOrder, (value) => {
    transactionSortOrderCookie.value = value.toString()
  })

  watch(forecastMonths, (value) => {
    forecastMonthsCookie.value = value.toString()
  })

  watch(settingsTab, (value) => {
    if (value !== null && SettingsTabOptions.includes(value)) {
      settingsTabCookie.value = value.toString()
    }
  })

  watch(skipOrganisationSwitchQuestion, (value) => {
    if (value !== null) {
      skipOrganisationSwitchQuestionCookie.value = value.toString()
    }
  })

  return {
    forecastShowRevenueDetails,
    forecastShowExpenseDetails,
    forecastShowChildDetails,
    forecastPerformance,
    forecastMonths,
    toggleEmployeeDisplayType,
    employeeDisplay,
    employeeSortBy,
    employeeSortOrder,
    toggleTransactionDisplayType,
    transactionDisplay,
    transactionSortBy,
    transactionSortOrder,
    toggleBankAccountDisplayType,
    bankAccountDisplay,
    bankAccountSortBy,
    bankAccountSortOrder,
    toggleCostOverviewDisplayType,
    costOverviewDisplay,
    toggleCostOverviewGroupingType,
    costOverviewGrouping,
    settingsTab,
    skipOrganisationSwitchQuestion,
  }
}
