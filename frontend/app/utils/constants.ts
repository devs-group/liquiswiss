import type { CookieOptions } from '#app'

export const Constants = {
  SESSION_EXPIRED_NAME: 'sessionExpired',
  EXPLICIT_LOGOUT: 'explicitLogout',
  REDIRECT_PATH_COOKIE: 'redirectPath',
  BASE_CURRENCY: 'CHF',
  BASE_LOCALE_CODE: 'de-CH',
}

export const LocalStorageKeys = {
  TRANSACTION_DISPLAY: 'transaction-display',
}

export const SettingsCookieProps = {
  maxAge: 60 * 60 * 24 * 365, // 1-year expiry
  path: '/',
  secure: true,
  sameSite: 'lax',
} as CookieOptions

export const RedirectCookieProps = {
  maxAge: 60 * 60 * 24 * 3, // 72-hour expiry
  path: '/',
  secure: true,
  sameSite: 'lax',
} as CookieOptions

export const Fallbacks = {
  CostLabel: '<Kein Label>',
}
