import {Routes} from "~/config/routes";

export default defineNuxtRouteMiddleware(async (to, from) => {
    const { getAccessToken, user } = useAuth()

    const routeToCheck = `/${Routes.LOGIN}`

    if (
        (from.path.includes(routeToCheck) && !to.path.includes(routeToCheck)) ||
        (!from.path.includes(routeToCheck) && to.path.includes(routeToCheck))
    ) {
        await getAccessToken()

        // If the user is not authenticated, redirect to the login page
        if (!user.value && to.path !== routeToCheck) {
            return navigateTo({name: Routes.LOGIN}, { replace: true })
        }

        // If the user is authenticated and tries to access /auth, redirect to the homepage
        if (user.value && to.path === routeToCheck) {
            return navigateTo({name: Routes.HOME}, { replace: true })
        }
    }
})