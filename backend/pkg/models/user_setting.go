package models

import "time"

type UserSetting struct {
	ID                            int64     `db:"id" json:"id"`
	UserID                        int64     `db:"user_id" json:"userId"`
	SettingsTab                   string    `db:"settings_tab" json:"settingsTab"`
	SkipOrganisationSwitchQuestion bool      `db:"skip_organisation_switch_question" json:"skipOrganisationSwitchQuestion"`
	CreatedAt                     time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt                     time.Time `db:"updated_at" json:"updatedAt"`
}

type UpdateUserSetting struct {
	SettingsTab                   *string `json:"settingsTab" validate:"omitempty"`
	SkipOrganisationSwitchQuestion *bool   `json:"skipOrganisationSwitchQuestion" validate:"omitempty"`
}
