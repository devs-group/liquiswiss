import { AuthRouteNames } from '~/config/routes'
import { Constants, RedirectCookieProps } from '~/utils/constants'

export default defineNuxtPlugin((_nuxtApp) => {
  globalThis.$fetch = $fetch.create({
    onResponseError({ response }) {
      const isOnAuthRoute = AuthRouteNames.includes(_nuxtApp._route.name as string)
      if (!isOnAuthRoute && response._data.logout === true) {
        // Save current path for redirect after re-login
        const redirectPathCookie = useCookie(Constants.REDIRECT_PATH_COOKIE, RedirectCookieProps)
        redirectPathCookie.value = _nuxtApp._route.fullPath
        localStorage.setItem(Constants.SESSION_EXPIRED_NAME, 'true')
        reloadNuxtApp({ force: true })
      }
    },
  })
})
