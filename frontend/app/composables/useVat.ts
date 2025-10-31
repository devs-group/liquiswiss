import type { VatFormData, VatResponse } from '~/models/vat'

export default function useVat() {
  const vats = useState<VatResponse[]>('vats', () => [])

  const useFetchListVats = async () => {
    const { data, error } = await useFetch<VatResponse[]>('/api/vats', {
      method: 'GET',
    })
    if (error.value) {
      return Promise.reject('Mehrwertsteuern konnten nicht geladen werden')
    }
    setVats(data.value, false)
  }

  const listVats = async () => {
    try {
      const data = await $fetch<VatResponse[]>('/api/vats', {
        method: 'GET',
      })
      setVats(data, false)
    }
    catch {
      return Promise.reject('Fehler beim Laden der Mehrwertsteuern')
    }
  }

  const getVat = async (vatID: number) => {
    try {
      return await $fetch<VatResponse>(`/api/vats/${vatID}`, {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject('Fehler beim Laden der Mehrwertsteuer')
    }
  }

  const createVat = async (payload: VatFormData) => {
    try {
      const vat = await $fetch<VatResponse>(`/api/vats`, {
        method: 'POST',
        body: {
          ...payload,
          value: AmountToInteger(payload.value as number),
        },
      })
      await listVats()
      return vat
    }
    catch {
      return Promise.reject('Fehler beim Erstellen der Mehrwertsteuer')
    }
  }

  const updateVat = async (payload: VatFormData) => {
    try {
      await $fetch<VatResponse>(`/api/vats/${payload.id}`, {
        method: 'PATCH',
        body: {
          ...payload,
          value: AmountToInteger(payload.value as number),
        },
      })
      await listVats()
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren der Mehrwertsteuer')
    }
  }

  const deleteVat = async (vatID: number) => {
    try {
      await $fetch(`/api/vats/${vatID}`, {
        method: 'DELETE',
      })
      await listVats()
    }
    catch {
      return Promise.reject('Fehler beim Löschen der Mehrwertsteuer')
    }
  }

  const setVats = (data: VatResponse[] | null, append: boolean) => {
    if (data) {
      if (append) {
        vats.value = vats.value.concat(data ?? [])
      }
      else {
        vats.value = data
      }
    }
    else {
      vats.value = []
    }
  }

  return {
    useFetchListVats,
    listVats,
    getVat,
    createVat,
    updateVat,
    deleteVat,
    setVats,
    vats,
  }
}
