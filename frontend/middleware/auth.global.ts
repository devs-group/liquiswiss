export default defineNuxtRouteMiddleware(async (to, from) => {
    const { getAccessToken, user } = useAuth()

    if (
        (from.path.includes('/auth') && !to.path.includes('/auth')) ||
        (!from.path.includes('/auth') && to.path.includes('/auth'))
    ) {
        await getAccessToken()

        // If the user is not authenticated, redirect to the login page
        if (!user.value && to.path !== '/auth') {
            return navigateTo('/auth', { replace: true })
        }

        // If the user is authenticated and tries to access /auth, redirect to the homepage
        if (user.value && to.path === '/auth') {
            return navigateTo('/', { replace: true })
        }
    }
})