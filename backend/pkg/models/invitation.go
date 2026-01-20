package models

import "time"

type Invitation struct {
	ID             int64     `db:"id" json:"id"`
	OrganisationID int64     `db:"organisation_id" json:"organisationId"`
	Email          string    `db:"email" json:"email"`
	Role           string    `db:"role" json:"role"`
	Token          string    `db:"token" json:"-"`
	InvitedBy      int64     `db:"invited_by" json:"invitedBy"`
	InvitedByName  string    `db:"invited_by_name" json:"invitedByName"`
	ExpiresAt      time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
}

type CreateInvitation struct {
	Email string `json:"email" validate:"required,email,max=100"`
	Role  string `json:"role" validate:"required,oneof=admin editor read-only"`
}

type AcceptInvitation struct {
	Token    string  `json:"token" validate:"required,max=100"`
	Password *string `json:"password" validate:"omitempty,min=8"`
}

type CheckInvitationResponse struct {
	Email            string `json:"email"`
	OrganisationName string `json:"organisationName"`
	InvitedByName    string `json:"invitedByName"`
	ExistingUser     bool   `json:"existingUser"`
}
