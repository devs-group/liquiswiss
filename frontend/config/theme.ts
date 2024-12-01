import Aura from "@primevue/themes/aura";
import { definePreset } from "@primevue/themes";

const LiquiswissTheme = definePreset(Aura, {
    components: {
        card: {
            colorScheme: {
                light: {
                    root: {
                        background: '{zinc.100}',
                    },
                },
                dark: {
                    root: {
                        background: '{zinc.800}',
                    },
                }
            }
        },
        menu: {
            colorScheme: {
                light: {
                    root: {
                        background: '{zinc.100}',
                    },
                },
                dark: {
                    root: {
                        background: '{zinc.800}',
                    },
                }
            }
        },
        dialog: {
            colorScheme: {
                light: {
                    root: {
                        background: '{zinc.100}',
                    },
                },
                dark: {
                    root: {
                        background: '{zinc.800}',
                    },
                }
            }
        },
    }
});

export default LiquiswissTheme;
