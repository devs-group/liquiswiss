// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
    modules: [
        'nuxt-primevue'
    ],
    primevue: {
        options: {
            ripple: true,
            inputStyle: 'filled'
        },
        cssLayerOrder: 'tailwind-base, primevue, tailwind-utilities',
    },
    css: [
        'primevue/resources/themes/aura-light-green/theme.css',
        'primeicons/primeicons.css',
    ],
    devtools: {enabled: true}
})
