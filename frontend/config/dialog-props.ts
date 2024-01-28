import type {DialogProps} from "primevue/dialog";

export const ModalConfig = {
    closable: true,
    closeOnEscape: true,
    blockScroll: true,
    keepInViewPort: true,
    maximizable: true,
    style: {
        width: '50vw',
    },
    breakpoints:{
        '960px': '75vw',
        '640px': '96vw'
    },
    modal: true,
    draggable: false,
} as DialogProps
