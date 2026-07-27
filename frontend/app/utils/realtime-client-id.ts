// Per-tab client id for server-computed "own" detection on SSE events.
// Module scope = one id per page load: the fetch interceptor sends it as
// X-Client-ID on every request and the SSE stream connects with it, so the
// backend can mark events caused by this exact tab with own=true.
let clientId: string | null = null

export const getRealtimeClientId = (): string => {
  if (!clientId) {
    clientId = typeof crypto !== 'undefined' && 'randomUUID' in crypto
      ? crypto.randomUUID()
      : `${Date.now()}-${Math.random().toString(36).slice(2)}`
  }
  return clientId
}
