import {DarkModeOptions} from "~/utils/types";

export default function useDarkMode() {
    const {darkModePreference} = useSettings()
    const isDarkMode = useState<boolean>('isDarkMode', () => false)

    watch(darkModePreference, (value) => {
        if (value !== null && DarkModeOptions.includes(value)) {
            setDarkMode()
        }
    })

    const setDarkMode = () => {
        switch (darkModePreference.value) {
            case 'dark':
                isDarkMode.value = true
                break
            case 'light':
                isDarkMode.value = false
                break
            default:
                if (import.meta.client) {
                    const darkModeQuery = window.matchMedia('(prefers-color-scheme: dark)');
                    isDarkMode.value = darkModeQuery.matches
                }
        }
    }

    onMounted(() => {
        // Also listen for changes
        const darkModeQuery = window.matchMedia('(prefers-color-scheme: dark)');
        darkModeQuery.addEventListener('change', (event) => {
            if (darkModePreference.value == 'system') {
                isDarkMode.value = event.matches
            }
        });
    });

    setDarkMode()

    return { darkModePreference, isDarkMode };
}