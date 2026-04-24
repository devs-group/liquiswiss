import type { User } from '~/models/auth'
import type { AcceptInvitationFormData, CheckInvitationResponse, CreateInvitationFormData, InvitationResponse, UserPendingInvitation } from '~/models/invitation'

export default function useInvitations() {
  const invitations = useState<InvitationResponse[]>('invitations', () => [])
  const myPendingInvitations = useState<UserPendingInvitation[]>('myPendingInvitations', () => [])
  const refreshInvitations = useState<(() => Promise<void>) | null>('refreshInvitations', () => null)

  const setRefreshInvitations = (fn: () => Promise<void>) => {
    refreshInvitations.value = fn
  }

  const createInvitation = async (organisationId: number, payload: CreateInvitationFormData) => {
    try {
      const invitation = await $fetch<InvitationResponse>(`/api/organisations/${organisationId}/invitations`, {
        method: 'POST',
        body: payload,
      })
      if (refreshInvitations.value) {
        await refreshInvitations.value()
      }
      return invitation
    }
    catch (error: unknown) {
      if ((error as { statusCode?: number })?.statusCode === 409) {
        return Promise.reject('Dieser Benutzer ist bereits Mitglied')
      }
      return Promise.reject('Fehler beim Erstellen der Einladung')
    }
  }

  const deleteInvitation = async (organisationId: number, invitationId: number) => {
    try {
      await $fetch(`/api/organisations/${organisationId}/invitations/${invitationId}`, {
        method: 'DELETE',
      })
      if (refreshInvitations.value) {
        await refreshInvitations.value()
      }
    }
    catch {
      return Promise.reject('Fehler beim Löschen der Einladung')
    }
  }

  const resendInvitation = async (organisationId: number, invitationId: number) => {
    try {
      await $fetch(`/api/organisations/${organisationId}/invitations/${invitationId}/resend`, {
        method: 'POST',
      })
    }
    catch {
      return Promise.reject('Fehler beim erneuten Senden der Einladung')
    }
  }

  const checkInvitation = async (token: string) => {
    try {
      return await $fetch<CheckInvitationResponse>(`/api/auth/invitation/check`, {
        method: 'GET',
        query: { token },
      })
    }
    catch (error: unknown) {
      if ((error as { statusCode?: number })?.statusCode === 410) {
        return Promise.reject('Einladung ist abgelaufen')
      }
      if ((error as { statusCode?: number })?.statusCode === 404) {
        return Promise.reject('Einladung nicht gefunden')
      }
      return Promise.reject('Fehler beim Überprüfen der Einladung')
    }
  }

  const acceptInvitation = async (payload: AcceptInvitationFormData) => {
    try {
      return await $fetch<User>(`/api/auth/invitation/accept`, {
        method: 'POST',
        body: payload,
      })
    }
    catch (error: unknown) {
      if ((error as { statusCode?: number })?.statusCode === 410) {
        return Promise.reject('Einladung ist abgelaufen')
      }
      if ((error as { statusCode?: number })?.statusCode === 401) {
        return Promise.reject('E-Mail oder Passwort ist falsch')
      }
      return Promise.reject('Fehler beim Annehmen der Einladung')
    }
  }

  const setInvitations = (data: InvitationResponse[] | null) => {
    invitations.value = data ?? []
  }

  const declineMyInvitation = async (invitationId: number) => {
    try {
      await $fetch(`/api/me/invitations/${invitationId}`, { method: 'DELETE' })
      myPendingInvitations.value = myPendingInvitations.value.filter(i => i.id !== invitationId)
    }
    catch {
      return Promise.reject('Fehler beim Ablehnen der Einladung')
    }
  }

  const loadMyPendingInvitations = async () => {
    try {
      // Forward the incoming request's cookies so the call succeeds during SSR.
      const headers = import.meta.server ? useRequestHeaders(['cookie']) : undefined
      const data = await $fetch<UserPendingInvitation[]>(`/api/me/invitations`, {
        method: 'GET',
        headers,
      })
      myPendingInvitations.value = data ?? []
    }
    catch {
      myPendingInvitations.value = []
    }
  }

  return {
    invitations,
    myPendingInvitations,
    setRefreshInvitations,
    createInvitation,
    deleteInvitation,
    resendInvitation,
    checkInvitation,
    acceptInvitation,
    setInvitations,
    loadMyPendingInvitations,
    declineMyInvitation,
  }
}
