import { AuthRouteNames } from '~/config/routes'

export default defineNuxtPlugin((_nuxtApp) => {
  globalThis.$fetch = $fetch.create({
    onResponseError({ response }) {
      const isOnAuthRoute = AuthRouteNames.includes(_nuxtApp._route.name as string)
      if (!isOnAuthRoute && response._data.logout === true) {
        localStorage.setItem(Constants.SESSION_EXPIRED_NAME, 'true')
        reloadNuxtApp({ force: true }) // , path: RoutePaths.AUTH
      }
    },
  })
})
