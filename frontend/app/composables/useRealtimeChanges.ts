// Central access to real-time (SSE) change events dispatched by the
// sse.client plugin. Consumers never touch the raw state or own-change
// detection themselves.

export interface RealtimeChange {
  entity: string
  action: string
  id?: number
  parentId?: number
  ts: number
  // true when this browser tab caused the change itself
  own: boolean
}

export type ExternalChangeKind = 'updated' | 'deleted' | null

export default function useRealtimeChanges() {
  const lastChange = useState<RealtimeChange | null>('sse-last-change', () => null)

  // Run a callback whenever one of the given entities changes.
  // Own changes are delivered too (data refresh should not depend on origin);
  // check change.own before showing any user-facing notification.
  const onEntityChange = (
    entities: string | string[],
    callback: (change: RealtimeChange) => void,
  ) => {
    const list = Array.isArray(entities) ? entities : [entities]
    watch(lastChange, (change) => {
      if (!change || !list.includes(change.entity)) return
      callback(change)
    })
  }

  // Standard external-change banner state for edit dialogs: set when the
  // edited record was changed/deleted by someone else, never by own actions.
  const useExternalChangeBanner = (
    entity: string,
    editedID: () => number | undefined | null,
  ) => {
    const externalChange = ref<ExternalChangeKind>(null)
    onEntityChange(entity, (change) => {
      if (change.own) return
      const id = editedID()
      if (!id || change.id !== id) return
      externalChange.value = change.action === 'deleted' ? 'deleted' : 'updated'
    })
    return externalChange
  }

  return { lastChange, onEntityChange, useExternalChangeBanner }
}
