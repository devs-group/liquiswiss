import { AuthRouteNames, RouteNames } from '~/config/routes'
import { Constants, RedirectCookieProps, SessionTrackingCookieProps } from '~/utils/constants'

export default defineNuxtRouteMiddleware(async (to) => {
  if (to.path?.startsWith('/api/')) {
    return
  }

  const { useFetchGetProfile, isAuthenticated, hasFetchedInitially } = useAuth()
  const { loadSettings, loadOrganisationSettings, settingsLoaded, organisationSettingsLoaded } = useSettings()
  const redirectPathCookie = useCookie(Constants.REDIRECT_PATH_COOKIE, RedirectCookieProps)
  const explicitLogoutCookie = useCookie(Constants.EXPLICIT_LOGOUT, RedirectCookieProps)
  const sessionExpiredCookie = useCookie<boolean | null>(Constants.SESSION_EXPIRED_COOKIE, RedirectCookieProps)
  const hadSessionCookie = useCookie<boolean | null>(Constants.HAD_SESSION_COOKIE, SessionTrackingCookieProps)

  // Note: We cannot detect cookie deletion via JavaScript because auth cookies are HTTP-only (secure).
  // Session expiry is detected through:
  // 1. API 401 responses with sessionExpired flag (handled by fetchInterceptor)
  // 2. Profile fetch failure on page load (checked below)

  let isSessionExpired = false

  if (!isAuthenticated.value && !hasFetchedInitially.value) {
    await useFetchGetProfile()
      .then(async () => {
        // Load settings after successful authentication
        if (!settingsLoaded.value) {
          await loadSettings()
        }
        if (!organisationSettingsLoaded.value) {
          await loadOrganisationSettings()
        }
      })
      .catch((err) => {
        // Check if session expired (Flow 2: page load with expired session)
        if (err?.sessionExpired) {
          isSessionExpired = true
        }
      })
  }
  else if (isAuthenticated.value) {
    // User already authenticated - ensure settings are loaded
    if (!settingsLoaded.value) {
      await loadSettings()
    }
    if (!organisationSettingsLoaded.value) {
      await loadOrganisationSettings()
    }
  }

  // Track that user has had a session (for detecting session expiry later)
  if (isAuthenticated.value) {
    hadSessionCookie.value = true
  }

  const isOnAuthRoute = AuthRouteNames.includes(to.name as string)

  // Unauthenticated user trying to access protected route
  if (!isAuthenticated.value && !isOnAuthRoute) {
    // Don't save redirect path if user explicitly logged out
    if (explicitLogoutCookie.value) {
      explicitLogoutCookie.value = null
      redirectPathCookie.value = null
      hadSessionCookie.value = null
      return navigateTo({ name: RouteNames.AUTH_LOGIN }, { replace: true })
    }

    // Save current path for redirect after login
    redirectPathCookie.value = to.fullPath

    // If user had a session before (cookies cleared/expired) or backend indicated session expired
    if (isSessionExpired || hadSessionCookie.value) {
      hadSessionCookie.value = null

      if (import.meta.client) {
        // Client-side navigation: show dialog on current page (user was actively using the app)
        // Set state directly to trigger the watch in app.vue
        const { sessionExpired } = useAuth()
        sessionExpired.value = true
        // Don't redirect - the dialog in app.vue will handle the reload
        return
      }
      else {
        // Page load/reload (SSR): redirect to login and show toast there
        sessionExpiredCookie.value = true
        return navigateTo({ name: RouteNames.AUTH_LOGIN }, { replace: true })
      }
    }

    return navigateTo({ name: RouteNames.AUTH_LOGIN }, { replace: true })
  }

  // Authenticated user on auth route - redirect to saved path or home
  if (isAuthenticated.value && isOnAuthRoute) {
    const wasExplicitLogout = !!explicitLogoutCookie.value
    const savedPath = redirectPathCookie.value

    // Always clear both cookies after login
    explicitLogoutCookie.value = null
    redirectPathCookie.value = null

    // Only redirect to saved path if it wasn't an explicit logout
    if (!wasExplicitLogout && savedPath && savedPath !== '/') {
      return navigateTo(savedPath, { replace: true })
    }
    return navigateTo({ name: RouteNames.HOME }, { replace: true })
  }
})
