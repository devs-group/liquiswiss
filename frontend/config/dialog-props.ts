import type {DialogProps} from "primevue/dialog";

export const ModalConfig = {
    closable: true,
    closeOnEscape: true,
    blockScroll: true,
    keepInViewPort: true,
    maximizable: false,
    style: {
        width: '50vw',
    },
    breakpoints:{
        '960px': '75vw',
        '640px': '96vw'
    },
    pt: {
        root: {
            class: "max-dialog-maximized-mobile",
        },
    },
    modal: true,
    draggable: false,
} as DialogProps
