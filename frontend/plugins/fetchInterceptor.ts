import {Constants} from "~/utils/constants";

export default defineNuxtPlugin((_nuxtApp) => {
    globalThis.$fetch = $fetch.create({
        onResponseError({ request, response, options }) {
            if (response._data.logout === true) {
                localStorage.setItem(Constants.SESSION_EXPIRED_NAME, 'true')
                reloadNuxtApp({force: true})
            }
        }
    })
})