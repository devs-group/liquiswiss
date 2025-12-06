import type { DarkModeType } from '~/utils/types'
import type { CurrencyResponse } from '~/models/currency'

export interface User {
  id: number
  name: string
  email: string
  currentOrganisationID: number
  currentScenarioID: number
  currency: CurrencyResponse
}

export interface RegistrationFormData {
  email: string
}

export interface ForgotPasswordFormData {
  email: string
}

export interface CheckResetPasswordCodeFormData {
  email: string
  code: string
}

export interface ResetPasswordFormData {
  email: string
  code: string
  password: string
}

export interface LoginFormData {
  email: string
  password: string
}

export interface CheckRegistrationCodeFormData {
  email: string
  code: string
}

export interface FinishRegistrationFormData {
  email: string
  code: string
  password: string
}

export interface Login {
  email: string
  password: string
}

export interface UserProfileFormData {
  id: number
  name: string
  email: string
}

export interface UserPasswordFormData {
  password: string
  passwordRepeat: string
}

export interface UserUpdateOrganisationFormData {
  organisationId: number
}
export interface UserUpdateScenarioFormData {
  scenarioId: number
}

export interface AppSettingsFormData {
  skipOrgSwitchQuestion: boolean
  darkMode: DarkModeType
}
