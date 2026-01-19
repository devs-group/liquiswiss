import { AuthRouteNames } from '~/config/routes'
import { Constants, RedirectCookieProps, SettingsCookieProps } from '~/utils/constants'

export default defineNuxtPlugin((_nuxtApp) => {
  globalThis.$fetch = $fetch.create({
    onResponseError({ response }) {
      const isOnAuthRoute = AuthRouteNames.includes(_nuxtApp._route.name as string)
      if (isOnAuthRoute) return

      // Check if this is a session expiry scenario:
      // 1. Backend explicitly indicates logout (refresh token was invalid)
      // 2. User had a session before and now gets 401 (cookies were deleted manually)
      const hadSessionCookie = useCookie<boolean | null>(Constants.HAD_SESSION_COOKIE, SettingsCookieProps)
      const isSessionExpired = response._data?.logout === true
        || (response.status === 401 && hadSessionCookie.value === true)

      if (isSessionExpired) {
        // Clear hadSession to prevent repeated dialogs
        hadSessionCookie.value = null
        // Save current path for redirect after re-login
        const redirectPathCookie = useCookie(Constants.REDIRECT_PATH_COOKIE, RedirectCookieProps)
        redirectPathCookie.value = _nuxtApp._route.fullPath
        // Set state to trigger session expired dialog in app.vue
        const { sessionExpired } = useAuth()
        sessionExpired.value = true
      }
    },
  })
})
