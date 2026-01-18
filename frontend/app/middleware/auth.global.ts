import { AuthRouteNames, RouteNames } from '~/config/routes'
import { Constants, RedirectCookieProps } from '~/utils/constants'

export default defineNuxtRouteMiddleware(async (to) => {
  if (to.path?.startsWith('/api/')) {
    return
  }

  const { useFetchGetProfile, isAuthenticated, hasFetchedInitially } = useAuth()
  const redirectPathCookie = useCookie(Constants.REDIRECT_PATH_COOKIE, RedirectCookieProps)
  const explicitLogoutCookie = useCookie(Constants.EXPLICIT_LOGOUT, RedirectCookieProps)

  if (!isAuthenticated.value && !hasFetchedInitially.value) {
    await useFetchGetProfile()
      .catch(() => {
        // Ignore because user is most likely not authenticated
      })
  }

  const isOnAuthRoute = AuthRouteNames.includes(to.name as string)

  // Unauthenticated user trying to access protected route
  if (!isAuthenticated.value && !isOnAuthRoute) {
    // Don't save redirect path if user explicitly logged out
    if (explicitLogoutCookie.value) {
      explicitLogoutCookie.value = null
      redirectPathCookie.value = null
    }
    else {
      // Save current path for redirect after login
      redirectPathCookie.value = to.fullPath
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
