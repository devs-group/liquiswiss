export default function useDarkMode() {
    const isDarkMode = ref(false);

    const checkDarkMode = () => {
        isDarkMode.value = window.matchMedia('(prefers-color-scheme: dark)').matches;
    };

    onMounted(() => {
        checkDarkMode();

        const darkModeQuery = window.matchMedia('(prefers-color-scheme: dark)');
        darkModeQuery.addEventListener('change', (event) => {
            isDarkMode.value = event.matches;
        });
    });

    return { isDarkMode };
}