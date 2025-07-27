import type { DialogProps } from 'primevue/dialog'

export const ModalConfig = {
  closable: false,
  closeOnEscape: true,
  blockScroll: true,
  keepInViewPort: true,
  maximizable: false,
  pt: {
    root: {
      class: 'max-dialog-maximized-mobile',
    },
  },
  modal: true,
  draggable: false,
} as DialogProps
