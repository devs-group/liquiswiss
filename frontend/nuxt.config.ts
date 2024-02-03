// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
    devtools: {
        enabled: false,
    },
    build: {
        transpile: ["primevue"],
    },
    runtimeConfig: {
        apiHost: '',
    },
    modules: [
        'nuxt-primevue',
        '@nuxtjs/tailwindcss',
    ],
    primevue: {
        options: {
            ripple: true,
            inputStyle: 'filled',
        },
        cssLayerOrder: 'tailwind-base, primevue, tailwind-utilities',
    },
    css: [
        'primevue/resources/themes/aura-light-green/theme.css',
        'primeicons/primeicons.css',
    ],
    compatibilityDate: '2024-07-16',
})