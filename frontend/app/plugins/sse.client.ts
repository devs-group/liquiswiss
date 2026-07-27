// Real-time refresh: listens on /api/events (SSE) and refreshes the matching
// composable state when data changes via MCP or another organisation member.
// Events only carry {entity, action, id}; data is refetched through the
// normal REST endpoints so all authorization stays server-side.

interface ChangeEvent {
  entity: string
  action: string
  id?: number
  parentId?: number
  // Server-computed per connection: true when this exact tab (user +
  // X-Client-ID) caused the change. MCP mutations are always own=false.
  own: boolean
}

// Sub-entities highlight their owning parent element when they have no own
// card in the DOM (e.g. salary costs live inside the salary card)
const PARENT_ENTITY: Record<string, string> = {
  salary_cost: 'salary',
}

const DEBOUNCE_MS = 400

export interface SseChangeState {
  entity: string
  action: string
  id?: number
  parentId?: number
  ts: number
  // true when this tab caused the change itself (no extra notifications then)
  own: boolean
}

export default defineNuxtPlugin((nuxtApp) => {
  const { isAuthenticated } = useAuth()
  // Pages with local state (e.g. employee detail) watch this to refetch
  const lastChange = useState<SseChangeState | null>('sse-last-change', () => null)

  let source: EventSource | null = null
  const timers = new Map<string, ReturnType<typeof setTimeout>>()

  const refreshForecast = async () => {
    const { listForecasts, listForecastDetails } = useForecasts()
    const { forecastMonths } = useSettings()
    const months = forecastMonths.value + 1
    await Promise.allSettled([listForecasts(months), listForecastDetails(months)])
  }

  const refreshCategories = async () => {
    const { categories } = useGlobalData()
    try {
      const data = await $fetch<{ data: typeof categories.value }>('/api/categories', {
        query: { page: 1, limit: 100 },
      })
      categories.value = data.data ?? []
    }
    catch (error) {
      console.error('SSE: Kategorien konnten nicht aktualisiert werden', error)
    }
  }

  const refreshers: Record<string, () => Promise<unknown>> = {
    // Categories piggyback on transaction changes: their inUse flag depends
    // on which transactions reference them
    transaction: () => Promise.allSettled([
      useTransactions().listTransactions(false),
      refreshCategories(),
    ]),
    employee: () => useEmployees().listEmployees(false),
    salary: () => useEmployees().listEmployees(false),
    salary_cost: () => useEmployees().listEmployees(false),
    salary_cost_label: () => useSalaryCostLabels().listSalaryCostsLabels(false),
    bank_account: () => useBankAccounts().listBankAccounts(false),
    vat: () => useVat().listVats(),
    vat_setting: () => useVatSettings().getVatSetting(),
    category: refreshCategories,
    organisation: () => useOrganisations().listOrganisations(),
    forecast: refreshForecast,
    // Data refresh handled by page watchers (organisation page useFetch)
    member: () => Promise.resolve(),
    invitation: () => Promise.resolve(),
  }

  const flashElement = flashRealtimeElement

  // Highlight the changed cards (data-realtime-id="entity:id"); when no card
  // matches (deletes, list views without ids) fall back to the page container
  // tagged with data-realtime="<entity> ..."
  // Per-element highlighting ONLY: the changed card itself, or its owning
  // parent element (e.g. salary card for salary costs). Whole areas never
  // blink; when nothing addressable exists the caller shows a toast instead.
  const flash = (entity: string, ids: Set<number>, parentIDs: Set<number>): boolean => {
    let matched = false
    ids.forEach((id) => {
      document.querySelectorAll(`[data-realtime-id="${entity}:${id}"]`).forEach((element) => {
        matched = true
        flashElement(element)
      })
    })
    if (matched) return true
    const parentEntity = PARENT_ENTITY[entity]
    if (parentEntity) {
      parentIDs.forEach((id) => {
        document.querySelectorAll(`[data-realtime-id="${parentEntity}:${id}"]`).forEach((element) => {
          matched = true
          flashElement(element)
        })
      })
    }
    return matched
  }

  // Entities without addressable elements get a toast instead of a blink;
  // silent entities are refreshed without any notification (their pages
  // handle field-level highlighting themselves, or the change is visible)
  const TOAST_LABELS: Record<string, string> = {
    category: 'Kategorien',
    vat: 'MwSt.-Sätze',
    vat_setting: 'MwSt.-Einstellungen',
    salary_cost_label: 'Lohnkosten-Labels',
    member: 'Mitglieder',
    invitation: 'Einladungen',
    organisation: 'Organisation',
  }
  const SILENT_ENTITIES = new Set(['forecast'])
  // Pages with their own field-level handling suppress the toast while open
  const selfHandled = (entity: string) => {
    if (['organisation', 'vat_setting', 'member', 'invitation', 'category'].includes(entity)) {
      return useRoute().path.startsWith('/organisation')
    }
    return false
  }

  const showUpdateToast = (entity: string) => {
    if (SILENT_ENTITIES.has(entity) || selfHandled(entity)) return
    const label = TOAST_LABELS[entity]
    if (!label) return
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const toast = (nuxtApp.vueApp.config.globalProperties as any).$toast
    toast?.add({
      severity: 'info',
      summary: 'Aktualisiert',
      detail: `${label} wurden soeben aktualisiert`,
      life: 3000,
    })
  }

  const pendingIDs = new Map<string, Set<number>>()
  const pendingParentIDs = new Map<string, Set<number>>()

  // Deleted rows blink red while still in the DOM, then the refetch removes
  // them. Returns true when at least one row was found and highlighted.
  const flashDeletes = (entity: string, ids: Set<number>) => {
    let matched = false
    ids.forEach((id) => {
      document.querySelectorAll(`[data-realtime-id="${entity}:${id}"]`).forEach((element) => {
        matched = true
        flashElement(element, 'realtime-flash-delete')
      })
    })
    return matched
  }

  const DELETE_FLASH_MS = 1600

  const dispatch = (event: ChangeEvent) => {
    // Deletes blink red IMMEDIATELY while the element is still in the DOM;
    // the refresh (and page watchers via lastChange) waits until the blink is
    // done so nothing removes the element early
    // Own events refresh data but never blink or toast: the acting tab
    // already sees its change directly
    let deleteFlashed = false
    if (!event.own && event.action === 'deleted' && event.id) {
      deleteFlashed = flashDeletes(event.entity, new Set([event.id]))
    }
    const notifyDelay = deleteFlashed ? DELETE_FLASH_MS : 0

    setTimeout(() => {
      lastChange.value = { ...event, ts: Date.now() }
      const refresh = refreshers[event.entity]
      if (!refresh) return
      // Debounce per entity so bulk MCP operations trigger one refetch;
      // collect all changed ids in the meantime for targeted highlighting
      if (!event.own && event.action !== 'deleted') {
        if (!pendingIDs.has(event.entity)) pendingIDs.set(event.entity, new Set())
        if (event.id) pendingIDs.get(event.entity)!.add(event.id)
      }
      if (!event.own && event.parentId) {
        if (!pendingParentIDs.has(event.entity)) pendingParentIDs.set(event.entity, new Set())
        pendingParentIDs.get(event.entity)!.add(event.parentId)
      }
      const existing = timers.get(event.entity)
      if (existing) clearTimeout(existing)
      timers.set(event.entity, setTimeout(() => {
        timers.delete(event.entity)
        const ids = pendingIDs.get(event.entity) ?? new Set<number>()
        pendingIDs.delete(event.entity)
        const parentIDs = pendingParentIDs.get(event.entity) ?? new Set<number>()
        pendingParentIDs.delete(event.entity)

        refresh()
          .then(async () => {
            await nextTick()
            if (deleteFlashed && ids.size === 0 && parentIDs.size === 0) return
            const matched = flash(event.entity, ids, parentIDs)
            // Page-level watchers (detail page, dialogs) refetch on the same
            // event and their re-render replaces the flashed elements. Run a
            // second pass once that render settled; still-blinking elements
            // are skipped so list pages don't restart their animation. When
            // nothing addressable exists at all, fall back to a toast.
            setTimeout(() => {
              const secondMatched = flash(event.entity, ids, parentIDs)
              // own events (server-computed) never toast: the acting tab
              // already sees its change
              if (!matched && !secondMatched && !deleteFlashed && !event.own) {
                showUpdateToast(event.entity)
              }
            }, 800)
          })
          .catch(error => console.error(`SSE: Refresh für ${event.entity} fehlgeschlagen`, error))
      }, DEBOUNCE_MS))
    }, notifyDelay)
  }

  const connect = () => {
    if (source) return
    // The client id lets the backend compute own=true for events this exact
    // tab caused (EventSource cannot send custom headers, hence the query)
    source = new EventSource(`/api/events?client=${getRealtimeClientId()}`)
    source.addEventListener('change', (message: MessageEvent) => {
      try {
        dispatch(JSON.parse(message.data) as ChangeEvent)
      }
      catch (error) {
        console.error('SSE: Ungültiges Event', error)
      }
    })
    // EventSource reconnects automatically; reconnects re-run the full auth
    // middleware on the backend. Nothing to do on error while authenticated.
    source.onerror = () => {
      if (!isAuthenticated.value) disconnect()
    }
  }

  const disconnect = () => {
    source?.close()
    source = null
    timers.forEach(timer => clearTimeout(timer))
    timers.clear()
  }

  watch(isAuthenticated, (authenticated) => {
    if (authenticated) connect()
    else disconnect()
  }, { immediate: true })

  // Free the server-side connection slot as early as possible on tab close
  window.addEventListener('beforeunload', disconnect)

  nuxtApp.hook('app:beforeMount', () => {
    if (isAuthenticated.value) connect()
  })
})
