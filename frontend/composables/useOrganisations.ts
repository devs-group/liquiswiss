import type {ListOrganisationResponse, OrganisationFormData, OrganisationResponse} from "~/models/organisation";

export default function useOrganisations() {
    const limitOrganisations = useState('limitOrganisations', () => 20)
    const pageOrganisations = useState('pageOrganisations', () => 1)
    const organisations = useState<OrganisationResponse[]>('organisations', () => []);

    const useFetchListOrganisations = async () => {
        const {data, error} = await useFetch<ListOrganisationResponse>('/api/organisations', {
            method: 'GET',
            query: {
                page: pageOrganisations.value,
                limit: limitOrganisations.value,
            }
        });
        if (error.value) {
            return Promise.reject('Organisationen konnten nicht geladen werden')
        }
        setOrganisations(data.value?.data ?? [], false)
    }

    const listOrganisations = async ()  => {
        try {
            const data = await $fetch<ListOrganisationResponse>('/api/organisations', {
                method: 'GET',
                query: {
                    page: pageOrganisations.value,
                    limit: limitOrganisations.value,
                }
            });
            setOrganisations(data.data ?? [], false)
        } catch (err) {
            return Promise.reject('Fehler beim Laden der Organisationen')
        }
    }

    const getOrganisation = async (organisationID: number) => {
        try {
            return await $fetch<OrganisationResponse>(`/api/organisations/${organisationID}`, {
                method: 'GET',
            });
        } catch (err) {
            return Promise.reject('Fehler beim Laden der Organisation')
        }
    }

    const createOrganisation = async (payload: OrganisationFormData) => {
        try {
            await $fetch<OrganisationResponse>(`/api/organisations`, {
                method: 'POST',
                body: payload,
            });
            await listOrganisations()
        } catch (err) {
            return Promise.reject('Fehler beim Erstellen der Organisation')
        }
    }

    const updateOrganisation = async (organisationID: number, payload: OrganisationFormData) => {
        try {
            await $fetch<OrganisationResponse>(`/api/organisations/${organisationID}`, {
                method: 'PATCH',
                body: payload,
            });
            await listOrganisations()
        } catch (err) {
            return Promise.reject('Fehler beim Aktualisieren der Organisation')
        }
    }

    const setOrganisations = (data: OrganisationResponse[]|null, append: boolean) => {
        if (data) {
            if (append) {
                organisations.value = organisations.value.concat(data ?? [])
            } else {
                organisations.value = data
            }
        } else {
            organisations.value = []
        }
    }

    return {
        useFetchListOrganisations,
        listOrganisations,
        getOrganisation,
        createOrganisation,
        updateOrganisation,
        setOrganisations,
        organisations,
    };
}
