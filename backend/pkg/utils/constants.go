package utils

import "time"

const (
	AccessTokenValidity = 20 * time.Minute
	//AccessTokenValidity = 1 * time.Second

	RefreshTokenValidity = 90 * 24 * time.Hour // 3 months validity
	//RefreshTokenValidity = 3 * time.Second

	RegistrationCodeValidity = 1 * time.Hour

	// Default anti-spam window between forgot-password requests for the same email.
	// Override via RESET_PASSWORD_DELAY_MINUTES env var.
	ResetPasswordDelay = 1 * time.Hour
	// Default validity window for an issued reset code.
	// Override via RESET_PASSWORD_VALIDITY_MINUTES env var.
	ResetPasswordValidity = 1 * time.Hour

	// Default anti-spam window between invitation resend attempts for the same invitation.
	// Override via INVITATION_RESEND_DELAY_MINUTES env var.
	InvitationResendDelay = 10 * time.Minute

	InvitationValidity = 7 * 24 * time.Hour // 7 days validity

	MaxForecastYears = 3

	AccessTokenName  = "liq-access-token"
	RefreshTokenName = "liq-refresh-token"

	TransactionsTableName = "transactions"
	SalariesTableName     = "salaries"
	SalaryCostsTableName  = "salary_costs"
)
