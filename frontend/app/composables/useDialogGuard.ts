// Close guard for dynamic dialogs: Escape and the header X close the dialog,
// but dirty forms first show a confirm ("changes will be lost") that can be
// aborted. Dirty values are additionally cached in sessionStorage so a page
// reload (dialogs reopen via query param) restores unsaved entries.

import { useConfirm } from 'primevue/useconfirm'

interface DialogDraft<T> {
  // Storage key, unique per dialog + record (e.g. "category:12", "category:new")
  key: string
  // Snapshot of the current form values (JSON-serializable)
  capture: () => T
  // Re-apply a saved snapshot (revive Dates etc. before setValues)
  restore: (saved: T) => void
}

interface DialogGuardOptions<T> {
  dirty: () => boolean
  close: (payload?: unknown) => void
  // Payload used when Escape or the header X trigger the close (e.g. a
  // requires-refresh flag the parent expects)
  payload?: () => unknown
  draft?: DialogDraft<T>
}

interface GuardEntry {
  requestClose: () => void
}

// Topmost dialog reacts to Escape; stacked dialogs close one by one
const guardStack: GuardEntry[] = []
let escapeListenerBound = false

// Overlays (dropdowns, datepickers, confirm popups) handle Escape themselves;
// the dialog must not close underneath them
const OVERLAY_SELECTOR = [
  '.p-select-overlay',
  '.p-multiselect-overlay',
  '.p-autocomplete-overlay',
  '.p-datepicker-panel:not(.p-datepicker-inline)',
  '.p-popover',
  '.p-confirmdialog',
  '.p-menu-overlay',
  '.p-tieredmenu-overlay',
].join(', ')

const bindEscapeListener = () => {
  if (escapeListenerBound || import.meta.server) return
  escapeListenerBound = true
  // Capture phase: runs BEFORE PrimeVue's own Escape handlers, so an open
  // overlay (dropdown, datepicker) is still in the DOM and detectable; the
  // overlay then closes itself and the dialog stays open for this press
  window.addEventListener('keydown', (event) => {
    if (event.key !== 'Escape' || guardStack.length === 0) return
    if (document.querySelector(OVERLAY_SELECTOR)) return
    guardStack[guardStack.length - 1]!.requestClose()
  }, { capture: true })
}

const draftStorageKey = (key: string) => `dialog-draft:${key}`

export default function useDialogGuard<T>(options: DialogGuardOptions<T>) {
  const confirm = useConfirm()

  const clearDraft = () => {
    if (options.draft) sessionStorage.removeItem(draftStorageKey(options.draft.key))
  }

  // Closes the dialog and drops the draft; use for cancel/success paths
  const close = (payload?: unknown) => {
    clearDraft()
    options.close(payload)
  }

  // Escape / X / Abbrechen: confirm first when the form is dirty
  const requestClose = (payload?: unknown) => {
    if (!options.dirty()) {
      close(payload)
      return
    }
    confirm.require({
      header: 'Ungespeicherte Änderungen',
      message: 'Die Änderungen gehen verloren. Trotzdem schliessen?',
      icon: 'pi pi-exclamation-triangle',
      acceptLabel: 'Verwerfen',
      acceptProps: { severity: 'danger' },
      rejectLabel: 'Weiter bearbeiten',
      rejectProps: { severity: 'secondary', outlined: true },
      accept: () => close(payload),
    })
  }

  const entry: GuardEntry = { requestClose: () => requestClose(options.payload?.()) }

  // Persist dirty values so a reload (dialog reopens via query param) can
  // restore them; cleared on every regular close
  let stopDraftWatch: (() => void) | undefined
  if (options.draft) {
    const draft = options.draft
    stopDraftWatch = watch(
      () => draft.capture(),
      (values) => {
        if (!options.dirty()) return
        try {
          sessionStorage.setItem(draftStorageKey(draft.key), JSON.stringify(values))
        }
        catch { /* storage full/unavailable: draft is best-effort */ }
      },
      { deep: true },
    )
  }

  const instance = getCurrentInstance()

  onMounted(() => {
    bindEscapeListener()
    guardStack.push(entry)

    // Restore a draft from before the reload; setValues makes the form dirty
    // again so the guard and the draft watcher stay active
    if (options.draft) {
      const raw = sessionStorage.getItem(draftStorageKey(options.draft.key))
      if (raw) {
        try {
          options.draft.restore(JSON.parse(raw) as T)
        }
        catch {
          clearDraft()
        }
      }
    }

    // The header X of the surrounding PrimeVue dialog closes without asking;
    // intercept it in the capture phase and route through the guard
    const root = (instance?.proxy?.$el as HTMLElement | null)?.closest('.p-dialog')
    const closeButton = root?.querySelector('.p-dialog-close-button')
    closeButton?.addEventListener('click', (event) => {
      event.preventDefault()
      event.stopImmediatePropagation()
      requestClose(options.payload?.())
    }, { capture: true })
  })

  onUnmounted(() => {
    stopDraftWatch?.()
    const index = guardStack.indexOf(entry)
    if (index >= 0) guardStack.splice(index, 1)
  })

  return { requestClose, close, clearDraft }
}
