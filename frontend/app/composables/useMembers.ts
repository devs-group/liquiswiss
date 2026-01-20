import type { OrganisationMemberResponse, UpdateMemberFormData } from '~/models/member'

export default function useMembers() {
  const members = useState<OrganisationMemberResponse[]>('members', () => [])
  const refreshMembers = useState<(() => Promise<void>) | null>('refreshMembers', () => null)

  const setRefreshMembers = (fn: () => Promise<void>) => {
    refreshMembers.value = fn
  }

  const updateMember = async (organisationId: number, memberUserId: number, payload: UpdateMemberFormData) => {
    try {
      await $fetch(`/api/organisations/${organisationId}/members/${memberUserId}`, {
        method: 'PATCH',
        body: payload,
      })
      if (refreshMembers.value) {
        await refreshMembers.value()
      }
    }
    catch (error: unknown) {
      if ((error as { statusCode?: number })?.statusCode === 403) {
        return Promise.reject('Keine Berechtigung')
      }
      if ((error as { statusCode?: number })?.statusCode === 409) {
        return Promise.reject('Der letzte Eigentümer kann nicht herabgestuft werden')
      }
      return Promise.reject('Fehler beim Aktualisieren des Mitglieds')
    }
  }

  const removeMember = async (organisationId: number, memberUserId: number) => {
    try {
      await $fetch(`/api/organisations/${organisationId}/members/${memberUserId}`, {
        method: 'DELETE',
      })
      if (refreshMembers.value) {
        await refreshMembers.value()
      }
    }
    catch (error: unknown) {
      if ((error as { statusCode?: number })?.statusCode === 403) {
        return Promise.reject('Keine Berechtigung')
      }
      if ((error as { statusCode?: number })?.statusCode === 409) {
        const errorMsg = (error as { data?: { error?: string } })?.data?.error
        if (errorMsg === 'cannot remove the last owner') {
          return Promise.reject('Der letzte Eigentümer kann nicht entfernt werden')
        }
        if (errorMsg === 'cannot remove yourself') {
          return Promise.reject('Sie können sich selbst nicht entfernen')
        }
        return Promise.reject('Konflikt beim Entfernen des Mitglieds')
      }
      return Promise.reject('Fehler beim Entfernen des Mitglieds')
    }
  }

  const setMembers = (data: OrganisationMemberResponse[] | null) => {
    members.value = data ?? []
  }

  return {
    members,
    setRefreshMembers,
    updateMember,
    removeMember,
    setMembers,
  }
}
