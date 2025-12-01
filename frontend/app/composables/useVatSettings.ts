import type { VatSettingFormData, VatSettingResponse } from '~/models/vat-setting'

export default function useVatSettings() {
  const vatSetting = useState<VatSettingResponse | null>('vatSetting', () => null)

  const useFetchGetVatSetting = async () => {
    const { data, error } = await useFetch<VatSettingResponse>('/api/vat-settings', {
      method: 'GET',
    })
    if (error.value) {
      // It's okay if settings don't exist yet
      vatSetting.value = null
      return null
    }
    vatSetting.value = data.value
    return data.value
  }

  const getVatSetting = async () => {
    try {
      const data = await $fetch<VatSettingResponse>('/api/vat-settings', {
        method: 'GET',
      })
      vatSetting.value = data
      return data
    }
    catch {
      // It's okay if settings don't exist yet
      vatSetting.value = null
      return null
    }
  }

  const createVatSetting = async (payload: VatSettingFormData) => {
    try {
      const setting = await $fetch<VatSettingResponse>(`/api/vat-settings`, {
        method: 'POST',
        body: payload,
      })
      vatSetting.value = setting
      return setting
    }
    catch {
      return Promise.reject('Fehler beim Erstellen der MwSt.-Einstellungen')
    }
  }

  const updateVatSetting = async (payload: Partial<VatSettingFormData>) => {
    try {
      const setting = await $fetch<VatSettingResponse>(`/api/vat-settings`, {
        method: 'PATCH',
        body: payload,
      })
      vatSetting.value = setting
      return setting
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren der MwSt.-Einstellungen')
    }
  }

  const deleteVatSetting = async () => {
    try {
      await $fetch(`/api/vat-settings`, {
        method: 'DELETE',
      })
      vatSetting.value = null
    }
    catch {
      return Promise.reject('Fehler beim Löschen der MwSt.-Einstellungen')
    }
  }

  const saveVatSetting = async (payload: VatSettingFormData) => {
    if (vatSetting.value) {
      return await updateVatSetting(payload)
    }
    else {
      return await createVatSetting(payload)
    }
  }

  return {
    useFetchGetVatSetting,
    getVatSetting,
    createVatSetting,
    updateVatSetting,
    deleteVatSetting,
    saveVatSetting,
    vatSetting,
  }
}
