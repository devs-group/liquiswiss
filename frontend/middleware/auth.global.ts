import {AuthRouteNames, RouteNames} from "~/config/routes";

export default defineNuxtRouteMiddleware(async (to, from) => {
    const { useFetchGetProfile, isAuthenticated, hasFetchedInitially } = useAuth()

    if (!isAuthenticated.value && !hasFetchedInitially.value) {
        await useFetchGetProfile()
            .catch(() => {
                // Ignore because user is most likely not authenticated
            })
    }

    const isOnAuthRoute = AuthRouteNames.includes(to.name as string)
    if (!isAuthenticated.value && !isOnAuthRoute) {
        return navigateTo({ name: RouteNames.AUTH_LOGIN}, { replace: true });
    }
    if (isAuthenticated.value && isOnAuthRoute) {
        return navigateTo({ name: RouteNames.HOME }, { replace: true });
    }
})