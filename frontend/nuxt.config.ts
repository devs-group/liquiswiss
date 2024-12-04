// https://nuxt.com/docs/api/configuration/nuxt-config

import LiquiswissTheme from "./config/theme";

export default defineNuxtConfig({
    devtools: {
        enabled: false,
        telemetry: false,
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
    ],
    primevue: {
        options: {
            ripple: true,
            inputVariant: 'filled',
            theme: {
                preset: LiquiswissTheme,
            },
        },
    },
    tailwindcss: {
        viewer: false,
    },
    experimental: {
        appManifest: true,
        checkOutdatedBuildInterval: 15 * 1000,
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