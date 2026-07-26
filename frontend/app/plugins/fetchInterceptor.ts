import { AuthRouteNames } from '~/config/routes'
import { Constants, RedirectCookieProps, SessionTrackingCookieProps } from '~/utils/constants'

export default defineNuxtPlugin((_nuxtApp) => {
  // Track this tab's own mutations so the SSE plugin can tell own changes
  // apart from external ones (no double notifications for own actions).
  // Recorded at REQUEST START: the server can only publish the SSE event
  // after receiving the request, so the recording deterministically precedes
  // the event - no timing assumptions needed. Failed requests are removed
  // again in onResponseError.
  const ownMutations = useState<{ path: string, ts: number }[]>('sse-own-mutations', () => [])

  const mutationPath = (request: unknown, method: string) => {
    if (method === 'GET') return null
    const path = typeof request === 'string' ? request : String(request)
    return path.startsWith('/api/') ? path : null
  }

  globalThis.$fetch = $fetch.create({
    onRequest({ request, options }) {
      const method = (options.method ?? 'GET').toUpperCase()
      const path = mutationPath(request, method)
      if (!path) return
      ownMutations.value = [...ownMutations.value.slice(-9), { path, ts: Date.now() }]
    },
    onResponseError({ request, response, options }) {
      // Failed mutation published no event: remove the record again so it
      // cannot suppress a real external change
      const method = (options.method ?? 'GET').toUpperCase()
      const path = mutationPath(request, method)
      if (path) {
        const index = ownMutations.value.findLastIndex(m => m.path === path)
        if (index >= 0) {
          ownMutations.value = ownMutations.value.toSpliced(index, 1)
        }
      }
      const isOnAuthRoute = AuthRouteNames.includes(_nuxtApp._route.name as string)
      if (isOnAuthRoute) return

      // Check if this is a session expiry scenario:
      // 1. Backend explicitly indicates logout (refresh token was invalid)
      // 2. User had a session before and now gets 401 (cookies were deleted manually)
      const hadSessionCookie = useCookie<boolean | null>(Constants.HAD_SESSION_COOKIE, SessionTrackingCookieProps)
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
