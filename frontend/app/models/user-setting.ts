export interface UserSettingResponse {
  id: number
  userId: number
  settingsTab: string
  skipOrganisationSwitchQuestion: boolean
  createdAt: string
  updatedAt: string
}

export interface UpdateUserSetting {
  settingsTab?: string
  skipOrganisationSwitchQuestion?: boolean
}
