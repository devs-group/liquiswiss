package models

import "time"

type OrganisationMember struct {
	UserID     int64             `db:"user_id" json:"userId"`
	Name       string            `db:"name" json:"name"`
	Email      string            `db:"email" json:"email"`
	Role       string            `db:"role" json:"role"`
	IsDefault  bool              `db:"is_default" json:"isDefault"`
	Permission *MemberPermission `json:"permission,omitempty"`
}

type MemberPermission struct {
	ID             int64      `db:"id" json:"id"`
	UserID         int64      `db:"user_id" json:"userId"`
	OrganisationID int64      `db:"organisation_id" json:"organisationId"`
	EntityType     *string    `db:"entity_type" json:"entityType"`
	CanView        bool       `db:"can_view" json:"canView"`
	CanEdit        bool       `db:"can_edit" json:"canEdit"`
	CanDelete      bool       `db:"can_delete" json:"canDelete"`
	CreatedAt      time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt      *time.Time `db:"updated_at" json:"updatedAt,omitempty"`
}

type UpdateMember struct {
	Role      *string `json:"role" validate:"omitempty,oneof=admin editor read-only"`
	CanView   *bool   `json:"canView"`
	CanEdit   *bool   `json:"canEdit"`
	CanDelete *bool   `json:"canDelete"`
}
