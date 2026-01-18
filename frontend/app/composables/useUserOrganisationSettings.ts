import type { UpdateUserOrganisationSetting, UserOrganisationSettingResponse } from '~/models/user-organisation-setting'

export default function useUserOrganisationSettings() {
  const userOrganisationSetting = useState<UserOrganisationSettingResponse | null>('userOrganisationSetting', () => null)

  // Call useRequestFetch at top level (synchronously) to get SSR cookie forwarding
  const requestFetch = useRequestFetch()

  const getUserOrganisationSetting = async () => {
    try {
      const data = await requestFetch<UserOrganisationSettingResponse>('/api/user-organisation-settings', {
        method: 'GET',
      })
      userOrganisationSetting.value = data
      return data
    }
    catch {
      userOrganisationSetting.value = null
      return null
    }
  }

  const updateUserOrganisationSetting = async (payload: UpdateUserOrganisationSetting) => {
    try {
      const setting = await $fetch<UserOrganisationSettingResponse>('/api/user-organisation-settings', {
        method: 'PATCH',
        body: payload,
      })
      userOrganisationSetting.value = setting
      return setting
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren der Organisationseinstellungen')
    }
  }

  return {
    getUserOrganisationSetting,
    updateUserOrganisationSetting,
    userOrganisationSetting,
  }
}
