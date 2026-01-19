import type { CookieOptions } from '#app'

// Helper: secure cookies only in production (HTTPS), not in dev/CI (HTTP)
export const isProduction = process.env.NODE_ENV === 'production'

export const Constants = {
  SESSION_EXPIRED_STATE: 'sessionExpiredState',
  SESSION_EXPIRED_COOKIE: 'sessionExpiredCookie',
  SESSION_EXPIRED_DIALOG_COOKIE: 'sessionExpiredDialog',
  EXPLICIT_LOGOUT: 'explicitLogout',
  REDIRECT_PATH_COOKIE: 'redirectPath',
  HAD_SESSION_COOKIE: 'hadSession',
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

// Session tracking cookies - secure only in production (HTTPS)
// In dev/CI (HTTP), secure cookies won't work
export const SessionTrackingCookieProps = {
  maxAge: 60 * 60 * 24 * 3, // 72-hour expiry
  path: '/',
  secure: isProduction,
  sameSite: 'lax',
} as CookieOptions

export const Fallbacks = {
  CostLabel: '<Kein Label>',
}
