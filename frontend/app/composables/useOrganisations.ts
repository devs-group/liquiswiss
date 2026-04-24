import type { ListOrganisationResponse, OrganisationFormData, OrganisationResponse } from '~/models/organisation'

export default function useOrganisations() {
  const limitOrganisations = useState('limitOrganisations', () => 20)
  const pageOrganisations = useState('pageOrganisations', () => 1)
  const organisations = useState<OrganisationResponse[]>('organisations', () => [])

  const useFetchListOrganisations = async () => {
    const { data, error } = await useFetch<ListOrganisationResponse>('/api/organisations', {
      method: 'GET',
      query: {
        page: pageOrganisations.value,
        limit: limitOrganisations.value,
      },
    })
    if (error.value) {
      return Promise.reject('Organisationen konnten nicht geladen werden')
    }
    setOrganisations(data.value?.data ?? [], false)
  }

  const listOrganisations = async () => {
    try {
      const data = await $fetch<ListOrganisationResponse>('/api/organisations', {
        method: 'GET',
        query: {
          page: pageOrganisations.value,
          limit: limitOrganisations.value,
        },
      })
      setOrganisations(data.data ?? [], false)
    }
    catch {
      return Promise.reject('Fehler beim Laden der Organisationen')
    }
  }

  const useFetchGetOrganisation = async (organisationID: number) => {
    const { data, error } = await useFetch<OrganisationResponse>(`/api/organisations/${organisationID}`, {
      method: 'GET',
    })
    if (error.value) {
      return Promise.reject('Fehler beim Laden der Organisation')
    }
    return data.value
  }

  const getOrganisation = async (organisationID: number) => {
    try {
      return await $fetch<OrganisationResponse>(`/api/organisations/${organisationID}`, {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject('Fehler beim Laden der Organisation')
    }
  }

  const createOrganisation = async (payload: OrganisationFormData) => {
    try {
      const organisation = await $fetch<OrganisationResponse>(`/api/organisations`, {
        method: 'POST',
        body: payload,
      })
      await listOrganisations()
      return organisation
    }
    catch {
      return Promise.reject('Fehler beim Erstellen der Organisation')
    }
  }

  const updateOrganisation = async (organisationID: number, payload: OrganisationFormData) => {
    try {
      await $fetch<OrganisationResponse>(`/api/organisations/${organisationID}`, {
        method: 'PATCH',
        body: payload,
      })
      await listOrganisations()
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren der Organisation')
    }
  }

  const setOrganisations = (data: OrganisationResponse[] | null, append: boolean) => {
    if (data) {
      if (append) {
        organisations.value = organisations.value.concat(data ?? [])
      }
      else {
        organisations.value = data
      }
    }
    else {
      organisations.value = []
    }
  }

  const { user } = useAuth()
  const currentOrganisationRole = computed(() => {
    const currentID = user.value?.currentOrganisationID
    return organisations.value.find(o => o.id === currentID)?.role ?? ''
  })
  const roleRank = (role: string): number => {
    switch (role) {
      case 'owner': return 3
      case 'admin': return 2
      case 'editor': return 1
      case 'read-only': return 0
      default: return -1
    }
  }
  const canEdit = computed(() => roleRank(currentOrganisationRole.value) >= roleRank('editor'))
  const canInvite = computed(() => roleRank(currentOrganisationRole.value) >= roleRank('admin'))
  const canManageOrganisation = computed(() => roleRank(currentOrganisationRole.value) >= roleRank('admin'))

  return {
    useFetchListOrganisations,
    listOrganisations,
    useFetchGetOrganisation,
    getOrganisation,
    createOrganisation,
    updateOrganisation,
    setOrganisations,
    organisations,
    currentOrganisationRole,
    canEdit,
    canInvite,
    canManageOrganisation,
  }
}
