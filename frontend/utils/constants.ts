import type {CookieOptions} from "#app";

export const Constants = {
    SESSION_EXPIRED_NAME: 'sessionExpired'
}

export const LocalStorageKeys = {
    TRANSACTION_DISPLAY: 'transaction-display'
}

export const SettingsCookieProps = {
    maxAge: 60 * 60 * 24 * 365,  // 1-year expiry
    path: '/',
    secure: true,
    sameSite: 'lax'
} as CookieOptions