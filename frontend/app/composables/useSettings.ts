import {
  BankAccountSortByOptions,
  type BankAccountSortByType,
  type DisplayType,
  DisplayTypeOptions,
  EmployeeSortByOptions,
  type EmployeeSortByType,
  type GroupingType,
  SettingsTabOptions,
  type SettingsTabType,
  SortOrderOptions,
  type SortOrderType,
  TransactionSortByOptions,
  type TransactionSortByType,
} from '~/utils/types'
import { RouteNames } from '~/config/routes'
import type { UpdateUserOrganisationSetting } from '~/models/user-organisation-setting'
import type { UpdateUserSetting } from '~/models/user-setting'

export default function useSettings() {
  const { getUserSetting, updateUserSetting, userSetting } = useUserSettings()
  const { getUserOrganisationSetting, updateUserOrganisationSetting, userOrganisationSetting } = useUserOrganisationSettings()

  // Organisation-scoped settings (per-project)
  const forecastShowRevenueDetails = useState('forecastShowRevenueDetails', () => false)
  const forecastShowExpenseDetails = useState('forecastShowExpenseDetails', () => false)
  const forecastShowChildDetails = useState<string[]>('forecastShowChildDetails', () => [])
  const forecastPerformance = useState('forecastPerformance', () => 100)
  const forecastMonths = useState('forecastMonths', () => 13)

  const employeeDisplay = useState<DisplayType>('employeeDisplay', () => 'grid')
  const employeeSortBy = useState<EmployeeSortByType>('employeeSortBy', () => 'name')
  const employeeSortOrder = useState<SortOrderType>('employeeSortOrder', () => 'ASC')
  const employeeHideTerminated = useState<boolean>('employeeHideTerminated', () => true)

  const transactionDisplay = useState<DisplayType>('transactionDisplay', () => 'grid')
  const transactionSortBy = useState<TransactionSortByType>('transactionSortBy', () => 'name')
  const transactionSortOrder = useState<SortOrderType>('transactionSortOrder', () => 'ASC')
  const transactionHideDisabled = useState<boolean>('transactionHideDisabled', () => true)

  const bankAccountDisplay = useState<DisplayType>('bankAccountDisplay', () => 'grid')
  const bankAccountSortBy = useState<BankAccountSortByType>('bankAccountSortBy', () => 'name')
  const bankAccountSortOrder = useState<SortOrderType>('bankAccountSortOrder', () => 'ASC')

  // User-scoped settings (global)
  const settingsTab = useState<SettingsTabType>('settingsTab', () => RouteNames.SETTINGS_PROFILE)
  const skipOrganisationSwitchQuestion = useState<boolean>('skipOrganisationSwitchQuestion', () => false)

  // Deprecated settings kept for backward compatibility
  const costOverviewDisplay = useState<DisplayType>('costOverviewDisplay', () => 'grid')
  const costOverviewGrouping = useState<GroupingType>('costOverviewGrouping', () => 'ungrouped')

  // Track if settings have been loaded
  const settingsLoaded = useState('settingsLoaded', () => false)
  const organisationSettingsLoaded = useState('organisationSettingsLoaded', () => false)

  // Track if settings are currently being loaded (prevent concurrent calls)
  const settingsLoading = useState('settingsLoading', () => false)
  const organisationSettingsLoading = useState('organisationSettingsLoading', () => false)

  // Debounce timers
  let userSettingDebounce: ReturnType<typeof setTimeout> | null = null
  let orgSettingDebounce: ReturnType<typeof setTimeout> | null = null

  // Cookie migration helpers
  const migrateUserSettingsFromCookies = async () => {
    const settingsTabCookie = CreateSettingsCookie('settings-tab')
    const skipOrgCookie = CreateSettingsCookie('skip-organisation-switch-question')

    const payload: UpdateUserSetting = {}
    let hasCookies = false

    if (settingsTabCookie.value !== undefined && settingsTabCookie.value !== null) {
      const val = settingsTabCookie.value as string
      if (SettingsTabOptions.includes(val as SettingsTabType)) {
        payload.settingsTab = val
        hasCookies = true
      }
    }

    if (skipOrgCookie.value !== undefined && skipOrgCookie.value !== null) {
      const val = skipOrgCookie.value
      payload.skipOrganisationSwitchQuestion = typeof val === 'boolean' ? val : val === 'true'
      hasCookies = true
    }

    if (hasCookies) {
      await updateUserSetting(payload).catch(() => {})
      // Clear cookies after migration
      settingsTabCookie.value = null
      skipOrgCookie.value = null
    }
  }

  // Load settings from API
  const loadSettings = async () => {
    if (settingsLoaded.value || settingsLoading.value) return
    settingsLoading.value = true

    try {
      const setting = await getUserSetting()
      if (setting) {
        // Check if this is a fresh record (just created with defaults)
        const isFreshRecord = setting.settingsTab === 'settings/profile'
          && setting.skipOrganisationSwitchQuestion === false

        if (isFreshRecord) {
          // Migrate from cookies if they exist
          await migrateUserSettingsFromCookies()
          // Re-fetch to get migrated values
          const updatedSetting = await getUserSetting()
          if (updatedSetting) {
            if (SettingsTabOptions.includes(updatedSetting.settingsTab as SettingsTabType)) {
              settingsTab.value = updatedSetting.settingsTab as SettingsTabType
            }
            skipOrganisationSwitchQuestion.value = updatedSetting.skipOrganisationSwitchQuestion
          }
        }
        else {
          if (SettingsTabOptions.includes(setting.settingsTab as SettingsTabType)) {
            settingsTab.value = setting.settingsTab as SettingsTabType
          }
          skipOrganisationSwitchQuestion.value = setting.skipOrganisationSwitchQuestion
        }
      }
      // Wait for watchers to flush before marking as loaded
      await nextTick()
      settingsLoaded.value = true
    }
    catch {
      // Settings will use defaults
    }
    finally {
      settingsLoading.value = false
    }
  }

  const migrateOrgSettingsFromCookies = async () => {
    const cookies = {
      forecastMonths: CreateSettingsCookie('forecast-months'),
      forecastPerformance: CreateSettingsCookie('forecast-performance'),
      forecastRevenueDetails: CreateSettingsCookie('forecast-revenue-details'),
      forecastExpenseDetails: CreateSettingsCookie('forecast-expense-details'),
      forecastChildDetails: CreateSettingsCookie('forecast-child-details'),
      employeeDisplay: CreateSettingsCookie('employee-display'),
      employeeSortBy: CreateSettingsCookie('employee-sort-by'),
      employeeSortOrder: CreateSettingsCookie('employee-sort-order'),
      employeeHideTerminated: CreateSettingsCookie('employee-hide-terminated'),
      transactionDisplay: CreateSettingsCookie('transaction-display'),
      transactionSortBy: CreateSettingsCookie('transaction-sort-by'),
      transactionSortOrder: CreateSettingsCookie('transaction-sort-order'),
      transactionHideDisabled: CreateSettingsCookie('transaction-hide-disabled'),
      bankAccountDisplay: CreateSettingsCookie('bank-account-display'),
      bankAccountSortBy: CreateSettingsCookie('bank-account-sort-by'),
      bankAccountSortOrder: CreateSettingsCookie('bank-account-sort-order'),
    }

    const payload: UpdateUserOrganisationSetting = {}
    let hasCookies = false

    if (cookies.forecastMonths.value) {
      const val = Number.parseInt(cookies.forecastMonths.value as string)
      if (!Number.isNaN(val)) {
        payload.forecastMonths = val
        hasCookies = true
      }
    }

    if (cookies.forecastPerformance.value) {
      const val = Number.parseInt(cookies.forecastPerformance.value as string)
      if (!Number.isNaN(val)) {
        payload.forecastPerformance = val
        hasCookies = true
      }
    }

    if (cookies.forecastRevenueDetails.value !== undefined && cookies.forecastRevenueDetails.value !== null) {
      const val = cookies.forecastRevenueDetails.value
      payload.forecastRevenueDetails = typeof val === 'boolean' ? val : val === 'true'
      hasCookies = true
    }

    if (cookies.forecastExpenseDetails.value !== undefined && cookies.forecastExpenseDetails.value !== null) {
      const val = cookies.forecastExpenseDetails.value
      payload.forecastExpenseDetails = typeof val === 'boolean' ? val : val === 'true'
      hasCookies = true
    }

    if (cookies.forecastChildDetails.value) {
      try {
        const val = cookies.forecastChildDetails.value as unknown
        if (val instanceof Array) {
          payload.forecastChildDetails = val as string[]
          hasCookies = true
        }
      }
      catch { /* ignore */ }
    }

    if (cookies.employeeDisplay.value && DisplayTypeOptions.includes(cookies.employeeDisplay.value as DisplayType)) {
      payload.employeeDisplay = cookies.employeeDisplay.value as DisplayType
      hasCookies = true
    }

    if (cookies.employeeSortBy.value && EmployeeSortByOptions.includes(cookies.employeeSortBy.value as EmployeeSortByType)) {
      payload.employeeSortBy = cookies.employeeSortBy.value as string
      hasCookies = true
    }

    if (cookies.employeeSortOrder.value && SortOrderOptions.includes(cookies.employeeSortOrder.value as SortOrderType)) {
      payload.employeeSortOrder = cookies.employeeSortOrder.value as SortOrderType
      hasCookies = true
    }

    if (cookies.employeeHideTerminated.value !== undefined && cookies.employeeHideTerminated.value !== null) {
      const val = cookies.employeeHideTerminated.value
      payload.employeeHideTerminated = typeof val === 'boolean' ? val : val === 'true'
      hasCookies = true
    }

    if (cookies.transactionDisplay.value && DisplayTypeOptions.includes(cookies.transactionDisplay.value as DisplayType)) {
      payload.transactionDisplay = cookies.transactionDisplay.value as DisplayType
      hasCookies = true
    }

    if (cookies.transactionSortBy.value && TransactionSortByOptions.includes(cookies.transactionSortBy.value as TransactionSortByType)) {
      payload.transactionSortBy = cookies.transactionSortBy.value as string
      hasCookies = true
    }

    if (cookies.transactionSortOrder.value && SortOrderOptions.includes(cookies.transactionSortOrder.value as SortOrderType)) {
      payload.transactionSortOrder = cookies.transactionSortOrder.value as SortOrderType
      hasCookies = true
    }

    if (cookies.transactionHideDisabled.value !== undefined && cookies.transactionHideDisabled.value !== null) {
      const val = cookies.transactionHideDisabled.value
      payload.transactionHideDisabled = typeof val === 'boolean' ? val : val === 'true'
      hasCookies = true
    }

    if (cookies.bankAccountDisplay.value && DisplayTypeOptions.includes(cookies.bankAccountDisplay.value as DisplayType)) {
      payload.bankAccountDisplay = cookies.bankAccountDisplay.value as DisplayType
      hasCookies = true
    }

    if (cookies.bankAccountSortBy.value && BankAccountSortByOptions.includes(cookies.bankAccountSortBy.value as BankAccountSortByType)) {
      payload.bankAccountSortBy = cookies.bankAccountSortBy.value as string
      hasCookies = true
    }

    if (cookies.bankAccountSortOrder.value && SortOrderOptions.includes(cookies.bankAccountSortOrder.value as SortOrderType)) {
      payload.bankAccountSortOrder = cookies.bankAccountSortOrder.value as SortOrderType
      hasCookies = true
    }

    if (hasCookies) {
      await updateUserOrganisationSetting(payload).catch(() => {})
      // Clear cookies after migration
      Object.values(cookies).forEach((cookie) => {
        cookie.value = null
      })
    }
  }

  const applyOrganisationSettings = (setting: typeof userOrganisationSetting.value) => {
    if (!setting) return

    forecastMonths.value = setting.forecastMonths
    forecastPerformance.value = setting.forecastPerformance
    forecastShowRevenueDetails.value = setting.forecastRevenueDetails
    forecastShowExpenseDetails.value = setting.forecastExpenseDetails
    forecastShowChildDetails.value = setting.forecastChildDetails || []

    if (DisplayTypeOptions.includes(setting.employeeDisplay)) {
      employeeDisplay.value = setting.employeeDisplay
    }
    if (EmployeeSortByOptions.includes(setting.employeeSortBy as EmployeeSortByType)) {
      employeeSortBy.value = setting.employeeSortBy as EmployeeSortByType
    }
    if (SortOrderOptions.includes(setting.employeeSortOrder)) {
      employeeSortOrder.value = setting.employeeSortOrder
    }
    employeeHideTerminated.value = setting.employeeHideTerminated

    if (DisplayTypeOptions.includes(setting.transactionDisplay)) {
      transactionDisplay.value = setting.transactionDisplay
    }
    if (TransactionSortByOptions.includes(setting.transactionSortBy as TransactionSortByType)) {
      transactionSortBy.value = setting.transactionSortBy as TransactionSortByType
    }
    if (SortOrderOptions.includes(setting.transactionSortOrder)) {
      transactionSortOrder.value = setting.transactionSortOrder
    }
    transactionHideDisabled.value = setting.transactionHideDisabled

    if (DisplayTypeOptions.includes(setting.bankAccountDisplay)) {
      bankAccountDisplay.value = setting.bankAccountDisplay
    }
    if (BankAccountSortByOptions.includes(setting.bankAccountSortBy as BankAccountSortByType)) {
      bankAccountSortBy.value = setting.bankAccountSortBy as BankAccountSortByType
    }
    if (SortOrderOptions.includes(setting.bankAccountSortOrder)) {
      bankAccountSortOrder.value = setting.bankAccountSortOrder
    }
  }

  const loadOrganisationSettings = async () => {
    if (organisationSettingsLoaded.value || organisationSettingsLoading.value) return
    organisationSettingsLoading.value = true

    try {
      const setting = await getUserOrganisationSetting()
      if (setting) {
        // Check if this is a fresh record (just created with defaults)
        const isFreshRecord = setting.forecastMonths === 13
          && setting.forecastPerformance === 100
          && setting.employeeDisplay === 'grid'
          && setting.transactionDisplay === 'grid'
          && setting.bankAccountDisplay === 'grid'

        if (isFreshRecord) {
          // Migrate from cookies if they exist
          await migrateOrgSettingsFromCookies()
          // Re-fetch to get migrated values
          const updatedSetting = await getUserOrganisationSetting()
          applyOrganisationSettings(updatedSetting)
        }
        else {
          applyOrganisationSettings(setting)
        }
      }
      // Wait for watchers to flush before marking as loaded
      // This prevents watchers from triggering saves during initial load
      await nextTick()
      organisationSettingsLoaded.value = true
    }
    catch {
      // Settings will use defaults
    }
    finally {
      organisationSettingsLoading.value = false
    }
  }

  // Immediate save functions (for discrete actions like sorting, toggles)
  const saveUserSettingNow = (payload: UpdateUserSetting) => {
    if (userSettingDebounce) clearTimeout(userSettingDebounce)
    updateUserSetting(payload).catch(() => {})
  }

  const saveOrganisationSettingNow = (payload: UpdateUserOrganisationSetting) => {
    if (orgSettingDebounce) clearTimeout(orgSettingDebounce)
    updateUserOrganisationSetting(payload).catch(() => {})
  }

  // Debounced save functions (for continuous inputs like sliders)
  const _saveUserSettingDebounced = (payload: UpdateUserSetting) => {
    if (userSettingDebounce) clearTimeout(userSettingDebounce)
    userSettingDebounce = setTimeout(() => {
      updateUserSetting(payload).catch(() => {})
    }, 500)
  }

  const saveOrganisationSettingDebounced = (payload: UpdateUserOrganisationSetting) => {
    if (orgSettingDebounce) clearTimeout(orgSettingDebounce)
    orgSettingDebounce = setTimeout(() => {
      updateUserOrganisationSetting(payload).catch(() => {})
    }, 500)
  }

  // Toggle functions (use immediate save)
  const toggleTransactionDisplayType = () => {
    transactionDisplay.value = transactionDisplay.value === 'grid' ? 'list' : 'grid'
    saveOrganisationSettingNow({ transactionDisplay: transactionDisplay.value })
  }

  const toggleEmployeeHideTerminated = () => {
    employeeHideTerminated.value = !employeeHideTerminated.value
    saveOrganisationSettingNow({ employeeHideTerminated: employeeHideTerminated.value })
  }

  const toggleTransactionHideDisabled = () => {
    transactionHideDisabled.value = !transactionHideDisabled.value
    saveOrganisationSettingNow({ transactionHideDisabled: transactionHideDisabled.value })
  }

  const toggleEmployeeDisplayType = () => {
    employeeDisplay.value = employeeDisplay.value === 'grid' ? 'list' : 'grid'
    saveOrganisationSettingNow({ employeeDisplay: employeeDisplay.value })
  }

  const toggleBankAccountDisplayType = () => {
    bankAccountDisplay.value = bankAccountDisplay.value === 'grid' ? 'list' : 'grid'
    saveOrganisationSettingNow({ bankAccountDisplay: bankAccountDisplay.value })
  }

  // Deprecated toggle functions kept for backward compatibility
  const toggleCostOverviewDisplayType = () => {
    costOverviewDisplay.value = costOverviewDisplay.value === 'grid' ? 'list' : 'grid'
  }

  const toggleCostOverviewGroupingType = () => {
    costOverviewGrouping.value = costOverviewGrouping.value === 'ungrouped' ? 'grouped' : 'ungrouped'
  }

  // Explicit setters that save to API (replaces watchers which caused duplicate calls)
  const setTransactionSort = (sortBy: TransactionSortByType, sortOrder: SortOrderType) => {
    transactionSortBy.value = sortBy
    transactionSortOrder.value = sortOrder
    if (organisationSettingsLoaded.value) {
      saveOrganisationSettingNow({ transactionSortBy: sortBy, transactionSortOrder: sortOrder })
    }
  }

  const setEmployeeSort = (sortBy: EmployeeSortByType, sortOrder: SortOrderType) => {
    employeeSortBy.value = sortBy
    employeeSortOrder.value = sortOrder
    if (organisationSettingsLoaded.value) {
      saveOrganisationSettingNow({ employeeSortBy: sortBy, employeeSortOrder: sortOrder })
    }
  }

  const setBankAccountSort = (sortBy: BankAccountSortByType, sortOrder: SortOrderType) => {
    bankAccountSortBy.value = sortBy
    bankAccountSortOrder.value = sortOrder
    if (organisationSettingsLoaded.value) {
      saveOrganisationSettingNow({ bankAccountSortBy: sortBy, bankAccountSortOrder: sortOrder })
    }
  }

  const setForecastPerformance = (value: number) => {
    forecastPerformance.value = value
    if (organisationSettingsLoaded.value) {
      saveOrganisationSettingDebounced({ forecastPerformance: value })
    }
  }

  const setForecastMonths = (value: number) => {
    forecastMonths.value = value
    if (organisationSettingsLoaded.value) {
      saveOrganisationSettingNow({ forecastMonths: value })
    }
  }

  const setForecastShowRevenueDetails = (value: boolean) => {
    forecastShowRevenueDetails.value = value
    if (organisationSettingsLoaded.value) {
      saveOrganisationSettingNow({ forecastRevenueDetails: value })
    }
  }

  const setForecastShowExpenseDetails = (value: boolean) => {
    forecastShowExpenseDetails.value = value
    if (organisationSettingsLoaded.value) {
      saveOrganisationSettingNow({ forecastExpenseDetails: value })
    }
  }

  const setForecastShowChildDetails = (value: string[]) => {
    forecastShowChildDetails.value = value
    if (organisationSettingsLoaded.value) {
      saveOrganisationSettingNow({ forecastChildDetails: value })
    }
  }

  const setSettingsTab = (value: SettingsTabType) => {
    settingsTab.value = value
    if (settingsLoaded.value && SettingsTabOptions.includes(value)) {
      saveUserSettingNow({ settingsTab: value })
    }
  }

  const setSkipOrganisationSwitchQuestion = (value: boolean) => {
    skipOrganisationSwitchQuestion.value = value
    if (settingsLoaded.value) {
      saveUserSettingNow({ skipOrganisationSwitchQuestion: value })
    }
  }

  return {
    // Organisation settings
    forecastShowRevenueDetails,
    forecastShowExpenseDetails,
    forecastShowChildDetails,
    forecastPerformance,
    forecastMonths,
    employeeDisplay,
    employeeSortBy,
    employeeSortOrder,
    transactionDisplay,
    transactionSortBy,
    transactionSortOrder,
    transactionHideDisabled,
    employeeHideTerminated,
    bankAccountDisplay,
    bankAccountSortBy,
    bankAccountSortOrder,
    // User settings
    settingsTab,
    skipOrganisationSwitchQuestion,
    // Setters (these save to API)
    setTransactionSort,
    setEmployeeSort,
    setBankAccountSort,
    setForecastPerformance,
    setForecastMonths,
    setForecastShowRevenueDetails,
    setForecastShowExpenseDetails,
    setForecastShowChildDetails,
    setSettingsTab,
    setSkipOrganisationSwitchQuestion,
    // Toggle functions
    toggleEmployeeDisplayType,
    toggleTransactionDisplayType,
    toggleTransactionHideDisabled,
    toggleEmployeeHideTerminated,
    toggleBankAccountDisplayType,
    // Deprecated (kept for backward compatibility)
    toggleCostOverviewDisplayType,
    costOverviewDisplay,
    toggleCostOverviewGroupingType,
    costOverviewGrouping,
    // Load functions
    loadSettings,
    loadOrganisationSettings,
    // State
    settingsLoaded,
    organisationSettingsLoaded,
    userSetting,
    userOrganisationSetting,
  }
}
