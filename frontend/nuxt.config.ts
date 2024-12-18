// https://nuxt.com/docs/api/configuration/nuxt-config

import LiquiswissTheme from "./config/theme";

export default defineNuxtConfig({
    devtools: {
        enabled: false,
        telemetry: false,
    },
    app: {
        head: {
            meta: [
                {name: 'apple-mobile-web-app-title', content: 'LiquiSwiss'},
            ],
            link: [
                {rel: 'icon', type: 'image/png', href: '/favicon-96x96.png', sizes: '96x96'},
                {rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg'},
                {rel: 'shortcut icon', href: '/favicon.ico'},
                {rel: 'apple-touch-icon', href: '/apple-touch-icon.png', sizes: '180x180'},
                {rel: 'manifest', href: '/site.webmanifest'},
            ]
        }
    },
    runtimeConfig: {
        apiHost: 'http://localhost:8080',
    },
    routeRules: {
        '/**': {
            headers: {
                'cache-control': 'no-cache',
            }
        }
    },
    modules: [
        '@primevue/nuxt-module',
        '@nuxtjs/tailwindcss',
        '@nuxtjs/color-mode',
    ],
    primevue: {
        options: {
            ripple: true,
            inputVariant: 'filled',
            theme: {
                preset: LiquiswissTheme,
                options: {
                    darkModeSelector: '.dark',
                }
            },
        },
    },
    colorMode: {
        classSuffix: '',
        storage: 'localStorage',
        storageKey: 'dark-mode-preference',
    },
    tailwindcss: {
        viewer: false,
    },
    experimental: {
        appManifest: true,
        // Every 5 minutes
        checkOutdatedBuildInterval: 300 * 1000,
    },
    vite: {
        esbuild: {
            drop: ['debugger'],
        }
    },
    css: ["@/assets/css/tailwind.css", "primeicons/primeicons.css"],
    watch: ["config/theme.ts"],
    compatibilityDate: '2024-07-16',
})