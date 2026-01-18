import type { UpdateUserSetting, UserSettingResponse } from '~/models/user-setting'

export default function useUserSettings() {
  const userSetting = useState<UserSettingResponse | null>('userSetting', () => null)

  // Call useRequestFetch at top level (synchronously) to get SSR cookie forwarding
  const requestFetch = useRequestFetch()

  const getUserSetting = async () => {
    try {
      const data = await requestFetch<UserSettingResponse>('/api/user-settings', {
        method: 'GET',
      })
      userSetting.value = data
      return data
    }
    catch {
      userSetting.value = null
      return null
    }
  }

  const updateUserSetting = async (payload: UpdateUserSetting) => {
    try {
      const setting = await $fetch<UserSettingResponse>('/api/user-settings', {
        method: 'PATCH',
        body: payload,
      })
      userSetting.value = setting
      return setting
    }
    catch {
      return Promise.reject('Fehler beim Aktualisieren der Benutzereinstellungen')
    }
  }

  return {
    getUserSetting,
    updateUserSetting,
    userSetting,
  }
}
