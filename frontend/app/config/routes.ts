export const RouteNames = {
  HOME: 'index',
  ORGANISATION: 'organisation',
  EMPLOYEES: 'employees',
  EMPLOYEES_EDIT: 'employees-id',
  TRANSACTIONS: 'transactions',
  FORECASTS: 'forecasts',
  BANK_ACCOUNTS: 'bank-accounts',
  SETTINGS: 'settings',
  SETTINGS_PROFILE: 'settings-profile',
  SETTINGS_ORGANISATIONS: 'settings-organisations',
  SETTINGS_APP: 'settings-app',
  SETTINGS_AUTOMATION: 'settings-automation',
  AUTH_LOGIN: 'auth',
  AUTH_REGISTRATION: 'auth-registration',
  AUTH_FORGOT_PASSWORD: 'auth-forgot-password',
  AUTH_RESET_PASSWORD: 'auth-reset-password',
  AUTH_VALIDATE: 'auth-validate',
  AUTH_INVITATION: 'auth-invitation',
}

export const RoutePaths = {
  HOME: '/',
  AUTH: '/auth',
}

export const AuthRouteNames = [
  RouteNames.AUTH_LOGIN,
  RouteNames.AUTH_REGISTRATION,
  RouteNames.AUTH_VALIDATE,
  RouteNames.AUTH_FORGOT_PASSWORD,
  RouteNames.AUTH_RESET_PASSWORD,
  RouteNames.AUTH_INVITATION,
]
