import type { DialogProps } from 'primevue/dialog'

export const ModalConfig = {
  closable: true,
  closeOnEscape: false,
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
