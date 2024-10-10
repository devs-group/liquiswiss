export const CreateSettingsCookie = (name: string) => {
    return useCookie(name, {
        maxAge: 60 * 60 * 24 * 365,  // 1-year expiry
        path: '/',
        secure: true,
        sameSite: 'lax'
    })
}