import type {
  CheckRegistrationCodeFormData,
  CheckResetPasswordCodeFormData,
  FinishRegistrationFormData,
  ForgotPasswordFormData,
  LoginFormData,
  RegistrationFormData,
  ResetPasswordFormData,
  User,
  UserPasswordFormData,
  UserProfileFormData,
  UserUpdateOrganisationFormData,
  UserUpdateScenarioFormData,
} from '~/models/auth'
import { Constants } from '~/utils/constants'

export default function useAuth() {
  const user = useState<User | null>('user')
  const hasFetchedInitially = useState('hasFetchedInitially', () => false)

  const login = async (payload: LoginFormData): Promise<void> => {
    await $fetch('/api/auth/login', {
      method: 'POST',
      body: payload,
    })
  }

  const registration = async (payload: RegistrationFormData): Promise<void> => {
    await $fetch('/api/auth/registration/create', {
      method: 'POST',
      body: payload,
    })
  }

  const registrationCheckCode = async (payload: CheckRegistrationCodeFormData): Promise<boolean> => {
    try {
      await $fetch('/api/auth/registration/check-code', {
        method: 'POST',
        body: payload,
      })
      return true
    }
    catch {
      return false
    }
  }

  const registrationFinish = async (payload: FinishRegistrationFormData): Promise<boolean> => {
    try {
      await $fetch('/api/auth/registration/finish', {
        method: 'POST',
        body: payload,
      })
      return true
    }
    catch {
      return false
    }
  }

  const forgotPassword = async (payload: ForgotPasswordFormData): Promise<void> => {
    await $fetch('/api/auth/forgot-password', {
      method: 'POST',
      body: payload,
    })
  }

  const resetPassword = async (payload: ResetPasswordFormData): Promise<void> => {
    await $fetch('/api/auth/reset-password', {
      method: 'POST',
      body: payload,
    })
  }

  const resetPasswordCheckCode = async (payload: CheckResetPasswordCodeFormData): Promise<boolean> => {
    try {
      await $fetch('/api/auth/reset-password-check-code', {
        method: 'POST',
        body: payload,
      })
      return true
    }
    catch {
      return false
    }
  }

  const logout = async () => {
    try {
      await $fetch('/api/auth/logout', {
        method: 'GET',
      })
      user.value = null
    }
    catch (error) {
      console.error('Error logging out:', error)
    }
  }

  // Only used to regain the AccessToken in case it expires
  const getAccessToken = async () => {
    try {
      await $fetch('/api/access-token', {
        method: 'GET',
      })
    }
    catch (error) {
      console.error('Error getting access token:', error)
    }
  }

  const useFetchGetProfile = async () => {
    hasFetchedInitially.value = true
    const { data, error } = await useFetch('/api/profile', {
      method: 'GET',
      retry: false,
    })
    if (error.value) {
      console.error(error.value)
      return Promise.reject('Benutzer konnte nicht geladen werden')
    }
    user.value = data.value
  }

  const updateProfile = async (payload: UserProfileFormData) => {
    try {
      user.value = await $fetch<User>(`/api/profile`, {
        method: 'PATCH',
        body: payload,
      })
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren des Profils')
    }
  }

  const updatePassword = async (payload: UserPasswordFormData) => {
    try {
      await $fetch(`/api/profile/password`, {
        method: 'POST',
        body: payload,
      })
    }
    catch {
      return Promise.reject('Fehler beim Ändern des Password')
    }
  }

  const updateCurrentOrganisation = async (payload: UserUpdateOrganisationFormData) => {
    try {
      await $fetch<User>(`/api/profile/organisation`, {
        method: 'PATCH',
        body: payload,
      })
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren der Organisation')
    }
  }

  const updateCurrentScenario = async (payload: UserUpdateScenarioFormData) => {
    try {
      await $fetch<User>(`/api/profile/scenario`, {
        method: 'PATCH',
        body: payload,
      })
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren des Szenarios')
    }
  }

  const isAuthenticated = computed(() => !!user.value)

  const getOrganisationCurrencyID = computed(() => {
    return user.value?.currency.id ?? null
  })

  const getOrganisationCurrencyCode = computed(() => {
    return user.value?.currency.code ?? Constants.BASE_CURRENCY
  })

  const getOrganisationCurrencyLocaleCode = computed(() => {
    return user.value?.currency.localeCode ?? Constants.BASE_LOCALE_CODE
  })

  return {
    user,
    hasFetchedInitially,
    isAuthenticated,
    getOrganisationCurrencyID,
    getOrganisationCurrencyCode,
    getOrganisationCurrencyLocaleCode,
    login,
    registration,
    registrationCheckCode,
    registrationFinish,
    forgotPassword,
    resetPassword,
    resetPasswordCheckCode,
    logout,
    getAccessToken,
    useFetchGetProfile,
    updateProfile,
    updatePassword,
    updateCurrentOrganisation,
    updateCurrentScenario,
  }
}
