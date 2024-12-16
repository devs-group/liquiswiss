export const RouteNames = {
    HOME: 'index',
    EMPLOYEES: 'employees',
    EMPLOYEE_EDIT: 'employees-id',
    TRANSACTIONS: 'transactions',
    FORECASTS: 'forecasts',
    BANK_ACCOUNTS: 'bank-accounts',
    SETTINGS: 'settings',
    PROFILE: 'settings-profile',
    ORGANISATIONS: 'settings-organisations',
    ORGANISATION_EDIT: 'settings-organisations-id',
    LOGIN: 'auth',
    REGISTRATION: 'auth-registration',
    FORGOT_PASSWORD: 'auth-forgot-password',
    RESET_PASSWORD: 'auth-reset-password',
    VALIDATE: 'auth-validate',
}

export const RoutePaths = {
    HOME: "/",
    AUTH: "/auth",
};

export const AuthRouteNames = [
    RouteNames.LOGIN,
    RouteNames.REGISTRATION,
    RouteNames.VALIDATE,
    RouteNames.FORGOT_PASSWORD,
    RouteNames.RESET_PASSWORD,
]
