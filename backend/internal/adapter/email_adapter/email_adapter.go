//go:generate mockgen -package=mocks -destination ../../mocks/email_adapter.go liquiswiss/internal/adapter/email_adapter IEmailAdapter
package email_adapter

import (
	"liquiswiss/config"
)

type IEmailAdapter interface {
	SendRegistrationMail(email, code string) error
	SendPasswordResetMail(email, code string) error
	SendInvitationMail(email, token, organisationName, invitedByName string) error
}

func NewEmailAdapter(cfg config.Config) IEmailAdapter {
	return newSMTPAdapter(cfg)
}
